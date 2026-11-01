package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

func madnessCostForCard(card *game.CardDef) (cost.Mana, bool) {
	if card == nil {
		return nil, false
	}
	return card.MadnessCost()
}

func (e *Engine) resolveMadnessTriggeredAbilityWithChoices(g *game.Game, obj *game.StackObject, ability *game.TriggeredAbility, agents [game.NumPlayers]PlayerAgent, log *TurnLog) string {
	cardID := obj.SourceID
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return "missing source"
	}
	player, ok := playerByID(g, obj.Controller)
	if !ok || !player.Exile.Contains(cardID) {
		return "resolved"
	}
	if !e.chooseMay(g, agents, obj.Controller, "Cast card for its madness cost?", log) {
		moveExiledCardToGraveyard(g, obj.Controller, cardID)
		return "declined"
	}
	manaCost, ok := game.BodyMadnessCost(ability)
	if !ok || !e.castMadnessSpellWithChoices(g, obj.Controller, card, manaCost, agents, log) {
		moveExiledCardToGraveyard(g, obj.Controller, cardID)
		return "resolved"
	}
	return "resolved"
}

// castMadnessSpellWithChoices casts card from exile for its madness cost, an
// alternative cost (CR 702.35a-b). The card moves to the stack first (CR 601.2a),
// then the total cost is determined from the madness cost through the shared
// spell-cost planner (CR 601.2f), so cost increases and reductions that match the
// spell, and the spell's additional costs, apply before the cost is paid
// (CR 601.2h). The cast is proposed when that total is payable now or after the
// caster activates payment-time mana abilities; the total is locked before those
// abilities are activated (CR 601.2g). An unpaid cost undoes the proposal and
// its reversible mana activations (CR 733) and leaves the card in exile for the
// caller to move to the graveyard.
func (e *Engine) castMadnessSpellWithChoices(g *game.Game, playerID game.PlayerID, card *game.CardInstance, manaCost cost.Mana, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	player, ok := playerByID(g, playerID)
	if !ok || !player.Exile.Contains(card.ID) {
		return false
	}
	spellDef := cardFaceOrDefault(card, game.FaceFront)
	if !cardCastRestrictionsSatisfied(g, playerID, spellDef) {
		return false
	}
	modes, targets, ok := firstLegalSpellCastChoice(g, playerID, spellDef)
	if !ok {
		return false
	}
	targetCounts, ok := spellTargetCounts(g, playerID, spellDef, modes, targets, game.CastBranch{})
	if !ok {
		panic("validated madness spell targets could not be segmented")
	}
	request := payment.SpellRequest{
		PlayerID:        playerID,
		CardID:          card.ID,
		SourceZone:      zone.Exile,
		Card:            spellDef,
		ChosenModes:     modes,
		Targets:         targets,
		CastPermissions: []payment.SpellCastPermission{payment.SpellCastPermissionDefault},
		Alternative:     opt.Val(cost.Alternative{Label: "Madness", ManaCost: opt.Val(manaCost)}),
	}
	// CR 601.2b: the madness cost is proposed, with its Phyrexian payments
	// announced, before the total is locked and paid.
	proposal, ok := e.proposeSpellCost(g, request, false, agents, log)
	if !ok {
		return false
	}
	stackObj := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackSpell,
		SourceID:     card.ID,
		Face:         game.FaceFront,
		Controller:   playerID,
		Targets:      append([]game.Target(nil), targets...),
		TargetCounts: targetCounts,
		ChosenModes:  append([]int(nil), modes...),
		SourceZone:   zone.Exile,
	}
	counterTransition := prepareCastCounterTransition(g, card, zone.Exile)
	if !player.Exile.Remove(card.ID) {
		return false
	}
	counterTransition.apply(g)
	g.Stack.Push(stackObj)
	rollback := func() {
		g.Stack.RemoveByID(stackObj.ID)
		restoreCastSourceCard(player, card.ID, zone.Exile)
		counterTransition.undo(g)
	}
	targetEvents := prepareTargetEvents(g, stackObj)
	prepared, ok := e.lockSpellCostProposal(g, request, proposal)
	if !ok {
		rollback()
		return false
	}
	request.PreparedCost = &prepared
	prefs, undoMana := e.preparedSpellPaymentPreferences(g, request, prepared, proposal, agents, log)
	request.Prefs = prefs
	riderSnapshot, _ := manaSpendRiderSnapshot(g, playerID)
	paymentResult, ok := paymentOrch.paySpellCosts(g, request)
	if !ok {
		undoMana()
		rollback()
		return false
	}
	poolSpent := paymentResult.PoolSpend
	stackObj.AdditionalCostsPaid = paymentResult.AdditionalCostsPaid
	stackObj.SacrificedAsCostIDs = paymentResult.SacrificedIDs
	stackObj.PaidCostSubjects = paymentResult.PaidCostSubjects
	stackObj.ColorsOfManaSpentToCast = distinctManaColorsSpent(poolSpent)
	stackObj.ManaSpentByColorToCast = manaSpentByColor(poolSpent)
	stackObj.ManaSpentToCast = totalManaSpent(poolSpent)
	stackObj.ManaFromCreaturesSpentToCast = creatureManaSpent(poolSpent)
	typeLine := stackSpellTypeLine(g, stackObj, spellDef)
	emitSpellCastEventsWithTargets(g, stackObj, game.Event{
		SourceID:                     card.ID,
		StackObjectID:                stackObj.ID,
		Controller:                   playerID,
		CardID:                       card.ID,
		Face:                         game.FaceFront,
		CardTypes:                    typeLine.types,
		CardSupertypes:               typeLine.supertypes,
		CardSubtypes:                 typeLine.subtypes,
		Colors:                       spellColors(spellDef),
		ManaValue:                    stackObjectKnownManaValue(g, stackObj),
		ManaSpentToCast:              opt.Val(totalManaSpent(poolSpent)),
		ManaFromCreaturesSpentToCast: opt.Val(creatureManaSpent(poolSpent)),
		FromZone:                     zone.Exile,
		ToZone:                       zone.Stack,
	}, targetEvents)
	resolveSpellCastManaSpendRiders(g, playerID, riderSnapshot, poolSpent, spellDef, stackObj)
	return true
}

func firstLegalSpellCastChoice(g *game.Game, playerID game.PlayerID, spellDef *game.CardDef) ([]int, []game.Target, bool) {
	if !isSupportedSpell(spellDef) {
		return nil, nil, false
	}
	for _, modes := range modeChoicesForSpellAt(g, playerID, spellDef) {
		targetResult := targetChoicesForSpell(g, playerID, spellDef, modes, game.CastBranch{})
		if targetResult.kind == targetInvalidSpec {
			continue
		}
		for _, targets := range targetResult.choices {
			if modesValidForSpellAt(g, playerID, spellDef, modes) && targetsValidForSpell(g, playerID, spellDef, modes, targets, game.CastBranch{}) {
				return append([]int(nil), modes...), append([]game.Target(nil), targets...), true
			}
		}
	}
	return nil, nil, false
}

func moveExiledCardToGraveyard(g *game.Game, playerID game.PlayerID, cardID id.ID) bool {
	return moveCardBetweenZonesInBatch(g, playerID, cardID, zone.Exile, zone.Graveyard, false, 0)
}
