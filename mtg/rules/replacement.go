package rules

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
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
		emitEvent(g, game.Event{
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

type zoneChangeReplacementResult struct {
	destination        zone.Type
	shuffleIntoLibrary bool
	revealSource       bool
}

func replacementZoneChangeDestination(g *game.Game, event game.Event) zone.Type {
	return replacementZoneChange(g, event).destination
}

func replacementZoneChange(g *game.Game, event game.Event) zoneChangeReplacementResult {
	result := zoneChangeReplacementResult{destination: event.ToZone}
	applied := make(map[id.ID]bool)
	for {
		event.ToZone = result.destination
		matches := matchingZoneReplacementEffects(g, event, applied)
		if len(matches) == 0 {
			return result
		}
		if len(matches) > 1 {
			recordReplacementDecision(g, replacementEventPlayer(event), replacementEffectLabels(matches))
		}
		replacement := matches[0]
		applied[replacement.ID] = true
		// After a replacement changes the event, the replacement process checks
		// the modified event again; the same effect cannot apply twice to one
		// event (CR 614.5, CR 616.1e).
		result.destination = replacement.ReplaceToZone
		result.shuffleIntoLibrary = replacement.ShuffleIntoLibrary && result.destination == zone.Library
		result.revealSource = result.revealSource || replacement.RevealSource
	}
}

func replacementTokenCreationAmount(g *game.Game, controller game.PlayerID, amount int) int {
	if amount <= 0 {
		return amount
	}
	event := game.Event{
		Kind:       game.EventTokenCreated,
		Controller: controller,
		Player:     controller,
		Amount:     amount,
	}
	applied := make(map[id.ID]bool)
	for {
		event.Amount = amount
		matches := matchingTokenCreationReplacementEffects(g, event, applied)
		if len(matches) == 0 {
			return amount
		}
		if len(matches) > 1 {
			recordReplacementDecision(g, replacementEventPlayer(event), replacementEffectLabels(matches))
		}
		replacement := matches[0]
		applied[replacement.ID] = true
		amount *= replacement.TokenMultiplier
	}
}

func replacementPermanentCounterPlacementAmount(g *game.Game, placementController game.PlayerID, permanent *game.Permanent, kind counter.Kind, amount int) int {
	if permanent == nil || amount <= 0 {
		return amount
	}
	event := game.Event{
		Kind:        game.EventCountersAdded,
		Controller:  placementController,
		Player:      permanent.Owner,
		CardID:      permanent.CardInstanceID,
		PermanentID: permanent.ObjectID,
		CounterKind: kind,
		Amount:      amount,
	}
	applied := make(map[id.ID]bool)
	for {
		event.Amount = amount
		matches := matchingCounterPlacementReplacementEffects(g, event, permanent, applied)
		if len(matches) == 0 {
			return amount
		}
		if len(matches) > 1 {
			recordReplacementDecision(g, replacementEventPlayer(event), replacementEffectLabels(matches))
		}
		replacement := matches[0]
		applied[replacement.ID] = true
		amount *= replacement.CounterMultiplier
	}
}

func replacementPlayerCounterPlacementAmount(g *game.Game, placementController, player game.PlayerID, kind counter.Kind, amount int) int {
	if amount <= 0 {
		return amount
	}
	event := game.Event{
		Kind:        game.EventCountersAdded,
		Controller:  placementController,
		Player:      player,
		CounterKind: kind,
		Amount:      amount,
	}
	applied := make(map[id.ID]bool)
	for {
		event.Amount = amount
		matches := matchingCounterPlacementReplacementEffects(g, event, nil, applied)
		if len(matches) == 0 {
			return amount
		}
		if len(matches) > 1 {
			recordReplacementDecision(g, replacementEventPlayer(event), replacementEffectLabels(matches))
		}
		replacement := matches[0]
		applied[replacement.ID] = true
		amount *= replacement.CounterMultiplier
	}
}

func revealZoneReplacementSource(g *game.Game, event game.Event, reveal bool) {
	if !reveal || event.CardID == 0 {
		return
	}
	emitEvent(g, game.Event{
		Kind:        game.EventCardRevealed,
		Controller:  event.Controller,
		Player:      event.Player,
		CardID:      event.CardID,
		Face:        event.Face,
		PermanentID: event.PermanentID,
		TokenName:   event.TokenName,
		TokenDef:    event.TokenDef,
	})
}

func applyEnterBattlefieldReplacementEffects(ctx enterBattlefieldContext, g *game.Game, permanent *game.Permanent, fromZone zone.Type) {
	event := game.Event{
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
	if !permanent.FaceDown {
		def, ok := permanentCardDef(g, permanent)
		if !ok {
			def = nil
		}
		if def != nil {
			staticMatches = staticETBReplacementEffects(ctx, g, permanent, def, event)
		}
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
			addCountersToPermanent(g, permanent, placement.Kind, placement.Amount)
		}
	}
}

func staticETBReplacementEffects(ctx enterBattlefieldContext, g *game.Game, permanent *game.Permanent, def *game.CardDef, event game.Event) []game.ReplacementEffect {
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
		if ability.UnlessPaid.Exists && enterBattlefieldPaymentPaid(ctx, g, event.Controller, event.CardID, ability.UnlessPaid.Val) {
			continue
		}
		if replacementEffectMatchesEvent(g, &replacement, event) {
			replacements = append(replacements, replacement)
		}
	}
	return replacements
}

func enterBattlefieldPaymentPaid(ctx enterBattlefieldContext, g *game.Game, playerID game.PlayerID, sourceCardID id.ID, res game.ResolutionPayment) bool {
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
		SourceCardID:    sourceCardID,
		Cost:            manaCostPtr(res.ManaCost),
		XValue:          res.XValue,
		AdditionalCosts: res.AdditionalCosts,
		Prefs:           prefs,
	})
}

func matchingZoneReplacementEffects(g *game.Game, event game.Event, applied map[id.ID]bool) []game.ReplacementEffect {
	var matches []game.ReplacementEffect
	for i := range g.ReplacementEffects {
		replacement := &g.ReplacementEffects[i]
		if applied[replacement.ID] || replacement.ReplaceToZone == zone.None || !replacementEffectMatchesEvent(g, replacement, event) {
			continue
		}
		matches = append(matches, *replacement)
	}
	matches = append(matches, matchingStaticSelfZoneReplacementEffects(g, event, applied)...)
	return matches
}

func matchingStaticSelfZoneReplacementEffects(g *game.Game, event game.Event, applied map[id.ID]bool) []game.ReplacementEffect {
	if event.CardID == 0 || event.FaceDown {
		return nil
	}
	card, ok := g.GetCardInstance(event.CardID)
	if !ok || card.Def == nil {
		return nil
	}
	def := cardFaceOrDefault(card, event.Face)
	var matches []game.ReplacementEffect
	for i := range def.ReplacementAbilities {
		ability := &def.ReplacementAbilities[i]
		replacement := ability.Replacement
		if replacement.ReplaceToZone == zone.None {
			continue
		}
		replacement.ID = event.CardID
		replacement.Controller = event.Controller
		replacement.SourceCardID = event.CardID
		replacement.SourceObjectID = 0
		if replacement.Description == "" {
			replacement.Description = ability.Text
		}
		if applied[replacement.ID] || !replacementEffectMatchesEvent(g, &replacement, event) {
			continue
		}
		matches = append(matches, replacement)
	}
	return matches
}

func matchingETBReplacementEffects(g *game.Game, event game.Event) []game.ReplacementEffect {
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

func matchingTokenCreationReplacementEffects(g *game.Game, event game.Event, applied map[id.ID]bool) []game.ReplacementEffect {
	var matches []game.ReplacementEffect
	for i := range g.ReplacementEffects {
		replacement := &g.ReplacementEffects[i]
		if replacement.TokenMultiplier <= 1 {
			continue
		}
		if applied[replacement.ID] || !replacementEffectMatchesEvent(g, replacement, event) {
			continue
		}
		matches = append(matches, *replacement)
	}
	return matches
}

func counterPlacementMatchEvent(g *game.Game, replacement *game.ReplacementEffect, event game.Event, recipient *game.Permanent) game.Event {
	if !replacement.CounterUseRecipientController {
		return event
	}
	event.Controller = counterRecipientController(g, event, recipient)
	return event
}

func counterRecipientController(g *game.Game, event game.Event, recipient *game.Permanent) game.PlayerID {
	if recipient != nil {
		return effectiveController(g, recipient)
	}
	if event.PermanentID != 0 {
		if permanent, ok := permanentByObjectID(g, event.PermanentID); ok {
			return effectiveController(g, permanent)
		}
	}
	if event.Player >= 0 && event.Player < game.NumPlayers {
		return event.Player
	}
	return event.Controller
}

func matchingCounterPlacementReplacementEffects(g *game.Game, event game.Event, recipient *game.Permanent, applied map[id.ID]bool) []game.ReplacementEffect {
	var matches []game.ReplacementEffect
	for i := range g.ReplacementEffects {
		replacement := &g.ReplacementEffects[i]
		if replacement.CounterMultiplier <= 1 {
			continue
		}
		if replacement.MatchCounterKind && replacement.CounterKindFilter != event.CounterKind {
			continue
		}
		if len(replacement.CounterRecipientTypes) > 0 && !counterRecipientPermanentMatches(g, event.PermanentID, recipient, replacement.CounterRecipientTypes) {
			continue
		}
		matchEvent := counterPlacementMatchEvent(g, replacement, event, recipient)
		if applied[replacement.ID] || !replacementEffectMatchesEvent(g, replacement, matchEvent) {
			continue
		}
		matches = append(matches, *replacement)
	}
	return matches
}

func counterRecipientPermanentMatches(g *game.Game, permanentID id.ID, permanent *game.Permanent, requiredTypes []types.Card) bool {
	if permanent == nil {
		if permanentID == 0 {
			return false
		}
		var ok bool
		permanent, ok = permanentByObjectID(g, permanentID)
		if !ok {
			return false
		}
	}
	for _, cardType := range requiredTypes {
		if !permanentHasType(g, permanent, cardType) {
			return false
		}
	}
	return true
}

func replacementEffectMatchesEvent(g *game.Game, replacement *game.ReplacementEffect, event game.Event) bool {
	if !replacementSourceStillApplies(g, replacement) {
		return false
	}
	if replacement.MatchEvent != game.EventUnknown && replacement.MatchEvent != event.Kind {
		return false
	}
	controller := replacementCurrentController(g, replacement)
	if replacement.ControllerFilter != game.TriggerControllerAny && !triggerControllerMatches(controller, replacement.ControllerFilter, event.Controller) {
		return false
	}
	if replacement.MatchFromZone && replacement.FromZone != event.FromZone {
		return false
	}
	if replacement.MatchToZone && replacement.ToZone != event.ToZone {
		return false
	}
	if !conditionSatisfied(g, conditionContext{
		controller: controller,
		event:      &event,
	}, replacement.Condition) {
		return false
	}
	return true
}

func replacementCurrentController(g *game.Game, replacement *game.ReplacementEffect) game.PlayerID {
	if replacement.SourceObjectID != 0 {
		if permanent, ok := permanentByObjectID(g, replacement.SourceObjectID); ok {
			return effectiveController(g, permanent)
		}
	}
	return replacement.Controller
}

func replacementSourceStillApplies(g *game.Game, replacement *game.ReplacementEffect) bool {
	if replacement.Duration != game.DurationPermanent || replacement.SourceObjectID == 0 {
		return true
	}
	_, ok := permanentByObjectID(g, replacement.SourceObjectID)
	return ok
}

func replacementEventPlayer(event game.Event) game.PlayerID {
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

func createPreventionShield(g *game.Game, obj *game.StackObject, amount int, object game.ObjectReference, player game.PlayerReference, duration game.EffectDuration) bool {
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
	if player.Kind() != game.PlayerReferenceNone {
		playerID, ok := resolvePlayerReference(g, obj, player)
		if !ok {
			return false
		}
		shield.Player = playerID
		g.PreventionShields = append(g.PreventionShields, shield)
		return true
	}
	resolved, ok := resolveObjectReference(g, obj, object)
	if !ok || resolved.permanent == nil {
		return false
	}
	shield.PermanentID = resolved.permanent.ObjectID
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
	emitEvent(g, game.Event{
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
	preventedEvent := game.Event{
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
		body, ok := abilities[i].(game.StaticAbility)
		if ok {
			colors = append(colors, game.StaticBodyProtectionColors(&body)...)
		}
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
