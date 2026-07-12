package rules

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

type conditionContext struct {
	controller             game.PlayerID
	source                 *game.Permanent
	event                  *game.Event
	obj                    *game.StackObject
	useBaseCharacteristics bool
	characteristicsBefore  game.ContinuousLayer
}

// conditionParametersNegative reports whether any numeric condition parameter is
// negative, which is structurally invalid and must fail closed.
func conditionParametersNegative(cond *game.Condition) bool {
	for _, agg := range cond.Aggregates {
		if agg.Value < 0 {
			return true
		}
	}
	return cond.AnyPlayerLifeAtMost < 0 ||
		cond.AnyOpponentPoisonAtLeast < 0 ||
		cond.ControllerGraveyardCardOfTypeCountAtLeast < 0 ||
		cond.ControllerGraveyardInstantOrSorceryCountAtLeast < 0 ||
		cond.ControlsMatching.Exists && cond.ControlsMatching.Val.MinCount < 0 ||
		cond.AnyOpponentControls.Exists && cond.AnyOpponentControls.Val.MinCount < 0 ||
		cond.OpponentsControl.Exists && cond.OpponentsControl.Val.MinCount < 0 ||
		cond.SourceLevelCountersAtLeast < 0 ||
		cond.SourceLevelCountersLessThan < 0 ||
		cond.SourceCountersAtLeast < 0
}

// aggregateValue evaluates a player- or board-derived quantity in the given
// condition context. It returns the value and whether it could be resolved; an
// unresolved quantity fails the comparison closed.
func aggregateValue(g *game.Game, ctx conditionContext, kind game.AggregateKind) (int, bool) {
	switch kind {
	case game.AggregateControllerLife:
		player, ok := playerByID(g, ctx.controller)
		if !ok {
			return 0, false
		}
		return player.Life, true
	case game.AggregateControllerLifeAboveStarting:
		player, ok := playerByID(g, ctx.controller)
		if !ok {
			return 0, false
		}
		return player.Life - player.StartingLife, true
	case game.AggregateControllerHandSize:
		player, ok := playerByID(g, ctx.controller)
		if !ok {
			return 0, false
		}
		return cardInstanceCount(g, player.Hand.All()), true
	case game.AggregateControllerLibrarySize:
		player, ok := playerByID(g, ctx.controller)
		if !ok {
			return 0, false
		}
		return cardInstanceCount(g, player.Library.All()), true
	case game.AggregateControllerGraveyardCardCount:
		player, ok := playerByID(g, ctx.controller)
		if !ok {
			return 0, false
		}
		return cardInstanceCount(g, player.Graveyard.All()), true
	case game.AggregateControllerGraveyardCardTypeCount:
		return controllerGraveyardCardTypeCount(g, ctx.controller), true
	case game.AggregateControllerGraveyardPermanentCardCount:
		return controllerGraveyardPermanentCardCount(g, ctx.controller), true
	case game.AggregateControllerGraveyardManaValueCount:
		return controllerGraveyardManaValueCount(g, ctx.controller), true
	case game.AggregateAnyOpponentGraveyardCardCount:
		return anyOpponentGraveyardCardCount(g, ctx.controller), true
	case game.AggregateControllerBasicLandTypeCount:
		return controllerBasicLandTypeCount(g, ctx), true
	case game.AggregateControllerCreaturePowerDiversity:
		return controllerCreaturePowerDiversity(g, ctx), true
	case game.AggregateOpponentCount:
		return len(aliveOpponents(g, ctx.controller)), true
	case game.AggregateAttackersAttackingController:
		return attackersAttackingPlayerCount(g, ctx.controller), true
	case game.AggregateControllerGainedLifeThisTurn:
		return lifeChangedThisTurn(g, ctx.controller, game.EventLifeGained), true
	case game.AggregateSpellX:
		if ctx.obj == nil {
			return 0, false
		}
		return ctx.obj.XValue, true
	case game.AggregateEventSpellManaSpentToCast:
		if ctx.event == nil || !ctx.event.ManaSpentToCast.Exists {
			return 0, false
		}
		return ctx.event.ManaSpentToCast.Val, true
	case game.AggregateEventPlayerHandSize:
		if ctx.event == nil {
			return 0, false
		}
		player, ok := playerByID(g, ctx.event.Player)
		if !ok {
			return 0, false
		}
		return cardInstanceCount(g, player.Hand.All()), true
	default:
		// AggregateNone carries no resolvable quantity; fail the comparison closed.
	}
	return 0, false
}

func conditionSatisfied(g *game.Game, ctx conditionContext, condition opt.V[game.Condition]) bool {
	if !condition.Exists || condition.Val.Empty() {
		return true
	}
	cond := condition.Val
	if conditionParametersNegative(&cond) {
		return false
	}
	matches := true
	if cond.ControlsMatching.Exists {
		matches = matches && controllerControlsMatchingSelection(g, ctx, cond.ControlsMatching.Val)
	}
	for _, agg := range cond.Aggregates {
		value, ok := aggregateValue(g, ctx, agg.Aggregate)
		matches = matches && ok && compare.Int{Op: agg.Op, Value: agg.Value}.Matches(value)
	}
	if cond.AnyOpponentPoisonAtLeast > 0 {
		matches = matches && anyOpponentPoisonAtLeast(g, ctx.controller, cond.AnyOpponentPoisonAtLeast)
	}
	if cond.AnyPlayerLifeAtMost > 0 {
		matches = matches && anyPlayerLifeAtMost(g, cond.AnyPlayerLifeAtMost)
	}
	if cond.ControllerHandEmpty {
		player, ok := playerByID(g, ctx.controller)
		matches = matches && ok && cardInstanceCount(g, player.Hand.All()) == 0
	}
	if cond.ControllerGraveyardCardOfTypeCountAtLeast > 0 {
		matches = matches && controllerGraveyardCardOfTypeCount(g, ctx.controller, cond.ControllerGraveyardCountCardType) >= cond.ControllerGraveyardCardOfTypeCountAtLeast
	}
	if cond.ControllerGraveyardInstantOrSorceryCountAtLeast > 0 {
		matches = matches && controllerGraveyardInstantOrSorceryCount(g, ctx.controller) >= cond.ControllerGraveyardInstantOrSorceryCountAtLeast
	}
	if cond.AnyOpponentControls.Exists {
		matches = matches && anyOpponentControlsMatchingSelection(g, ctx, cond.AnyOpponentControls.Val)
	}
	if cond.OpponentsControl.Exists {
		matches = matches && playersControlMatchingSelection(g, ctx, aliveOpponents(g, ctx.controller), cond.OpponentsControl.Val)
	}
	if cond.ControlComparison.Exists {
		matches = matches && controlCountComparisonSatisfied(g, ctx, cond.ControlComparison.Val)
	}
	if cond.Object.Exists || cond.ObjectMatches.Exists || len(cond.Types) > 0 {
		matches = matches && conditionObjectMatches(g, ctx, &cond)
	}
	if cond.EventPermanentNameUniqueAmongControlledAndGraveyardCreatures {
		matches = matches && eventPermanentNameUniqueAmongControlledAndGraveyardCreatures(g, ctx)
	}
	if cond.SourceClassLevelAtLeast > 0 {
		matches = matches && ctx.source != nil && ctx.source.ClassLevel >= cond.SourceClassLevelAtLeast
	}
	if cond.SourceClassLevelLessThan > 0 {
		matches = matches && ctx.source != nil && ctx.source.ClassLevel < cond.SourceClassLevelLessThan
	}
	if cond.SourceLevelCountersAtLeast > 0 {
		matches = matches && ctx.source != nil && ctx.source.Counters.Get(counter.Level) >= cond.SourceLevelCountersAtLeast
	}
	if cond.SourceLevelCountersLessThan > 0 {
		matches = matches && ctx.source != nil && ctx.source.Counters.Get(counter.Level) < cond.SourceLevelCountersLessThan
	}
	if cond.SourceCountersAtLeast > 0 {
		matches = matches && cond.SourceCounterKindKnown && ctx.source != nil &&
			ctx.source.Counters.Get(cond.SourceCounterKind) >= cond.SourceCountersAtLeast
	}
	if cond.SourceAttachedCombatCounterpartSubtypes != [2]types.Sub{} {
		matches = matches && sourceAttachedCombatCounterpartMatches(g, ctx.source, cond.SourceAttachedCombatCounterpartSubtypes)
	}
	if cond.SourceNotMonstrous {
		matches = matches && ctx.source != nil && !ctx.source.Monstrous
	}

	if cond.SourceSaddled {
		matches = matches && ctx.source != nil && ctx.source.Saddled
	}
	if cond.SourceTributeNotPaid {
		matches = matches && ctx.source != nil && !ctx.source.TributePaid
	}
	if cond.ControllerHasMaxSpeed {
		player, ok := playerByID(g, ctx.controller)
		matches = matches && ok && player.Speed >= 4
	}
	if cond.TargetEnteredThisTurn.Exists {
		matches = matches && conditionTargetEnteredThisTurn(g, ctx, cond.TargetEnteredThisTurn.Val)
	}
	if cond.CastFromZone.Exists {
		matches = matches && ctx.obj != nil && !ctx.obj.Copy && ctx.obj.SourceZone == cond.CastFromZone.Val
	}
	if cond.CastDuringControllerMainPhase {
		matches = matches && ctx.obj != nil && !ctx.obj.Copy && ctx.obj.CastDuringControllerMainPhase
	}
	if cond.SpellWasKicked {
		matches = matches && ctx.obj != nil && !ctx.obj.Copy && ctx.obj.KickerPaid
	}
	if cond.GiftPromised {
		matches = matches && ctx.obj != nil && ctx.obj.GiftPromised
	}
	if cond.EventPermanentWasKicked {
		matches = matches && ctx.event != nil && ctx.event.KickerPaid
	}
	if cond.EventPermanentWasCastFromControllerHand {
		matches = matches &&
			ctx.event != nil &&
			ctx.event.EnterWasCast &&
			ctx.event.EnterHasCastController &&
			ctx.event.EnterCastController == ctx.controller &&
			ctx.event.EnterCastFromZone == zone.Hand
	}
	if cond.SpellColorManaSpent.Count > 0 {
		matches = matches && ctx.obj != nil && !ctx.obj.Copy &&
			ctx.obj.ManaSpentByColorToCast[cond.SpellColorManaSpent.Color] >= cond.SpellColorManaSpent.Count
	}
	if cond.SpellSameColorManaSpentAtLeast > 0 {
		matches = matches && ctx.obj != nil && !ctx.obj.Copy &&
			greatestSameColorManaSpent(ctx.obj.ManaSpentByColorToCast) >= cond.SpellSameColorManaSpentAtLeast
	}
	if cond.LandEnteredThisTurnOrControlsBasicLand {
		matches = matches && sourceLandEnteredThisTurnOrControlsBasicLand(g, ctx)
	}
	if cond.ControllerCreatedTokenThisTurn {
		matches = matches && controllerCreatedTokenThisTurn(g, ctx.controller)
	}
	if cond.EventHistory.Exists {
		matches = matches && conditionEventHistorySatisfied(g, ctx, &cond.EventHistory.Val)
	}
	if cond.ControllerControlsCommander {
		matches = matches && playerControlsCommander(g, ctx.controller)
	}
	if len(cond.ControllerControlsNamed) > 0 {
		matches = matches && controllerControlsNamed(g, ctx, cond.ControllerControlsNamed)
	}
	if cond.FirstCombatPhaseOfTurn {
		matches = matches && g.Turn.CombatPhasesThisTurn <= 1
	}
	if cond.ControllerControlsGreatestPowerCreature {
		matches = matches && controllerControlsGreatestPowerCreature(g, ctx)
	}
	if cond.ControllerControlsGreatestToughnessCreature {
		matches = matches && controllerControlsGreatestToughnessCreature(g, ctx)
	}
	if cond.ControllerIsMonarch {
		player, ok := playerByID(g, ctx.controller)
		matches = matches && ok && player.IsMonarch
	}
	if cond.ControllerWasMonarchAtTurnStart {
		matches = matches && g.Turn.MonarchAtTurnStart.Exists &&
			g.Turn.MonarchAtTurnStart.Val == ctx.controller
	}
	if cond.AnOpponentIsMonarch {
		matches = matches && anyAliveOpponentIsMonarch(g, ctx.controller)
	}
	if cond.NoMonarch {
		matches = matches && !anyPlayerIsMonarch(g)
	}
	if cond.EventDefendingPlayerIsMonarch {
		monarch := livingMonarch(g)
		matches = matches && ctx.event != nil && monarch.Exists && monarch.Val == ctx.event.Player
	}
	if cond.ControllerHasInitiative {
		player, ok := playerByID(g, ctx.controller)
		matches = matches && ok && player.HasInitiative
	}
	if cond.ControllerHasCityBlessing {
		player, ok := playerByID(g, ctx.controller)
		matches = matches && ok && player.HasCityBlessing
	}
	if cond.SourceControllerTurn {
		matches = matches && g.Turn.ActivePlayer == ctx.controller
	}
	if cond.SourceAbilityResolutionOrdinalThisTurn > 0 {
		matches = matches && sourceAbilityResolutionOrdinalMatches(g, ctx, cond.SourceAbilityResolutionOrdinalThisTurn)
	}
	if cond.Negate {
		return !matches
	}
	return matches
}

func sourceAttachedCombatCounterpartMatches(g *game.Game, source *game.Permanent, subtypes [2]types.Sub) bool {
	if source == nil || !source.AttachedTo.Exists || g.Combat == nil {
		return false
	}
	attachedID := source.AttachedTo.Val
	for _, block := range g.Combat.Blockers {
		var counterpartID game.ObjectID
		switch {
		case block.Blocker == attachedID:
			counterpartID = block.Blocking
		case block.Blocking == attachedID:
			counterpartID = block.Blocker
		default:
			continue
		}
		counterpart, ok := permanentByObjectID(g, counterpartID)
		if !ok {
			continue
		}
		for _, subtype := range subtypes {
			if subtype != "" && permanentHasSubtype(g, counterpart, subtype) {
				return true
			}
		}
	}
	return false
}

// anyAliveOpponentIsMonarch reports whether any of the controller's alive
// opponents currently holds the monarch designation (CR 720). Exactly one
// player is the monarch at a time, so this is true when the monarch exists and
// is one of the controller's opponents.
func anyAliveOpponentIsMonarch(g *game.Game, controller game.PlayerID) bool {
	for _, opponentID := range aliveOpponents(g, controller) {
		if opponent, ok := playerByID(g, opponentID); ok && opponent.IsMonarch {
			return true
		}
	}
	return false
}

// anyPlayerIsMonarch reports whether any living player currently holds the
// monarch designation, backing the "there is no monarch" condition (Crown of
// Gondor, Archivist of Gondor). It filters to living players so a monarch who
// has left the game (whose IsMonarch flag is not cleared on elimination) no
// longer counts, matching anyAliveOpponentIsMonarch.
func anyPlayerIsMonarch(g *game.Game) bool {
	for i := range g.Players {
		if g.Players[i].IsMonarch && g.Players[i].IsAlive() {
			return true
		}
	}
	return false
}

func cardInstanceCount(g *game.Game, objectIDs []id.ID) int {
	count := 0
	for _, objectID := range objectIDs {
		if _, ok := g.GetCardInstance(objectID); ok {
			count++
		}
	}
	return count
}

// attackersAttackingPlayerCount counts the attackers declared this combat that
// are attacking the given player directly or one of that player's planeswalkers
// ("attacking you and/or planeswalkers you control"). Battle attacks are
// excluded. The full attacker declaration is authoritative in g.Combat.
func attackersAttackingPlayerCount(g *game.Game, player game.PlayerID) int {
	if g.Combat == nil {
		return 0
	}
	count := 0
	for _, declaration := range g.Combat.Attackers {
		if declaration.Target.Player == player && declaration.Target.BattleID == 0 {
			count++
		}
	}
	return count
}

func controllerGraveyardCardTypeCount(g *game.Game, controller game.PlayerID) int {
	player, ok := playerByID(g, controller)
	if !ok {
		return 0
	}
	distinct := make(map[types.Card]bool)
	for _, cardID := range player.Graveyard.All() {
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		for _, cardType := range cardFaceOrDefault(card, game.FaceFront).Types {
			distinct[cardType] = true
		}
		if card.Def.Layout == game.LayoutSplit && card.Def.Alternate.Exists {
			for _, cardType := range card.Def.Alternate.Val.Types {
				distinct[cardType] = true
			}
		}
	}
	return len(distinct)
}

// controllerGraveyardPermanentCardCount counts the cards in the controller's
// graveyard that are permanent cards ("there are four or more permanent cards in
// your graveyard", the Descend ability word).
func controllerGraveyardPermanentCardCount(g *game.Game, controller game.PlayerID) int {
	player, ok := playerByID(g, controller)
	if !ok {
		return 0
	}
	count := 0
	for _, cardID := range player.Graveyard.All() {
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		if cardFaceOrDefault(card, game.FaceFront).IsPermanent() {
			count++
			continue
		}
		if card.Def.Layout == game.LayoutSplit && card.Def.Alternate.Exists &&
			card.Def.Alternate.Val.IsPermanent() {
			count++
		}
	}
	return count
}

// controllerGraveyardManaValueCount counts the distinct mana values among cards
// in the controller's graveyard ("there are five or more mana values among cards
// in your graveyard", Syndicate Infiltrator).
func controllerGraveyardManaValueCount(g *game.Game, controller game.PlayerID) int {
	player, ok := playerByID(g, controller)
	if !ok {
		return 0
	}
	distinct := make(map[int]bool)
	for _, cardID := range player.Graveyard.All() {
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		manaValue := cardFaceOrDefault(card, game.FaceFront).ManaValue()
		if card.Def.Layout == game.LayoutSplit && card.Def.Alternate.Exists {
			manaValue += card.Def.Alternate.Val.ManaValue()
		}
		distinct[manaValue] = true
	}
	return len(distinct)
}

// anyOpponentGraveyardCardCount returns the largest graveyard size among the
// controller's living opponents ("an opponent has eight or more cards in their
// graveyard", Nimana Skitter-Sneak). The existential "an opponent has N or more"
// gate is satisfied exactly when this maximum is at least N.
func anyOpponentGraveyardCardCount(g *game.Game, controller game.PlayerID) int {
	largest := 0
	for _, opponentID := range aliveOpponents(g, controller) {
		opponent, ok := playerByID(g, opponentID)
		if !ok {
			continue
		}
		if size := cardInstanceCount(g, opponent.Graveyard.All()); size > largest {
			largest = size
		}
	}
	return largest
}

// graveyard whose card types include cardType ("twenty or more creature cards
// are in your graveyard", Mortal Combat).
func controllerGraveyardCardOfTypeCount(g *game.Game, controller game.PlayerID, cardType types.Card) int {
	player, ok := playerByID(g, controller)
	if !ok {
		return 0
	}
	count := 0
	for _, cardID := range player.Graveyard.All() {
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		if slices.Contains(cardFaceOrDefault(card, game.FaceFront).Types, cardType) {
			count++
			continue
		}
		if card.Def.Layout == game.LayoutSplit && card.Def.Alternate.Exists &&
			slices.Contains(card.Def.Alternate.Val.Types, cardType) {
			count++
		}
	}
	return count
}

// controllerGraveyardInstantOrSorceryCount counts the cards in the controller's
// graveyard that are instants and/or sorceries, backing the "instant and/or
// sorcery cards in your graveyard" count condition (Spell mastery). A split card
// counts once when either half is an instant or sorcery.
func controllerGraveyardInstantOrSorceryCount(g *game.Game, controller game.PlayerID) int {
	player, ok := playerByID(g, controller)
	if !ok {
		return 0
	}
	isInstantOrSorcery := func(cardTypes []types.Card) bool {
		return slices.Contains(cardTypes, types.Instant) || slices.Contains(cardTypes, types.Sorcery)
	}
	count := 0
	for _, cardID := range player.Graveyard.All() {
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		if isInstantOrSorcery(cardFaceOrDefault(card, game.FaceFront).Types) {
			count++
			continue
		}
		if card.Def.Layout == game.LayoutSplit && card.Def.Alternate.Exists &&
			isInstantOrSorcery(card.Def.Alternate.Val.Types) {
			count++
		}
	}
	return count
}

// without a source permanent.
func sourceLandEnteredThisTurnOrControlsBasicLand(g *game.Game, ctx conditionContext) bool {
	if ctx.source != nil && permanentEnteredThisTurn(g, ctx.source.ObjectID) {
		return true
	}
	selection := game.Selection{
		RequiredTypes: []types.Card{types.Land},
		Supertypes:    []types.Super{types.Basic},
	}
	return countPlayerMatchingSelection(g, ctx, ctx.controller, selection) >= 1
}

// sourceAbilityResolutionOrdinalMatches reports whether the resolving triggered
// ability has resolved exactly ordinal times this turn, counting the current
// resolution. It reads the resolving stack object's (source, ability) tally from
// Game.ResolvedTriggeredAbilitiesThisTurn, which the ability increments as it
// begins resolving ("if this is the second time this ability has resolved this
// turn"; Prowl, Pursuit Vehicle). It fails closed when no resolving triggered
// ability is in context.
func sourceAbilityResolutionOrdinalMatches(g *game.Game, ctx conditionContext, ordinal int) bool {
	if ctx.obj == nil {
		return false
	}
	key := game.TriggeredAbilityUse{SourceID: ctx.obj.SourceID, AbilityIndex: ctx.obj.AbilityIndex}
	return g.ResolvedTriggeredAbilitiesThisTurn[key] == ordinal
}

func controllerBasicLandTypeCount(g *game.Game, ctx conditionContext) int {
	distinct := make(map[types.Sub]bool)
	for _, permanent := range g.Battlefield {
		if permanent.PhasedOut {
			continue
		}
		values := permanentValuesForCondition(g, permanent, ctx)
		if values.controller != ctx.controller || !slices.Contains(values.types, types.Land) {
			continue
		}
		for _, subtype := range basicLandSubtypes {
			if slices.Contains(values.subtypes, subtype) {
				distinct[subtype] = true
			}
		}
	}
	return len(distinct)
}

func controllerCreaturePowerDiversity(g *game.Game, ctx conditionContext) int {
	distinct := make(map[int]bool)
	for _, permanent := range g.Battlefield {
		if permanent.PhasedOut {
			continue
		}
		values := permanentValuesForCondition(g, permanent, ctx)
		if values.controller == ctx.controller &&
			slices.Contains(values.types, types.Creature) &&
			values.powerOK {
			distinct[values.power] = true
		}
	}
	return len(distinct)
}

// controllerControlsGreatestPowerCreature reports whether the context controller
// controls a creature whose power equals the greatest power among all creatures
// on the battlefield ("you control the creature with the greatest power or tied
// for the greatest power"). It is false when no creatures with a defined power
// exist.
func controllerControlsGreatestPowerCreature(g *game.Game, ctx conditionContext) bool {
	greatest := 0
	haveGreatest := false
	controllerGreatest := 0
	haveControllerGreatest := false
	for _, permanent := range g.Battlefield {
		if permanent.PhasedOut {
			continue
		}
		values := permanentValuesForCondition(g, permanent, ctx)
		if !slices.Contains(values.types, types.Creature) || !values.powerOK {
			continue
		}
		if !haveGreatest || values.power > greatest {
			greatest = values.power
			haveGreatest = true
		}
		if values.controller == ctx.controller && (!haveControllerGreatest || values.power > controllerGreatest) {
			controllerGreatest = values.power
			haveControllerGreatest = true
		}
	}
	return haveGreatest && haveControllerGreatest && controllerGreatest >= greatest
}

// controllerControlsGreatestToughnessCreature reports whether the context
// controller controls a creature whose toughness equals the greatest toughness
// among all creatures on the battlefield ("you control the creature with the
// greatest toughness or tied for the greatest toughness"). It is false when no
// creatures with a defined toughness exist.
func controllerControlsGreatestToughnessCreature(g *game.Game, ctx conditionContext) bool {
	greatest := 0
	haveGreatest := false
	controllerGreatest := 0
	haveControllerGreatest := false
	for _, permanent := range g.Battlefield {
		if permanent.PhasedOut {
			continue
		}
		values := permanentValuesForCondition(g, permanent, ctx)
		if !slices.Contains(values.types, types.Creature) || !values.toughnessOK {
			continue
		}
		if !haveGreatest || values.toughness > greatest {
			greatest = values.toughness
			haveGreatest = true
		}
		if values.controller == ctx.controller && (!haveControllerGreatest || values.toughness > controllerGreatest) {
			controllerGreatest = values.toughness
			haveControllerGreatest = true
		}
	}
	return haveGreatest && haveControllerGreatest && controllerGreatest >= greatest
}

func permanentValuesForCondition(g *game.Game, permanent *game.Permanent, ctx conditionContext) permanentEffectiveValues {
	switch {
	case ctx.useBaseCharacteristics:
		return basePermanentValues(g, permanent)
	case ctx.characteristicsBefore != 0:
		values := permanentValuesBeforeLayer(g, permanent, ctx.characteristicsBefore)
		applyCounterAndTemporaryValues(permanent, &values)
		return values
	default:
		return effectivePermanentValues(g, permanent)
	}
}

func conditionObjectMatches(g *game.Game, ctx conditionContext, cond *game.Condition) bool {
	if cond == nil || cond.ObjectMatches.Exists && !cond.Object.Exists {
		return false
	}
	obj := ctx.obj
	if obj == nil && (ctx.event != nil || ctx.source != nil) {
		obj = &game.StackObject{Controller: ctx.controller}
		if ctx.event != nil {
			obj.HasTriggerEvent = true
			obj.TriggerEvent = *ctx.event
		}
		if ctx.source != nil {
			obj.SourceID = ctx.source.ObjectID
		}
	}
	ref := game.EventPermanentReference()
	if cond.Object.Exists {
		ref = cond.Object.Val
	}
	if obj == nil {
		return false
	}
	resolved, ok := resolveObjectReference(g, obj, ref)
	if !ok {
		return false
	}
	if cond.ObjectMatches.Exists &&
		!resolvedObjectMatchesConditionSelection(g, ctx, &resolved, &cond.ObjectMatches.Val) {
		return false
	}
	for _, cardType := range cond.Types {
		if !resolvedObjectHasType(g, &resolved, cardType) {
			return false
		}
	}
	return true
}

func resolvedObjectMatchesConditionSelection(
	g *game.Game,
	ctx conditionContext,
	resolved *resolvedObjectReference,
	selection *game.Selection,
) bool {
	if resolved == nil || selection == nil {
		return false
	}
	if resolved.permanent != nil {
		values := permanentValuesForCondition(g, resolved.permanent, ctx)
		subject := selectionSubject{
			kind:      subjectPermanent,
			g:         g,
			permanent: resolved.permanent,
			values:    values,
			viewer:    ctx.controller,
		}
		if selection.Controller != game.ControllerAny {
			subject.controller = values.controller
		}
		if ctx.source != nil {
			subject.sourceObjectID = ctx.source.ObjectID
		}
		return matchSelection(&subject, selection)
	}
	if resolved.stack != nil {
		colors, ok := stackObjectColors(g, resolved.stack)
		if !ok {
			return false
		}
		subject := selectionSubject{
			kind:   subjectCastSpell,
			g:      g,
			event:  game.Event{Colors: colors},
			viewer: ctx.controller,
		}
		return matchSelection(&subject, selection)
	}
	if resolved.snapshot.ObjectID == 0 {
		return false
	}
	subject := selectionSubject{
		kind:   subjectEventPermanent,
		g:      g,
		event:  game.Event{PermanentID: resolved.snapshot.ObjectID},
		viewer: ctx.controller,
	}
	if selection.Controller != game.ControllerAny {
		subject.controller = resolved.snapshot.Controller
	}
	if ctx.source != nil {
		subject.sourceObjectID = ctx.source.ObjectID
	}
	return matchSelection(&subject, selection)
}

func resolvedObjectHasType(g *game.Game, resolved *resolvedObjectReference, cardType types.Card) bool {
	if resolved.permanent != nil {
		return permanentHasType(g, resolved.permanent, cardType)
	}
	return slices.Contains(resolved.snapshot.Types, cardType)
}

func controllerControlsMatchingSelection(g *game.Game, ctx conditionContext, control game.SelectionCount) bool {
	return playersControlMatchingSelection(g, ctx, []game.PlayerID{ctx.controller}, control)
}

// controllerControlsNamed reports whether the context controller controls, for
// each requested name, at least one active battlefield permanent whose
// effective name matches. Names are compared case-insensitively with hyphens
// and spaces treated alike, so the printed Oracle spelling ("Urza's
// Power-Plant") matches the canonical card name ("Urza's Power Plant").
func controllerControlsNamed(g *game.Game, ctx conditionContext, names []string) bool {
	for _, name := range names {
		want := normalizeControlledName(name)
		if want == "" {
			return false
		}
		found := false
		for _, permanent := range g.Battlefield {
			if !activeBattlefieldPermanent(permanent) ||
				effectiveController(g, permanent) != ctx.controller {
				continue
			}
			if normalizeControlledName(permanentEffectiveName(g, permanent)) == want {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// normalizeControlledName canonicalizes a permanent name for the
// control-a-named-permanent predicate: lowercased, with every hyphen treated as
// a space and runs of whitespace collapsed. This reconciles the Oracle-text
// spelling used in conditions with the printed card name.
func normalizeControlledName(name string) string {
	var builder strings.Builder
	lastSpace := true
	for _, r := range strings.ToLower(name) {
		if r == '-' || r == ' ' || r == '\t' || r == '\n' {
			if !lastSpace {
				_ = builder.WriteByte(' ')
				lastSpace = true
			}
			continue
		}
		_, _ = builder.WriteRune(r)
		lastSpace = false
	}
	return strings.TrimSpace(builder.String())
}

func anyOpponentControlsMatchingSelection(g *game.Game, ctx conditionContext, control game.SelectionCount) bool {
	for _, opponent := range aliveOpponents(g, ctx.controller) {
		if playersControlMatchingSelection(g, ctx, []game.PlayerID{opponent}, control) {
			return true
		}
	}
	return false
}

func playersControlMatchingSelection(g *game.Game, ctx conditionContext, controllers []game.PlayerID, control game.SelectionCount) bool {
	if control.MinCount < 0 {
		return false
	}
	want := control.MinCount
	if want <= 0 {
		want = 1
	}
	allowed := make(map[game.PlayerID]bool, len(controllers))
	for _, controller := range controllers {
		allowed[controller] = true
	}
	count := 0
	totalPower := 0
	sel := control.Selection
	var distinctNames map[string]bool
	if control.DistinctNames.Exists {
		distinctNames = make(map[string]bool)
	}
	for _, permanent := range g.Battlefield {
		if permanent.PhasedOut {
			continue
		}
		if ctx.useBaseCharacteristics {
			if !allowed[permanent.Controller] {
				continue
			}
		}
		values := permanentValuesForCondition(g, permanent, ctx)
		if !ctx.useBaseCharacteristics && !allowed[values.controller] {
			continue
		}
		subject := selectionSubject{
			kind:      subjectPermanent,
			g:         g,
			permanent: permanent,
			values:    values,
			viewer:    ctx.controller,
			useBase:   ctx.useBaseCharacteristics,
		}
		if sel.Controller != game.ControllerAny {
			if ctx.useBaseCharacteristics {
				subject.controller = permanent.Controller
			} else {
				subject.controller = values.controller
			}
		}
		if ctx.source != nil {
			subject.sourceObjectID = ctx.source.ObjectID
		}
		if !matchSelection(&subject, &sel) {
			continue
		}
		count++
		if control.TotalPower.Exists {
			powerValues := &values
			if ctx.useBaseCharacteristics {
				effective := effectivePermanentValues(g, permanent)
				powerValues = &effective
			}
			if powerValues.powerOK {
				totalPower += powerValues.power
			}
		}
		if distinctNames != nil && values.name != "" {
			distinctNames[values.name] = true
		}
		if count >= want && !control.DistinctNames.Exists {
			if !control.TotalPower.Exists {
				return true
			}
			// Only a lower-bound total-power comparison is monotonic in the
			// running sum: once it matches it stays matched as more permanents
			// are counted, so an early match is final. An upper-bound comparison
			// ("total power N or less") can be satisfied by a partial sum yet
			// fail once every matching permanent is counted, so it must fall
			// through to the post-loop check.
			if totalPowerLowerBound(control.TotalPower.Val) && control.TotalPower.Val.Matches(totalPower) {
				return true
			}
		}
	}
	if count < want {
		return false
	}
	if control.TotalPower.Exists && !control.TotalPower.Val.Matches(totalPower) {
		return false
	}
	if control.DistinctNames.Exists && !control.DistinctNames.Val.Matches(len(distinctNames)) {
		return false
	}
	return control.TotalPower.Exists || control.DistinctNames.Exists
}

// totalPowerLowerBound reports whether a total-power comparison is a lower bound
// (">= n" or "> n"). A lower bound is monotonic in the running sum used by
// playersControlMatchingSelection, so an early match while iterating the
// battlefield is final. Upper-bound and exact comparisons are not, so their
// callers must count every matching permanent before concluding.
func totalPowerLowerBound(c compare.Int) bool {
	return c.Op == compare.GreaterOrEqual || c.Op == compare.GreaterThan
}

// countPlayerMatchingSelection counts permanents matching sel controlled by the
// given player, mirroring the subject construction of
// playersControlMatchingSelection without any count threshold.
func countPlayerMatchingSelection(g *game.Game, ctx conditionContext, player game.PlayerID, sel game.Selection) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if permanent.PhasedOut {
			continue
		}
		if ctx.useBaseCharacteristics {
			if permanent.Controller != player {
				continue
			}
		}
		values := permanentValuesForCondition(g, permanent, ctx)
		if !ctx.useBaseCharacteristics && values.controller != player {
			continue
		}
		subject := selectionSubject{
			kind:      subjectPermanent,
			g:         g,
			permanent: permanent,
			values:    values,
			viewer:    ctx.controller,
			useBase:   ctx.useBaseCharacteristics,
		}
		if sel.Controller != game.ControllerAny {
			if ctx.useBaseCharacteristics {
				subject.controller = permanent.Controller
			} else {
				subject.controller = values.controller
			}
		}
		if ctx.source != nil {
			subject.sourceObjectID = ctx.source.ObjectID
		}
		if matchSelection(&subject, &sel) {
			count++
		}
	}
	return count
}

// controlCountComparisonSatisfied evaluates a cross-player control-count
// comparison. The opponent-scoped side is quantified existentially
// (ControlPlayerAnyOpponent) or universally (ControlPlayerEachOpponent); with no
// opponents the universal form is vacuously true and the existential form false.
// A ControlPlayerTriggeringPlayer side names the specific player tied to the
// triggering event, so the comparison resolves against that one player.
func controlCountComparisonSatisfied(g *game.Game, ctx conditionContext, cmp game.ControlCountComparison) bool {
	youCount := countPlayerMatchingSelection(g, ctx, ctx.controller, cmp.Selection)
	if cmp.Left == game.ControlPlayerTriggeringPlayer || cmp.Right == game.ControlPlayerTriggeringPlayer {
		if ctx.event == nil {
			return false
		}
		triggeringCount := countPlayerMatchingSelection(g, ctx, ctx.event.Controller, cmp.Selection)
		left := youCount
		if cmp.Left != game.ControlPlayerController {
			left = triggeringCount
		}
		right := youCount
		if cmp.Right != game.ControlPlayerController {
			right = triggeringCount
		}
		return compare.Int{Op: cmp.Op, Value: right}.Matches(left)
	}
	universal := cmp.Left == game.ControlPlayerEachOpponent || cmp.Right == game.ControlPlayerEachOpponent
	opponents := aliveOpponents(g, ctx.controller)
	if len(opponents) == 0 {
		return universal
	}
	for _, opponent := range opponents {
		opponentCount := countPlayerMatchingSelection(g, ctx, opponent, cmp.Selection)
		left := youCount
		if cmp.Left != game.ControlPlayerController {
			left = opponentCount
		}
		right := youCount
		if cmp.Right != game.ControlPlayerController {
			right = opponentCount
		}
		satisfied := compare.Int{Op: cmp.Op, Value: right}.Matches(left)
		if universal && !satisfied {
			return false
		}
		if !universal && satisfied {
			return true
		}
	}
	return universal
}
func anyPlayerLifeAtMost(g *game.Game, maximum int) bool {
	for playerID := range game.PlayerID(game.NumPlayers) {
		player, ok := playerByID(g, playerID)
		if ok && !player.Eliminated && player.Life <= maximum {
			return true
		}
	}
	return false
}

func anyOpponentPoisonAtLeast(g *game.Game, controller game.PlayerID, minimum int) bool {
	for _, opponent := range aliveOpponents(g, controller) {
		player, ok := playerByID(g, opponent)
		if ok && player.PoisonCounters >= minimum {
			return true
		}
	}
	return false
}

func conditionTargetEnteredThisTurn(g *game.Game, ctx conditionContext, targetIndex int) bool {
	if ctx.obj == nil {
		return false
	}
	permanent, ok := effectPermanentAt(g, ctx.obj, targetIndex)
	if !ok {
		return false
	}
	return permanentEnteredThisTurn(g, permanent.ObjectID)
}

// permanentEnteredThisTurn reports whether the permanent identified by id entered
// the battlefield during the current turn, scanning this turn's events for its
// enter-the-battlefield event.
func permanentEnteredThisTurn(g *game.Game, permanentID id.ID) bool {
	return eventsThisTurnWindow(g).any(func(event game.Event) bool {
		return event.Kind == game.EventPermanentEnteredBattlefield && event.PermanentID == permanentID
	})
}

// permanentDealtDamageThisTurn reports whether the permanent identified by id was
// dealt damage during the current turn, scanning this turn's events for a
// damage-dealt event whose recipient permanent is it (CR 120).
func permanentDealtDamageThisTurn(g *game.Game, permanentID id.ID) bool {
	return eventsThisTurnWindow(g).any(func(event game.Event) bool {
		return event.Kind == game.EventDamageDealt &&
			event.DamageRecipient == game.DamageRecipientPermanent &&
			event.PermanentID == permanentID
	})
}

func eventPermanentNameUniqueAmongControlledAndGraveyardCreatures(g *game.Game, ctx conditionContext) bool {
	if ctx.event == nil || ctx.event.PermanentID == 0 {
		return false
	}
	resolved, ok := resolvePermanentOrLastKnown(g, ctx.event.PermanentID)
	if !ok {
		return false
	}
	name := resolvedObjectName(g, &resolved)
	if name == "" {
		return false
	}
	for _, permanent := range g.Battlefield {
		if !activeBattlefieldPermanent(permanent) ||
			permanent.ObjectID == ctx.event.PermanentID ||
			effectiveController(g, permanent) != ctx.controller ||
			!permanentHasType(g, permanent, types.Creature) {
			continue
		}
		if def, ok := permanentCardDef(g, permanent); ok && def.Name == name {
			return false
		}
	}
	player, ok := playerByID(g, ctx.controller)
	if !ok {
		return false
	}
	for _, cardID := range player.Graveyard.All() {
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		def := cardFaceOrDefault(card, game.FaceFront)
		if def.Name == name && def.HasType(types.Creature) {
			return false
		}
	}
	return true
}

func resolvedObjectName(g *game.Game, resolved *resolvedObjectReference) string {
	if resolved.permanent != nil {
		if resolved.permanent.Token {
			return permanentTokenName(resolved.permanent)
		}
		return permanentEffectiveName(g, resolved.permanent)
	}
	return resolved.snapshot.Name
}

func activationConditionSatisfied(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, condition opt.V[game.Condition]) bool {
	return conditionSatisfied(g, conditionContext{
		controller: playerID,
		source:     permanent,
	}, condition)
}

// controllerCreatedTokenThisTurn reports whether the controller created at least
// one token during the current turn. Token creation is recorded in the per-turn
// event log as a battlefield-entry event carrying the token's definition, so the
// scan reuses that log rather than tracking separate mutable state.
func controllerCreatedTokenThisTurn(g *game.Game, controller game.PlayerID) bool {
	return eventsThisTurnWindow(g).any(func(event game.Event) bool {
		return event.Kind == game.EventPermanentEnteredBattlefield &&
			event.TokenDef != nil &&
			event.Controller == controller
	})
}

// conditionEventHistorySatisfied returns true when the chosen turn's event
// history contains at least hist.MinCount events matching hist.Pattern (at least
// one when MinCount is zero). The source permanent is passed to
// triggerMatchesEvent so controller-relative filters (TriggerControllerYou,
// TriggerPlayerYou, etc.) resolve correctly. A nil source permanent fails closed
// for any pattern that references the source (such filters can never match
// without one); source-agnostic patterns, such as "a creature died this turn"
// gating a resolving instant, evaluate against the event log directly.
func conditionEventHistorySatisfied(g *game.Game, ctx conditionContext, hist *game.EventHistoryCondition) bool {
	if ctx.source == nil && eventHistoryPatternNeedsSource(&hist.Pattern) {
		return false
	}
	var events eventWindow
	switch hist.Window {
	case game.EventHistoryCurrentTurn:
		events = eventsThisTurnWindow(g)
	case game.EventHistoryPreviousTurn:
		events = eventsPreviousTurnWindow(g)
	default:
		return false
	}
	want := max(hist.MinCount, 1)
	matches := 0
	for i := range events {
		if triggerMatchesEvent(g, ctx.source, &hist.Pattern, events[i]) {
			matches++
			if matches >= want {
				return true
			}
		}
	}
	return false
}

// eventHistoryPatternNeedsSource reports whether a trigger pattern consults the
// ability's source permanent (its controller, identity, or attachment) to match
// an event. Such patterns can never match without a source, so a source-agnostic
// caller (a resolving instant gating on event history) fails them closed instead
// of dereferencing a nil source.
func eventHistoryPatternNeedsSource(pattern *game.TriggerPattern) bool {
	return pattern.Controller != game.TriggerControllerAny ||
		pattern.CauseController != game.TriggerControllerAny ||
		pattern.Player != game.TriggerPlayerAny ||
		pattern.Source != game.TriggerSourceAny ||
		pattern.ExcludeSelf ||
		pattern.SubjectSelectionOrSelf ||
		pattern.DamageSourceSelectionOrSelf ||
		pattern.DamageRecipientIsSource ||
		pattern.SpellTargetsSource ||
		!pattern.StepPlayerSourceAttachedSelection.Empty()
}
