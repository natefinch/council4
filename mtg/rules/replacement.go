package rules

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/rules/payment"
)

type enterBattlefieldContext struct {
	engine *Engine
	agents [game.NumPlayers]PlayerAgent
	log    *TurnLog
}

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
	if event.amount <= 0 {
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
	if permanent == nil {
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
			FromZone:    zone.Battlefield,
			ToZone:      zone.Graveyard,
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

func replacementZoneChangeDestination(g *game.Game, event game.GameEvent) zone.Type {
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

func applyEnterBattlefieldReplacementEffects(ctx enterBattlefieldContext, g *game.Game, permanent *game.Permanent, fromZone zone.Type) {
	event := game.GameEvent{
		Kind:        game.EventPermanentEnteredBattlefield,
		Controller:  effectiveController(g, permanent),
		Player:      permanent.Owner,
		CardID:      permanent.CardInstanceID,
		PermanentID: permanent.ObjectID,
		TokenName:   permanentTokenName(permanent),
		TokenDef:    permanent.TokenDef,
		FromZone:    fromZone,
		ToZone:      zone.Battlefield,
	}
	var staticMatches []game.ReplacementEffect
	if def, ok := permanentCardDef(g, permanent); ok {
		staticMatches = staticETBReplacementEffects(ctx, g, permanent, def, event)
	}
	matches := matchingETBReplacementEffects(g, event)
	matches = append(matches, staticMatches...)
	if len(matches) > 1 {
		recordReplacementDecision(g, replacementEventPlayer(event), replacementEffectLabels(matches))
	}
	for i := range matches {
		replacement := &matches[i]
		if replacement.EntersTapped {
			setPermanentTapped(g, permanent, true)
		}
		for _, placement := range replacement.EntersWithCounters {
			permanent.Counters.Add(placement.Kind, placement.Amount)
		}
	}
}

func staticETBReplacementEffects(ctx enterBattlefieldContext, g *game.Game, permanent *game.Permanent, def *game.CardDef, event game.GameEvent) []game.ReplacementEffect {
	var replacements []game.ReplacementEffect
	for i := range def.ReplacementAbilities {
		ability := &def.ReplacementAbilities[i]
		replacement := ability.Replacement
		replacement.Controller = event.Controller
		replacement.SourceCardID = permanent.CardInstanceID
		replacement.SourceObjectID = 0
		replacement.CreatedTurn = g.Turn.TurnNumber
		if replacement.Description == "" {
			replacement.Description = ability.Text
		}
		if ability.UnlessPaid.Exists && enterBattlefieldPaymentPaid(ctx, g, event.Controller, ability.UnlessPaid.Val) {
			continue
		}
		if replacementEffectMatchesEvent(g, &replacement, event) {
			replacements = append(replacements, replacement)
		}
	}
	return replacements
}

func enterBattlefieldPaymentPaid(ctx enterBattlefieldContext, g *game.Game, playerID game.PlayerID, res game.ResolutionPayment) bool {
	if !canPayResolutionPayment(g, playerID, &res) {
		return false
	}
	engine := ctx.engine
	if engine == nil {
		engine = NewEngine(nil)
	}
	prompt := res.Prompt
	if prompt == "" {
		prompt = "Pay enter-the-battlefield cost?"
	}
	if !engine.chooseMay(g, ctx.agents, playerID, prompt, ctx.log) {
		return false
	}
	prefs := engine.paymentPreferencesForCost(g, playerID, manaCostPtr(res.ManaCost), res.AdditionalCosts, ctx.agents, ctx.log)
	return paymentOrch.payGenericCost(g, payment.GenericRequest{
		PlayerID:        playerID,
		Cost:            manaCostPtr(res.ManaCost),
		XValue:          res.XValue,
		AdditionalCosts: res.AdditionalCosts,
		Prefs:           prefs,
	})
}

func matchingZoneReplacementEffects(g *game.Game, event game.GameEvent, applied map[id.ID]bool) []game.ReplacementEffect {
	var matches []game.ReplacementEffect
	for i := range g.ReplacementEffects {
		replacement := &g.ReplacementEffects[i]
		if applied[replacement.ID] || replacement.ReplaceToZone == zone.None || !replacementEffectMatchesEvent(g, replacement, event) {
			continue
		}
		matches = append(matches, *replacement)
	}
	return matches
}

func matchingETBReplacementEffects(g *game.Game, event game.GameEvent) []game.ReplacementEffect {
	var matches []game.ReplacementEffect
	for i := range g.ReplacementEffects {
		replacement := &g.ReplacementEffects[i]
		if !replacement.EntersTapped && len(replacement.EntersWithCounters) == 0 {
			continue
		}
		if !replacementEffectMatchesEvent(g, replacement, event) {
			continue
		}
		matches = append(matches, *replacement)
	}
	return matches
}

func replacementEffectMatchesEvent(g *game.Game, replacement *game.ReplacementEffect, event game.GameEvent) bool {
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
	if !conditionSatisfied(g, conditionContext{
		controller: replacement.Controller,
		event:      &event,
	}, replacement.Condition) {
		return false
	}
	return true
}

func replacementSourceStillApplies(g *game.Game, replacement *game.ReplacementEffect) bool {
	if replacement.Duration != game.DurationPermanent || replacement.SourceObjectID == 0 {
		return true
	}
	_, ok := permanentByObjectID(g, replacement.SourceObjectID)
	return ok
}

func replacementEventPlayer(event game.GameEvent) game.PlayerID {
	if event.Player >= 0 && event.Player < game.NumPlayers {
		return event.Player
	}
	return event.Controller
}

func replacementEffectLabels(replacements []game.ReplacementEffect) []string {
	labels := make([]string, 0, len(replacements))
	for i := range replacements {
		replacement := &replacements[i]
		if replacement.Description != "" {
			labels = append(labels, replacement.Description)
			continue
		}
		labels = append(labels, fmt.Sprintf("replacement %d", replacement.ID))
	}
	return labels
}

func createPreventionShield(g *game.Game, obj *game.StackObject, amount, targetIndex int, duration game.EffectDuration) bool {
	if amount <= 0 {
		return false
	}
	shield := game.PreventionShield{
		ID:          g.IDGen.Next(),
		Controller:  obj.Controller,
		Amount:      amount,
		Duration:    effectDurationOrDefault(duration, game.DurationUntilEndOfTurn),
		CreatedTurn: g.Turn.TurnNumber,
	}
	if targetIndex == game.TargetIndexController {
		shield.Player = obj.Controller
		g.PreventionShields = append(g.PreventionShields, shield)
		return true
	}
	if targetIndex < 0 || targetIndex >= len(obj.Targets) {
		return false
	}
	target := obj.Targets[targetIndex]
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
	if len(g.PreventionShields) == 0 {
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
	if len(g.ReplacementEffects) == 0 {
		return
	}
	kept := g.ReplacementEffects[:0]
	for i := range g.ReplacementEffects {
		replacement := &g.ReplacementEffects[i]
		if replacement.Duration == game.DurationUntilEndOfTurn || replacement.Duration == game.DurationThisTurn {
			continue
		}
		if !replacementSourceStillApplies(g, replacement) {
			continue
		}
		kept = append(kept, *replacement)
	}
	g.ReplacementEffects = kept
}

func replaceDestroyWithRegeneration(g *game.Game, permanent *game.Permanent) bool {
	if permanent.RegenerationShields <= 0 {
		return false
	}
	permanent.RegenerationShields--
	setPermanentTapped(g, permanent, true)
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
		FromZone:    zone.Battlefield,
		ToZone:      zone.Graveyard,
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
	source, ok := damageSourceDef(g, sourceID, sourceObjectID)
	return ok && permanentProtectedFromSourceDef(g, permanent, source)
}

func permanentProtectedFromSourceDef(g *game.Game, permanent *game.Permanent, source *game.CardDef) bool {
	if permanent == nil || source == nil {
		return false
	}
	for _, clr := range permanentProtectionColors(g, permanent) {
		if slices.Contains(source.Colors, clr) {
			return true
		}
	}
	return false
}

func permanentProtectionColors(g *game.Game, permanent *game.Permanent) []color.Color {
	var colors []color.Color
	abilities := permanentEffectiveAbilities(g, permanent)
	for i := range abilities {
		ability := &abilities[i]
		if !abilityHasKeyword(ability, game.Protection) {
			continue
		}
		colors = append(colors, ability.ProtectionColors()...)
	}
	return colors
}

func damageSourceDef(g *game.Game, sourceID, sourceObjectID id.ID) (*game.CardDef, bool) {
	if sourceID != 0 {
		if card, ok := g.GetCardInstance(sourceID); ok {
			return card.Def, true
		}
	}
	if permanent, ok := permanentByObjectID(g, sourceObjectID); ok {
		return permanentCardDef(g, permanent)
	}
	return nil, false
}
