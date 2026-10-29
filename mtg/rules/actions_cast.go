package rules

import (
	"maps"
	"math"
	"slices"
	"strconv"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

func (e *Engine) applyPlayLand(g *game.Game, playerID game.PlayerID, cardID id.ID) bool {
	return e.applyPlayLandFace(g, playerID, cardID, game.FaceFront)
}

func (e *Engine) applyPlayLandFace(g *game.Game, playerID game.PlayerID, cardID id.ID, face game.FaceIndex) bool {
	return e.applyPlayLandFaceWithChoices(g, playerID, cardID, face, [game.NumPlayers]PlayerAgent{}, nil)
}

func (e *Engine) applyPlayLandFaceWithChoices(g *game.Game, playerID game.PlayerID, cardID id.ID, face game.FaceIndex, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	return e.applyPlayLandFaceFromZoneWithChoices(g, playerID, cardID, zone.Hand, face, agents, log)
}

func (e *Engine) applyPlayLandFaceFromZoneWithChoices(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, agents [game.NumPlayers]PlayerAgent, log *TurnLog, announced ...game.PlayPermission) bool {
	if !canPlayAnyLand(g, playerID) {
		return false
	}

	player := g.Players[playerID]
	sourcePlayer := castSourcePlayer(g, player, cardID, sourceZone)
	card, ok := landCardInstanceFaceFromZone(g, sourcePlayer, cardID, sourceZone, face)
	if !ok {
		return false
	}
	var landPermission game.RuleEffect
	var permission game.PlayPermission
	if len(announced) > 0 {
		permission = announced[0]
	}
	if sourceZone != zone.Hand {
		routes := playLandFromZoneRoutes(g, playerID, cardID, sourceZone)
		permission, ok = e.selectPermanentTypePlay(g, playerID, permission, sourceZone, cardFaceOrDefault(card, face), len(routes) > 0, agents, log)
		if !ok {
			return false
		}
		if permission == (game.PlayPermission{}) {
			landPermission = e.choosePlayLandRoute(g, playerID, routes, agents, log)
		}
	} else if permission != (game.PlayPermission{}) {
		return false
	}
	source, ok := playerCardsInZone(sourcePlayer, sourceZone)
	if !ok || !source.Remove(cardID) {
		return false
	}

	if _, ok := createCardPermanentFaceWithChoices(e, g, card, playerID, sourceZone, face, agents, log); !ok {
		restoreCastSourceCard(sourcePlayer, cardID, sourceZone)
		return false
	}
	g.Turn.LandsPlayedThisTurn++
	recordPermanentTypePlay(g, playerID, permission)
	if sourceZone != zone.Hand {
		recordExilePlayPermissionUse(g, landPermission)
	}
	emitCardPlayedFromExileEvent(g, playerID, cardID, sourceZone)
	emitLandPlayedEvent(g, playerID, cardID)
	return true
}

// choosePlayLandRoute returns the permission that authorizes a land play from
// routes, ordered by playLandFromZoneRoutes. A route that spends no
// once-per-turn use is always taken when one exists; otherwise the player
// chooses which once-per-turn permission to spend (CR 305.1, as for a spell in
// CR 601.2b).
func (e *Engine) choosePlayLandRoute(g *game.Game, playerID game.PlayerID, routes []game.RuleEffect, agents [game.NumPlayers]PlayerAgent, log *TurnLog) game.RuleEffect {
	if len(routes) == 0 {
		return game.RuleEffect{}
	}
	if len(routes) == 1 || !routes[0].OncePerTurn {
		return routes[0]
	}
	options := make([]game.ChoiceOption, len(routes))
	for i, route := range routes {
		options[i] = game.ChoiceOption{Index: i, Label: "Play land " + castRouteSourceLabel(g, castRoute{permission: route}, i)}
	}
	selected := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           playerID,
		Prompt:           "Choose land play permission",
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}, log)
	if len(selected) == 1 && selected[0] >= 0 && selected[0] < len(routes) {
		return routes[selected[0]]
	}
	return routes[0]
}

func (e *Engine) applyCastSpell(g *game.Game, playerID game.PlayerID, cast action.CastSpellAction) bool {
	return e.applyCastSpellWithChoices(g, playerID, cast, [game.NumPlayers]PlayerAgent{}, nil)
}

func (e *Engine) applyCastSpellWithChoices(g *game.Game, playerID game.PlayerID, cast action.CastSpellAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if !castGiftRecipientValid(g, playerID, cast) {
		return false
	}
	if cast.Mutate {
		return e.applyMutateCastWithChoices(g, playerID, cast, agents, log)
	}
	sourceZone := normalizedCastSourceZone(cast)
	if sourceZone == zone.Battlefield {
		if cast.Permission != (game.PlayPermission{}) {
			return false
		}
		return e.applyPreparedCopyWithChoices(g, playerID, cast, agents, log)
	}

	branch := castBranchForCast(cast)
	var validationPermission []game.PlayPermission
	if cast.Permission != (game.PlayPermission{}) {
		validationPermission = []game.PlayPermission{cast.Permission}
	}
	if !e.canCastSpellFaceFromZoneWithOptions(g, playerID, cast.CardID, sourceZone, cast.Face, cast.Targets, cast.XValue, cast.ChosenModes, effectiveKickerCount(cast.KickerPaid, cast.KickerCount), cast.Overloaded, cast.GiftPromised, cast.Bargained, cast.Bestowed, cast.Offspring, validationPermission...) {
		return false
	}

	player := g.Players[playerID]
	card, ok := g.GetCardInstance(cast.CardID)
	if !ok {
		return false
	}
	spellDef := cardFaceOrDefault(card, cast.Face)
	announcementDef := spellDef
	if cast.Overloaded {
		announcementDef = overloadSpellDef(spellDef)
	}
	completedTargets, ok := e.completeSpellAnnouncementTargets(g, playerID, announcementDef, cast.ChosenModes, cast.Targets, agents, log, branch)
	if !ok || !e.canCastSpellFaceFromZoneWithOptions(g, playerID, cast.CardID, sourceZone, cast.Face, completedTargets, cast.XValue, cast.ChosenModes, effectiveKickerCount(cast.KickerPaid, cast.KickerCount), cast.Overloaded, cast.GiftPromised, cast.Bargained, cast.Bestowed, cast.Offspring, validationPermission...) {
		return false
	}
	cast.Targets = completedTargets
	permissions := independentCastPermissionsForZone(g, playerID, card.ID, sourceZone, cast.Face)
	cast.Permission, ok = e.selectPermanentTypePlay(g, playerID, cast.Permission, sourceZone, castSelectionFace(g, playerID, spellDef, cast.Bestowed), len(permissions) > 0, agents, log)
	if !ok {
		return false
	}
	if cast.Permission != (game.PlayPermission{}) {
		permissions = []payment.SpellCastPermission{payment.SpellCastPermissionRuleEffect}
	}
	targetCounts, ok := spellTargetCounts(g, playerID, announcementDef, cast.ChosenModes, cast.Targets, branch)
	if !ok {
		panic("validated spell targets could not be segmented")
	}
	// Splice onto Arcane (CR 702.47): as an Arcane spell is cast from hand for its
	// normal cost, the caster may reveal splice cards in hand, pay their mana
	// splice costs as additional costs, and append their spell effects to this
	// spell. Overloaded casts and non-Arcane spells are never spliceable. When no
	// card is spliced these are all nil, leaving the cast unchanged.
	var splice spliceCastResult
	if sourceZone == zone.Hand && !cast.Overloaded && spellIsArcane(spellDef) {
		splicePermissions := castPermissionsForZone(g, playerID, card.ID, sourceZone, cast.Face)
		splice = e.chooseSpliceOntoArcane(g, playerID, card.ID, spellDef, cast, splicePermissions, agents, log)
	}
	// castRoutes are the permissions this cast may use, read before the card
	// leaves its source zone because plotted, foretold, and linked-exile status,
	// a permission's exile-counter filter (Evelyn, the Covetous), and a
	// top-of-library permission's position filter (Bolas's Citadel) no longer
	// match once the card moves. A free route (plot, a linked free cast such as
	// Court of Locthwain's, or a per-card "play it without paying its mana cost"
	// permission such as Dauthi Voidwalker's) replaces the mana cost with an
	// empty payment; a foretell route pays the foretell cost; a pay-life route
	// pays life equal to the spell's mana value. Each route carries only its own
	// permission's cost replacement, spending rider and usage (CR 118.9a,
	// 118.14); only the route actually cast through has its once-per-turn use
	// recorded or its one-shot linked permission consumed once the spell is put
	// on the stack.
	mode := castCostMode{overloaded: cast.Overloaded, bestowed: cast.Bestowed}
	castRoutes := spellCastRoutes(g, playerID, card.ID, sourceZone, cast.Face, mode)

	// CR 601.2a: proposing a cast first moves the card from its source zone to
	// the stack as the topmost object. Doing this before the spell's costs are
	// determined and paid (CR 601.2f-h) is what makes it impossible for a cost
	// paid from the source zone — "discard a card", "exile a blue card from your
	// hand", and the like — to ever select the very card being cast.
	obj := &game.StackObject{
		ID:                            g.IDGen.Next(),
		Kind:                          game.StackSpell,
		SourceID:                      cast.CardID,
		Face:                          cast.Face,
		Controller:                    playerID,
		Targets:                       append([]game.Target(nil), cast.Targets...),
		TargetCounts:                  targetCounts,
		ChosenModes:                   append([]int(nil), cast.ChosenModes...),
		XValue:                        cast.XValue,
		KickerPaid:                    cast.KickerPaid,
		KickerCount:                   cast.KickerCount,
		GiftPromised:                  cast.GiftPromised,
		GiftRecipient:                 cast.GiftRecipient,
		Bargained:                     cast.Bargained,
		OffspringPaid:                 cast.Offspring,
		Overloaded:                    cast.Overloaded,
		SourceZone:                    sourceZone,
		SplicedContent:                splice.contents,
		SplicedTargets:                splice.targets,
		SplicedTargetCounts:           splice.targetCounts,
		CastDuringControllerMainPhase: playerID == g.Turn.ActivePlayer && g.Turn.IsMainPhase(),
	}
	counterTransition := prepareCastCounterTransition(g, card, sourceZone)
	if !removeCastSourceCard(g, castSourcePlayer(g, player, cast.CardID, sourceZone), cast.CardID, sourceZone) {
		return false
	}
	counterTransition.apply(g)
	g.Stack.Push(obj)

	targetEvents := prepareTargetEvents(g, obj)
	riderSnapshot, _ := manaSpendRiderSnapshot(g, playerID)
	request := payment.SpellRequest{
		PlayerID:        playerID,
		CardID:          card.ID,
		SourceZone:      sourceZone,
		Card:            spellDef,
		XValue:          cast.XValue,
		KickerPaid:      cast.KickerPaid,
		KickerCount:     cast.KickerCount,
		Bargained:       cast.Bargained,
		Offspring:       cast.Offspring,
		ChosenModes:     cast.ChosenModes,
		CastPermissions: permissions,
		Targets:         cast.Targets,
		SpliceManaCosts: splice.manaCosts,
		Bestowed:        cast.Bestowed,
	}
	requests := mode.castRouteRequests(request, castRoutes)
	route, prefs := e.chooseCastRoute(g, castRoutes, requests, agents, log)
	request = requests[route]
	request.Prefs = prefs
	paymentResult, ok := paymentOrch.paySpellCosts(g, request)
	if !ok {
		// CR 728: the proposed cast is illegal because its costs can't be paid,
		// so the proposal is undone — take the spell back off the stack and
		// return the card to the zone it was cast from.
		g.Stack.RemoveByID(obj.ID)
		restoreCastSourceCard(castSourcePlayer(g, player, cast.CardID, sourceZone), cast.CardID, sourceZone)
		counterTransition.undo(g)
		return false
	}
	if sourceZone == zone.Command {
		recordCommanderCast(g, playerID, cast.CardID)
	}
	consumeCastLinkedExileForFreePermission(g, castRoutes[route].linkedFree.ID)
	recordExilePlayPermissionUse(g, castRoutes[route].permission)
	recordPermanentTypePlay(g, playerID, cast.Permission)
	obj.Evoked = !cast.Overloaded && evokeAlternativeChosen(spellDef, paymentResult.AlternativeIndex)
	obj.Converted = !cast.Overloaded && convertedAlternativeChosen(spellDef, paymentResult.AlternativeIndex)
	obj.Dashed = !cast.Overloaded && dashAlternativeChosen(spellDef, paymentResult.AlternativeIndex)
	obj.Bestowed = cast.Bestowed
	obj.BuybackPaid = paymentResult.BuybackPaid
	obj.Flashback = paymentResult.CastPermission == payment.SpellCastPermissionFlashback
	obj.AdditionalCostsPaid = paymentResult.AdditionalCostsPaid
	obj.SacrificedAsCostIDs = paymentResult.SacrificedIDs
	obj.PaidCostSubjects = paymentResult.PaidCostSubjects
	obj.ColorsOfManaSpentToCast = distinctManaColorsSpent(paymentResult.PoolSpend)
	obj.ManaSpentByColorToCast = manaSpentByColor(paymentResult.PoolSpend)
	obj.ManaSpentToCast = totalManaSpent(paymentResult.PoolSpend)
	obj.ManaFromCreaturesSpentToCast = creatureManaSpent(paymentResult.PoolSpend)

	// CR 601.2i: the costs are paid, so the spell becomes cast. The card is
	// already on the stack; emit its zone-change and spell-cast events now so
	// "when you cast" triggers fire after payment.
	typeLine := stackSpellTypeLine(g, obj, spellDef)
	emitSpellCastEventsWithTargets(g, obj, game.Event{
		SourceID:                     cast.CardID,
		StackObjectID:                obj.ID,
		Controller:                   playerID,
		CardID:                       cast.CardID,
		Face:                         cast.Face,
		CardTypes:                    typeLine.types,
		CardSupertypes:               typeLine.supertypes,
		CardSubtypes:                 typeLine.subtypes,
		Colors:                       spellColors(spellDef),
		ManaValue:                    stackObjectKnownManaValue(g, obj),
		ManaSpentToCast:              opt.Val(totalManaSpent(paymentResult.PoolSpend)),
		ManaFromCreaturesSpentToCast: opt.Val(creatureManaSpent(paymentResult.PoolSpend)),
		KickerPaid:                   cast.KickerPaid,
		FromZone:                     sourceZone,
		ToZone:                       zone.Stack,
	}, targetEvents)
	resolveSpellCastManaSpendRiders(g, playerID, riderSnapshot, paymentResult.PoolSpend, spellDef, obj)
	return true
}

func (e *Engine) applyMutateCastWithChoices(g *game.Game, playerID game.PlayerID, cast action.CastSpellAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	sourceZone := normalizedCastSourceZone(cast)
	var validationPermission []game.PlayPermission
	if cast.Permission != (game.PlayPermission{}) {
		validationPermission = []game.PlayPermission{cast.Permission}
	}
	if !canCastMutateSpell(g, playerID, cast.CardID, sourceZone, cast.MutateTargetID, validationPermission...) {
		return false
	}
	player := g.Players[playerID]
	card, ok := g.GetCardInstance(cast.CardID)
	if !ok {
		return false
	}
	spellDef := cardFaceOrDefault(card, game.FaceFront)
	permissions := independentCastPermissionsForZone(g, playerID, card.ID, sourceZone, game.FaceFront)
	cast.Permission, ok = e.selectPermanentTypePlay(g, playerID, cast.Permission, sourceZone, spellDef, len(permissions) > 0, agents, log)
	if !ok {
		return false
	}
	if cast.Permission != (game.PlayPermission{}) {
		permissions = []payment.SpellCastPermission{payment.SpellCastPermissionRuleEffect}
	}
	mutateCost, ok := spellDef.MutateCost()
	if !ok {
		return false
	}
	alternative := mutateAlternativeCost(mutateCost)
	// The permissions authorizing a cast from exile, with their spending riders
	// and usage, are read before the card leaves exile: an exile-counter filter
	// no longer matches once the card moves (Evelyn, the Covetous). Mutate is an
	// alternative cost, so the cast is never one made without paying its mana
	// cost (CR 118.9a).
	castRoutes := spellCastRoutes(g, playerID, card.ID, sourceZone, game.FaceFront, mutateCastCostMode)

	// CR 601.2a: move the card to the stack before its costs are paid, so a
	// from-zone cost can't select the card being cast (see applyCastSpellWithChoices).
	obj := &game.StackObject{
		ID:             g.IDGen.Next(),
		Kind:           game.StackSpell,
		SourceID:       cast.CardID,
		Face:           game.FaceFront,
		Controller:     playerID,
		Targets:        []game.Target{game.PermanentTarget(cast.MutateTargetID)},
		TargetCounts:   []int{1},
		Mutate:         true,
		MutateTargetID: cast.MutateTargetID,
		SourceZone:     sourceZone,
	}
	counterTransition := prepareCastCounterTransition(g, card, sourceZone)
	if !removeCastSourceCard(g, castSourcePlayer(g, player, cast.CardID, sourceZone), cast.CardID, sourceZone) {
		return false
	}
	counterTransition.apply(g)
	g.Stack.Push(obj)

	targetEvents := prepareTargetEvents(g, obj)
	riderSnapshot, _ := manaSpendRiderSnapshot(g, playerID)
	request := payment.SpellRequest{
		PlayerID:        playerID,
		CardID:          card.ID,
		SourceZone:      sourceZone,
		Card:            spellDef,
		Alternative:     opt.Val(alternative),
		CastPermissions: permissions,
	}
	requests := mutateCastRouteRequests(request, castRoutes)
	route, prefs := e.chooseCastRoute(g, castRoutes, requests, agents, log)
	request = requests[route]
	request.Prefs = prefs
	paymentResult, ok := paymentOrch.paySpellCosts(g, request)
	if !ok {
		// CR 728: undo the proposal when costs can't be paid.
		g.Stack.RemoveByID(obj.ID)
		restoreCastSourceCard(castSourcePlayer(g, player, cast.CardID, sourceZone), cast.CardID, sourceZone)
		counterTransition.undo(g)
		return false
	}
	if sourceZone == zone.Command {
		recordCommanderCast(g, playerID, cast.CardID)
	}
	recordExilePlayPermissionUse(g, castRoutes[route].permission)
	recordPermanentTypePlay(g, playerID, cast.Permission)
	obj.AdditionalCostsPaid = paymentResult.AdditionalCostsPaid
	obj.SacrificedAsCostIDs = paymentResult.SacrificedIDs
	obj.PaidCostSubjects = paymentResult.PaidCostSubjects
	typeLine := stackSpellTypeLine(g, obj, spellDef)
	emitSpellCastEventsWithTargets(g, obj, game.Event{
		SourceID:                     cast.CardID,
		StackObjectID:                obj.ID,
		Controller:                   playerID,
		CardID:                       cast.CardID,
		Face:                         game.FaceFront,
		CardTypes:                    typeLine.types,
		CardSupertypes:               typeLine.supertypes,
		CardSubtypes:                 typeLine.subtypes,
		Colors:                       spellColors(spellDef),
		ManaValue:                    stackObjectKnownManaValue(g, obj),
		ManaSpentToCast:              opt.Val(totalManaSpent(paymentResult.PoolSpend)),
		ManaFromCreaturesSpentToCast: opt.Val(creatureManaSpent(paymentResult.PoolSpend)),
		FromZone:                     sourceZone,
		ToZone:                       zone.Stack,
	}, targetEvents)
	obj.ColorsOfManaSpentToCast = distinctManaColorsSpent(paymentResult.PoolSpend)
	obj.ManaSpentByColorToCast = manaSpentByColor(paymentResult.PoolSpend)
	obj.ManaSpentToCast = totalManaSpent(paymentResult.PoolSpend)
	obj.ManaFromCreaturesSpentToCast = creatureManaSpent(paymentResult.PoolSpend)
	resolveSpellCastManaSpendRiders(g, playerID, riderSnapshot, paymentResult.PoolSpend, spellDef, obj)
	return true
}

func canCastMutateSpell(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, targetID id.ID, announced ...game.PlayPermission) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return false
	}
	player := g.Players[playerID]
	card, ok := g.GetCardInstance(cardID)
	if !ok || !castSourceContains(castSourcePlayer(g, player, cardID, sourceZone), cardID, sourceZone) {
		return false
	}
	switch sourceZone {
	case zone.Hand:
	case zone.Command:
		if !isCommanderOwnedBy(g, playerID, cardID) {
			return false
		}
	case zone.Exile:
		// CR 702.140a: Mutate is an alternative cost, usable from exile under
		// whatever permission authorizes casting the card there (CR 118.9a):
		// an Adventure card's CR 715.3d permission or an independent per-card
		// or group cast/play permission. The mutate-mode routes below leave out
		// permissions that only cast the card without paying its mana cost.
		if !exileCastAuthorized(g, playerID, cardID, game.FaceFront) {
			return false
		}
	case zone.Graveyard:
		if len(castPermissionsForZone(g, playerID, cardID, sourceZone, game.FaceFront)) == 0 {
			return false
		}
	default:
		return false
	}
	spellDef := cardFaceOrDefault(card, game.FaceFront)
	castFace := castSelectionFace(g, playerID, spellDef, false)
	permissions, valid := selectedSpellPermissions(g, playerID, cardID, sourceZone, game.FaceFront, castFace, announced)
	if !valid || castFromZoneProhibited(g, playerID, sourceZone) ||
		spellCastProhibited(g, playerID, castFace) || spellCastLimitReached(g, playerID, castFace) {
		return false
	}
	mutateCost, ok := spellDef.MutateCost()
	if !ok ||
		!spellDef.HasType(types.Creature) ||
		!isSupportedSpell(spellDef) ||
		!canCastAtCurrentTiming(g, playerID, castFace) ||
		!mutateTargetLegal(g, playerID, card.Owner, spellDef, targetID) {
		return false
	}
	routes := spellCastRoutes(g, playerID, cardID, sourceZone, game.FaceFront, mutateCastCostMode)
	return anyCastRouteAvailable(g, spellDef, 0, mutateCastRouteRequests(payment.SpellRequest{
		PlayerID:        playerID,
		CardID:          cardID,
		SourceZone:      sourceZone,
		Card:            spellDef,
		Alternative:     opt.Val(mutateAlternativeCost(mutateCost)),
		CastPermissions: permissions,
	}, routes))
}

// mutateCastCostMode is the cost mode of a mutate cast: Mutate is an
// alternative cost (CR 702.140a), so no route casts it without paying its mana
// cost (CR 118.9a).
var mutateCastCostMode = castCostMode{mutate: true}

// mutateCastRouteRequests returns one mutate payment request per route, each
// paying the mutate alternative cost with only its own route's spending rider.
func mutateCastRouteRequests(base payment.SpellRequest, routes []castRoute) []payment.SpellRequest {
	requests := make([]payment.SpellRequest, len(routes))
	for i, route := range routes {
		requests[i] = base
		requests[i].SpendAnyMana = route.spendAs
	}
	return requests
}

func legalMutateTargets(g *game.Game, playerID, owner game.PlayerID, spellDef *game.CardDef) []*game.Permanent {
	var targets []*game.Permanent
	for _, permanent := range g.Battlefield {
		if mutateTargetLegal(g, playerID, owner, spellDef, permanent.ObjectID) {
			targets = append(targets, permanent)
		}
	}
	return targets
}

func mutateTargetLegal(g *game.Game, playerID, owner game.PlayerID, spellDef *game.CardDef, targetID id.ID) bool {
	target, ok := permanentByObjectID(g, targetID)
	return ok &&
		!target.PhasedOut &&
		target.Owner == owner &&
		permanentHasType(g, target, types.Creature) &&
		!permanentHasSubtype(g, target, types.Human) &&
		!targetProtectedFromSource(g, playerID, spellDef, 0, game.PermanentTarget(targetID))
}

func mutateAlternativeCost(manaCost cost.Mana) cost.Alternative {
	return cost.Alternative{
		Label:    "Mutate",
		ManaCost: opt.Val(append(cost.Mana(nil), manaCost...)),
	}
}

func canCastPreparedCopy(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, targets []game.Target, xValue int, chosenModes []int) bool {
	if !canAct(g, playerID) ||
		playerID != g.Turn.PriorityPlayer ||
		permanent == nil ||
		!permanent.Prepared ||
		permanent.PhasedOut ||
		effectiveController(g, permanent) != playerID ||
		xValue < 0 {
		return false
	}
	sourceID, sourceDef, ok := preparedSpellSource(g, permanent)
	if !ok {
		return false
	}
	spellDef, ok := sourceDef.FaceDef(game.FaceAlternate)
	if !ok {
		return false
	}
	if xValue != 0 &&
		!costHasVariableMana(manaCostPtr(spellDef.ManaCost)) &&
		!additionalCostsUseX(spellDef.AdditionalCosts) {
		return false
	}
	branch := game.CastBranch{XValue: opt.Val(xValue), KickerCount: opt.Val(0)}
	if !modesValidForSpellAt(g, playerID, spellDef, chosenModes) ||
		!isSupportedSpell(spellDef) ||
		(!spellDef.HasType(types.Instant) && !spellDef.HasType(types.Sorcery)) ||
		!spellTargetsValidForCast(g, playerID, spellDef, chosenModes, targets, xValue, 0, branch) ||
		!canCastAtCurrentTiming(g, playerID, castSelectionFace(g, playerID, spellDef, false)) {
		return false
	}
	return paymentOrch.canPaySpellCosts(g, payment.SpellRequest{
		PlayerID:   playerID,
		CardID:     sourceID,
		SourceZone: zone.Battlefield,
		Card:       spellDef,
		XValue:     xValue,
		Targets:    targets,
	})
}

func (e *Engine) applyPreparedCopyWithChoices(g *game.Game, playerID game.PlayerID, cast action.CastSpellAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if cast.Face != game.FaceAlternate || cast.KickerPaid {
		return false
	}
	permanent := preparedSourcePermanent(g, cast.CardID)
	if !canCastPreparedCopy(g, playerID, permanent, cast.Targets, cast.XValue, cast.ChosenModes) {
		return false
	}
	sourceID, sourceDef, ok := preparedSpellSource(g, permanent)
	if !ok {
		return false
	}
	spellDef, ok := sourceDef.FaceDef(game.FaceAlternate)
	if !ok {
		return false
	}
	branch := castBranchForCast(cast)
	completedTargets, ok := e.completeSpellAnnouncementTargets(g, playerID, spellDef, cast.ChosenModes, cast.Targets, agents, log, branch)
	if !ok || !canCastPreparedCopy(g, playerID, permanent, completedTargets, cast.XValue, cast.ChosenModes) {
		return false
	}
	cast.Targets = completedTargets
	targetCounts, ok := spellTargetCounts(g, playerID, spellDef, cast.ChosenModes, cast.Targets, branch)
	if !ok {
		panic("validated prepared spell targets could not be segmented")
	}
	stackID := g.IDGen.Next()
	targetEvents := prepareTargetEvents(g, &game.StackObject{
		ID: stackID, Kind: game.StackSpell, Controller: playerID,
		SourceID: sourceID, SourceCardID: permanent.CardInstanceID, SourceTokenDef: permanent.TokenDef,
		Face: game.FaceAlternate, SourceZone: zone.Battlefield, Copy: true,
		Targets: cast.Targets, TargetCounts: targetCounts,
		ChosenModes: cast.ChosenModes, XValue: cast.XValue,
	})
	prefs := e.paymentPreferencesForSpellFromZone(g, playerID, sourceID, zone.Battlefield, game.FaceAlternate, spellDef, cast.XValue, agents, log)
	riderSnapshot, _ := manaSpendRiderSnapshot(g, playerID)
	paymentResult, ok := paymentOrch.paySpellCosts(g, payment.SpellRequest{
		PlayerID:   playerID,
		CardID:     sourceID,
		SourceZone: zone.Battlefield,
		Card:       spellDef,
		XValue:     cast.XValue,
		Targets:    cast.Targets,
		Prefs:      prefs,
	})
	if !ok {
		return false
	}
	permanent.Prepared = false
	obj := &game.StackObject{
		ID:                  stackID,
		Kind:                game.StackSpell,
		SourceID:            sourceID,
		Face:                game.FaceAlternate,
		SourceCardID:        permanent.CardInstanceID,
		SourceTokenDef:      permanent.TokenDef,
		Controller:          playerID,
		Targets:             append([]game.Target(nil), cast.Targets...),
		TargetCounts:        targetCounts,
		ChosenModes:         append([]int(nil), cast.ChosenModes...),
		XValue:              cast.XValue,
		Copy:                true,
		CastCopy:            true,
		BuybackPaid:         paymentResult.BuybackPaid,
		AdditionalCostsPaid: paymentResult.AdditionalCostsPaid,
		SacrificedAsCostIDs: paymentResult.SacrificedIDs,
		PaidCostSubjects:    paymentResult.PaidCostSubjects,
		SourceZone:          zone.Battlefield,

		CastDuringControllerMainPhase: playerID == g.Turn.ActivePlayer && g.Turn.IsMainPhase(),
	}
	g.Stack.Push(obj)
	emitCastTargetEvents(g, obj, targetEvents)
	typeLine := stackSpellTypeLine(g, obj, spellDef)
	emitEvent(g, game.Event{
		Kind:                         game.EventSpellCast,
		SourceID:                     sourceID,
		StackObjectID:                obj.ID,
		Controller:                   playerID,
		CardID:                       permanent.CardInstanceID,
		Face:                         game.FaceAlternate,
		PermanentID:                  permanent.ObjectID,
		TokenDef:                     permanent.TokenDef,
		CardTypes:                    typeLine.types,
		CardSupertypes:               typeLine.supertypes,
		CardSubtypes:                 typeLine.subtypes,
		Colors:                       spellColors(spellDef),
		CardCharacteristicsKnown:     true,
		ManaValue:                    stackObjectKnownManaValue(g, obj),
		ManaSpentToCast:              opt.Val(totalManaSpent(paymentResult.PoolSpend)),
		ManaFromCreaturesSpentToCast: opt.Val(creatureManaSpent(paymentResult.PoolSpend)),
		FromZone:                     zone.Battlefield,
		ToZone:                       zone.Stack,

		PlayerEventOrdinalThisTurn: nextSpellCastOrdinalThisTurn(g, playerID),
	})
	obj.ColorsOfManaSpentToCast = distinctManaColorsSpent(paymentResult.PoolSpend)
	obj.ManaSpentByColorToCast = manaSpentByColor(paymentResult.PoolSpend)
	obj.ManaSpentToCast = totalManaSpent(paymentResult.PoolSpend)
	obj.ManaFromCreaturesSpentToCast = creatureManaSpent(paymentResult.PoolSpend)
	resolveSpellCastManaSpendRiders(g, playerID, riderSnapshot, paymentResult.PoolSpend, spellDef, obj)
	return true
}

func preparedSourcePermanent(g *game.Game, sourceID id.ID) *game.Permanent {
	for _, permanent := range g.Battlefield {
		if permanent != nil &&
			(permanent.CardInstanceID == sourceID || permanent.Token && permanent.ObjectID == sourceID) {
			return permanent
		}
	}
	return nil
}

func preparedSpellSource(g *game.Game, permanent *game.Permanent) (id.ID, *game.CardDef, bool) {
	if permanent == nil {
		return 0, nil, false
	}
	if permanent.CardInstanceID != 0 {
		card, ok := g.GetCardInstance(permanent.CardInstanceID)
		if !ok || card.Def.Layout != game.LayoutPrepare || !card.Def.Alternate.Exists {
			return 0, nil, false
		}
		return permanent.CardInstanceID, card.Def, true
	}
	if !permanent.Token || permanent.TokenDef == nil ||
		permanent.TokenDef.Layout != game.LayoutPrepare ||
		!permanent.TokenDef.Alternate.Exists {
		return 0, nil, false
	}
	return permanent.ObjectID, permanent.TokenDef, true
}

func (e *Engine) canCastSpell(g *game.Game, playerID game.PlayerID, cardID id.ID, targets []game.Target, xValue int, chosenModes []int) bool {
	return e.canCastSpellWithKicker(g, playerID, cardID, targets, xValue, chosenModes, false)
}

func (e *Engine) canCastSpellWithKicker(g *game.Game, playerID game.PlayerID, cardID id.ID, targets []game.Target, xValue int, chosenModes []int, kickerPaid bool) bool {
	return e.canCastSpellFaceFromZoneWithKicker(g, playerID, cardID, zone.Hand, game.FaceFront, targets, xValue, chosenModes, kickerPaid)
}

func (e *Engine) canCastSpellFromZoneWithKicker(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, targets []game.Target, xValue int, chosenModes []int, kickerPaid bool) bool {
	return e.canCastSpellFaceFromZoneWithKicker(g, playerID, cardID, sourceZone, game.FaceFront, targets, xValue, chosenModes, kickerPaid)
}

func (e *Engine) canCastSpellFaceFromZoneWithKicker(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, targets []game.Target, xValue int, chosenModes []int, kickerPaid bool) bool {
	return e.canCastSpellFaceFromZoneWithOptions(g, playerID, cardID, sourceZone, face, targets, xValue, chosenModes, effectiveKickerCount(kickerPaid, 0), false, false, false, false, false)
}

// canCastBestowSpellFaceFromZone validates a cast whose Bestow keyword is used
// (CR 702.103): the spell is cast for its Bestow alternative cost as an Aura
// spell, so its targets are validated on the bestowed branch (which requires the
// enchant-creature target) rather than the default branch.
func (e *Engine) canCastBestowSpellFaceFromZone(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, targets []game.Target, xValue int, chosenModes []int) bool {
	return e.canCastSpellFaceFromZoneWithOptions(g, playerID, cardID, sourceZone, face, targets, xValue, chosenModes, 0, false, false, false, true, false)
}

// canCastGiftSpellFaceFromZone validates a cast whose Gift keyword action
// promises a gift (CR 702.171). Promising a gift activates the spell's
// gift-promised target specs, so its targets are validated on the promised
// branch rather than the default branch.
func (e *Engine) canCastGiftSpellFaceFromZone(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, targets []game.Target, xValue int, chosenModes []int) bool {
	return e.canCastSpellFaceFromZoneWithOptions(g, playerID, cardID, sourceZone, face, targets, xValue, chosenModes, 0, false, true, false, false, false)
}

// canCastSpellFaceFromZoneWithMultikick validates a Multikicker cast whose
// kicker cost is paid kickerCount times (CR 702.32).
func (e *Engine) canCastSpellFaceFromZoneWithMultikick(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, targets []game.Target, xValue int, chosenModes []int, kickerCount int) bool {
	return e.canCastSpellFaceFromZoneWithOptions(g, playerID, cardID, sourceZone, face, targets, xValue, chosenModes, kickerCount, false, false, false, false, false)
}

// canCastBargainedSpellFaceFromZone validates a cast whose Bargain additional
// cost is paid (CR 702.166b). Bargaining activates the spell's bargained target
// specs, so its targets are validated on the bargained branch, and the payment
// planner requires the caster to be able to sacrifice an artifact, enchantment,
// or token.
func (e *Engine) canCastBargainedSpellFaceFromZone(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, targets []game.Target, xValue int, chosenModes []int) bool {
	return e.canCastSpellFaceFromZoneWithOptions(g, playerID, cardID, sourceZone, face, targets, xValue, chosenModes, 0, false, false, true, false, false)
}

// canCastOffspringSpellFaceFromZone validates a cast whose Offspring additional
// cost is paid (CR 702.171). Paying the offspring cost activates the spell's
// offspring branch, so its targets are validated on the offspring branch, and
// the payment planner adds the offspring mana cost to the spell's total cost.
func (e *Engine) canCastOffspringSpellFaceFromZone(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, targets []game.Target, xValue int, chosenModes []int) bool {
	return e.canCastSpellFaceFromZoneWithOptions(g, playerID, cardID, sourceZone, face, targets, xValue, chosenModes, 0, false, false, false, false, true)
}

// effectiveKickerCount resolves the number of times the kicker cost is paid from
// the binary kicker flag and the explicit Multikicker count: an explicit count
// wins, otherwise a paid ordinary kicker counts once.
func effectiveKickerCount(kickerPaid bool, kickerCount int) int {
	if kickerCount > 0 {
		return kickerCount
	}
	if kickerPaid {
		return 1
	}
	return 0
}

func (e *Engine) canCastOverloadedSpellFaceFromZone(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, chosenModes []int) bool {
	return e.canCastOverloadedSpellFaceFromZoneWithOptions(g, playerID, cardID, sourceZone, face, 0, chosenModes, false)
}

func (e *Engine) canCastOverloadedSpellFaceFromZoneWithOptions(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, xValue int, chosenModes []int, kickerPaid bool) bool {
	return e.canCastSpellFaceFromZoneWithOptions(g, playerID, cardID, sourceZone, face, nil, xValue, chosenModes, effectiveKickerCount(kickerPaid, 0), true, false, false, false, false)
}

func (*Engine) canCastSpellFaceFromZoneWithOptions(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, targets []game.Target, xValue int, chosenModes []int, kickerCount int, overloaded bool, giftPromised bool, bargained bool, bestowed bool, offspring bool, announced ...game.PlayPermission) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return false
	}
	if kickerCount < 0 {
		return false
	}
	kickerPaid := kickerCount > 0
	branch := game.CastBranch{
		GiftPromised: giftPromised, Kicked: kickerPaid, Bargained: bargained, Bestowed: bestowed, Offspring: offspring,
		XValue: opt.Val(xValue), KickerCount: opt.Val(kickerCount),
	}
	if xValue < 0 {
		return false
	}
	player := g.Players[playerID]
	card, ok := g.GetCardInstance(cardID)
	if !ok || !castSourceContains(castSourcePlayer(g, player, cardID, sourceZone), cardID, sourceZone) {
		return false
	}
	spellDef, ok := cardFaceDef(card, face)
	if !ok || !card.Def.CanChooseCastFace(face) {
		return false
	}
	if face == game.FaceFused && (!canFuseFromZone(playerID, card, sourceZone) || !slices.Equal(chosenModes, game.FusedModes())) {
		return false
	}
	permissions, valid := selectedSpellPermissions(g, playerID, cardID, sourceZone, face, castSelectionFace(g, playerID, spellDef, bestowed), announced)
	if !valid {
		return false
	}
	if castFromZoneProhibited(g, playerID, sourceZone) {
		return false
	}
	switch sourceZone {
	case zone.Command:
		if !isCommanderOwnedBy(g, playerID, cardID) {
			return false
		}
	case zone.Hand:
	case zone.Graveyard:
		if len(permissions) == 0 {
			return false
		}
	case zone.Exile:
		if !exileCastAuthorized(g, playerID, cardID, face) {
			return false
		}
	case zone.Library:
		if !canCastSpellsFromZoneByRuleEffect(g, playerID, cardID, sourceZone, face, bestowed) {
			return false
		}
	default:
		return false
	}
	announcementDef := spellDef
	announcedManaCost := manaCostPtr(spellDef.ManaCost)
	if overloaded {
		if !spellDef.Overload.Exists {
			return false
		}
		announcementDef = overloadSpellDef(spellDef)
		overloadCost := spellDef.Overload.Val.Cost
		announcedManaCost = &overloadCost
	}
	if bestowed && !spellHasBestow(spellDef) {
		return false
	}
	if xValue != 0 &&
		!costHasVariableMana(announcedManaCost) &&
		!additionalCostsUseX(spellDef.AdditionalCosts) {
		return false
	}
	if !modesValidForSpellAtBranch(g, playerID, announcementDef, chosenModes, branch) || !isSupportedSpell(spellDef) ||
		!spellTargetsValidForCast(g, playerID, announcementDef, chosenModes, targets, xValue, kickerCount, branch) {
		return false
	}
	castFace := castSelectionFace(g, playerID, spellDef, bestowed)
	if !canCastAtCurrentTiming(g, playerID, castFace) {
		return false
	}
	if spellCastProhibited(g, playerID, castFace) {
		return false
	}
	if spellCastLimitReached(g, playerID, castFace) {
		return false
	}
	if kickerPaid && !spellHasKicker(spellDef) {
		return false
	}
	if giftPromised && !spellHasGift(spellDef) {
		return false
	}
	if bargained && !spellHasBargain(spellDef) {
		return false
	}
	if offspring && !spellHasOffspring(spellDef) {
		return false
	}
	if kickerCount > 1 && !spellHasMultikicker(spellDef) {
		return false
	}
	request := payment.SpellRequest{
		PlayerID:        playerID,
		CardID:          card.ID,
		SourceZone:      sourceZone,
		Card:            spellDef,
		XValue:          xValue,
		KickerPaid:      kickerPaid,
		KickerCount:     kickerCount,
		Bargained:       bargained,
		Offspring:       offspring,
		ChosenModes:     chosenModes,
		CastPermissions: permissions,
		Targets:         targets,
		Bestowed:        bestowed,
	}
	mode := castCostMode{overloaded: overloaded, bestowed: bestowed}
	routes := spellCastRoutes(g, playerID, card.ID, sourceZone, face, mode)
	return anyCastRouteAvailable(g, spellDef, xValue, mode.castRouteRequests(request, routes))
}

func overloadAlternativeCost(manaCost cost.Mana) cost.Alternative {
	return cost.Alternative{
		Label:    "Overload",
		ManaCost: opt.Val(append(cost.Mana(nil), manaCost...)),
	}
}

// spellHasBestow reports whether a spell face carries the Bestow keyword
// (CR 702.103) so it may be cast for its Bestow alternative cost as an Aura.
func spellHasBestow(spellDef *game.CardDef) bool {
	_, ok := game.CardDefBestow(spellDef)
	return ok
}

// bestowAlternativeCost is the alternative cost of casting a card with Bestow for
// its Bestow cost rather than its mana cost (CR 702.103b). It replaces the mana
// cost with the Bestow keyword's cost; a card reaching this path always has a
// Bestow keyword, so a missing keyword falls back to an empty (free) cost.
func bestowAlternativeCost(spellDef *game.CardDef) cost.Alternative {
	bestow, _ := game.CardDefBestow(spellDef)
	return cost.Alternative{
		Label:    "Bestow",
		ManaCost: opt.Val(append(cost.Mana(nil), bestow.Cost...)),
	}
}

// freeCastAlternativeCost is the alternative cost of a cast that pays no mana
// cost ("without paying its mana cost"): the mana cost is emptied and no
// additional cost replaces it. It backs the plotted cast from exile (CR 718).
func freeCastAlternativeCost() cost.Alternative {
	return cost.Alternative{
		Label:    "Without paying its mana cost",
		ManaCost: opt.Val(cost.Mana{}),
	}
}

// foretellAlternativeCost is the alternative cost of casting a foretold card from
// exile for its foretell cost rather than its mana cost (CR 702.144). It replaces
// the mana cost with the card's Foretell cost; a card reaching this path always
// has a Foretell keyword, so a missing cost falls back to an empty (free) cost.
func foretellAlternativeCost(card *game.CardDef) cost.Alternative {
	foretellCost, _ := card.ForetellCost()
	return cost.Alternative{
		Label:    "Foretell",
		ManaCost: opt.Val(append(cost.Mana(nil), foretellCost...)),
	}
}

// castRoute is one permission through which a card may be cast from exile
// or the top of a library. free marks a permission that casts it without
// paying its mana cost (plot, a linked free cast, or a free play permission);
// foretell marks a cast for its foretell cost; payLife marks a permission that
// replaces the mana cost with paying life equal to the spell's mana value
// (Bolas's Citadel); spendAs is the spending permission riding on that same
// permission (CR 118.14, 609.4b); permission is the authorizing rule effect,
// whose use is recorded once the spell is cast through it so a once-per-turn
// permission spends its use (Evelyn, the Covetous); linkedFree is the one-shot
// linked free-cast permission (Court of Locthwain) consumed only when the spell
// is cast through this route. The zero route is a cast authorized by no such
// rule effect (an Adventure's permanent face). A rider is not an alternative
// cost: it governs paying whichever total cost the caster chooses, so it
// combines with overload, bestow, evoke, kicker, and additional costs.
type castRoute struct {
	free       bool
	foretell   bool
	payLife    bool
	spendAs    game.ManaSpendPermission
	permission game.RuleEffect
	linkedFree game.RuleEffect
}

type castRouteKey struct {
	player game.PlayerID
	card   id.ID
	zone   zone.Type
	face   game.FaceIndex
	mode   castCostMode
}

// zoneCastXValues returns the X candidates legal-action enumeration validates
// for casting face of card from sourceZone, overloaded or not. The candidates
// are sampled from the same permission routes and cost options cast validation
// pays through: the normal cost, each alternative cost the spell offers from
// sourceZone (flashback, escape, evoke), and each route's own alternative cost
// (a free or foretell cast from exile, Bolas's Citadel's life payment) and
// spending rider (Evelyn, the Covetous; Court of Locthwain) (CR 601.2b, 609.4b).
func zoneCastXValues(g *game.Game, playerID game.PlayerID, card *game.CardInstance, sourceZone zone.Type, face game.FaceIndex, spellDef *game.CardDef, overloaded bool) []int {
	announced := manaCostPtr(spellDef.ManaCost)
	if overloaded {
		overloadCost := spellDef.Overload.Val.Cost
		announced = &overloadCost
	}
	if !costHasVariableMana(announced) {
		return legalXValuesForCostAndAdditional(g, playerID, announced, spellDef.AdditionalCosts)
	}
	mode := castCostMode{overloaded: overloaded}
	base := payment.SpellRequest{
		PlayerID:        playerID,
		CardID:          card.ID,
		SourceZone:      sourceZone,
		Card:            spellDef,
		CastPermissions: castPermissionsForZone(g, playerID, card.ID, sourceZone, face),
	}
	return spellCastXValues(g, spellDef, mode.castRouteRequests(base, spellCastRoutes(g, playerID, card.ID, sourceZone, face, mode)))
}

// preparedSpellXValues returns the X candidates legal-action enumeration
// validates for casting the copy of spellDef, the prepare spell of the prepared
// permanent's card sourceID, through the request cast validation pays.
func preparedSpellXValues(g *game.Game, playerID game.PlayerID, sourceID id.ID, spellDef *game.CardDef) []int {
	manaCost := manaCostPtr(spellDef.ManaCost)
	if !costHasVariableMana(manaCost) {
		return legalXValuesForCostAndAdditional(g, playerID, manaCost, spellDef.AdditionalCosts)
	}
	return spellCastXValues(g, spellDef, []payment.SpellRequest{{
		PlayerID:   playerID,
		CardID:     sourceID,
		SourceZone: zone.Battlefield,
		Card:       spellDef,
	}})
}

// spellCastXValues returns the X candidates legal-action enumeration validates
// for casting spellDef, whose announced mana cost has an X, through any of
// requests, the payment requests of the permission routes cast validation may
// pay through. Every candidate is still checked by full cast validation, so the
// candidates only need to include every X that validation accepts: the union,
// over every cost option of every request, of the X values that option could
// pay. An option is any total cost the caster may choose (CR 601.2b, 601.2f):
// the mana cost or an alternative cost, which need not share the printed cost's
// {X} (Firecat Blitz's flashback cost "{R}{R}, Sacrifice X Mountains" bounds X
// by its additional cost, not by mana), or any mana at all. A route whose alternative cost
// has no {X} announces only X=0. Each option's mana at a given X is lowered by
// the most generic mana any applicable cost reduction could remove (CR 601.2f,
// 118.7a), since a reduction lowers the total cost whichever option is chosen,
// and is paid as cast payment pays it: under the request's spending rider, in
// spellDef's spell-cast spending context, so mana restricted to spells of its
// kind ("Spend this mana only to cast a creature spell.", Humble Naturalist)
// counts, and with the convoke, improvise or delve payment the spell may use
// once its total cost is determined (CR 702.51a-b, 702.66a-b, 702.126a-b).
func spellCastXValues(g *game.Game, spellDef *game.CardDef, requests []payment.SpellRequest) []int {
	if len(requests) == 0 {
		return nil
	}
	playerID, cardID, sourceZone := requests[0].PlayerID, requests[0].CardID, requests[0].SourceZone
	reduction, noGeneric, replacesCost := maxSpellGenericCostReduction(g, playerID, cardID, sourceZone, spellDef)
	if noGeneric {
		reduction = math.MaxInt
	}
	candidates := map[int]bool{}
	for _, request := range requests {
		for _, option := range paymentOrch.spellCostOptions(g, request) {
			var optionMana cost.Mana
			if option.ManaCost != nil {
				optionMana = *option.ManaCost
			}
			values := legalXValuesWithManaCheck(g, playerID, costHasVariableMana(option.ManaCost), option.AdditionalCosts, func(x int) bool {
				if replacesCost {
					return true
				}
				lowered := genericReducedManaCost(optionMana, x, reduction)
				return paymentOrch.canPaySpellManaCost(g, payment.SpellManaRequest{
					PlayerID:   playerID,
					CardID:     cardID,
					SourceZone: sourceZone,
					Card:       request.Card,
					SpendAs:    request.SpendAnyMana,
					Bestowed:   request.Bestowed,
					Cost:       &lowered,
				})
			})
			for _, x := range values {
				if castRouteAllowsX(spellDef, x, request) {
					candidates[x] = true
				}
			}
		}
	}
	return slices.Sorted(maps.Keys(candidates))
}

// maxSpellGenericCostReduction bounds the generic mana the spell cost modifiers
// applicable to casting spellDef from sourceZone could remove from its total
// cost (CR 601.2f), whatever targets, Bargain or Bestow choice the cast makes.
// The modifiers are gathered as the payment planner gathers them, once per
// optional cost-determining choice, with every battlefield permanent offered as
// a target so each target-gated reduction ("Spells you cast that target this
// creature cost {2} less to cast.", Elderwood Scion) counts. Increases are
// ignored. noGeneric reports that some modifier replaces the generic amount, so
// no generic mana need be payable at all; replacesCost reports that some
// modifier replaces the whole mana cost, so no printed symbol bounds X and every
// sampled candidate is left to full validation.
func maxSpellGenericCostReduction(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, spellDef *game.CardDef) (reduction int, noGeneric, replacesCost bool) {
	targets := make([]game.Target, 0, len(g.Battlefield))
	for _, permanent := range g.Battlefield {
		targets = append(targets, game.PermanentTarget(permanent.ObjectID))
	}
	state := &rulesPaymentState{g: g}
	for _, bargained := range []bool{false, spellHasBargain(spellDef)} {
		for _, bestowed := range []bool{false, spellHasBestow(spellDef)} {
			total := 0
			for _, modifier := range state.CostModifiersForSpell(playerID, spellDef, cardID, sourceZone, targets, bargained, bestowed) {
				noGeneric = noGeneric || modifier.SetGeneric.Exists
				replacesCost = replacesCost || modifier.SetManaCost.Exists
				total += max(0, modifier.GenericReduction)
			}
			reduction = max(reduction, total)
		}
	}
	return reduction, noGeneric, replacesCost
}

// genericReducedManaCost resolves manaCost's X symbols to xValue and lowers its
// combined generic amount by reduction, never below zero, leaving every other
// symbol unchanged. Generic cost reductions affect only the generic component
// and can't take it below zero (CR 118.7a, 601.2f).
func genericReducedManaCost(manaCost cost.Mana, xValue, reduction int) cost.Mana {
	lowered := make(cost.Mana, 0, len(manaCost)+1)
	generic := 0
	for _, symbol := range manaCost {
		switch symbol.Kind {
		case cost.VariableSymbol:
			generic += xValue
		case cost.GenericSymbol:
			generic += symbol.Generic
		default:
			lowered = append(lowered, symbol)
		}
	}
	if generic > reduction {
		lowered = append(lowered, cost.O(generic-reduction))
	}
	return lowered
}

// castCostMode records the cost a cast already uses before any permission
// route applies: overload, bestow, mutate, or a Bolas's Citadel life payment is
// chosen with the cast. A plotted, foretold, or linked free cast is instead one
// exileCastRoute among the permissions the caster may choose to cast through.
type castCostMode struct {
	mutate     bool
	overloaded bool
	bestowed   bool
}

// otherAlternative reports a cast already using another alternative cost, which
// cannot also be cast without paying its mana cost or for another cost "rather
// than pay its mana cost" (CR 118.9a).
func (m castCostMode) otherAlternative() bool {
	return m.mutate || m.overloaded || m.bestowed
}

// alternative returns the alternative cost a cast in mode m pays through route,
// or none when the spell is cast for its normal mana cost or chooses among its
// own alternative costs.
func (m castCostMode) alternative(spellDef *game.CardDef, xValue int, route *castRoute) opt.V[cost.Alternative] {
	switch {
	case m.overloaded:
		return opt.Val(overloadAlternativeCost(spellDef.Overload.Val.Cost))
	case m.bestowed:
		return opt.Val(bestowAlternativeCost(spellDef))
	case route.payLife:
		return opt.Val(payLifeManaValueAlternativeCost(spellDef, xValue))
	case route.free:
		return opt.Val(freeCastAlternativeCost())
	case route.foretell:
		return opt.Val(foretellAlternativeCost(spellDef))
	default:
		return opt.V[cost.Alternative]{}
	}
}

// castRouteRequests returns one payment request per route, each carrying only
// its own route's alternative cost and spending rider.
func (m castCostMode) castRouteRequests(base payment.SpellRequest, routes []castRoute) []payment.SpellRequest {
	requests := make([]payment.SpellRequest, len(routes))
	for i := range routes {
		route := &routes[i]
		request := base
		request.SpendAnyMana = route.spendAs
		request.Alternative = m.alternative(base.Card, base.XValue, route)
		requests[i] = request
	}
	return requests
}

// castFromZonePermissionRoutes returns one route per active rule-effect
// permission that lets playerID cast the face of cardID from sourceZone: each
// per-card or zone-wide play/cast permission and each matching group
// RuleEffectCastSpellsFromZone permission. Every one is a separate permission
// the caster may choose to cast through, with only its own rider and usage
// (CR 118.14). bestowed selects the characteristics a bestowed cast has as an
// Aura spell (CR 702.103b) for the group permissions' type filters. allowFree
// reports whether the cast may use a permission's own alternative cost. A "without paying its mana cost" permission authorizes
// only that free cast, so when the cast already uses another alternative cost
// (CR 118.9a) it offers no route; excluded then reports that such a matching
// permission was left out.
func castFromZonePermissionRoutes(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, bestowed, allowFree bool) (routes []castRoute, excluded bool) {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if !castFromZoneRuleEffectMatches(g, effect, playerID, cardID, sourceZone, face) {
			continue
		}
		route := castRoute{spendAs: effect.SpendAnyMana, permission: *effect}
		if ruleEffectCastsWithoutPayingManaCost(effect) {
			if !allowFree {
				excluded = true
				continue
			}
			route.free = true
		}
		routes = append(routes, route)
	}
	groupRoutes, groupExcluded := castSpellsFromZoneRoutes(g, playerID, cardID, sourceZone, face, bestowed, allowFree)
	return orderedCastRoutes(append(routes, groupRoutes...)), excluded || groupExcluded
}

// castSpellsFromZoneRoutes returns one route per matching group
// RuleEffectCastSpellsFromZone permission that lets playerID cast the face of
// cardID from sourceZone in the selected cast mode (bestowed or not), in effect
// order. A permission that replaces the mana
// cost with paying life equal to the spell's mana value (Bolas's Citadel)
// imposes that cost only on a cast made through it; it is an alternative cost
// (CR 118.9a), so when allowFree is false (the cast already uses another
// alternative cost) it offers no route and excluded reports it was left out.
func castSpellsFromZoneRoutes(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, bestowed, allowFree bool) (routes []castRoute, excluded bool) {
	for _, effect := range matchingCastSpellsFromZoneEffects(g, playerID, cardID, sourceZone, face, bestowed) {
		if effect.PayLifeEqualToManaValue && !allowFree {
			excluded = true
			continue
		}
		routes = append(routes, castRoute{payLife: effect.PayLifeEqualToManaValue, spendAs: effect.SpendAnyMana, permission: effect})
	}
	return routes, excluded
}

// orderedCastRoutes merges routes that differ only in which permission without
// a usage limit authorizes them, and orders free routes that spend nothing
// first, then a one-shot linked free cast, then paying routes without a usage
// limit, then once-per-turn routes, so the default choice never spends a
// limited use or one-shot permission when another route offers the same option.
func orderedCastRoutes(routes []castRoute) []castRoute {
	rank := func(route castRoute) int {
		switch {
		case route.free && route.linkedFree.ID == 0 && !route.permission.OncePerTurn:
			return 0
		case route.free && !route.permission.OncePerTurn:
			return 1
		case route.permission.OncePerTurn:
			return 3
		default:
			return 2
		}
	}
	slices.SortStableFunc(routes, func(a, b castRoute) int { return rank(a) - rank(b) })
	type routeKey struct {
		free     bool
		foretell bool
		payLife  bool
		spendAs  game.ManaSpendPermission
		limited  id.ID
	}
	seen := make(map[routeKey]bool, len(routes))
	merged := routes[:0]
	for _, route := range routes {
		key := routeKey{free: route.free, foretell: route.foretell, payLife: route.payLife, spendAs: route.spendAs}
		if route.permission.OncePerTurn {
			key.limited = route.permission.SourceObjectID
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		merged = append(merged, route)
	}
	return merged
}

// spellCastRoutes returns the permission routes through which playerID may cast
// the face of cardID from sourceZone in mode. Each permission is a separate
// route the caster may choose (CR 601.2b), carrying only its own cost, rider,
// and usage (CR 118.14), so the routes never mix one permission's free cast
// with another permission's rider or usage, and no permission hides another's
// cost options. A plotted card (CR 702.170d, at sorcery speed) and a linked
// free cast (Court of Locthwain, for any legal spell face) offer a free route,
// and a foretold card (CR 702.143a) a foretell-cost route; these, like a free
// play permission, are alternative costs, so a cast already using another
// alternative cost (CR 118.9a) cannot use them. An Adventure card's CR 715.3d
// permission covers only its permanent face (the front face) and adds a zero
// route. A cast from the top of a library has one route per matching
// cast-from-library permission, so a pay-life permission (Bolas's Citadel)
// never forces its cost on a cast another permission also authorizes; that
// pay-life cost is an alternative cost too, so a cast already using another
// one cannot use it. A cast from another zone has the single zero route; a
// cast from exile or a library with no applicable route has none. A cast from
// exile whose permission this model does not track at all keeps the zero
// route.
// The returned routes are read-only and shared only within a pure-read frame.
func spellCastRoutes(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, mode castCostMode) []castRoute {
	fc := frameCacheFor(g)
	key := castRouteKey{player: playerID, card: cardID, zone: sourceZone, face: face, mode: mode}
	if fc != nil {
		if routes, ok := fc.castRoutes[key]; ok {
			return routes
		}
	}
	routes := buildSpellCastRoutes(g, playerID, cardID, sourceZone, face, mode)
	if fc != nil {
		if fc.castRoutes == nil {
			fc.castRoutes = make(map[castRouteKey][]castRoute)
		}
		fc.castRoutes[key] = routes[:len(routes):len(routes)]
		return fc.castRoutes[key]
	}
	return routes
}

func buildSpellCastRoutes(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, mode castCostMode) []castRoute {
	switch sourceZone {
	case zone.Exile:
		routes, excluded := exilePermissionCastRoutes(g, playerID, cardID, face, mode.bestowed, !mode.otherAlternative())
		if len(routes) == 0 && !excluded {
			return []castRoute{{}}
		}
		return routes
	case zone.Library:
		routes, _ := castSpellsFromZoneRoutes(g, playerID, cardID, zone.Library, face, mode.bestowed, !mode.otherAlternative())
		return orderedCastRoutes(routes)
	default:
		return []castRoute{{}}
	}
}

// exileCastAuthorized reports whether some exile cast route authorizes
// playerID to cast the face of cardID from exile under any cost mode: the
// routes are those spellCastRoutes offers a cast that uses no other
// alternative cost, normally or (for a face with Bestow) bestowed, a superset
// of every mode's routes. Cast validation and
// legal-action face enumeration share it, so a face is listed exactly when a
// permission can authorize it; each concrete cast is still fully validated
// against the routes of its own mode.
func exileCastAuthorized(g *game.Game, playerID game.PlayerID, cardID id.ID, face game.FaceIndex) bool {
	if routes, _ := exilePermissionCastRoutes(g, playerID, cardID, face, false, true); len(routes) > 0 {
		return true
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok || !spellHasBestow(cardFaceOrDefault(card, face)) {
		return false
	}
	routes, _ := exilePermissionCastRoutes(g, playerID, cardID, face, true, true)
	return len(routes) > 0
}

// exilePermissionCastRoutes returns, in spellCastRoutes order, the permission
// routes through which playerID may cast the face of cardID from exile, bestowed
// or not. allowFree is false when the cast already uses another alternative cost, so
// free and foretell routes are left out; excluded then reports that such a
// matching permission was left out. A plotted or foretold card and the
// CR 715.3d Adventure permission authorize only the front face. A linked free
// cast ("cast a spell from among cards exiled with this enchantment", Court of
// Locthwain) authorizes every legal spell face: an adventurer card as its
// Adventure (CR 715.3a, 715.3d) or either half of a split card (CR 709.3).
func exilePermissionCastRoutes(g *game.Game, playerID game.PlayerID, cardID id.ID, face game.FaceIndex, bestowed, allowFree bool) (routes []castRoute, excluded bool) {
	routes, excluded = castFromZonePermissionRoutes(g, playerID, cardID, zone.Exile, face, bestowed, allowFree)
	if face == game.FaceFront {
		if cardIsPlottedInExile(g, playerID, cardID) {
			excluded = true
			if allowFree && isSorcerySpeed(g, playerID) {
				routes = append(routes, castRoute{free: true})
			}
		}
		if cardIsForetoldInExile(g, playerID, cardID) {
			excluded = true
			if allowFree {
				routes = append(routes, castRoute{foretell: true})
			}
		}
		if hasAdventureExilePermission(g, playerID, cardID) {
			routes = append(routes, castRoute{})
		}
	}
	if permission, ok := castLinkedExileForFreePermission(g, playerID, cardID); ok {
		excluded = true
		if allowFree {
			routes = append(routes, castRoute{free: true, linkedFree: permission})
		}
	}
	return orderedCastRoutes(routes), excluded
}

// castRouteAllowsX reports whether request, built for one route, can announce
// xValue for spellDef: a route whose alternative cost has no {X} cannot
// announce a nonzero X.
func castRouteAllowsX(spellDef *game.CardDef, xValue int, request payment.SpellRequest) bool {
	return xValue == 0 || !costHasVariableMana(manaCostPtr(spellDef.ManaCost)) || !request.Alternative.Exists ||
		costHasVariableMana(manaCostPtr(request.Alternative.Val.ManaCost))
}

// anyCastRouteAvailable reports whether any route's request can pay to cast
// spellDef with xValue.
func anyCastRouteAvailable(g *game.Game, spellDef *game.CardDef, xValue int, requests []payment.SpellRequest) bool {
	for _, request := range requests {
		if castRouteAllowsX(spellDef, xValue, request) && paymentOrch.canPaySpellCosts(g, request) {
			return true
		}
	}
	return false
}

// chooseCastRoute lets the caster choose the permission to cast through
// together with that permission's cost option (CR 601.2b, 601.2f, 118.9): every
// payable option of every route is offered in one choice, each route offering
// only its own alternative costs and rider. It returns the chosen route and the
// payment preferences for its option. A single route keeps the ordinary
// spell-cost choice.
func (e *Engine) chooseCastRoute(g *game.Game, routes []castRoute, requests []payment.SpellRequest, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (int, *payment.Preferences) {
	if len(requests) == 1 {
		return 0, e.paymentPreferencesForSpellRequest(g, requests[0], agents, log)
	}
	type routeOption struct {
		route  int
		option payment.SpellOptionSummary
	}
	var options []routeOption
	labels := map[string]int{}
	for i, request := range requests {
		if !castRouteAllowsX(request.Card, request.XValue, request) {
			continue
		}
		for _, option := range paymentOrch.planner(g).PayableSpellOptions(request) {
			options = append(options, routeOption{route: i, option: option})
			labels[option.Label]++
		}
	}
	if len(options) == 0 {
		return 0, e.paymentPreferencesForSpellRequest(g, requests[0], agents, log)
	}
	selected := options[0]
	if len(options) > 1 {
		choiceOptions := make([]game.ChoiceOption, 0, len(options))
		for i, option := range options {
			label := option.option.Label
			if labels[label] > 1 {
				label += " (" + castRouteSourceLabel(g, routes[option.route], option.route) + ")"
			}
			choiceOptions = append(choiceOptions, game.ChoiceOption{Index: i, Label: label})
		}
		chosen := e.chooseChoice(g, agents, game.ChoiceRequest{
			Kind:             game.ChoicePayment,
			Player:           requests[0].PlayerID,
			Prompt:           "Choose spell cost",
			Options:          choiceOptions,
			MinChoices:       1,
			MaxChoices:       1,
			DefaultSelection: []int{0},
		}, log)
		if len(chosen) == 1 && chosen[0] >= 0 && chosen[0] < len(options) {
			selected = options[chosen[0]]
		}
	}
	return selected.route, e.paymentPreferencesForSpellOption(g, requests[selected.route], selected.option, agents, log)
}

// castRouteSourceLabel names the permission a route casts through, to tell
// apart otherwise identical cost options offered by different permissions.
func castRouteSourceLabel(g *game.Game, route castRoute, index int) string {
	source := route.permission.SourceCardID
	if source == 0 {
		source = route.linkedFree.SourceCardID
	}
	if source != 0 {
		if card, ok := g.GetCardInstance(source); ok && card.Def != nil {
			return "via " + card.Def.Name
		}
	}
	return "permission " + strconv.Itoa(index+1)
}

// payLifeManaValueAlternativeCost is the alternative cost imposed when a spell is
// cast from the top of the library under a permission that replaces its mana cost
// with paying life equal to its mana value ("If you cast a spell this way, pay
// life equal to its mana value rather than pay its mana cost.", Bolas's Citadel,
// Gwenom, Remorseless). The mana cost is emptied and a pay-life additional cost
// equal to the cast spell's mana value (counting the announced X) takes its place;
// the spell's own additional costs are appended by the payment planner.
func payLifeManaValueAlternativeCost(spellDef *game.CardDef, xValue int) cost.Alternative {
	alt := cost.Alternative{
		Label:    "Pay life equal to mana value",
		ManaCost: opt.Val(cost.Mana{}),
	}
	if manaValue := stackManaValue(spellDef, xValue); manaValue > 0 {
		alt.AdditionalCosts = []cost.Additional{{
			Kind:   cost.AdditionalPayLife,
			Text:   "pay life equal to its mana value",
			Amount: manaValue,
		}}
	}
	return alt
}

func overloadSpellDef(card *game.CardDef) *game.CardDef {
	overloaded := *card
	overloaded.SpellAbility = opt.Val(card.Overload.Val.SpellAbility)
	return &overloaded
}

func legalCastFacesForZone(g *game.Game, playerID game.PlayerID, card *game.CardInstance, sourceZone zone.Type) []game.FaceIndex {
	if sourceZone == zone.Exile {
		var faces []game.FaceIndex
		for _, face := range card.Def.LegalCastFaces() {
			if exileCastAuthorized(g, playerID, card.ID, face) {
				faces = append(faces, face)
			}
		}
		return faces
	}
	if sourceZone == zone.Graveyard {
		var faces []game.FaceIndex
		for _, face := range card.Def.LegalCastFaces() {
			if canCastFromZoneByRuleEffect(g, playerID, card.ID, sourceZone, face) {
				faces = append(faces, face)
			}
		}
		return faces
	}
	if sourceZone == zone.Library {
		var faces []game.FaceIndex
		for _, face := range card.Def.LegalCastFaces() {
			if castSpellsFromZonePermitsSomeCast(g, playerID, card.ID, sourceZone, face) {
				faces = append(faces, face)
			}
		}
		return faces
	}
	faces := card.Def.LegalCastFaces()
	if canFuseFromZone(playerID, card, sourceZone) {
		faces = append(faces, game.FaceFused)
	}
	return faces
}

// canFuseFromZone reports whether playerID may choose to cast both halves of a
// fuse split card as one fused split spell. Fuse functions only while the card
// is in its caster's own hand (CR 702.102a), so the choice is never offered
// from another zone or another player's hand, even with a cast permission.
func canFuseFromZone(playerID game.PlayerID, card *game.CardInstance, sourceZone zone.Type) bool {
	return sourceZone == zone.Hand && card.Owner == playerID && card.Def.CanFuse()
}

// hasAdventureExilePermission reports whether playerID holds the CR 715.3d
// permission to play cardID: playerID controlled the Adventure spell as it
// resolved into exile, and that same exiled object (same incarnation, in its
// owner's exile) is still there. The owner gains nothing merely by owning it.
func hasAdventureExilePermission(g *game.Game, playerID game.PlayerID, cardID id.ID) bool {
	adventure, ok := g.AdventureCards[cardID]
	if !ok || adventure.Player != playerID {
		return false
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok || card.ZoneVersion != adventure.ZoneVersion {
		return false
	}
	owner, ok := playerByID(g, card.Owner)
	return ok && owner.Exile.Contains(cardID)
}

// castSourcePlayer returns the player whose zone physically holds the card being
// played or cast from sourceZone. A player plays or casts from their own hand,
// graveyard, library, or command zone, but exile is a shared zone: a card exiled
// from another player's library — e.g. a card Court of Locthwain exiled from an
// opponent's library — rests in its owner's exile bucket even though a different
// player is permitted to play or cast it. Resolving the owner keeps the
// containment, removal, and restore steps pointed at the bucket that actually
// holds the card. For every other card the owner is the actor, so this changes
// nothing.
func castSourcePlayer(g *game.Game, caster *game.Player, cardID id.ID, sourceZone zone.Type) *game.Player {
	if sourceZone != zone.Exile {
		return caster
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return caster
	}
	if owner := g.Players[card.Owner]; owner != nil {
		return owner
	}
	return caster
}

// foreignExileCastableCards returns the cards resting in another player's exile
// bucket that playerID currently holds an active permission to play or cast. A
// card Court of Locthwain exiled from an opponent's library lives in that
// opponent's exile bucket even though the controller may play it (for as long as
// it remains exiled) or, while monarch, free-cast it from the accumulated pool.
// The action enumerators scan the acting player's own exile bucket directly, so
// this surfaces only the cross-player cards those scans would otherwise miss.
//
// It gathers the affected card of every active RuleEffectPlayFromZone /
// RuleEffectCastFromZone whose exile permission reaches playerID, plus the
// linked-exile pool of every active RuleEffectCastLinkedExileForFree the player
// controls (the monarch-gated free cast — the effect only exists while the gate
// is satisfied), plus every exiled card whose CR 715.3d Adventure permission
// belongs to playerID (for example a stolen Adventure owned by another player),
// in card-ID order, plus, when a group RuleEffectPlayLandsFromZone /
// RuleEffectCastSpellsFromZone exile permission reaches playerID ("you may play
// a card from exile with a collection counter on it", Evelyn, the Covetous),
// every card in another player's exile that permission currently covers for
// some face, in card-ID order. Group matching reuses the validation matchers,
// so their exile-counter, provenance, ownership and once-per-turn filters apply.
// Cards already in playerID's own exile are omitted, cards no longer in exile
// are skipped, and each card is returned at most once. The result is empty for
// any player who holds no such cross-player exile permission, so it adds
// nothing to enumeration for every other card.
func foreignExileCastableCards(g *game.Game, playerID game.PlayerID) []id.ID {
	own, ok := playerByID(g, playerID)
	if !ok {
		return nil
	}
	var cards []id.ID
	seen := make(map[id.ID]struct{})
	consider := func(cardID id.ID) {
		if cardID == 0 || own.Exile.Contains(cardID) {
			return
		}
		if _, dup := seen[cardID]; dup {
			return
		}
		if z, ok := cardZone(g, cardID); !ok || z != zone.Exile {
			return
		}
		seen[cardID] = struct{}{}
		cards = append(cards, cardID)
	}
	groupExilePermission := false
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if !playerRelationMatches(effect.Controller, playerID, effect.AffectedPlayer) {
			continue
		}
		switch effect.Kind {
		case game.RuleEffectPlayFromZone, game.RuleEffectCastFromZone:
			if effect.CastFromZone == zone.Exile {
				consider(effect.AffectedCardID)
			}
		case game.RuleEffectPlayLandsFromZone, game.RuleEffectCastSpellsFromZone:
			if effect.CastFromZone == zone.Exile {
				groupExilePermission = true
			}
		case game.RuleEffectCastLinkedExileForFree:
			if effect.ExiledLinkKey == "" {
				continue
			}
			key := game.LinkedObjectKey{SourceID: effect.SourceCardID, LinkID: string(effect.ExiledLinkKey)}
			for _, ref := range linkedObjects(g, key) {
				consider(ref.CardID)
			}
		default:
			continue
		}
	}
	var adventures []id.ID
	for cardID := range g.AdventureCards {
		if hasAdventureExilePermission(g, playerID, cardID) {
			adventures = append(adventures, cardID)
		}
	}
	slices.Sort(adventures)
	for _, cardID := range adventures {
		consider(cardID)
	}
	if groupExilePermission {
		for _, cardID := range foreignExileGroupPermittedCards(g, own) {
			consider(cardID)
		}
	}
	return cards
}

// foreignExileGroupPermittedCards returns, in card-ID order, the cards in other
// players' exile buckets that a continuous group play-lands or cast-spells
// permission lets own's player play or cast as some face right now.
func foreignExileGroupPermittedCards(g *game.Game, own *game.Player) []id.ID {
	var cards []id.ID
	for _, other := range g.Players {
		if other == nil || other.ID == own.ID {
			continue
		}
		for _, cardID := range other.Exile.All() {
			card, ok := g.GetCardInstance(cardID)
			if !ok {
				continue
			}
			if canPlayLandFromZoneByRuleEffect(g, own.ID, cardID, zone.Exile) ||
				slices.ContainsFunc(card.Def.LegalCastFaces(), func(face game.FaceIndex) bool {
					return castSpellsFromZonePermitsSomeCast(g, own.ID, cardID, zone.Exile, face)
				}) {
				cards = append(cards, cardID)
			}
		}
	}
	slices.Sort(cards)
	return cards
}

func castSourceContains(player *game.Player, cardID id.ID, sourceZone zone.Type) bool {
	switch sourceZone {
	case zone.Hand:
		return player.Hand.Contains(cardID)
	case zone.Command:
		return player.CommandZone.Contains(cardID)
	case zone.Graveyard:
		return player.Graveyard.Contains(cardID)
	case zone.Exile:
		return player.Exile.Contains(cardID)
	case zone.Library:
		top, ok := player.Library.Top()
		return ok && top == cardID
	default:
		return false
	}
}

func castSourceZoneCards(player *game.Player, sourceZone zone.Type) []id.ID {
	switch sourceZone {
	case zone.Hand:
		return player.Hand.All()
	case zone.Command:
		return player.CommandZone.All()
	case zone.Graveyard:
		return player.Graveyard.All()
	case zone.Exile:
		return player.Exile.All()
	case zone.Library:
		if topID, ok := player.Library.Top(); ok {
			return []id.ID{topID}
		}
		return nil
	default:
		return nil
	}
}

func removeCastSourceCard(g *game.Game, player *game.Player, cardID id.ID, sourceZone zone.Type) bool {
	switch sourceZone {
	case zone.Hand:
		return player.Hand.Remove(cardID)
	case zone.Command:
		return player.CommandZone.Remove(cardID)
	case zone.Graveyard:
		return player.Graveyard.Remove(cardID)
	case zone.Exile:
		return player.Exile.Remove(cardID)
	case zone.Library:
		return player.Library.Remove(cardID)
	default:
		return false
	}
}

// restoreCastSourceCard returns a card to the zone it was cast from. It reverses
// removeCastSourceCard when a proposed cast is abandoned because its costs can't
// be paid (CR 601.2h / 728, handling an illegal action by undoing the
// proposal), so the card is never lost.
func restoreCastSourceCard(player *game.Player, cardID id.ID, sourceZone zone.Type) {
	switch sourceZone {
	case zone.Hand:
		player.Hand.Add(cardID)
	case zone.Command:
		player.CommandZone.Add(cardID)
	case zone.Graveyard:
		player.Graveyard.Add(cardID)
	case zone.Exile:
		player.Exile.Add(cardID)
	case zone.Library:
		player.Library.Add(cardID)
	default:
	}
}
