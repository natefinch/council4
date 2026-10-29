package rules

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

type freeCastOption struct {
	cardID id.ID
	cast   action.CastSpellAction
	// payable reports whether the proposal's costs can be paid from current
	// resources without first activating a payment-time mana ability.
	payable bool
}

// freeCastOptionsForCard lists the legal proposals to cast card without paying
// its mana cost. A proposal whose remaining costs are not currently payable is
// still offered when they could be paid after activating payment-time mana
// abilities (CR 601.2g); an unpaid proposal is then reversed (CR 733).
func (e *Engine) freeCastOptionsForCard(g *game.Game, playerID game.PlayerID, card *game.CardInstance, fromZone zone.Type, selection game.Selection, maxManaValue opt.V[int]) []freeCastOption {
	if card == nil || card.Def == nil || castFromZoneProhibited(g, playerID, fromZone) {
		return nil
	}
	var options []freeCastOption
	manualMana := e.hasManualPaymentManaAbility(g, playerID)
	for _, face := range freeCastFaces(playerID, card, fromZone) {
		spellDef := cardFaceOrDefault(card, face)
		if !freeCastFaceAvailable(g, playerID, card, spellDef, selection, maxManaValue) {
			continue
		}
		for _, cast := range freeCastBranchOptions(g, playerID, card.ID, fromZone, face, spellDef) {
			branch := castBranchForCast(cast)
			for _, modes := range modeChoicesForSpellAtBranch(g, playerID, spellDef, branch) {
				result := targetChoicesForSpellWithKickerCount(g, playerID, spellDef, modes, branch, cast.KickerCount)
				if result.kind == targetInvalidSpec || result.kind == targetNoLegalChoices {
					continue
				}
				for _, targets := range result.choices {
					candidate := cast
					candidate.ChosenModes = append([]int(nil), modes...)
					candidate.Targets = append([]game.Target(nil), targets...)
					if !freeCastProposalLegal(g, playerID, spellDef, candidate) {
						continue
					}
					payable, feasible := spellCostFeasibility(g, freeCastCostRequest(playerID, spellDef, candidate), manualMana)
					if feasible {
						options = append(options, freeCastOption{cardID: card.ID, cast: candidate, payable: payable})
					}
				}
			}
		}
	}
	return options
}

// freeCastFaces lists the faces a resolving effect may let playerID cast
// without paying its mana cost. A fuse card cast from its caster's own hand may
// be cast fused even without paying its mana cost (CR 702.102a; Wear // Tear
// ruling 2013-04-15), and the fused spell's combined mana value is what any
// mana value limit sees (CR 709.4d). From any other zone only the halves are
// offered.
func freeCastFaces(playerID game.PlayerID, card *game.CardInstance, fromZone zone.Type) []game.FaceIndex {
	faces := card.Def.LegalCastFaces()
	if canFuseFromZone(playerID, card, fromZone) {
		faces = append(faces, game.FaceFused)
	}
	return faces
}

func freeCastFaceAvailable(g *game.Game, player game.PlayerID, card *game.CardInstance, def *game.CardDef, selection game.Selection, maxManaValue opt.V[int]) bool {
	castFace := castSelectionFace(g, player, def, false)
	return freeCastFaceMatchesSelection(g, card, def, selection, player) &&
		(!maxManaValue.Exists || stackManaValue(def, 0) <= maxManaValue.Val) &&
		cardCastRestrictionsSatisfied(g, player, def) && !spellCastProhibited(g, player, castFace) &&
		!spellCastLimitReached(g, player, castFace) && isSupportedSpell(def)
}

func freeCastFaceMatchesSelection(g *game.Game, card *game.CardInstance, spellDef *game.CardDef, selection game.Selection, viewer game.PlayerID) bool {
	faceCard := *card
	faceCard.Def = spellDef
	return handCardMatchesSelection(g, &faceCard, selection, viewer)
}

func freeCastBranchOptions(g *game.Game, playerID game.PlayerID, cardID id.ID, fromZone zone.Type, face game.FaceIndex, spellDef *game.CardDef) []action.CastSpellAction {
	base := []action.CastSpellAction{{
		CardID:     cardID,
		SourceZone: fromZone,
		Face:       face,
	}}
	if spellHasKicker(spellDef) {
		base = append(base, action.CastSpellAction{
			CardID:      cardID,
			SourceZone:  fromZone,
			Face:        face,
			KickerPaid:  true,
			KickerCount: 1,
		})
	}
	base = expandFreeCastBooleanBranch(base, spellHasBargain(spellDef), func(cast *action.CastSpellAction) {
		cast.Bargained = true
	})
	base = expandFreeCastBooleanBranch(base, spellHasOffspring(spellDef), func(cast *action.CastSpellAction) {
		cast.Offspring = true
	})
	if !spellHasGift(spellDef) {
		return base
	}
	withGift := append([]action.CastSpellAction(nil), base...)
	for _, cast := range base {
		for _, opponent := range aliveOpponents(g, playerID) {
			gift := cast
			gift.GiftPromised = true
			gift.GiftRecipient = opponent
			withGift = append(withGift, gift)
		}
	}
	return withGift
}

func expandFreeCastBooleanBranch(base []action.CastSpellAction, enabled bool, apply func(*action.CastSpellAction)) []action.CastSpellAction {
	if !enabled {
		return base
	}
	expanded := append([]action.CastSpellAction(nil), base...)
	for _, cast := range base {
		branch := cast
		apply(&branch)
		expanded = append(expanded, branch)
	}
	return expanded
}

// freeCastOptionLegal reports whether cast may be proposed without paying its
// mana cost: its choices must be legal and its remaining costs payable, either
// now or after activating payment-time mana abilities (CR 601.2g).
func (e *Engine) freeCastOptionLegal(g *game.Game, playerID game.PlayerID, spellDef *game.CardDef, cast action.CastSpellAction) bool {
	if !freeCastProposalLegal(g, playerID, spellDef, cast) {
		return false
	}
	_, feasible := spellCostFeasibility(g, freeCastCostRequest(playerID, spellDef, cast), e.hasManualPaymentManaAbility(g, playerID))
	return feasible
}

func freeCastProposalLegal(g *game.Game, playerID game.PlayerID, spellDef *game.CardDef, cast action.CastSpellAction) bool {
	if cast.XValue < 0 || cast.KickerCount < 0 ||
		cast.XValue != 0 && costHasVariableMana(manaCostPtr(spellDef.ManaCost)) {
		return false
	}
	branch := castBranchForCast(cast)
	if !modesValidForSpellAtBranch(g, playerID, spellDef, cast.ChosenModes, branch) ||
		!spellTargetsValidForCast(g, playerID, spellDef, cast.ChosenModes, cast.Targets, cast.XValue,
			effectiveKickerCount(cast.KickerPaid, cast.KickerCount), branch) {
		return false
	}
	return true
}

func freeCastCostRequest(playerID game.PlayerID, spellDef *game.CardDef, cast action.CastSpellAction) payment.SpellRequest {
	return payment.SpellRequest{
		PlayerID:        playerID,
		CardID:          cast.CardID,
		SourceZone:      cast.SourceZone,
		Card:            spellDef,
		XValue:          cast.XValue,
		KickerPaid:      cast.KickerPaid,
		KickerCount:     cast.KickerCount,
		Bargained:       cast.Bargained,
		Offspring:       cast.Offspring,
		ChosenModes:     cast.ChosenModes,
		Targets:         cast.Targets,
		CastPermissions: []payment.SpellCastPermission{payment.SpellCastPermissionDefault},
		Alternative:     opt.Val(freeCastAlternativeCost()),
	}
}

func (e *Engine) chooseAndCastFreeSpellFromSource(g *game.Game, sourcePlayer *game.Player, controllerID game.PlayerID, cardID id.ID, fromZone zone.Type, selection game.Selection, maxManaValue opt.V[int], exileOnResolution bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	return e.chooseAndCastFreeSpellFromSourceMay(g, sourcePlayer, controllerID, cardID, fromZone, selection, maxManaValue, exileOnResolution, false, agents, log)
}

// chooseAndCastFreeSpellFromSourceMay offers controllerID the legal ways to cast
// cardID without paying its mana cost. When mayDecline is set the cast is
// optional ("you may cast"), so the same choice also offers declining, and
// declining casts nothing; it is only offered when a legal cast exists.
func (e *Engine) chooseAndCastFreeSpellFromSourceMay(g *game.Game, sourcePlayer *game.Player, controllerID game.PlayerID, cardID id.ID, fromZone zone.Type, selection game.Selection, maxManaValue opt.V[int], exileOnResolution, mayDecline bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	return e.chooseAndCastFreeSpell(g, sourcePlayer, controllerID, cardID, fromZone, selection, maxManaValue, freeCastSpell{exileOnResolution: exileOnResolution}, mayDecline, agents, log)
}

// freeCastSpell records how a free-cast route marks the spell it puts on the
// stack.
type freeCastSpell struct {
	exileOnResolution bool
	// copyCard casts a copy of the card rather than the card itself.
	copyCard bool
	// suspend marks a spell cast by its suspend ability; a creature spell cast
	// this way gains haste (CR 702.62a).
	suspend bool
}

func (e *Engine) chooseAndCastFreeSpell(g *game.Game, sourcePlayer *game.Player, controllerID game.PlayerID, cardID id.ID, fromZone zone.Type, selection game.Selection, maxManaValue opt.V[int], spell freeCastSpell, mayDecline bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if sourcePlayer == nil || !castSourceContains(sourcePlayer, cardID, fromZone) || castFromZoneProhibited(g, controllerID, fromZone) {
		return false
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	options := e.freeCastOptionsForCard(g, controllerID, card, fromZone, selection, maxManaValue)
	if len(options) == 0 {
		for _, face := range card.Def.LegalCastFaces() {
			def := cardFaceOrDefault(card, face)
			if freeCastFaceAvailable(g, controllerID, card, def, selection, maxManaValue) &&
				(spellHasMultikicker(def) || !costHasVariableMana(manaCostPtr(def.ManaCost)) && additionalCostsUseX(def.AdditionalCosts)) {
				options = append(options, freeCastOption{cardID: card.ID, cast: action.CastSpellAction{
					CardID: card.ID, SourceZone: fromZone, Face: face,
				}})
			}
		}
	}
	// CR 733: a proposal whose costs are not paid is reversed to the moment
	// before it was proposed, so the caster chooses again among the proposals
	// not yet tried rather than losing the opportunity.
	for len(options) > 0 {
		if !castSourceContains(sourcePlayer, cardID, fromZone) {
			return false
		}
		index, ok := e.chooseFreeCastOption(g, controllerID, card, fromZone, options, mayDecline, agents, log)
		if !ok {
			return false
		}
		cast := options[index].cast
		spellDef := cardFaceOrDefault(card, cast.Face)
		cast, ok = e.chooseCastNumbers(g, controllerID, spellDef, cast, true, agents, log)
		if !ok || !e.freeCastOptionLegal(g, controllerID, spellDef, cast) {
			g.Execution.InvalidChoices++
			return false
		}
		if e.castFreeSpell(g, sourcePlayer, controllerID, card, cast, spell, agents, log) {
			return true
		}
		options = slices.Delete(options, index, index+1)
	}
	return false
}

// chooseFreeCastOption asks controllerID to pick one of options, defaulting to
// the first proposal payable from current resources. It reports false when the
// caster declines or answers invalidly.
func (e *Engine) chooseFreeCastOption(g *game.Game, controllerID game.PlayerID, card *game.CardInstance, fromZone zone.Type, options []freeCastOption, mayDecline bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (int, bool) {
	defaultIndex := 0
	for i := range options {
		if options[i].payable {
			defaultIndex = i
			break
		}
	}
	choiceOptions := make([]game.ChoiceOption, len(options))
	for i := range options {
		spellDef := cardFaceOrDefault(card, options[i].cast.Face)
		choiceOptions[i] = game.ChoiceOption{
			Index:   i,
			Label:   freeCastOptionLabel(spellDef, options[i].cast),
			Card:    cardChoiceInfo(g, card.ID),
			Targets: append([]game.Target(nil), options[i].cast.Targets...),
		}
	}
	if mayDecline {
		choiceOptions = append(choiceOptions, game.ChoiceOption{
			Index: len(options),
			Label: "Don't cast " + cardChoiceLabel(g, card.ID),
			Card:  cardChoiceInfo(g, card.ID),
		})
	}
	selected := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           controllerID,
		Prompt:           "Choose how to cast " + cardChoiceLabel(g, card.ID) + " without paying its mana cost",
		Options:          choiceOptions,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{defaultIndex},
		Private:          fromZone == zone.Hand || fromZone == zone.Library,
	}, log)
	if len(selected) != 1 || selected[0] < 0 || selected[0] >= len(options) {
		return 0, false
	}
	return selected[0], true
}

func freeCastOptionLabel(spellDef *game.CardDef, cast action.CastSpellAction) string {
	label := spellDef.Name
	if cast.Face == game.FaceFused {
		label += " (fused)"
	} else if len(cast.ChosenModes) > 0 {
		label += fmt.Sprintf(" modes %v", cast.ChosenModes)
	}
	if cast.KickerCount > 1 {
		label += fmt.Sprintf(" with multikicker %d", cast.KickerCount)
	} else if cast.KickerPaid {
		label += " with kicker"
	}
	if cast.Bargained {
		label += " bargained"
	}
	if cast.Offspring {
		label += " with offspring"
	}
	if cast.GiftPromised {
		label += fmt.Sprintf(" promising gift to player %d", cast.GiftRecipient)
	}
	if len(cast.Targets) > 0 {
		label += fmt.Sprintf(" targeting %v", cast.Targets)
	}
	return label
}

func (e *Engine) castFreeOption(g *game.Game, sourcePlayer *game.Player, controllerID game.PlayerID, card *game.CardInstance, cast action.CastSpellAction, exileOnResolution bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	return e.castFreeSpell(g, sourcePlayer, controllerID, card, cast, freeCastSpell{exileOnResolution: exileOnResolution}, agents, log)
}

func (e *Engine) castFreeOptionWithCopy(g *game.Game, sourcePlayer *game.Player, controllerID game.PlayerID, card *game.CardInstance, cast action.CastSpellAction, exileOnResolution, copyCard bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	return e.castFreeSpell(g, sourcePlayer, controllerID, card, cast, freeCastSpell{exileOnResolution: exileOnResolution, copyCard: copyCard}, agents, log)
}

func (e *Engine) castFreeSpell(g *game.Game, sourcePlayer *game.Player, controllerID game.PlayerID, card *game.CardInstance, cast action.CastSpellAction, spell freeCastSpell, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	copyCard := spell.copyCard
	if castFromZoneProhibited(g, controllerID, cast.SourceZone) {
		return false
	}
	// A cast copy has its own costs, not the identity of the card remaining in
	// exile. Payment must not reserve or consume that card as the casting source.
	paymentCardID := card.ID
	if copyCard {
		paymentCardID = 0
		cast.CardID = 0
	}
	spellDef := cardFaceOrDefault(card, cast.Face)
	branch := castBranchForCast(cast)
	completedTargets, ok := e.completeSpellAnnouncementTargets(g, controllerID, spellDef, cast.ChosenModes, cast.Targets, agents, log, branch)
	if !ok {
		return false
	}
	cast.Targets = completedTargets
	if !e.freeCastOptionLegal(g, controllerID, spellDef, cast) {
		return false
	}
	targetCounts, ok := spellTargetCounts(g, controllerID, spellDef, cast.ChosenModes, cast.Targets, branch)
	if !ok {
		panic("validated free-cast spell targets could not be segmented")
	}
	var splice spliceCastResult
	if cast.SourceZone == zone.Hand && spellIsArcane(spellDef) {
		splice = e.chooseSpliceOntoArcane(g, controllerID, card.ID, spellDef, cast, nil, agents, log)
	}
	obj := &game.StackObject{
		ID:                            g.IDGen.Next(),
		Kind:                          game.StackSpell,
		SourceID:                      card.ID,
		Face:                          cast.Face,
		Controller:                    controllerID,
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
		ExileOnResolution:             spell.exileOnResolution,
		Suspend:                       spell.suspend,
		SourceZone:                    cast.SourceZone,
		SplicedContent:                splice.contents,
		SplicedTargets:                splice.targets,
		SplicedTargetCounts:           splice.targetCounts,
		CastDuringControllerMainPhase: controllerID == g.Turn.ActivePlayer && g.Turn.IsMainPhase(),
	}
	var counterTransition castCounterTransition
	if copyCard {
		obj.Copy = true
		obj.CastCopy = true
		obj.CopyOwner = opt.Val(controllerID)
		obj.SourceCardID = card.ID
		obj.SourceTokenDef = copyCardDef(card.Def)
	} else {
		counterTransition = prepareCastCounterTransition(g, card, cast.SourceZone)
		if !removeCastSourceCard(g, sourcePlayer, card.ID, cast.SourceZone) {
			return false
		}
		counterTransition.apply(g)
	}
	g.Stack.Push(obj)
	rollback := func() {
		g.Stack.RemoveByID(obj.ID)
		if !copyCard {
			restoreCastSourceCard(sourcePlayer, card.ID, cast.SourceZone)
			counterTransition.undo(g)
		}
	}
	targetEvents := prepareTargetEvents(g, obj)
	request := payment.SpellRequest{
		PlayerID:        controllerID,
		CardID:          paymentCardID,
		SourceZone:      cast.SourceZone,
		Card:            spellDef,
		XValue:          cast.XValue,
		KickerPaid:      cast.KickerPaid,
		KickerCount:     cast.KickerCount,
		Bargained:       cast.Bargained,
		Offspring:       cast.Offspring,
		ChosenModes:     cast.ChosenModes,
		Targets:         cast.Targets,
		CastPermissions: []payment.SpellCastPermission{payment.SpellCastPermissionDefault},
		SpliceManaCosts: splice.manaCosts,
		Alternative:     opt.Val(freeCastAlternativeCost()),
	}
	// CR 601.2f-h: lock the total cost, then, if it includes a mana payment,
	// let the caster activate mana abilities before paying it. An unpaid
	// proposal is reversed (CR 733).
	prepared, ok := e.choosePreparedSpellCost(g, request, agents, log)
	if !ok {
		rollback()
		return false
	}
	request.PreparedCost = &prepared
	paymentOption := prepared.Summary
	undoMana := func() {}
	if prepared.IncludesManaPayment() {
		undoMana = e.activateManualPaymentManaAbilities(g, controllerID, agents, log)
	}
	prefs := e.paymentPreferencesForCostFromSource(
		g,
		controllerID,
		paymentOption.ManaCost,
		paymentOption.AdditionalCosts,
		cast.XValue,
		paymentCardID,
		cast.SourceZone,
		agents,
		log,
	)
	prefs.AlternativeIndex = paymentOption.Index
	request.Prefs = prefs
	riderSnapshot, _ := manaSpendRiderSnapshot(g, controllerID)
	paymentResult, ok := paymentOrch.paySpellCosts(g, request)
	if !ok {
		undoMana()
		rollback()
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
	stormCopies := stormCopyCount(g, spellDef)
	typeLine := stackSpellTypeLine(g, obj, spellDef)
	emitSpellCastEventsWithTargets(g, obj, game.Event{
		SourceID:                     card.ID,
		StackObjectID:                obj.ID,
		Controller:                   controllerID,
		CardID:                       paymentCardID,
		TokenDef:                     obj.SourceTokenDef,
		Face:                         cast.Face,
		CardTypes:                    typeLine.types,
		CardSupertypes:               typeLine.supertypes,
		CardSubtypes:                 typeLine.subtypes,
		Colors:                       spellColors(spellDef),
		ManaValue:                    stackObjectKnownManaValue(g, obj),
		ManaSpentToCast:              opt.Val(totalManaSpent(paymentResult.PoolSpend)),
		ManaFromCreaturesSpentToCast: opt.Val(creatureManaSpent(paymentResult.PoolSpend)),
		KickerPaid:                   cast.KickerPaid,
		FromZone:                     cast.SourceZone,
		ToZone:                       zone.Stack,
	}, targetEvents)
	createStormCopies(g, obj, spellDef, stormCopies)
	resolveSpellCastManaSpendRiders(g, controllerID, riderSnapshot, paymentResult.PoolSpend, spellDef, obj)
	return true
}
