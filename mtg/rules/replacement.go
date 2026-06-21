package rules

import (
	"fmt"
	"slices"

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
	engine *Engine
	agents [game.NumPlayers]PlayerAgent
	log    *TurnLog
	xValue int
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
	permanentProtected := event.permanent != nil &&
		permanentProtectedFromSource(g, event.permanent, event.sourceID, event.sourceObjectID)
	if permanentProtected {
		amount = 0
	}
	if event.permanent == nil &&
		playerProtectedFromSource(g, event.player, event.sourceID, event.sourceObjectID, nil) {
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

func replacementDamageAmount(g *game.Game, event damageEvent) int {
	if event.amount <= 0 {
		return event.amount
	}
	applied := make(map[id.ID]bool)
	for {
		matches := matchingDamageReplacementEffects(g, event, applied)
		if len(matches) == 0 {
			return event.amount
		}
		replacement := matches[0]
		if len(matches) > 1 {
			decision := recordReplacementDecision(g, replacementDecisionPlayer(g, event), replacementEffectLabels(matches))
			replacement = selectedReplacementEffect(matches, decision)
		}
		applied[replacement.ID] = true
		if replacement.DamageMultiplier > 1 {
			event.amount *= replacement.DamageMultiplier
		}
		event.amount += replacement.DamageAddend
		if event.amount <= 0 {
			return 0
		}
	}
}

func orderedPreventionShieldIndices(g *game.Game, event damageEvent) []int {
	var indices []int
	var options []string
	for i, shield := range g.PreventionShields {
		if !preventionShieldApplies(shield, event) || (!shield.All && shield.Amount <= 0) {
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

func replaceDestroyPermanent(g *game.Game, permanent *game.Permanent, preventRegeneration bool) bool {
	if permanent == nil {
		return false
	}
	hasShieldCounter := permanent.Counters.Get(counter.Shield) > 0
	hasRegeneration := permanent.RegenerationShields > 0 && !preventRegeneration
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

func replacementDecisionPlayer(g *game.Game, event damageEvent) game.PlayerID {
	if event.permanent != nil {
		return effectiveController(g, event.permanent)
	}
	return event.player
}

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
		if replacement.CounterMultiplier > 1 {
			amount *= replacement.CounterMultiplier
		}
		amount += replacement.CounterAddend
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
		if replacement.EntersAsCopy {
			applyEntersAsCopy(ctx, g, permanent, replacement)
		}
	}
	if hasKeyword(g, permanent, game.Riot) {
		applyEntryRiotChoice(ctx, g, permanent)
	}
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
	effectID := g.IDGen.Next()
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:               effectID,
		SourceObjectID:   permanent.ObjectID,
		SourceCardID:     permanent.CardInstanceID,
		Controller:       controller,
		Timestamp:        permanent.Timestamp(),
		Duration:         game.DurationForAsLongAsSourceOnBattlefield,
		CreatedTurn:      g.Turn.TurnNumber,
		AffectedObjectID: permanent.ObjectID,
		Layer:            game.LayerCopy,
		CopyValues:       opt.Val(values),
	})
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
		if replacementEffectMatchesEventWithSource(g, &replacement, event, permanent) {
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
			if !entersTappedGroupTypeMatches(g, replacement, source) {
				continue
			}
		}
		if !replacementEffectMatchesEventWithSource(g, replacement, event, source) {
			continue
		}
		matches = append(matches, *replacement)
	}
	return matches
}

// entersTappedGroupTypeMatches reports whether the entering permanent satisfies
// the permanent-type filter of a group enters-tapped replacement. An empty
// filter taps every entering permanent.
func entersTappedGroupTypeMatches(g *game.Game, replacement *game.ReplacementEffect, permanent *game.Permanent) bool {
	if len(replacement.EntersTappedTypes) == 0 {
		return true
	}
	if permanent == nil {
		return false
	}
	for _, cardType := range replacement.EntersTappedTypes {
		if permanentHasType(g, permanent, cardType) {
			return true
		}
	}
	return false
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
		if replacement.CounterMultiplier <= 1 && replacement.CounterAddend == 0 {
			continue
		}
		if replacement.MatchCounterKind && replacement.CounterKindFilter != event.CounterKind {
			continue
		}
		if len(replacement.CounterRecipientTypes) > 0 && !counterRecipientPermanentMatches(g, event.PermanentID, recipient, replacement.CounterRecipientTypes) {
			continue
		}
		if replacement.CounterRecipientAnyPermanent && !counterRecipientPermanentMatches(g, event.PermanentID, recipient, nil) {
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
	return replacementEffectMatchesEventWithSource(g, replacement, event, nil)
}

// replacementEffectMatchesEventWithSource reports whether the replacement
// applies to the event. A non-nil source supplies the source permanent to the
// condition context so source-relative condition predicates (such as the
// EventHistory "this turn" conditions on enters-with-counters replacements)
// resolve "you" against the replacement's own permanent. The source-less
// replacementEffectMatchesEvent wrapper preserves the prior behavior for every
// other replacement category.
func replacementEffectMatchesEventWithSource(g *game.Game, replacement *game.ReplacementEffect, event game.Event, source *game.Permanent) bool {
	if !replacementSourceIsActive(g, replacement) {
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
		source:     source,
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
		Duration:    effectDurationOrDefault(duration, game.DurationUntilEndOfTurn),
		CreatedTurn: g.Turn.TurnNumber,
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

func applyPreventionShields(g *game.Game, event damageEvent, amount int) int {
	order := orderedPreventionShieldIndices(g, event)
	for _, i := range order {
		shield := &g.PreventionShields[i]
		if !preventionShieldApplies(*shield, event) || amount <= 0 {
			continue
		}
		if shield.All {
			amount = 0
			continue
		}
		if shield.Amount <= 0 {
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
	if shield.CombatOnly && !event.combatDamage {
		return false
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
