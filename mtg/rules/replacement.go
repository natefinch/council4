package rules

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

type damageEvent struct {
	sourceID       id.ID
	sourceObjectID id.ID
	controller     game.PlayerID
	player         game.PlayerID
	permanent      *game.Permanent
	amount         int
	combatDamage   bool
}

func applyDamagePrevention(g *game.Game, event damageEvent) int {
	if g == nil || event.amount <= 0 {
		return 0
	}
	amount := event.amount
	if event.permanent != nil && permanentProtectedFromSource(g, event.permanent, event.sourceID, event.sourceObjectID) {
		amount = 0
	}
	if amount > 0 {
		amount = applyPreventionShields(g, event, amount)
	}
	if amount > 0 && event.permanent != nil && event.permanent.Counters.Remove(counter.Shield, 1) > 0 {
		amount = 0
	}
	if prevented := event.amount - amount; prevented > 0 {
		emitDamagePreventedEvent(g, event, prevented)
	}
	return amount
}

func orderedPreventionShieldIndices(g *game.Game, event damageEvent) []int {
	var indices []int
	var options []string
	for i, shield := range g.PreventionShields {
		if !preventionShieldApplies(shield, event) || shield.Amount <= 0 {
			continue
		}
		indices = append(indices, i)
		options = append(options, "prevention shield")
	}
	if len(indices) > 1 {
		recordReplacementDecision(g, replacementDecisionPlayer(g, event), options)
	}
	return indices
}

func replaceDestroyPermanent(g *game.Game, permanent *game.Permanent) bool {
	if g == nil || permanent == nil {
		return false
	}
	hasShieldCounter := permanent.Counters.Get(counter.Shield) > 0
	hasRegeneration := permanent.RegenerationShields > 0
	if hasShieldCounter && hasRegeneration {
		recordReplacementDecision(g, effectiveController(g, permanent), []string{"shield counter", "regeneration shield"})
	}
	if hasShieldCounter {
		permanent.Counters.Remove(counter.Shield, 1)
		emitEvent(g, game.GameEvent{
			Kind:        game.EventDestroyReplaced,
			Controller:  effectiveController(g, permanent),
			Player:      permanent.Owner,
			CardID:      permanent.CardInstanceID,
			PermanentID: permanent.ObjectID,
			TokenName:   permanentTokenName(permanent),
			TokenDef:    permanent.TokenDef,
			FromZone:    game.ZoneBattlefield,
			ToZone:      game.ZoneGraveyard,
		})
		return true
	}
	return replaceDestroyWithRegeneration(g, permanent)
}

func replacementDecisionPlayer(g *game.Game, event damageEvent) game.PlayerID {
	if event.permanent != nil {
		return effectiveController(g, event.permanent)
	}
	return event.player
}

func recordReplacementDecision(g *game.Game, player game.PlayerID, options []string) {
	selected := make([]int, len(options))
	for i := range options {
		selected[i] = i
	}
	g.ReplacementDecisions = append(g.ReplacementDecisions, game.ReplacementDecision{
		Player:       player,
		Options:      append([]string(nil), options...),
		Selected:     selected,
		UsedFallback: true,
	})
}

func createReplacementEffect(g *game.Game, obj *game.StackObject, effect game.Effect) bool {
	if g == nil || obj == nil || effect.Replacement == nil {
		return false
	}
	replacement := *effect.Replacement
	replacement.ID = g.IDGen.Next()
	replacement.Controller = obj.Controller
	replacement.SourceCardID, replacement.SourceObjectID = damageSourceIDs(g, obj)
	replacement.CreatedTurn = g.Turn.TurnNumber
	if effect.Duration != game.DurationPermanent {
		replacement.Duration = effect.Duration
	}
	if replacement.Duration == game.DurationPermanent && effect.UntilEndOfTurn {
		replacement.Duration = game.DurationUntilEndOfTurn
	}
	g.ReplacementEffects = append(g.ReplacementEffects, replacement)
	return true
}

func replacementZoneChangeDestination(g *game.Game, event game.GameEvent) game.ZoneType {
	if g == nil {
		return event.ToZone
	}
	destination := event.ToZone
	applied := make(map[id.ID]bool)
	for {
		event.ToZone = destination
		matches := matchingZoneReplacementEffects(g, event, applied)
		if len(matches) == 0 {
			return destination
		}
		if len(matches) > 1 {
			recordReplacementDecision(g, replacementEventPlayer(event), replacementEffectLabels(matches))
		}
		replacement := matches[0]
		applied[replacement.ID] = true
		// After a replacement changes the event, the replacement process checks
		// the modified event again; the same effect cannot apply twice to one
		// event (CR 614.5, CR 616.1e).
		destination = replacement.ReplaceToZone
	}
}

func applyEnterBattlefieldReplacementEffects(g *game.Game, permanent *game.Permanent, fromZone game.ZoneType) {
	if g == nil || permanent == nil {
		return
	}
	event := game.GameEvent{
		Kind:        game.EventPermanentEnteredBattlefield,
		Controller:  effectiveController(g, permanent),
		Player:      permanent.Owner,
		CardID:      permanent.CardInstanceID,
		PermanentID: permanent.ObjectID,
		TokenName:   permanentTokenName(permanent),
		TokenDef:    permanent.TokenDef,
		FromZone:    fromZone,
		ToZone:      game.ZoneBattlefield,
	}
	matches := matchingETBReplacementEffects(g, event)
	if len(matches) > 1 {
		recordReplacementDecision(g, replacementEventPlayer(event), replacementEffectLabels(matches))
	}
	for _, replacement := range matches {
		if replacement.EntersTapped {
			permanent.Tapped = true
		}
		for _, placement := range replacement.EntersWithCounters {
			permanent.Counters.Add(placement.Kind, placement.Amount)
		}
	}
}

func matchingZoneReplacementEffects(g *game.Game, event game.GameEvent, applied map[id.ID]bool) []game.ReplacementEffect {
	var matches []game.ReplacementEffect
	for _, replacement := range g.ReplacementEffects {
		if applied[replacement.ID] || replacement.ReplaceToZone == game.ZoneNone || !replacementEffectMatchesEvent(g, replacement, event) {
			continue
		}
		matches = append(matches, replacement)
	}
	return matches
}

func matchingETBReplacementEffects(g *game.Game, event game.GameEvent) []game.ReplacementEffect {
	var matches []game.ReplacementEffect
	for _, replacement := range g.ReplacementEffects {
		if !replacement.EntersTapped && len(replacement.EntersWithCounters) == 0 {
			continue
		}
		if !replacementEffectMatchesEvent(g, replacement, event) {
			continue
		}
		matches = append(matches, replacement)
	}
	return matches
}

func replacementEffectMatchesEvent(g *game.Game, replacement game.ReplacementEffect, event game.GameEvent) bool {
	if !replacementSourceStillApplies(g, replacement) {
		return false
	}
	if replacement.MatchEvent != game.EventUnknown && replacement.MatchEvent != event.Kind {
		return false
	}
	if replacement.ControllerFilter != game.TriggerControllerAny && !triggerControllerMatches(replacement.Controller, replacement.ControllerFilter, event.Controller) {
		return false
	}
	if replacement.MatchFromZone && replacement.FromZone != event.FromZone {
		return false
	}
	if replacement.MatchToZone && replacement.ToZone != event.ToZone {
		return false
	}
	return true
}

func replacementSourceStillApplies(g *game.Game, replacement game.ReplacementEffect) bool {
	if replacement.Duration != game.DurationPermanent || replacement.SourceObjectID == 0 {
		return true
	}
	return permanentByObjectID(g, replacement.SourceObjectID) != nil
}

func replacementEventPlayer(event game.GameEvent) game.PlayerID {
	if event.Player >= 0 && event.Player < game.NumPlayers {
		return event.Player
	}
	return event.Controller
}

func replacementEffectLabels(replacements []game.ReplacementEffect) []string {
	labels := make([]string, 0, len(replacements))
	for _, replacement := range replacements {
		if replacement.Description != "" {
			labels = append(labels, replacement.Description)
			continue
		}
		labels = append(labels, fmt.Sprintf("replacement %d", replacement.ID))
	}
	return labels
}

func createPreventionShield(g *game.Game, obj *game.StackObject, effect game.Effect) bool {
	if g == nil || obj == nil || effect.Amount <= 0 {
		return false
	}
	shield := game.PreventionShield{
		ID:          g.IDGen.Next(),
		Controller:  obj.Controller,
		Amount:      effect.Amount,
		Duration:    effectDurationOrDefault(effect.Duration, game.DurationUntilEndOfTurn),
		CreatedTurn: g.Turn.TurnNumber,
	}
	if effect.TargetIndex == -1 {
		shield.Player = obj.Controller
		g.PreventionShields = append(g.PreventionShields, shield)
		return true
	}
	if effect.TargetIndex < 0 || effect.TargetIndex >= len(obj.Targets) {
		return false
	}
	target := obj.Targets[effect.TargetIndex]
	switch target.Kind {
	case game.TargetPlayer:
		shield.Player = target.PlayerID
	case game.TargetPermanent:
		shield.PermanentID = target.PermanentID
	default:
		return false
	}
	g.PreventionShields = append(g.PreventionShields, shield)
	return true
}

func applyPreventionShields(g *game.Game, event damageEvent, amount int) int {
	order := orderedPreventionShieldIndices(g, event)
	for _, i := range order {
		shield := &g.PreventionShields[i]
		if !preventionShieldApplies(*shield, event) || shield.Amount <= 0 || amount <= 0 {
			continue
		}
		prevented := min(amount, shield.Amount)
		amount -= prevented
		shield.Amount -= prevented
	}
	g.PreventionShields = compactPreventionShields(g.PreventionShields)
	return amount
}

func preventionShieldApplies(shield game.PreventionShield, event damageEvent) bool {
	if event.permanent != nil {
		return shield.PermanentID == event.permanent.ObjectID
	}
	return shield.PermanentID == 0 && shield.Player == event.player
}

func compactPreventionShields(shields []game.PreventionShield) []game.PreventionShield {
	kept := shields[:0]
	for _, shield := range shields {
		if shield.Amount > 0 {
			kept = append(kept, shield)
		}
	}
	return kept
}

func expirePreventionShields(g *game.Game) {
	if g == nil || len(g.PreventionShields) == 0 {
		return
	}
	kept := g.PreventionShields[:0]
	for _, shield := range g.PreventionShields {
		if shield.Duration == game.DurationUntilEndOfTurn || shield.Duration == game.DurationThisTurn {
			continue
		}
		kept = append(kept, shield)
	}
	g.PreventionShields = kept
}

func expireReplacementEffects(g *game.Game) {
	if g == nil || len(g.ReplacementEffects) == 0 {
		return
	}
	kept := g.ReplacementEffects[:0]
	for _, replacement := range g.ReplacementEffects {
		if replacement.Duration == game.DurationUntilEndOfTurn || replacement.Duration == game.DurationThisTurn {
			continue
		}
		if !replacementSourceStillApplies(g, replacement) {
			continue
		}
		kept = append(kept, replacement)
	}
	g.ReplacementEffects = kept
}

func replaceDestroyWithRegeneration(g *game.Game, permanent *game.Permanent) bool {
	if permanent.RegenerationShields <= 0 {
		return false
	}
	permanent.RegenerationShields--
	permanent.Tapped = true
	permanent.MarkedDamage = 0
	permanent.MarkedDeathtouchDamage = false
	removePermanentFromCombat(g, permanent.ObjectID)
	emitEvent(g, game.GameEvent{
		Kind:        game.EventDestroyReplaced,
		Controller:  effectiveController(g, permanent),
		Player:      permanent.Owner,
		CardID:      permanent.CardInstanceID,
		PermanentID: permanent.ObjectID,
		TokenName:   permanentTokenName(permanent),
		TokenDef:    permanent.TokenDef,
		FromZone:    game.ZoneBattlefield,
		ToZone:      game.ZoneGraveyard,
	})
	return true
}

func emitDamagePreventedEvent(g *game.Game, event damageEvent, prevented int) {
	preventedEvent := game.GameEvent{
		Kind:            game.EventDamagePrevented,
		SourceID:        event.sourceID,
		SourceObjectID:  event.sourceObjectID,
		Controller:      event.controller,
		Player:          event.player,
		Amount:          prevented,
		DamageRecipient: game.DamageRecipientPlayer,
		CombatDamage:    event.combatDamage,
	}
	if event.permanent != nil {
		preventedEvent.Player = event.permanent.Owner
		preventedEvent.PermanentID = event.permanent.ObjectID
		preventedEvent.CardID = event.permanent.CardInstanceID
		preventedEvent.TokenName = permanentTokenName(event.permanent)
		preventedEvent.TokenDef = event.permanent.TokenDef
		preventedEvent.DamageRecipient = game.DamageRecipientPermanent
	}
	emitEvent(g, preventedEvent)
}

func permanentProtectedFromSource(g *game.Game, permanent *game.Permanent, sourceID, sourceObjectID id.ID) bool {
	source := damageSourceDef(g, sourceID, sourceObjectID)
	return permanentProtectedFromSourceDef(g, permanent, source)
}

func permanentProtectedFromSourceDef(g *game.Game, permanent *game.Permanent, source *game.CardDef) bool {
	if permanent == nil || source == nil {
		return false
	}
	for _, color := range permanentProtectionColors(g, permanent) {
		if slices.Contains(source.Colors, color) {
			return true
		}
	}
	return false
}

func permanentProtectionColors(g *game.Game, permanent *game.Permanent) []mana.Color {
	var colors []mana.Color
	for _, ability := range permanentEffectiveAbilities(g, permanent) {
		if !abilityHasKeyword(&ability, game.Protection) {
			continue
		}
		colors = append(colors, ability.ProtectionFromColors...)
	}
	return colors
}

func damageSourceDef(g *game.Game, sourceID, sourceObjectID id.ID) *game.CardDef {
	if g == nil {
		return nil
	}
	if sourceID != 0 {
		if card := g.GetCardInstance(sourceID); card != nil {
			return card.Def
		}
	}
	return permanentCardDef(g, permanentByObjectID(g, sourceObjectID))
}
