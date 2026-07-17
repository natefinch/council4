package rules

import (
	"fmt"

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
}

func (*Engine) freeCastOptionsForCard(g *game.Game, playerID game.PlayerID, card *game.CardInstance, fromZone zone.Type, selection game.Selection, maxManaValue opt.V[int]) []freeCastOption {
	if card == nil || card.Def == nil || castFromZoneProhibited(g, playerID, fromZone) {
		return nil
	}
	var options []freeCastOption
	for _, face := range card.Def.LegalCastFaces() {
		spellDef := cardFaceOrDefault(card, face)
		if !freeCastFaceMatchesSelection(g, card, spellDef, selection, playerID) ||
			maxManaValue.Exists && stackManaValue(spellDef, 0) > maxManaValue.Val ||
			!cardCastRestrictionsSatisfied(g, playerID, spellDef) ||
			spellCastProhibited(g, playerID, spellDef) ||
			spellCastLimitReached(g, playerID, spellDef) ||
			!isSupportedSpell(spellDef) {
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
					if freeCastOptionLegal(g, playerID, spellDef, candidate) {
						options = append(options, freeCastOption{cardID: card.ID, cast: candidate})
					}
				}
			}
		}
	}
	return options
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
	if spellHasMultikicker(spellDef) {
		for count := 1; count <= maxLegalMultikickCount; count++ {
			base = append(base, action.CastSpellAction{
				CardID:      cardID,
				SourceZone:  fromZone,
				Face:        face,
				KickerPaid:  true,
				KickerCount: count,
			})
		}
	} else if spellHasKicker(spellDef) {
		base = append(base, action.CastSpellAction{
			CardID:     cardID,
			SourceZone: fromZone,
			Face:       face,
			KickerPaid: true,
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

func freeCastOptionLegal(g *game.Game, playerID game.PlayerID, spellDef *game.CardDef, cast action.CastSpellAction) bool {
	branch := castBranchForCast(cast)
	if !modesValidForSpellAtBranch(g, playerID, spellDef, cast.ChosenModes, branch) ||
		!targetsValidForSpell(g, playerID, spellDef, cast.ChosenModes, cast.Targets, branch) ||
		!spellTargetCountsMatchX(g, playerID, spellDef, cast.ChosenModes, cast.Targets, 0, branch) ||
		!spellTargetCountsMatchKicker(g, playerID, spellDef, cast.ChosenModes, cast.Targets, effectiveKickerCount(cast.KickerPaid, cast.KickerCount), branch) ||
		!spellTargetsSatisfyManaValueX(g, playerID, spellDef, cast.ChosenModes, cast.Targets, 0, branch) {
		return false
	}
	return paymentOrch.canPaySpellCosts(g, payment.SpellRequest{
		PlayerID:        playerID,
		CardID:          cast.CardID,
		SourceZone:      cast.SourceZone,
		Card:            spellDef,
		XValue:          0,
		KickerPaid:      cast.KickerPaid,
		KickerCount:     cast.KickerCount,
		Bargained:       cast.Bargained,
		Offspring:       cast.Offspring,
		ChosenModes:     cast.ChosenModes,
		Targets:         cast.Targets,
		CastPermissions: []payment.SpellCastPermission{payment.SpellCastPermissionDefault},
		Alternative:     opt.Val(freeCastAlternativeCost()),
	})
}

func (e *Engine) chooseAndCastFreeSpellFromSource(g *game.Game, sourcePlayer *game.Player, controllerID game.PlayerID, cardID id.ID, fromZone zone.Type, selection game.Selection, maxManaValue opt.V[int], exileOnResolution bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if sourcePlayer == nil || !castSourceContains(sourcePlayer, cardID, fromZone) {
		return false
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	options := e.freeCastOptionsForCard(g, controllerID, card, fromZone, selection, maxManaValue)
	if len(options) == 0 {
		return false
	}
	choiceOptions := make([]game.ChoiceOption, len(options))
	for i := range options {
		spellDef := cardFaceOrDefault(card, options[i].cast.Face)
		choiceOptions[i] = game.ChoiceOption{
			Index:   i,
			Label:   freeCastOptionLabel(spellDef, options[i].cast),
			Card:    cardChoiceInfo(g, cardID),
			Targets: append([]game.Target(nil), options[i].cast.Targets...),
		}
	}
	selected := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           controllerID,
		Prompt:           "Choose how to cast " + cardChoiceLabel(g, cardID) + " without paying its mana cost",
		Options:          choiceOptions,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: firstChoiceIndices(1),
	}, log)
	if len(selected) != 1 || selected[0] < 0 || selected[0] >= len(options) {
		return false
	}
	return e.castFreeOption(g, sourcePlayer, controllerID, card, options[selected[0]].cast, exileOnResolution, agents, log)
}

func freeCastOptionLabel(spellDef *game.CardDef, cast action.CastSpellAction) string {
	label := spellDef.Name
	if len(cast.ChosenModes) > 0 {
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

func (e *Engine) chooseFreeCastPaymentOption(g *game.Game, controllerID game.PlayerID, request payment.SpellRequest, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (payment.SpellOptionSummary, bool) {
	options := paymentOrch.planner(g).PayableSpellOptions(request)
	if len(options) == 0 {
		return payment.SpellOptionSummary{}, false
	}
	if len(options) == 1 {
		return options[0], true
	}
	choiceOptions := make([]game.ChoiceOption, len(options))
	for i, option := range options {
		choiceOptions[i] = game.ChoiceOption{Index: option.Index, Label: option.Label}
	}
	selected := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind:             game.ChoicePayment,
		Player:           controllerID,
		Prompt:           "Choose additional spell cost",
		Options:          choiceOptions,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{options[0].Index},
	}, log)
	if len(selected) == 1 {
		for _, option := range options {
			if option.Index == selected[0] {
				return option, true
			}
		}
	}
	return options[0], true
}

func (e *Engine) castFreeOption(g *game.Game, sourcePlayer *game.Player, controllerID game.PlayerID, card *game.CardInstance, cast action.CastSpellAction, exileOnResolution bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	spellDef := cardFaceOrDefault(card, cast.Face)
	branch := castBranchForCast(cast)
	completedTargets, ok := e.completeSpellAnnouncementTargets(g, controllerID, spellDef, cast.ChosenModes, cast.Targets, agents, log, branch)
	if !ok {
		return false
	}
	cast.Targets = completedTargets
	if !freeCastOptionLegal(g, controllerID, spellDef, cast) {
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
		KickerPaid:                    cast.KickerPaid,
		KickerCount:                   cast.KickerCount,
		GiftPromised:                  cast.GiftPromised,
		GiftRecipient:                 cast.GiftRecipient,
		Bargained:                     cast.Bargained,
		OffspringPaid:                 cast.Offspring,
		ExileOnResolution:             exileOnResolution,
		SourceZone:                    cast.SourceZone,
		SplicedContent:                splice.contents,
		SplicedTargets:                splice.targets,
		SplicedTargetCounts:           splice.targetCounts,
		CastDuringControllerMainPhase: controllerID == g.Turn.ActivePlayer && g.Turn.IsMainPhase(),
	}
	if !removeCastSourceCard(g, sourcePlayer, card.ID, cast.SourceZone) {
		return false
	}
	g.Stack.Push(obj)
	request := payment.SpellRequest{
		PlayerID:        controllerID,
		CardID:          card.ID,
		SourceZone:      cast.SourceZone,
		Card:            spellDef,
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
	paymentOption, ok := e.chooseFreeCastPaymentOption(g, controllerID, request, agents, log)
	if !ok {
		g.Stack.RemoveByID(obj.ID)
		restoreCastSourceCard(sourcePlayer, card.ID, cast.SourceZone)
		return false
	}
	prefs := e.paymentPreferencesForCostFromSource(
		g,
		controllerID,
		paymentOption.ManaCost,
		paymentOption.AdditionalCosts,
		0,
		card.ID,
		cast.SourceZone,
		agents,
		log,
	)
	prefs.AlternativeIndex = paymentOption.Index
	request.Prefs = prefs
	riderSnapshot, _ := manaSpendRiderSnapshot(g, controllerID)
	paymentResult, ok := paymentOrch.paySpellCosts(g, request)
	if !ok {
		g.Stack.RemoveByID(obj.ID)
		restoreCastSourceCard(sourcePlayer, card.ID, cast.SourceZone)
		return false
	}
	obj.AdditionalCostsPaid = paymentResult.AdditionalCostsPaid
	obj.SacrificedAsCostIDs = paymentResult.SacrificedIDs
	obj.ColorsOfManaSpentToCast = distinctManaColorsSpent(paymentResult.PoolSpend)
	obj.ManaSpentByColorToCast = manaSpentByColor(paymentResult.PoolSpend)
	obj.ManaSpentToCast = totalManaSpent(paymentResult.PoolSpend)
	obj.ManaFromCreaturesSpentToCast = creatureManaSpent(paymentResult.PoolSpend)
	stormCopies := stormCopyCount(g, spellDef)
	emitSpellCastEvents(g, obj, game.Event{
		SourceID:                     card.ID,
		StackObjectID:                obj.ID,
		Controller:                   controllerID,
		CardID:                       card.ID,
		Face:                         cast.Face,
		CardTypes:                    stackObjectCardTypes(obj, spellDef),
		CardSupertypes:               cardSupertypes(spellDef),
		CardSubtypes:                 stackObjectCardSubtypes(obj, spellDef),
		Colors:                       spellColors(spellDef),
		ManaValue:                    opt.Val(stackManaValue(spellDef, 0)),
		ManaSpentToCast:              opt.Val(totalManaSpent(paymentResult.PoolSpend)),
		ManaFromCreaturesSpentToCast: opt.Val(creatureManaSpent(paymentResult.PoolSpend)),
		KickerPaid:                   cast.KickerPaid,
		FromZone:                     cast.SourceZone,
		ToZone:                       zone.Stack,
	})
	createStormCopies(g, obj, spellDef, stormCopies)
	resolveSpellCastManaSpendRiders(g, controllerID, riderSnapshot, paymentResult.PoolSpend, spellDef, obj)
	e.resolveCascadeForCast(g, obj, spellDef, agents, log)
	return true
}
