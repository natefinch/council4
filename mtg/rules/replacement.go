package rules

import (
	"fmt"
	"slices"
	"strings"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

type enterBattlefieldContext struct {
	engine            *Engine
	agents            [game.NumPlayers]PlayerAgent
	log               *TurnLog
	xValue            int
	kickCount         int
	kickerPaid        bool
	colorsOfManaSpent int
	manaSpentByColor  map[color.Color]int
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

// applyDamageModifications applies the replacement and prevention effects that
// modify a damage event and returns the amount of damage actually dealt. It
// interleaves damage replacement effects (CR 614) with prevention shields and
// shield counters (CR 615) in a single CR 616.1 selection loop: while more than
// one modifying effect applies, the affected object's controller or the affected
// player chooses which to apply next, the chosen effect is applied (consuming a
// finite shield's amount or a shield counter), and the applicable set is then
// recomputed (CR 616.1f). Each replacement effect affects the event at most once
// (CR 614.5). Protection from the source prevents all of the damage (CR 702.16e)
// and is applied first, since it costs nothing and zeroes the event regardless of
// order. Each prevention reports the amount it prevented via a damage-prevented
// event.
func applyDamageModifications(g *game.Game, event damageEvent) int {
	if event.amount <= 0 {
		return 0
	}
	if damageEventProtected(g, event) {
		emitDamagePreventedEvent(g, event, event.amount)
		return 0
	}
	chooser := replacementDecisionPlayer(g, event)
	appliedReplacements := make(map[id.ID]bool)
	amount := event.amount
	for amount > 0 {
		event.amount = amount
		shieldIndices := applicablePreventionShieldIndices(g, event)
		hasShieldCounter := event.permanent != nil && event.permanent.Counters.Get(counter.Shield) > 0
		replacements := matchingDamageReplacementEffects(g, event, appliedReplacements)

		// Candidates are listed prevention-first so the deterministic fallback
		// (no agent) prevents before replacing, but an agent may choose any
		// applicable effect, including a replacement before a prevention.
		labels := make([]string, 0, len(shieldIndices)+1+len(replacements))
		for range shieldIndices {
			labels = append(labels, "prevention shield")
		}
		if hasShieldCounter {
			labels = append(labels, "shield counter")
		}
		labels = append(labels, replacementEffectLabels(replacements)...)
		if len(labels) == 0 {
			break
		}

		chosen := 0
		if len(labels) > 1 {
			decision := chooseReplacementDecision(g, chooser, labels)
			chosen = decision.Selected[0]
		}

		shieldCounterIndex := len(shieldIndices)
		replacementsStart := shieldCounterIndex
		if hasShieldCounter {
			replacementsStart++
		}
		switch {
		case chosen < len(shieldIndices):
			shield := &g.PreventionShields[shieldIndices[chosen]]
			prevented := amount
			if !shield.All {
				prevented = min(amount, shield.Amount)
				shield.Amount -= prevented
			}
			amount -= prevented
			g.PreventionShields = compactPreventionShields(g.PreventionShields)
			if prevented > 0 {
				emitDamagePreventedEvent(g, event, prevented)
			}
		case hasShieldCounter && chosen == shieldCounterIndex:
			// A shield counter prevents all of the remaining damage to its
			// permanent and is removed (CR 122.1c).
			event.permanent.Counters.Remove(counter.Shield, 1)
			emitDamagePreventedEvent(g, event, amount)
			amount = 0
		default:
			replacement := replacements[chosen-replacementsStart]
			appliedReplacements[replacement.ID] = true
			if replacement.DamageMultiplier > 1 {
				amount *= replacement.DamageMultiplier
			}
			amount += replacement.DamageAddend
			if amount < 0 {
				amount = 0
			}
		}
	}
	return amount
}

// damageEventProtected reports whether the damage event's recipient has protection
// from the source, which prevents all of the damage (CR 702.16e).
func damageEventProtected(g *game.Game, event damageEvent) bool {
	if event.permanent != nil {
		return permanentProtectedFromSource(g, event.permanent, event.sourceID, event.sourceObjectID)
	}
	return playerProtectedFromSource(g, event.player, event.sourceID, event.sourceObjectID, nil)
}

// applicablePreventionShieldIndices returns the indices of the prevention shields
// that currently apply to a damage event (CR 615.1): a global or matching shield
// that either prevents all damage or still has prevention left.
func applicablePreventionShieldIndices(g *game.Game, event damageEvent) []int {
	var indices []int
	for i := range g.PreventionShields {
		shield := g.PreventionShields[i]
		if !preventionShieldApplies(shield, event) || (!shield.All && shield.Amount <= 0) {
			continue
		}
		indices = append(indices, i)
	}
	return indices
}

// replaceDestroyPermanent applies effects that replace destroying a permanent
// (CR 614): a shield counter (CR 122.1c) or a regeneration shield (a
// destruction-replacement effect, CR 614.8) each replace the destroy event. When
// both are available, the permanent's controller chooses which to apply
// (CR 616.1). Returns whether the destroy was replaced.
func replaceDestroyPermanent(g *game.Game, permanent *game.Permanent, preventRegeneration bool) bool {
	if permanent == nil {
		return false
	}
	hasShieldCounter := permanent.Counters.Get(counter.Shield) > 0
	hasRegeneration := permanent.RegenerationShields > 0 && !preventRegeneration
	useShieldCounter := hasShieldCounter
	if hasShieldCounter && hasRegeneration {
		decision := chooseReplacementDecision(g, effectiveController(g, permanent), []string{"shield counter", "regeneration shield"})
		useShieldCounter = decision.Selected[0] == 0
	}
	if useShieldCounter {
		permanent.Counters.Remove(counter.Shield, 1)
		emitEvent(g, game.Event{
			Kind:        game.EventDestroyReplaced,
			Controller:  effectiveController(g, permanent),
			Player:      permanent.Owner,
			CardID:      permanent.CardInstanceID,
			FaceDown:    permanent.FaceDown,
			PermanentID: permanent.ObjectID,
			TokenName:   permanentTokenName(permanent),
			TokenDef:    permanent.TokenDef,
			FromZone:    zone.Battlefield,
			ToZone:      zone.Graveyard,
		})
		return true
	}
	if preventRegeneration {
		return false
	}
	return replaceDestroyWithRegeneration(g, permanent)
}

// replacementDecisionPlayer returns the CR 616.1 chooser for a damage event: the
// affected object's controller, or the affected player when no object is
// involved (CR 616.1: the affected object's controller, or owner if it has no
// controller, or the affected player).
func replacementDecisionPlayer(g *game.Game, event damageEvent) game.PlayerID {
	if event.permanent != nil {
		return effectiveController(g, event.permanent)
	}
	return event.player
}

// recordReplacementDecision records a deterministic replacement choice point for
// the turn log without prompting. It is used where the order among several
// applicable effects does not change the outcome (prevention shields, which
// reduce damage commutatively, and the independent entering-the-battlefield entry
// modifications, which all apply). Order-sensitive selections instead use
// chooseReplacementDecision, which makes a real CR 616.1 choice.
func recordReplacementDecision(g *game.Game, player game.PlayerID, options []string) game.ReplacementDecision {
	selected := make([]int, len(options))
	for i := range options {
		selected[i] = i
	}
	decision := game.ReplacementDecision{
		Player:       player,
		Options:      append([]string(nil), options...),
		Selected:     selected,
		UsedFallback: true,
	}
	g.ReplacementDecisions = append(g.ReplacementDecisions, decision)
	return decision
}

func selectedReplacementEffect(matches []game.ReplacementEffect, decision game.ReplacementDecision) game.ReplacementEffect {
	for _, selected := range decision.Selected {
		if selected >= 0 && selected < len(matches) {
			return matches[selected]
		}
	}
	return matches[0]
}

type zoneChangeReplacementResult struct {
	destination        zone.Type
	shuffleIntoLibrary bool
	revealSource       bool
}

func replacementZoneChangeDestination(g *game.Game, event game.Event) zone.Type {
	return replacementZoneChange(g, event).destination
}

// replacementZoneChange applies zone-change replacement effects to determine
// where a card actually goes (CR 614.1a, "instead" effects; e.g. a card that
// would be put into a graveyard is exiled instead). It loops because a replaced
// destination can match further replacement effects, tracking applied effects so
// each applies at most once (CR 614.5) and repeating until none remain
// (CR 616.1f); when several match at once the affected object's controller or the
// affected player chooses one (CR 616.1, falling back to the first match when no
// agent is available).
func replacementZoneChange(g *game.Game, event game.Event) zoneChangeReplacementResult {
	result := zoneChangeReplacementResult{destination: event.ToZone}
	applied := make(map[id.ID]bool)
	for {
		event.ToZone = result.destination
		matches := matchingZoneReplacementEffects(g, event, applied)
		if len(matches) == 0 {
			return result
		}
		replacement := matches[0]
		if len(matches) > 1 {
			decision := chooseReplacementDecision(g, replacementEventPlayer(event), replacementEffectLabels(matches))
			replacement = selectedReplacementEffect(matches, decision)
		}
		applied[replacement.ID] = true
		// After a replacement changes the event, the replacement process checks
		// the modified event again; the same effect cannot apply twice to one
		// event (CR 614.5, CR 616.1e).
		result.destination = replacement.ReplaceToZone
		result.shuffleIntoLibrary = replacement.ShuffleIntoLibrary && result.destination == zone.Library
		result.revealSource = result.revealSource || replacement.RevealSource
	}
}

// replacementTokenCreationTypes applies token-type replacement effects (Academy
// Manufactor: "If you would create a Clue, Food, or Treasure token, instead
// create one of each.") to a single token the controller would create. It
// returns the token definitions to create instead, which is just the original
// token when no replacement matches. The substitute tokens are created directly
// by callers without re-entering this function, so a matched replacement cannot
// recursively re-trigger itself.
func replacementTokenCreationTypes(g *game.Game, controller game.PlayerID, token *game.CardDef) []*game.CardDef {
	if token == nil {
		return nil
	}
	event := game.Event{
		Kind:       game.EventTokenCreated,
		Controller: controller,
		Player:     controller,
		Amount:     1,
	}
	for i := range g.ReplacementEffects {
		replacement := &g.ReplacementEffects[i]
		if replacement.TokenReplaceDef != nil {
			if !tokenHasAllSubtypes(token, replacement.TokenRequiredSubtypes) ||
				!tokenHasAllTypes(token, replacement.TokenRequiredTypes) {
				continue
			}
			if !replacementEffectMatchesEvent(g, replacement, event) {
				continue
			}
			// The identity substitution replaces the matched token with one copy
			// of the spelled-out substitute (Divine Visitation: a 4/4 Angel for
			// each creature token), so the substitute itself is created without
			// re-entering this matched replacement.
			return []*game.CardDef{replacement.TokenReplaceDef}
		}
		if len(replacement.CreateOneOfEachTokens) == 0 {
			continue
		}
		if !replacementEffectMatchesEvent(g, replacement, event) {
			continue
		}
		if !tokenNameInSet(token, replacement.CreateOneOfEachTokens) {
			continue
		}
		return replacement.CreateOneOfEachTokens
	}
	return []*game.CardDef{token}
}

// tokenNameInSet reports whether token shares a name with any definition in set,
// the trigger condition for a one-of-each token-type replacement.
func tokenNameInSet(token *game.CardDef, set []*game.CardDef) bool {
	for _, def := range set {
		if def != nil && def.Name == token.Name {
			return true
		}
	}
	return false
}

// replacementTokenCreationAmount applies replacement effects that change how many
// tokens are created (CR 614.16, effects that apply when an effect would create
// one or more tokens; e.g. doubling the number created). Each effect applies to
// the event at most once (CR 614.5); when several apply, the creating player
// chooses one (CR 616.1, falling back to the first match when no agent is
// available).
func replacementTokenCreationAmount(g *game.Game, controller game.PlayerID, token *game.CardDef, amount int) int {
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
		matches := matchingTokenCreationReplacementEffects(g, event, token, applied)
		if len(matches) == 0 {
			return amount
		}
		replacement := matches[0]
		if len(matches) > 1 {
			decision := chooseReplacementDecision(g, controller, replacementEffectLabels(matches))
			replacement = selectedReplacementEffect(matches, decision)
		}
		applied[replacement.ID] = true
		if replacement.TokenMultiplier > 1 {
			amount *= replacement.TokenMultiplier
		}
		amount += replacement.TokenAddend
	}
}

// replacementPermanentCounterPlacementAmount applies replacement effects that
// change how many counters are placed on a permanent (CR 614.16, e.g. doubling
// effects). Each effect applies at most once (CR 614.5); when several apply, the
// permanent's controller chooses one (CR 616.1, falling back to the first match
// when no agent is available).
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
		replacement := matches[0]
		if len(matches) > 1 {
			decision := chooseReplacementDecision(g, effectiveController(g, permanent), replacementEffectLabels(matches))
			replacement = selectedReplacementEffect(matches, decision)
		}
		applied[replacement.ID] = true
		if replacement.CounterMultiplier > 1 {
			amount *= replacement.CounterMultiplier
		}
		amount += replacement.CounterAddend
	}
}

// replacementPlayerCounterPlacementAmount applies replacement effects that change
// how many counters are placed on a player (CR 614.16). Each effect applies at
// most once (CR 614.5); when several apply, the affected player chooses one
// (CR 616.1, falling back to the first match when no agent is available).
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
		replacement := matches[0]
		if len(matches) > 1 {
			decision := chooseReplacementDecision(g, player, replacementEffectLabels(matches))
			replacement = selectedReplacementEffect(matches, decision)
		}
		applied[replacement.ID] = true
		if replacement.CounterMultiplier > 1 {
			amount *= replacement.CounterMultiplier
		}
		amount += replacement.CounterAddend
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

// applyEnterBattlefieldReplacementEffects applies the replacement effects that
// modify how a permanent enters the battlefield: "enters with" / "as this
// enters" / "enters as" effects (CR 614.1c) and "enters tapped"-style continuous
// effects (CR 614.1d). These include entering tapped, with counters, as a copy,
// and the keyword entry choices (unleash, riot, devour, tribute); required
// choices are made before the permanent enters (CR 614.12a). These entry
// modifications are independent and all apply, so the engine applies them all in a
// deterministic order rather than making the CR 616.1 one-at-a-time choice (the
// order does not change the result); the choice point is recorded for the log.
func applyEnterBattlefieldReplacementEffects(ctx enterBattlefieldContext, g *game.Game, permanent *game.Permanent, fromZone zone.Type) {
	event := game.Event{
		Kind:        game.EventPermanentEnteredBattlefield,
		Controller:  effectiveController(g, permanent),
		Player:      permanent.Owner,
		CardID:      permanent.CardInstanceID,
		PermanentID: permanent.ObjectID,
		TokenName:   permanentTokenName(permanent),
		TokenDef:    permanent.TokenDef,
		KickerPaid:  ctx.kickerPaid,
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
	matches := matchingETBReplacementEffects(g, permanent, event)
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
			amount := placement.Amount
			if placement.AmountFromX {
				amount = ctx.xValue
			}
			if placement.Dynamic.Exists && placement.Dynamic.Val != nil {
				// A group enters-with-counters replacement ("Each other creature
				// you control enters with ... equal to <source>'s toughness." —
				// Arwen, Weaver of Hope) scales by a characteristic of the
				// replacement's source permanent, not the entering one, so the
				// dynamic amount resolves against the source object. A self
				// replacement reads the entering permanent itself.
				sourceID := permanent.ObjectID
				sourceCardID := permanent.CardInstanceID
				if replacement.EntersWithCountersOthers {
					sourceID = replacement.SourceObjectID
					sourceCardID = replacement.SourceCardID
				}
				obj := &game.StackObject{
					SourceID:                sourceID,
					SourceCardID:            sourceCardID,
					Controller:              replacement.Controller,
					ColorsOfManaSpentToCast: ctx.colorsOfManaSpent,
					KickerCount:             ctx.kickCount,
				}
				amount = dynamicAmountValue(g, obj, replacement.Controller, *placement.Dynamic.Val)
			}
			if amount <= 0 {
				continue
			}
			addCountersToPermanent(g, permanent, placement.Kind, amount)
		}
		if replacement.EntryColorChoice {
			applyEntryColorChoice(ctx, g, permanent, replacement.Controller, replacement.EntryColorChoiceExclude)
		}
		if replacement.EntryTypeChoice {
			applyEntryTypeChoice(ctx, g, permanent, replacement.Controller)
		}
		if replacement.EntryDevourMultiplier > 0 {
			applyEntryDevour(ctx, g, permanent, replacement.Controller, replacement.EntryDevourMultiplier, replacement.EntryDevourType, replacement.EntryDevourSubtype)
		}
		if replacement.EntryTributeCount > 0 {
			applyEntryTribute(ctx, g, permanent, replacement.Controller, replacement.EntryTributeCount)
		}
		if replacement.EntersAsCopy {
			applyEntersAsCopy(ctx, g, permanent, replacement)
		}
	}
	if hasKeyword(g, permanent, game.Riot) {
		applyEntryRiotChoice(ctx, g, permanent)
	}
	if hasKeyword(g, permanent, game.Unleash) {
		applyEntryUnleashChoice(ctx, g, permanent)
	}
}

// applyEntryUnleashChoice resolves the unleash keyword for an entering permanent
// (CR 702.86): its controller may have it enter with a +1/+1 counter on it. A
// permanent that entered with such a counter can't block while it still has one;
// that restriction is enforced where block legality is determined.
func applyEntryUnleashChoice(ctx enterBattlefieldContext, g *game.Game, permanent *game.Permanent) {
	engine := ctx.engine
	if engine == nil {
		engine = NewEngine(nil)
	}
	controller := effectiveController(g, permanent)
	request := game.ChoiceRequest{
		Kind:   game.ChoiceModal,
		Player: controller,
		Prompt: "Unleash: choose whether to enter with a +1/+1 counter.",
		Options: []game.ChoiceOption{
			{Index: 0, Label: "Enter with a +1/+1 counter"},
			{Index: 1, Label: "Enter without a counter"},
		},
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}
	selected := engine.chooseChoice(g, ctx.agents, request, ctx.log)
	if len(selected) == 1 && selected[0] == 1 {
		return
	}
	addCountersToPermanent(g, permanent, counter.PlusOnePlusOne, 1)
}

// applyEntryRiotChoice resolves the riot keyword for an entering permanent
// (CR 702.137): its controller chooses for it to enter with a +1/+1 counter or
// to gain haste. The haste choice clears summoning sickness, which is equivalent
// to granting haste for a permanent that is entering now and uniformly enables
// both attacking and activating tap abilities this turn.
func applyEntryRiotChoice(ctx enterBattlefieldContext, g *game.Game, permanent *game.Permanent) {
	engine := ctx.engine
	if engine == nil {
		engine = NewEngine(nil)
	}
	controller := effectiveController(g, permanent)
	request := game.ChoiceRequest{
		Kind:   game.ChoiceModal,
		Player: controller,
		Prompt: "Riot: choose to enter with a +1/+1 counter or gain haste.",
		Options: []game.ChoiceOption{
			{Index: 0, Label: "Enter with a +1/+1 counter"},
			{Index: 1, Label: "Gain haste"},
		},
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}
	selected := engine.chooseChoice(g, ctx.agents, request, ctx.log)
	if len(selected) == 1 && selected[0] == 1 {
		permanent.SummoningSick = false
		return
	}
	addCountersToPermanent(g, permanent, counter.PlusOnePlusOne, 1)
}

// applyEntryColorChoice prompts the permanent's controller to choose a color as
// the permanent enters and stores the result on the permanent under
// EntryColorChoiceKey (CR 614.12). Later abilities such as "{T}: Add one mana of
// the chosen color." read the stored color. A non-empty exclude color is removed
// from the prompt ("choose a color other than <color>").
func applyEntryColorChoice(ctx enterBattlefieldContext, g *game.Game, permanent *game.Permanent, controller game.PlayerID, exclude mana.Color) {
	engine := ctx.engine
	if engine == nil {
		engine = NewEngine(nil)
	}
	colors := make([]mana.Color, 0, 5)
	for _, c := range []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G} {
		if c != exclude {
			colors = append(colors, c)
		}
	}
	choice := game.ResolutionChoice{
		Kind:   game.ResolutionChoiceMana,
		Prompt: entryColorChoicePrompt(exclude),
		Colors: colors,
	}
	result, ok := engine.chooseEntryColor(g, ctx.agents, controller, &choice, ctx.log)
	if !ok {
		return
	}
	if permanent.EntryChoices == nil {
		permanent.EntryChoices = make(map[game.ChoiceKey]game.ResolutionChoiceResult)
	}
	permanent.EntryChoices[game.EntryColorChoiceKey] = result
}

// applyEntryTypeChoice prompts the permanent's controller to choose a creature
// type as the permanent enters and stores the result on the permanent under
// EntryTypeChoiceKey (CR 614.12). Later abilities that reference "the chosen
// type" read the stored subtype.
func applyEntryTypeChoice(ctx enterBattlefieldContext, g *game.Game, permanent *game.Permanent, controller game.PlayerID) {
	engine := ctx.engine
	if engine == nil {
		engine = NewEngine(nil)
	}
	choice := game.ResolutionChoice{
		Kind:          game.ResolutionChoiceSubtype,
		Prompt:        "Choose a creature type.",
		SubtypeOfType: types.Creature,
	}
	result, ok := engine.chooseEntryColor(g, ctx.agents, controller, &choice, ctx.log)
	if !ok {
		return
	}
	if permanent.EntryChoices == nil {
		permanent.EntryChoices = make(map[game.ChoiceKey]game.ResolutionChoiceResult)
	}
	permanent.EntryChoices[game.EntryTypeChoiceKey] = result
}

// applyEntryDevour resolves the Devour keyword for an entering permanent (CR
// 702.81): its controller may sacrifice any number of other matching permanents
// they control as it enters, and it enters with multiplier +1/+1 counters on it
// for each one sacrificed this way. The matching permanents are creatures for the
// plain "Devour N" form; the typed variants restrict the choice to a card type
// (cardType, for "Devour artifact N"/"Devour land N") or a subtype (subtype, for
// "Devour Food N"). Choosing to sacrifice nothing is legal and is the default.
func applyEntryDevour(ctx enterBattlefieldContext, g *game.Game, permanent *game.Permanent, controller game.PlayerID, multiplier int, cardType types.Card, subtype types.Sub) {
	engine := ctx.engine
	if engine == nil {
		engine = NewEngine(nil)
	}
	var candidates []*game.Permanent
	for _, candidate := range g.Battlefield {
		if candidate.ObjectID == permanent.ObjectID {
			continue
		}
		if effectiveController(g, candidate) != controller {
			continue
		}
		if !devourCandidateMatches(g, candidate, cardType, subtype) {
			continue
		}
		candidates = append(candidates, candidate)
	}
	if len(candidates) == 0 {
		return
	}
	options := make([]game.ChoiceOption, len(candidates))
	for i, candidate := range candidates {
		options[i] = game.ChoiceOption{Index: i, Label: permanentChoiceLabel(g, candidate), Card: permanentChoiceInfo(g, candidate)}
	}
	request := game.ChoiceRequest{
		Kind:       game.ChoicePayment,
		Player:     controller,
		Prompt:     devourPrompt(cardType, subtype),
		Options:    options,
		MinChoices: 0,
		MaxChoices: len(candidates),
	}
	selected := engine.chooseChoice(g, ctx.agents, request, ctx.log)
	sacrificed := make([]*game.Permanent, 0, len(selected))
	for _, index := range selected {
		if index >= 0 && index < len(candidates) {
			sacrificed = append(sacrificed, candidates[index])
		}
	}
	if len(sacrificed) == 0 {
		return
	}
	sacrificePermanentsSimultaneously(g, sacrificed)
	addCountersToPermanent(g, permanent, counter.PlusOnePlusOne, multiplier*len(sacrificed))
}

// devourCandidateMatches reports whether a permanent may be sacrificed to a
// Devour replacement. A non-empty subtype restricts the choice to permanents
// with that subtype ("Devour Food N"); otherwise a non-empty cardType restricts
// it to that card type ("Devour artifact N"/"Devour land N"); the plain creature
// form ("Devour N") leaves both empty and matches creatures.
func devourCandidateMatches(g *game.Game, candidate *game.Permanent, cardType types.Card, subtype types.Sub) bool {
	if subtype != "" {
		return permanentHasSubtype(g, candidate, subtype)
	}
	if cardType != "" {
		return permanentHasType(g, candidate, cardType)
	}
	return permanentHasType(g, candidate, types.Creature)
}

// devourPrompt builds the sacrifice prompt for a Devour replacement, naming the
// permanents that may be sacrificed for the creature form or a typed variant.
func devourPrompt(cardType types.Card, subtype types.Sub) string {
	noun := "creatures"
	if subtype != "" {
		noun = string(subtype) + "s"
	} else if cardType != "" {
		noun = strings.ToLower(string(cardType)) + "s"
	}
	return "Devour: choose any number of " + noun + " to sacrifice."
}

// applyEntryTribute resolves the Tribute keyword for an entering permanent (CR
// 702.110): an opponent of the controller's choice may put count +1/+1 counters
// on it as it enters. The controller first chooses which opponent (when more
// than one is alive); that opponent then decides whether to add the counters.
// Adding them sets the permanent's TributePaid flag so a paired "if tribute
// wasn't paid" intervening-if can react. Declining (or having no opponent) leaves
// the flag unset.
func applyEntryTribute(ctx enterBattlefieldContext, g *game.Game, permanent *game.Permanent, controller game.PlayerID, count int) {
	engine := ctx.engine
	if engine == nil {
		engine = NewEngine(nil)
	}
	opponents := aliveOpponents(g, controller)
	if len(opponents) == 0 {
		return
	}
	chosenOpponent := opponents[0]
	if len(opponents) > 1 {
		options := make([]game.ChoiceOption, len(opponents))
		for i, opponent := range opponents {
			options[i] = game.ChoiceOption{Index: i, Label: fmt.Sprintf("Player %d", opponent+1)}
		}
		request := game.ChoiceRequest{
			Kind:             game.ChoiceModal,
			Player:           controller,
			Prompt:           "Tribute: choose an opponent.",
			Options:          options,
			MinChoices:       1,
			MaxChoices:       1,
			DefaultSelection: []int{0},
		}
		selected := engine.chooseChoice(g, ctx.agents, request, ctx.log)
		if len(selected) == 1 && selected[0] >= 0 && selected[0] < len(opponents) {
			chosenOpponent = opponents[selected[0]]
		}
	}
	request := game.ChoiceRequest{
		Kind:   game.ChoiceModal,
		Player: chosenOpponent,
		Prompt: fmt.Sprintf("Tribute: put %d +1/+1 counters on it, or decline.", count),
		Options: []game.ChoiceOption{
			{Index: 0, Label: "Put the counters on it"},
			{Index: 1, Label: "Decline"},
		},
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}
	selected := engine.chooseChoice(g, ctx.agents, request, ctx.log)
	if len(selected) == 1 && selected[0] == 1 {
		return
	}
	addCountersToPermanent(g, permanent, counter.PlusOnePlusOne, count)
	permanent.TributePaid = true
}

// applyEntersAsCopy has the entering permanent enter as a copy of a permanent
// its controller chooses among those matching the replacement's selection
// (CR 706, CR 614). The optional form first asks whether to copy at all; the
// copy is registered as a layer-1 continuous effect lasting as long as the
// entering permanent stays on the battlefield, with the recognized copiable
// riders ("isn't legendary", "is an <type> in addition to its other types")
// baked into the copied values.
func applyEntersAsCopy(ctx enterBattlefieldContext, g *game.Game, permanent *game.Permanent, replacement *game.ReplacementEffect) {
	if replacement.EntersAsCopySelection == nil {
		return
	}
	engine := ctx.engine
	if engine == nil {
		engine = NewEngine(nil)
	}
	controller := effectiveController(g, permanent)
	obj := &game.StackObject{Controller: controller, SourceID: permanent.ObjectID}
	resolver := newReferenceResolverWithSource(g, obj, permanent)
	var candidates []*game.Permanent
	for _, candidate := range g.Battlefield {
		if candidate.ObjectID == permanent.ObjectID {
			continue
		}
		if !resolver.permanentMatchesGroupSelection(replacement.EntersAsCopySelection, permanent, candidate) {
			continue
		}
		candidates = append(candidates, candidate)
	}
	if len(candidates) == 0 {
		return
	}
	if replacement.EntersAsCopyOptional {
		selected := engine.chooseChoice(g, ctx.agents, mayChoiceRequest(controller, "Have this permanent enter as a copy?"), ctx.log)
		if len(selected) != 1 || selected[0] != 1 {
			return
		}
	}
	chosen := candidates[0]
	if len(candidates) > 1 {
		options := make([]game.ChoiceOption, len(candidates))
		for i, candidate := range candidates {
			options[i] = game.ChoiceOption{Index: i, Label: permanentChoiceLabel(g, candidate), Card: permanentChoiceInfo(g, candidate)}
		}
		request := game.ChoiceRequest{
			Kind:             game.ChoiceTarget,
			Player:           controller,
			Prompt:           "Choose a permanent to copy",
			Options:          options,
			MinChoices:       1,
			MaxChoices:       1,
			DefaultSelection: []int{0},
		}
		selected := engine.chooseChoice(g, ctx.agents, request, ctx.log)
		if len(selected) == 1 && selected[0] >= 0 && selected[0] < len(candidates) {
			chosen = candidates[selected[0]]
		}
	}
	def, ok := permanentCopyDef(g, chosen)
	if !ok {
		return
	}
	values := copyableValuesFromDef(def)
	if replacement.EntersAsCopyNotLegendary {
		values.Supertypes = slices.DeleteFunc(values.Supertypes, func(super types.Super) bool {
			return super == types.Legendary
		})
	}
	for _, cardType := range replacement.EntersAsCopyAddTypes {
		if !slices.Contains(values.Types, cardType) {
			values.Types = append(values.Types, cardType)
		}
	}
	for _, subtype := range replacement.EntersAsCopyAddSubtypes {
		if !slices.Contains(values.Subtypes, subtype) {
			values.Subtypes = append(values.Subtypes, subtype)
		}
	}
	for _, keyword := range replacement.EntersAsCopyAddKeywords {
		body, ok := game.KeywordStaticBody(keyword)
		if !ok {
			continue
		}
		static := body
		values.Abilities = append(values.Abilities, &static)
	}
	duration := game.DurationForAsLongAsSourceOnBattlefield
	if replacement.EntersAsCopyUntilEndOfTurn {
		duration = game.DurationUntilEndOfTurn
	}
	effectID := g.IDGen.Next()
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:               effectID,
		SourceObjectID:   permanent.ObjectID,
		SourceCardID:     permanent.CardInstanceID,
		Controller:       controller,
		Timestamp:        permanent.Timestamp(),
		Duration:         duration,
		CreatedTurn:      g.Turn.TurnNumber,
		AffectedObjectID: permanent.ObjectID,
		Layer:            game.LayerCopy,
		CopyValues:       opt.Val(values),
	})
	for _, placement := range replacement.EntersAsCopyConditionalCounters {
		if placement.Amount <= 0 {
			continue
		}
		if slices.Contains(values.Types, placement.IfType) {
			addCountersToPermanent(g, permanent, placement.Kind, placement.Amount)
		}
	}
	if replacement.EntersAsCopyTapped {
		setPermanentTapped(g, permanent, true)
	}
}

// copyableValuesFromDef snapshots a card definition's copiable characteristics
// (CR 706.2): name, colors, supertypes, types, subtypes, printed power and
// toughness, abilities, and oracle text.
func copyableValuesFromDef(def *game.CardDef) game.CopyableValues {
	values := game.CopyableValues{
		Name:             def.Name,
		Colors:           append([]color.Color(nil), def.Colors...),
		Supertypes:       append([]types.Super(nil), def.Supertypes...),
		Types:            append([]types.Card(nil), def.Types...),
		Subtypes:         append([]types.Sub(nil), def.Subtypes...),
		Power:            def.Power,
		Toughness:        def.Toughness,
		DynamicPower:     def.DynamicPower,
		DynamicToughness: def.DynamicToughness,
		OracleText:       def.OracleText,
	}
	values.Abilities = make([]game.Ability, 0, def.AbilityCount())
	for i := 0; i < def.AbilityCount(); i++ {
		values.Abilities = append(values.Abilities, def.BodyAt(i))
	}
	return values
}

// entryColorChoicePrompt builds the prompt for an entry-time color choice,
// naming the excluded color when the choice is constrained.
func entryColorChoicePrompt(exclude mana.Color) string {
	if name, ok := colorWordForMana(exclude); ok {
		return "Choose a color other than " + name + "."
	}
	return "Choose a color."
}

// colorWordForMana maps a basic mana color to its English color word.
func colorWordForMana(c mana.Color) (string, bool) {
	switch c {
	case mana.W:
		return "white", true
	case mana.U:
		return "blue", true
	case mana.B:
		return "black", true
	case mana.R:
		return "red", true
	case mana.G:
		return "green", true
	default:
		return "", false
	}
}

func staticETBReplacementEffects(ctx enterBattlefieldContext, g *game.Game, permanent *game.Permanent, def *game.CardDef, event game.Event) []game.ReplacementEffect {
	var replacements []game.ReplacementEffect
	for i := range def.ReplacementAbilities {
		ability := &def.ReplacementAbilities[i]
		// Optional entry-to-zone replacements (Mox Diamond) are resolved before
		// entry by optionalEntryReplacementDeclined; skip them here so the
		// payment is not prompted twice.
		if ability.Replacement.ReplaceToZone != zone.None {
			continue
		}
		replacement := ability.Replacement
		// Group enters-with-counters replacements ("Each other creature you
		// control enters ...") apply to OTHER entering permanents through the
		// registered global matching path, not to the source itself; skip them
		// here so the source does not gain its own group counters on entry.
		if replacement.EntersWithCountersOthers {
			continue
		}
		replacement.Controller = event.Controller
		replacement.SourceCardID = permanent.CardInstanceID
		replacement.SourceObjectID = 0
		replacement.CreatedTurn = g.Turn.TurnNumber
		if replacement.Description == "" {
			replacement.Description = ability.Text
		}
		if ability.UnlessPaid.Exists && enterBattlefieldPaymentPaid(ctx, g, event.Controller, permanent, ability.UnlessPaid.Val) {
			continue
		}
		// A self enters-with-counters replacement may carry an Adamant condition
		// ("if at least three <color> mana was spent to cast this spell"), which
		// reads the resolving spell's per-color mana spend. The cast-time tallies
		// travel on the entry context, so synthesize an object that carries them
		// into the condition evaluation.
		obj := &game.StackObject{
			Controller:              event.Controller,
			ColorsOfManaSpentToCast: ctx.colorsOfManaSpent,
			ManaSpentByColorToCast:  ctx.manaSpentByColor,
			KickerCount:             ctx.kickCount,
		}
		if replacementEffectMatchesEventWithSource(g, &replacement, event, permanent, obj) {
			replacements = append(replacements, replacement)
		}
	}
	return replacements
}

// optionalEntryReplacementDeclined resolves an optional self enters-the-
// battlefield replacement that lets the controller pay an alternative cost to
// keep the permanent on the battlefield, or sends it to another zone if the
// cost is not paid (Mox Diamond). It returns true when the permanent must not
// enter because the controller declined or could not pay; the card has then
// been moved from fromZone to the replacement zone.
func optionalEntryReplacementDeclined(ctx enterBattlefieldContext, g *game.Game, card *game.CardInstance, permanent *game.Permanent, def *game.CardDef, fromZone zone.Type) bool {
	if def == nil || card == nil || permanent.FaceDown {
		return false
	}
	for i := range def.ReplacementAbilities {
		ability := &def.ReplacementAbilities[i]
		if !ability.UnlessPaid.Exists || ability.Replacement.ReplaceToZone == zone.None {
			continue
		}
		if enterBattlefieldPaymentPaid(ctx, g, permanent.Controller, permanent, ability.UnlessPaid.Val) {
			return false
		}
		moveCardBetweenZones(g, card.Owner, card.ID, fromZone, ability.Replacement.ReplaceToZone)
		return true
	}
	return false
}

func enterBattlefieldPaymentPaid(ctx enterBattlefieldContext, g *game.Game, playerID game.PlayerID, source *game.Permanent, res game.ResolutionPayment) bool {
	if source == nil {
		return false
	}
	entryObject := &game.StackObject{
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   playerID,
		XValue:       res.XValue,
	}
	resolved, ok := materializeResolutionPayment(g, entryObject, source, &res)
	if !ok {
		return false
	}
	res = resolved
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
	prefs := engine.paymentPreferencesForCost(g, playerID, manaCostPtr(res.ManaCost), res.AdditionalCosts, res.XValue, ctx.agents, ctx.log)
	return paymentOrch.payGenericCost(g, payment.GenericRequest{
		PlayerID:        playerID,
		SourceCardID:    source.CardInstanceID,
		Cost:            manaCostPtr(res.ManaCost),
		XValue:          res.XValue,
		AdditionalCosts: res.AdditionalCosts,
		Prefs:           prefs,
	})
}

// matchingZoneReplacementEffects returns the zone-change replacement effects that
// apply to an event and have not already been applied to it (CR 614.5). It
// includes both registered replacement effects and a card's own static
// zone-change replacement effects ("if this would be put into a graveyard, exile
// it instead"). When several apply, the affected player chooses which to apply
// (CR 616.1); the engine does not model CR 614.15 self-replacement effects, so the
// CR 616.1a self-replacement-first ordering does not arise here.
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
		// Continuous graveyard-redirect replacements are registered into
		// g.ReplacementEffects while their source is on the battlefield and
		// matched there against every moving card; the self-zone path must not
		// re-apply them from the printed card definition.
		if replacement.ContinuousZoneRedirect {
			continue
		}
		// Optional entry-to-zone replacements are handled by the entry-payment
		// path (optionalEntryReplacementDeclined); they must not redirect a
		// generic zone change without prompting for their cost.
		if ability.UnlessPaid.Exists {
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

func matchingETBReplacementEffects(g *game.Game, permanent *game.Permanent, event game.Event) []game.ReplacementEffect {
	source := permanent
	if source == nil {
		source, _ = permanentByObjectID(g, event.PermanentID)
	}
	var matches []game.ReplacementEffect
	for i := range g.ReplacementEffects {
		replacement := &g.ReplacementEffects[i]
		if !replacement.EntersTapped && len(replacement.EntersWithCounters) == 0 {
			continue
		}
		if replacement.EntersTappedOthers {
			if replacement.SourceObjectID != 0 && replacement.SourceObjectID == event.PermanentID {
				continue
			}
			if !entersTappedGroupSelectionMatches(g, replacement, source) {
				continue
			}
		}
		if replacement.EntersWithCountersOthers {
			if replacement.SourceObjectID != 0 && replacement.SourceObjectID == event.PermanentID {
				continue
			}
			if !entersWithCountersGroupMatches(g, replacement, source) {
				continue
			}
		}
		if !replacementEffectMatchesEventWithSource(g, replacement, event, source, nil) {
			continue
		}
		matches = append(matches, *replacement)
	}
	return matches
}

// entersTappedGroupSelectionMatches reports whether the entering permanent
// satisfies the permanent characteristic filter of a group enters-tapped
// replacement, matched through the canonical matchSelection. A nil selection
// taps every entering permanent.
func entersTappedGroupSelectionMatches(g *game.Game, replacement *game.ReplacementEffect, permanent *game.Permanent) bool {
	if replacement.EntersTappedSelection == nil {
		return true
	}
	if permanent == nil {
		return false
	}
	return permanentMatchesReplacementSelection(g, permanent, replacement.EntersTappedSelection)
}

// entersWithCountersGroupMatches reports whether the entering permanent is a
// recipient of a group enters-with-counters replacement ("Each other creature
// you control enters ..."). The recipient selection carries the controller
// scope and the source-exclusion ("other") rider, evaluated relative to the
// replacement's source permanent and controller.
func entersWithCountersGroupMatches(g *game.Game, replacement *game.ReplacementEffect, permanent *game.Permanent) bool {
	if replacement.EntersWithCountersRecipient == nil || permanent == nil {
		return false
	}
	sourcePermanent, _ := permanentByObjectID(g, replacement.SourceObjectID)
	controller := replacementCurrentController(g, replacement)
	resolver := newReferenceResolver(g, &game.StackObject{
		SourceID:     replacement.SourceObjectID,
		SourceCardID: replacement.SourceCardID,
		Controller:   controller,
	})
	return resolver.permanentMatchesGroupSelection(replacement.EntersWithCountersRecipient, sourcePermanent, permanent)
}

func matchingTokenCreationReplacementEffects(g *game.Game, event game.Event, token *game.CardDef, applied map[id.ID]bool) []game.ReplacementEffect {
	var matches []game.ReplacementEffect
	for i := range g.ReplacementEffects {
		replacement := &g.ReplacementEffects[i]
		if replacement.TokenMultiplier <= 1 && replacement.TokenAddend == 0 {
			continue
		}
		// Cross-type addends create a different predefined token (Tippy-Toe's Food)
		// rather than more of the matched token, so they are handled separately and
		// must not inflate the matched-token count here.
		if replacement.TokenAddendDef != nil {
			continue
		}
		if !tokenHasAllSubtypes(token, replacement.TokenRequiredSubtypes) {
			continue
		}
		if !tokenHasAllTypes(token, replacement.TokenRequiredTypes) {
			continue
		}
		if applied[replacement.ID] || !replacementEffectMatchesEvent(g, replacement, event) {
			continue
		}
		matches = append(matches, *replacement)
	}
	return matches
}

// tokenHasAllSubtypes reports whether the created token carries every subtype in
// the replacement's filter. An empty filter matches any token; a nil token never
// satisfies a non-empty filter.
func tokenHasAllSubtypes(token *game.CardDef, required []types.Sub) bool {
	if len(required) == 0 {
		return true
	}
	if token == nil {
		return false
	}
	for _, sub := range required {
		if !slices.Contains(token.Subtypes, sub) {
			return false
		}
	}
	return true
}

// tokenHasAllTypes reports whether the created token carries every card type in
// the replacement's filter. An empty filter matches any token; a nil token never
// satisfies a non-empty filter.
func tokenHasAllTypes(token *game.CardDef, required []types.Card) bool {
	if len(required) == 0 {
		return true
	}
	if token == nil {
		return false
	}
	for _, cardType := range required {
		if !slices.Contains(token.Types, cardType) {
			return false
		}
	}
	return true
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
		if replacement.CounterMultiplier <= 1 && replacement.CounterAddend == 0 {
			continue
		}
		if replacement.MatchCounterKind && replacement.CounterKindFilter != event.CounterKind {
			continue
		}
		if replacement.CounterRecipientSelection != nil && !counterRecipientMatchesSelection(g, event.PermanentID, recipient, replacement.CounterRecipientSelection, replacement.SourceObjectID) {
			continue
		}
		if replacement.CounterRecipientAnyPermanent && !counterRecipientMatchesSelection(g, event.PermanentID, recipient, &game.Selection{}, replacement.SourceObjectID) {
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

func matchingDamageReplacementEffects(g *game.Game, event damageEvent, applied map[id.ID]bool) []game.ReplacementEffect {
	matchEvent := damageReplacementEvent(g, event)
	var matches []game.ReplacementEffect
	for i := range g.ReplacementEffects {
		replacement := &g.ReplacementEffects[i]
		if replacement.DamageMultiplier <= 1 && replacement.DamageAddend == 0 {
			continue
		}
		if len(replacement.DamageSourceColors) > 0 && !damageSourceHasAnyColor(g, event, replacement.DamageSourceColors) {
			continue
		}
		if len(replacement.DamageSourceTypes) > 0 && !damageSourceHasAllTypes(g, event, replacement.DamageSourceTypes) {
			continue
		}
		if replacement.DamageNoncombatOnly && event.combatDamage {
			continue
		}
		if replacement.DamageRecipientOpponent && !damageRecipientIsOpponent(g, event, replacement) {
			continue
		}
		if replacement.DamageExcludeSource && damageSourceIsReplacementSource(event, replacement) {
			continue
		}
		if applied[replacement.ID] || !replacementEffectMatchesEvent(g, replacement, matchEvent) {
			continue
		}
		matches = append(matches, *replacement)
	}
	return matches
}

func damageRecipientIsOpponent(g *game.Game, event damageEvent, replacement *game.ReplacementEffect) bool {
	controller := replacementCurrentController(g, replacement)
	var recipient game.PlayerID
	if event.permanent != nil {
		recipient = effectiveController(g, event.permanent)
	} else {
		recipient = event.player
	}
	return recipient != controller
}

func damageSourceHasAllTypes(g *game.Game, event damageEvent, requiredTypes []types.Card) bool {
	if event.sourceObjectID != 0 {
		if permanent, ok := permanentByObjectID(g, event.sourceObjectID); ok {
			for _, cardType := range requiredTypes {
				if !permanentHasType(g, permanent, cardType) {
					return false
				}
			}
			return true
		}
	}
	def, ok := damageSourceDef(g, event.sourceID, event.sourceObjectID)
	if !ok || def == nil {
		return false
	}
	for _, cardType := range requiredTypes {
		if !slices.Contains(def.Types, cardType) {
			return false
		}
	}
	return true
}

func damageSourceIsReplacementSource(event damageEvent, replacement *game.ReplacementEffect) bool {
	return (event.sourceObjectID != 0 && event.sourceObjectID == replacement.SourceObjectID) ||
		(event.sourceID != 0 && event.sourceID == replacement.SourceCardID)
}

func damageReplacementEvent(g *game.Game, event damageEvent) game.Event {
	matchEvent := game.Event{
		Kind:           game.EventDamageDealt,
		SourceID:       event.sourceID,
		SourceObjectID: event.sourceObjectID,
		Controller:     event.controller,
		Player:         event.player,
		Amount:         event.amount,
		Colors:         damageSourceColors(g, event),
		CombatDamage:   event.combatDamage,
	}
	if event.permanent != nil {
		matchEvent.Player = event.permanent.Owner
		matchEvent.PermanentID = event.permanent.ObjectID
		matchEvent.CardID = event.permanent.CardInstanceID
		matchEvent.TokenName = permanentTokenName(event.permanent)
		matchEvent.TokenDef = event.permanent.TokenDef
		matchEvent.DamageRecipient = game.DamageRecipientPermanent
	} else {
		matchEvent.DamageRecipient = game.DamageRecipientPlayer
	}
	return matchEvent
}

func damageSourceHasAnyColor(g *game.Game, event damageEvent, colors []color.Color) bool {
	sourceColors := damageSourceColors(g, event)
	return slices.ContainsFunc(colors, func(c color.Color) bool {
		return slices.Contains(sourceColors, c)
	})
}

func damageSourceColors(g *game.Game, event damageEvent) []color.Color {
	if event.sourceObjectID != 0 {
		if permanent, ok := permanentByObjectID(g, event.sourceObjectID); ok {
			return permanentEffectiveColors(g, permanent)
		}
		if snapshot, ok := lastKnownObject(g, event.sourceObjectID); ok {
			return append([]color.Color(nil), snapshot.Colors...)
		}
	}
	if source, ok := damageSourceDef(g, event.sourceID, event.sourceObjectID); ok && source != nil {
		return append([]color.Color(nil), source.Colors...)
	}
	return nil
}

// counterRecipientPermanent resolves the permanent a counter-placement event
// would add counters to, preferring the supplied recipient and falling back to
// the event's permanent ID. A player recipient (no permanent) does not resolve.
func counterRecipientPermanent(g *game.Game, permanentID id.ID, permanent *game.Permanent) (*game.Permanent, bool) {
	if permanent != nil {
		return permanent, true
	}
	if permanentID == 0 {
		return nil, false
	}
	return permanentByObjectID(g, permanentID)
}

// counterRecipientMatchesSelection reports whether the counter recipient
// permanent satisfies sel, matched through the canonical matchSelection. It
// reads the recipient's effective characteristics, the same values the legacy
// per-type checks read through permanentHasType.
func counterRecipientMatchesSelection(g *game.Game, permanentID id.ID, permanent *game.Permanent, sel *game.Selection, sourceObjectID id.ID) bool {
	recipient, ok := counterRecipientPermanent(g, permanentID, permanent)
	if !ok {
		return false
	}
	return permanentMatchesReplacementSelectionFromSource(g, recipient, sel, sourceObjectID)
}

// permanentMatchesReplacementSelection matches a replacement's object-characteristic
// Selection against a live permanent through the shared matchSelection, reading
// the permanent's effective values. Replacement recipient filters carry no
// controller relativity (controller scope lives outside the Selection), so the
// subject's viewer is irrelevant.
func permanentMatchesReplacementSelection(g *game.Game, permanent *game.Permanent, sel *game.Selection) bool {
	return permanentMatchesReplacementSelectionFromSource(g, permanent, sel, 0)
}

// permanentMatchesReplacementSelectionFromSource matches as
// permanentMatchesReplacementSelection but supplies the replacement's own source
// object so an ExcludeSource recipient filter ("another creature you control",
// Benevolent Hydra) drops the source permanent from the match.
func permanentMatchesReplacementSelectionFromSource(g *game.Game, permanent *game.Permanent, sel *game.Selection, sourceObjectID id.ID) bool {
	values := effectivePermanentValues(g, permanent)
	subject := selectionSubject{
		kind:           subjectPermanent,
		g:              g,
		permanent:      permanent,
		values:         &values,
		sourceObjectID: sourceObjectID,
	}
	return matchSelection(&subject, sel)
}

func replacementEffectMatchesEvent(g *game.Game, replacement *game.ReplacementEffect, event game.Event) bool {
	return replacementEffectMatchesEventWithSource(g, replacement, event, nil, nil)
}

// replacementEffectMatchesEventWithSource reports whether the replacement
// applies to the event. A non-nil source supplies the source permanent to the
// condition context so source-relative condition predicates (such as the
// EventHistory "this turn" conditions on enters-with-counters replacements)
// resolve "you" against the replacement's own permanent. A non-nil obj supplies
// the resolving spell's cast-time data (mana spent by color) so an Adamant self
// enters-with-counters condition resolves as the permanent enters. The
// source-less replacementEffectMatchesEvent wrapper preserves the prior
// behavior for every other replacement category.
func replacementEffectMatchesEventWithSource(g *game.Game, replacement *game.ReplacementEffect, event game.Event, source *game.Permanent, obj *game.StackObject) bool {
	if !replacementSourceIsActive(g, replacement) {
		return false
	}
	if replacement.MatchEvent != game.EventUnknown && replacement.MatchEvent != event.Kind {
		return false
	}
	if replacement.AffectedObjectID != 0 && replacement.AffectedObjectID != event.PermanentID {
		return false
	}
	if replacement.AffectedCardID != 0 && replacement.AffectedCardID != event.CardID {
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
	if replacement.ContinuousZoneRedirect && !continuousZoneRedirectMatchesEvent(g, replacement, event, controller) {
		return false
	}
	if !conditionSatisfied(g, conditionContext{
		controller: controller,
		source:     source,
		event:      &event,
		obj:        obj,
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

// continuousZoneRedirectMatchesEvent reports whether a continuous
// graveyard-redirect replacement (CR 614) applies to the zone-change event. The
// watched graveyard belongs to the moving card's owner (event.Player), matched
// relative to the replacement's controller; a "would die" form additionally
// restricts by the dying permanent's controller (event.Controller); and the
// moving card must carry one of the replacement's required card types when the
// filter is non-empty.
func continuousZoneRedirectMatchesEvent(g *game.Game, replacement *game.ReplacementEffect, event game.Event, controller game.PlayerID) bool {
	if replacement.RedirectOwnerFilter != game.TriggerControllerAny &&
		!triggerControllerMatches(controller, replacement.RedirectOwnerFilter, event.Player) {
		return false
	}
	if replacement.RedirectControlFilter != game.TriggerControllerAny &&
		!triggerControllerMatches(controller, replacement.RedirectControlFilter, event.Controller) {
		return false
	}
	if len(replacement.RedirectTypeFilter) == 0 {
		return true
	}
	return movingCardHasAnyType(g, event, replacement.RedirectTypeFilter)
}

// movingCardHasAnyType reports whether the card moved by a zone-change event has
// any of the given card types, read from the card's moving face. Tokens (which
// carry no card instance) never match a non-empty type filter.
func movingCardHasAnyType(g *game.Game, event game.Event, cardTypes []types.Card) bool {
	if event.CardID == 0 {
		return false
	}
	card, ok := g.GetCardInstance(event.CardID)
	if !ok || card.Def == nil {
		return false
	}
	def := cardFaceOrDefault(card, event.Face)
	for _, cardType := range cardTypes {
		if slices.Contains(def.Types, cardType) {
			return true
		}
	}
	return false
}

// drawFromEmptyLibraryWins reports whether an active replacement effect causes
// playerID to win the game when they would draw from an empty library
// (Laboratory Maniac, Jace, Wielder of Mysteries).
func drawFromEmptyLibraryWins(g *game.Game, playerID game.PlayerID) bool {
	for i := range g.ReplacementEffects {
		replacement := &g.ReplacementEffects[i]
		if !replacement.DrawFromEmptyLibraryWins {
			continue
		}
		if !replacementSourceIsActive(g, replacement) {
			continue
		}
		if replacementCurrentController(g, replacement) == playerID {
			return true
		}
	}
	return false
}

// drawCardMultiplier reports how many cards a single "would draw a card" event
// by playerID becomes after applying registered draw-doubling replacements
// (CR 614). firstInDrawStep marks the controller's first draw in their own draw
// step, which DrawCardExceptFirstInDrawStep replacements (Teferi's Ageless
// Insight) leave unmultiplied. Multiple draw multipliers compound. The result is
// always at least one.
func drawCardMultiplier(g *game.Game, playerID game.PlayerID, firstInDrawStep bool) int {
	multiplier := 1
	for i := range g.ReplacementEffects {
		replacement := &g.ReplacementEffects[i]
		if replacement.DrawCardMultiplier <= 1 {
			continue
		}
		if firstInDrawStep && replacement.DrawCardExceptFirstInDrawStep {
			continue
		}
		if !replacementSourceIsActive(g, replacement) {
			continue
		}
		if replacementCurrentController(g, replacement) != playerID {
			continue
		}
		multiplier *= replacement.DrawCardMultiplier
	}
	if multiplier < 1 {
		return 1
	}
	return multiplier
}

// replacementLifeGainAmount reports how much life a single "you would gain life"
// event by playerID becomes after applying registered life-gain replacements
// (CR 614), as on Boon Reflection ("twice that much") and Angel of Vitality
// ("that much plus 1"). Multiple replacements compound in registration order;
// each multiplies the running amount and then adds its addend. The result is
// never negative.
func replacementLifeGainAmount(g *game.Game, playerID game.PlayerID, amount int) int {
	for i := range g.ReplacementEffects {
		replacement := &g.ReplacementEffects[i]
		if replacement.LifeGainMultiplier <= 1 && replacement.LifeGainAddend == 0 {
			continue
		}
		if !replacementSourceIsActive(g, replacement) {
			continue
		}
		if replacementCurrentController(g, replacement) != playerID {
			continue
		}
		if replacement.LifeGainMultiplier > 1 {
			amount *= replacement.LifeGainMultiplier
		}
		amount += replacement.LifeGainAddend
	}
	if amount < 0 {
		return 0
	}
	return amount
}

// replacementLifeLossAmount reports how much life a single "would lose life"
// event by playerID becomes after applying registered life-loss replacements
// (CR 614), as on Bloodletter of Aclazotz ("twice that much"). Multiple
// replacements compound in registration order; each multiplies the running
// amount and then adds its addend. Replacements restricted to opponents skip
// loss by their own controller, and turn-restricted replacements apply only on
// their controller's turn. The result is never negative.
func replacementLifeLossAmount(g *game.Game, playerID game.PlayerID, amount int) int {
	for i := range g.ReplacementEffects {
		replacement := &g.ReplacementEffects[i]
		if replacement.LifeLossMultiplier <= 1 && replacement.LifeLossAddend == 0 {
			continue
		}
		if !replacementSourceIsActive(g, replacement) {
			continue
		}
		controller := replacementCurrentController(g, replacement)
		if replacement.LifeLossRecipientOpponent && playerID == controller {
			continue
		}
		if replacement.LifeLossDuringControllerTurn && g.Turn.ActivePlayer != controller {
			continue
		}
		if replacement.LifeLossMultiplier > 1 {
			amount *= replacement.LifeLossMultiplier
		}
		amount += replacement.LifeLossAddend
	}
	if amount < 0 {
		return 0
	}
	return amount
}

func replacementSourceStillApplies(g *game.Game, replacement *game.ReplacementEffect) bool {
	if replacement.Duration != game.DurationPermanent || replacement.SourceObjectID == 0 {
		return true
	}
	_, ok := permanentByObjectID(g, replacement.SourceObjectID)
	return ok
}

func replacementSourceIsActive(g *game.Game, replacement *game.ReplacementEffect) bool {
	if !replacementSourceStillApplies(g, replacement) {
		return false
	}
	if replacement.Duration != game.DurationPermanent || replacement.SourceObjectID == 0 {
		return true
	}
	source, ok := permanentByObjectID(g, replacement.SourceObjectID)
	return ok && activeBattlefieldPermanent(source)
}

// replacementEventPlayer returns the player who chooses among several applicable
// replacement effects for an event (CR 616.1): the affected object's controller,
// or the affected player when there is no controlled object. A permanent leaving
// the battlefield and a spell or ability leaving the stack both carry their
// controller in event.Controller; a card in another zone has only its owner
// (event.Player).
func replacementEventPlayer(event game.Event) game.PlayerID {
	if (event.PermanentID != 0 || event.StackObjectID != 0) && event.Controller >= 0 && event.Controller < game.NumPlayers {
		return event.Controller
	}
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

func createPreventionShield(g *game.Game, obj *game.StackObject, amount int, prim game.PreventDamage, duration game.EffectDuration) bool {
	if !prim.All && amount <= 0 {
		return false
	}
	shield := game.PreventionShield{
		ID:          g.IDGen.Next(),
		Controller:  obj.Controller,
		Amount:      amount,
		All:         prim.All,
		CombatOnly:  prim.CombatOnly,
		Global:      prim.Global,
		Duration:    effectDurationOrDefault(duration, game.DurationUntilEndOfTurn),
		CreatedTurn: g.Turn.TurnNumber,
	}
	if prim.Global {
		g.PreventionShields = append(g.PreventionShields, shield)
		return true
	}
	if _, anyTarget := prim.AnyTarget.AnyTargetObjectReference(); anyTarget {
		// An any-target shield names one target slot the controller chose as
		// either a player or a permanent. Resolve the player half first (it
		// succeeds only when the slot holds a player), then fall back to the
		// permanent half, mirroring any-target damage resolution.
		if playerRef, ok := prim.AnyTarget.AnyTargetPlayerReference(); ok {
			if playerID, ok := resolvePlayerReference(g, obj, playerRef); ok {
				shield.Player = playerID
				g.PreventionShields = append(g.PreventionShields, shield)
				return true
			}
		}
		objectRef, _ := prim.AnyTarget.AnyTargetObjectReference()
		resolved, ok := resolveObjectReference(g, obj, objectRef)
		if !ok || resolved.permanent == nil {
			return false
		}
		shield.PermanentID = resolved.permanent.ObjectID
		g.PreventionShields = append(g.PreventionShields, shield)
		return true
	}
	if prim.Player.Kind() != game.PlayerReferenceNone {
		playerID, ok := resolvePlayerReference(g, obj, prim.Player)
		if !ok {
			return false
		}
		shield.Player = playerID
		g.PreventionShields = append(g.PreventionShields, shield)
		return true
	}
	resolved, ok := resolveObjectReference(g, obj, prim.Object)
	if !ok || resolved.permanent == nil {
		return false
	}
	if prim.BySource {
		shield.SourcePermanentID = resolved.permanent.ObjectID
	} else {
		shield.PermanentID = resolved.permanent.ObjectID
	}
	g.PreventionShields = append(g.PreventionShields, shield)
	return true
}

func preventionShieldApplies(shield game.PreventionShield, event damageEvent) bool {
	if shield.CombatOnly && !event.combatDamage {
		return false
	}
	if shield.Global {
		return true
	}
	if shield.SourcePermanentID != 0 {
		return shield.SourcePermanentID == event.sourceObjectID
	}
	if event.permanent != nil {
		return shield.PermanentID == event.permanent.ObjectID
	}
	return shield.PermanentID == 0 && shield.Player == event.player
}

func compactPreventionShields(shields []game.PreventionShield) []game.PreventionShield {
	kept := shields[:0]
	for _, shield := range shields {
		if shield.All || shield.Amount > 0 {
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
	if permanent == nil {
		return false
	}
	// Use effective characteristics for permanents on the battlefield (CR 702.16c).
	if sourceObjectID != 0 {
		if sourcePermanent, ok := permanentByObjectID(g, sourceObjectID); ok {
			vals := effectivePermanentValues(g, sourcePermanent)
			return permanentProtectedFromChars(g, permanent, sourceChars{
				colors:   vals.colors,
				types:    vals.types,
				subtypes: vals.subtypes,
			})
		}
		// If the object ID resolves to a stack object, use the selected face's
		// characteristics so that alternate-face spells (adventures, MDFCs)
		// use the face that was actually cast.
		if stackObj, ok := stackObjectByID(g, sourceObjectID); ok {
			if chars, ok := stackObjectSourceChars(g, stackObj); ok {
				return permanentProtectedFromChars(g, permanent, chars)
			}
		}
		// LKI fallback: covers departed permanents and resolved spells whose
		// stack object was already removed (CR 702.16c, 800.4a).
		if snapshot, ok := lastKnownObject(g, sourceObjectID); ok {
			return permanentProtectedFromChars(g, permanent, sourceChars{
				colors:   snapshot.Colors,
				types:    snapshot.Types,
				subtypes: snapshot.Subtypes,
			})
		}
	}
	// Fall back to card def for instants/sorceries identified by card instance.
	// Use the card's selected face if one is on the stack; otherwise root def.
	if sourceID != 0 {
		if card, ok := g.GetCardInstance(sourceID); ok {
			return permanentProtectedFromSourceDef(g, permanent, cardFaceOrDefault(card, selectedFaceForCardInstance(g, card)))
		}
	}
	return false
}

// stackObjectSourceChars returns the effective source characteristics of a
// spell stack object, using the selected face (obj.Face) so that
// alternate-face spells carry the correct types/colors/subtypes.
func stackObjectSourceChars(g *game.Game, obj *game.StackObject) (sourceChars, bool) {
	if obj.SourceTokenDef != nil {
		return sourceChars{
			colors:   obj.SourceTokenDef.Colors,
			types:    obj.SourceTokenDef.Types,
			subtypes: obj.SourceTokenDef.Subtypes,
		}, true
	}
	if obj.SourceID != 0 {
		if card, ok := g.GetCardInstance(obj.SourceID); ok {
			faceDef := cardFaceOrDefault(card, obj.Face)
			return sourceChars{
				colors:   faceDef.Colors,
				types:    faceDef.Types,
				subtypes: faceDef.Subtypes,
			}, true
		}
	}
	return sourceChars{}, false
}

// selectedFaceForCardInstance finds the face the card instance is currently
// using on the stack (if any); falls back to FaceFront.
func selectedFaceForCardInstance(g *game.Game, card *game.CardInstance) game.FaceIndex {
	for _, obj := range g.Stack.Objects() {
		if obj.SourceID == card.ID {
			return obj.Face
		}
	}
	return game.FaceFront
}

// sourceChars holds the effective qualities used for protection evaluation.
type sourceChars struct {
	colors   []color.Color
	types    []types.Card
	subtypes []types.Sub
}

func sourceCharsForProtection(g *game.Game, sourceID, sourceObjectID id.ID, fallback *game.CardDef) (sourceChars, bool) {
	if sourceObjectID != 0 {
		if sourcePermanent, ok := permanentByObjectID(g, sourceObjectID); ok {
			vals := effectivePermanentValues(g, sourcePermanent)
			return sourceChars{colors: vals.colors, types: vals.types, subtypes: vals.subtypes}, true
		}
		if stackObj, ok := stackObjectByID(g, sourceObjectID); ok {
			if chars, ok := stackObjectSourceChars(g, stackObj); ok {
				return chars, true
			}
		}
		if snapshot, ok := lastKnownObject(g, sourceObjectID); ok {
			return sourceChars{
				colors:   snapshot.Colors,
				types:    snapshot.Types,
				subtypes: snapshot.Subtypes,
			}, true
		}
	}
	if sourceID != 0 {
		if card, ok := g.GetCardInstance(sourceID); ok {
			def := cardFaceOrDefault(card, selectedFaceForCardInstance(g, card))
			return sourceChars{colors: def.Colors, types: def.Types, subtypes: def.Subtypes}, true
		}
	}
	if fallback != nil {
		return sourceChars{
			colors:   fallback.Colors,
			types:    fallback.Types,
			subtypes: fallback.Subtypes,
		}, true
	}
	return sourceChars{}, false
}

func playerProtectedFromSource(g *game.Game, player game.PlayerID, sourceID, sourceObjectID id.ID, fallback *game.CardDef) bool {
	effects := activeRuleEffects(g)
	hasScopedProtection := false
	for i := range effects {
		effect := &effects[i]
		if effect.Kind == game.RuleEffectPlayerProtection &&
			playerRelationMatches(effect.Controller, player, effect.AffectedPlayer) {
			if effect.Protection.Everything {
				return true
			}
			hasScopedProtection = true
		}
	}
	if !hasScopedProtection {
		return false
	}
	chars, ok := sourceCharsForProtection(g, sourceID, sourceObjectID, fallback)
	if !ok {
		return false
	}
	for i := range effects {
		effect := &effects[i]
		if effect.Kind == game.RuleEffectPlayerProtection &&
			playerRelationMatches(effect.Controller, player, effect.AffectedPlayer) &&
			protectionMatchesSource(effect.Protection, chars) {
			return true
		}
	}
	return false
}

func permanentProtectedFromSourceDef(g *game.Game, permanent *game.Permanent, source *game.CardDef) bool {
	if permanent == nil || source == nil {
		return false
	}
	return permanentProtectedFromChars(g, permanent, sourceChars{
		colors:   source.Colors,
		types:    source.Types,
		subtypes: source.Subtypes,
	})
}

func permanentProtectedFromChars(g *game.Game, permanent *game.Permanent, source sourceChars) bool {
	values := effectivePermanentValues(g, permanent)
	// Check the effective keyword map first: if Protection was removed via
	// RemoveKeywords (e.g., "loses all abilities"), it will be false here even
	// though the ability body may still appear in values.abilities.
	if !values.keywords[game.Protection] {
		return false
	}
	for i := range values.abilities {
		body, ok := values.abilities[i].(*game.StaticAbility)
		if !ok {
			continue
		}
		prot, ok := game.StaticBodyProtectionKeyword(body)
		if !ok {
			continue
		}
		if protectionMatchesSource(prot, source) {
			return true
		}
	}
	return false
}

func protectionMatchesSource(prot game.ProtectionKeyword, source sourceChars) bool {
	if prot.Everything {
		return true
	}
	if prot.EachColor && len(source.colors) > 0 {
		return true
	}
	if prot.Multicolored && len(source.colors) >= 2 {
		return true
	}
	if prot.Monocolored && len(source.colors) == 1 {
		return true
	}
	for _, clr := range prot.FromColors {
		if slices.Contains(source.colors, clr) {
			return true
		}
	}
	for _, t := range prot.FromTypes {
		if slices.Contains(source.types, t) {
			return true
		}
	}
	for _, sub := range prot.FromSubtypes {
		if slices.Contains(source.subtypes, sub) {
			return true
		}
	}
	return false
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

// permanentProtectedFromPermanentEffective reports whether permanent has
// protection from sourcePermanent using sourcePermanent's effective
// characteristics. Used for blocking checks (CR 702.16b).
func permanentProtectedFromPermanentEffective(g *game.Game, permanent, sourcePermanent *game.Permanent) bool {
	if permanent == nil || sourcePermanent == nil {
		return false
	}
	vals := effectivePermanentValues(g, sourcePermanent)
	return permanentProtectedFromChars(g, permanent, sourceChars{
		colors:   vals.colors,
		types:    vals.types,
		subtypes: vals.subtypes,
	})
}
