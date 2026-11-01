package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

// castPaidTargetedSpell casts a specific targeted card from fromZone for
// controllerID during the resolution of an ability, paying the card's normal
// mana and additional costs. It locates whichever player's zone currently holds
// the card (like castFreeTargetedSpell), so a card targeted in any player's zone
// is cast under controllerID's control. Unlike a normal cast, priority and
// sorcery-speed timing are ignored because the cast happens mid-resolution, but
// printed cast restrictions, cast prohibitions, and per-turn cast limits still
// apply. It returns false (casting nothing) when the card is no longer in a
// player's fromZone, the cast is restricted or prohibited, the spell has no
// legal cast choice, or its costs are not paid in full.
func (e *Engine) castPaidTargetedSpell(g *game.Game, controllerID game.PlayerID, cardID id.ID, fromZone zone.Type, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	source, ok := playerHoldingCastSource(g, cardID, fromZone)
	if !ok {
		return false
	}
	return e.castPaidSpellFromSource(g, source, controllerID, cardID, fromZone, agents, log)
}

// castPaidSpellFromSource casts cardID out of sourcePlayer's fromZone for
// controllerID during resolution, paying its normal mana and additional costs
// and pushing the spell to the stack under controllerID's control. It mirrors
// castFreeSpellFromSource but runs the spell-cost payment step (CR 601.2f-h)
// with the standard rollback on an unpaid or declined cost (CR 733): the spell
// is taken back off the stack and the card returned to its source zone. The
// source player and controller differ only for a targeted cast from another
// player's zone.
func (e *Engine) castPaidSpellFromSource(g *game.Game, sourcePlayer *game.Player, controllerID game.PlayerID, cardID id.ID, fromZone zone.Type, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if sourcePlayer == nil {
		return false
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	searchCast := fromZone == zone.Library && sourcePlayer == g.Players[controllerID] && searchCastCardAllowed(g, controllerID, card)
	if !searchCast && !castSourceContains(sourcePlayer, cardID, fromZone) {
		return false
	}
	spellDef := cardFaceOrDefault(card, game.FaceFront)
	// The cast ignores priority and sorcery-speed timing but still obeys printed
	// cast restrictions, cast prohibitions, and per-turn cast limits (CR 601.3e).
	castFace := castSelectionFace(g, controllerID, spellDef, false)
	if castFromZoneProhibited(g, controllerID, fromZone) || !cardCastRestrictionsSatisfied(g, controllerID, spellDef) ||
		spellCastProhibited(g, controllerID, castFace) ||
		spellCastLimitReached(g, controllerID, castFace) || !isSupportedSpell(spellDef) {
		return false
	}
	cast, ok := e.chooseCastAnnouncements(g, controllerID, spellDef, action.CastSpellAction{
		CardID: cardID, SourceZone: fromZone, Face: game.FaceFront,
	}, false, agents, log)
	if !ok {
		g.Execution.InvalidChoices++
		return false
	}
	request := payment.SpellRequest{
		PlayerID: controllerID, CardID: cardID, SourceZone: fromZone, Card: spellDef,
		XValue: cast.XValue, KickerPaid: cast.KickerPaid, KickerCount: cast.KickerCount,
		ChosenModes:     cast.ChosenModes,
		CastPermissions: castPermissionsForZone(g, controllerID, cardID, fromZone, game.FaceFront),
	}
	// CR 601.2b-c: the cost is proposed, with its Phyrexian payments announced,
	// before the spell's targets are chosen.
	proposal, ok := e.proposeSpellCost(g, request, true, agents, log)
	if !ok {
		return false
	}
	cast, ok = e.chooseCastTargets(g, controllerID, spellDef, cast, agents, log)
	if !ok {
		g.Execution.InvalidChoices++
		return false
	}
	modes, targets := cast.ChosenModes, cast.Targets
	targetCounts, ok := spellTargetCounts(g, controllerID, spellDef, modes, targets, castBranchForCast(cast))
	if !ok {
		panic("validated paid-cast spell targets could not be segmented")
	}
	request.Targets = targets
	if _, feasible := e.lockSpellCostProposal(g, request, proposal); !feasible {
		return false
	}
	// CR 601.2a: proposing the cast moves the card to the stack before its costs
	// are determined and paid, so a cost paid from the source zone can never
	// select the very card being cast.
	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackSpell,
		SourceID:     cardID,
		Face:         game.FaceFront,
		Controller:   controllerID,
		Targets:      append([]game.Target(nil), targets...),
		TargetCounts: targetCounts,
		ChosenModes:  append([]int(nil), modes...),
		XValue:       cast.XValue,
		KickerPaid:   cast.KickerPaid,
		KickerCount:  cast.KickerCount,
		SourceZone:   fromZone,
	}
	counterTransition := prepareCastCounterTransition(g, card, fromZone)
	libraryBefore := sourcePlayer.Library.All()
	restoreSource := func() {
		if fromZone != zone.Library {
			restoreCastSourceCard(sourcePlayer, cardID, fromZone)
			return
		}
		// Costs can remove other cards from the library. Restore before the next
		// surviving original neighbour rather than moving the proposal to the top.
		for _, next := range libraryBefore[slices.Index(libraryBefore, cardID)+1:] {
			if index := slices.Index(sourcePlayer.Library.All(), next); index >= 0 {
				sourcePlayer.Library.InsertAt(index, cardID)
				return
			}
		}
		sourcePlayer.Library.AddToBottom(cardID)
	}
	if !removeCastSourceCard(g, sourcePlayer, cardID, fromZone) {
		return false
	}
	counterTransition.apply(g)
	g.Stack.Push(obj)
	targetEvents := prepareTargetEvents(g, obj)
	prepared, ok := e.lockSpellCostProposal(g, request, proposal)
	if !ok {
		g.Stack.RemoveByID(obj.ID)
		restoreSource()
		counterTransition.undo(g)
		return false
	}
	request.PreparedCost = &prepared
	// CR 601.2g: mana abilities may be activated while paying only when the
	// locked total, with its announced Phyrexian payments, includes mana.
	prefs, undoMana := e.preparedSpellPaymentPreferences(g, request, prepared, proposal, agents, log)
	request.Prefs = prefs
	riderSnapshot, _ := manaSpendRiderSnapshot(g, controllerID)
	paymentResult, ok := paymentOrch.paySpellCosts(g, request)
	if !ok {
		// CR 733: the proposed cast is illegal because its costs can't or won't be
		// paid, so the proposal is undone.
		undoMana()
		g.Stack.RemoveByID(obj.ID)
		restoreSource()
		counterTransition.undo(g)
		return false
	}
	obj.BuybackPaid = paymentResult.BuybackPaid
	obj.AdditionalCostsPaid = paymentResult.AdditionalCostsPaid
	obj.SacrificedAsCostIDs = paymentResult.SacrificedIDs
	obj.PaidCostSubjects = paymentResult.PaidCostSubjects
	obj.ColorsOfManaSpentToCast = distinctManaColorsSpent(paymentResult.PoolSpend)
	obj.ManaSpentByColorToCast = manaSpentByColor(paymentResult.PoolSpend)
	obj.ManaSpentToCast = totalManaSpent(paymentResult.PoolSpend)
	obj.ManaFromCreaturesSpentToCast = creatureManaSpent(paymentResult.PoolSpend)
	typeLine := stackSpellTypeLine(g, obj, spellDef)
	emitSpellCastEventsWithTargets(g, obj, game.Event{
		SourceID:                     cardID,
		StackObjectID:                obj.ID,
		Controller:                   controllerID,
		CardID:                       cardID,
		Face:                         game.FaceFront,
		CardTypes:                    typeLine.types,
		CardSupertypes:               typeLine.supertypes,
		CardSubtypes:                 typeLine.subtypes,
		Colors:                       spellColors(spellDef),
		ManaValue:                    stackObjectKnownManaValue(g, obj),
		ManaSpentToCast:              opt.Val(totalManaSpent(paymentResult.PoolSpend)),
		ManaFromCreaturesSpentToCast: opt.Val(creatureManaSpent(paymentResult.PoolSpend)),
		FromZone:                     fromZone,
		ToZone:                       zone.Stack,
	}, targetEvents)
	resolveSpellCastManaSpendRiders(g, controllerID, riderSnapshot, paymentResult.PoolSpend, spellDef, obj)
	return true
}
