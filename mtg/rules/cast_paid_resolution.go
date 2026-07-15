package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

// castPaidTargetedSpell casts a specific targeted card from fromZone for
// controllerID during the resolution of an ability, paying the card's normal
// mana and additional costs. It locates whichever player's zone currently holds
// the card (like castFreeTargetedSpell), so a card targeted in any player's zone
// is cast under controllerID's control. Unlike a normal cast the timing checks
// are skipped — the cast happens mid-resolution ignoring priority and
// sorcery-speed timing — but the spell still obeys cast prohibitions and
// per-turn cast limits. It returns false (casting nothing) when the card is no
// longer in a player's fromZone, the cast is prohibited, the spell has no legal
// cast choice, or its costs are not paid in full.
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
// with the standard rollback on an unpaid or declined cost (CR 728): the spell
// is taken back off the stack and the card returned to its source zone. The
// source player and controller differ only for a targeted cast from another
// player's zone.
func (e *Engine) castPaidSpellFromSource(g *game.Game, sourcePlayer *game.Player, controllerID game.PlayerID, cardID id.ID, fromZone zone.Type, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if sourcePlayer == nil || !castSourceContains(sourcePlayer, cardID, fromZone) {
		return false
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	spellDef := cardFaceOrDefault(card, game.FaceFront)
	// The cast ignores timing but still obeys cast prohibitions and per-turn cast
	// limits (CR 601.3e): a "can't cast" restriction or a reached cast limit
	// forbids it even though priority and sorcery-speed timing are bypassed.
	if spellCastProhibited(g, controllerID, spellDef) || spellCastLimitReached(g, controllerID, spellDef) {
		return false
	}
	modes, targets, ok := firstLegalSpellCastChoice(g, controllerID, spellDef)
	if !ok {
		return false
	}
	targetCounts, ok := spellTargetCounts(g, controllerID, spellDef, modes, targets, game.CastBranch{})
	if !ok {
		panic("validated paid-cast spell targets could not be segmented")
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
		SourceZone:   fromZone,
	}
	if !removeCastSourceCard(g, sourcePlayer, cardID, fromZone) {
		return false
	}
	g.Stack.Push(obj)
	prefs := e.paymentPreferencesForSpellFromZone(g, controllerID, cardID, fromZone, game.FaceFront, spellDef, 0, agents, log)
	riderSnapshot, _ := manaSpendRiderSnapshot(g, controllerID)
	paymentResult, ok := paymentOrch.paySpellCosts(g, payment.SpellRequest{
		PlayerID:        controllerID,
		CardID:          cardID,
		SourceZone:      fromZone,
		Card:            spellDef,
		ChosenModes:     modes,
		CastPermissions: castPermissionsForZone(g, controllerID, cardID, fromZone, game.FaceFront),
		Targets:         targets,
		Prefs:           prefs,
	})
	if !ok {
		// CR 728: the proposed cast is illegal because its costs can't or won't be
		// paid, so the proposal is undone.
		g.Stack.RemoveByID(obj.ID)
		restoreCastSourceCard(sourcePlayer, cardID, fromZone)
		return false
	}
	obj.AdditionalCostsPaid = paymentResult.AdditionalCostsPaid
	obj.SacrificedAsCostIDs = paymentResult.SacrificedIDs
	obj.ColorsOfManaSpentToCast = distinctManaColorsSpent(paymentResult.PoolSpend)
	obj.ManaSpentByColorToCast = manaSpentByColor(paymentResult.PoolSpend)
	obj.ManaSpentToCast = totalManaSpent(paymentResult.PoolSpend)
	// stormCopyCount must be read before the spell-cast event is emitted, since
	// that event increments the storm count for later spells this turn.
	stormCopies := stormCopyCount(g, spellDef)
	emitSpellCastEvents(g, obj, game.Event{
		SourceID:        cardID,
		StackObjectID:   obj.ID,
		Controller:      controllerID,
		CardID:          cardID,
		Face:            game.FaceFront,
		CardTypes:       stackObjectCardTypes(obj, spellDef),
		CardSupertypes:  cardSupertypes(spellDef),
		CardSubtypes:    stackObjectCardSubtypes(obj, spellDef),
		Colors:          spellColors(spellDef),
		ManaValue:       opt.Val(stackManaValue(spellDef, 0)),
		ManaSpentToCast: opt.Val(totalManaSpent(paymentResult.PoolSpend)),
		FromZone:        fromZone,
		ToZone:          zone.Stack,
	})
	createStormCopies(g, obj, spellDef, stormCopies)
	resolveSpellCastManaSpendRiders(g, controllerID, riderSnapshot, paymentResult.PoolSpend, spellDef, obj)
	e.resolveCascadeForCast(g, obj, spellDef, agents, log)
	return true
}
