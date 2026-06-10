package rules

import (
	"strings"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

const maxLegalXValue = 20

func canPayCost(g *game.Game, playerID game.PlayerID, manaCost *cost.Mana) bool {
	return paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: manaCost})
}

func canPayCostWithX(g *game.Game, playerID game.PlayerID, manaCost *cost.Mana, xValue int) bool {
	return paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: manaCost, XValue: xValue})
}

func (e *Engine) legalActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canAct(g, playerID) {
		return []action.Action{actionBuild.pass()}
	}
	if splitSecondOnStack(g) {
		actions := e.legalManaAbilityActions(g, playerID)
		actions = append(actions, e.legalTurnFaceUpActions(g, playerID)...)
		actions = append(actions, actionBuild.pass())
		return actions
	}

	actions := e.legalLandActions(g, playerID)
	actions = append(actions, e.legalCastActions(g, playerID)...)
	actions = append(actions, legalPreparedSpellActions(g, playerID)...)
	actions = append(actions, e.legalFaceDownCastActions(g, playerID)...)
	actions = append(actions, e.legalCommanderCastActions(g, playerID)...)
	actions = append(actions, e.legalActivateAbilityActions(g, playerID)...)
	actions = append(actions, e.legalCyclingActions(g, playerID)...)
	actions = append(actions, e.legalNinjutsuActions(g, playerID)...)
	actions = append(actions, e.legalSuspendActions(g, playerID)...)
	actions = append(actions, e.legalTurnFaceUpActions(g, playerID)...)
	actions = append(actions, actionBuild.pass())
	return actions
}

func legalPreparedSpellActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return nil
	}
	var actions []action.Action
	for _, permanent := range g.Battlefield {
		if !permanent.Prepared || permanent.PhasedOut || effectiveController(g, permanent) != playerID {
			continue
		}
		sourceID, sourceDef, ok := preparedSpellSource(g, permanent)
		if !ok {
			continue
		}
		spellDef, ok := sourceDef.FaceDef(game.FaceAlternate)
		if !ok {
			continue
		}
		for _, xValue := range legalXValuesForCost(g, playerID, manaCostPtr(spellDef.ManaCost)) {
			for _, modes := range modeChoicesForSpell(spellDef) {
				targetResult := targetChoicesForSpell(g, playerID, spellDef, modes)
				if targetResult.kind == targetInvalidSpec {
					continue
				}
				for _, targets := range targetResult.choices {
					if canCastPreparedCopy(g, playerID, permanent, targets, xValue, modes) {
						actions = append(actions, actionBuild.castSpell(sourceID, zone.Battlefield, game.FaceAlternate, targets, xValue, modes))
					}
				}
			}
		}
	}
	return actions
}

func normalizedCastSourceZone(cast action.CastSpellAction) zone.Type {
	if cast.SourceZone == zone.None {
		return zone.Hand
	}
	return cast.SourceZone
}

func splitSecondOnStack(g *game.Game) bool {
	for _, obj := range g.Stack.Objects() {
		if obj.Kind != game.StackSpell {
			continue
		}
		card, ok := g.GetCardInstance(obj.SourceID)
		if ok && cardFaceOrDefault(card, obj.Face).HasKeyword(game.SplitSecond) {
			return true
		}
	}
	return false
}

func (*Engine) legalLandActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canPlayAnyLand(g, playerID) {
		return nil
	}

	player := g.Players[playerID]
	var actions []action.Action
	for _, cardID := range player.Hand.All() {
		card, ok := g.GetCardInstance(cardID)
		if !ok || !player.Hand.Contains(cardID) {
			continue
		}
		for _, face := range card.Def.FaceIndexes() {
			if _, ok := landCardInstanceFace(g, player, cardID, face); ok {
				actions = append(actions, actionBuild.playLand(cardID, face))
			}
		}
	}
	return actions
}

func (e *Engine) legalCastActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return nil
	}

	player := g.Players[playerID]
	var actions []action.Action
	for _, sourceZone := range castableZonesForPlayer(g, playerID) {
		for _, cardID := range castSourceZoneCards(player, sourceZone) {
			card, ok := g.GetCardInstance(cardID)
			if !ok {
				continue
			}
			for _, face := range legalCastFacesForZone(g, playerID, card, sourceZone) {
				spellDef := cardFaceOrDefault(card, face)
				if face == game.FaceFront && (sourceZone == zone.Hand || sourceZone == zone.Exile) {
					if _, ok := spellDef.MutateCost(); ok {
						for _, target := range legalMutateTargets(g, playerID, card.Owner, spellDef) {
							if canCastMutateSpell(g, playerID, cardID, sourceZone, target.ObjectID) {
								actions = append(actions, actionBuild.castMutateSpell(cardID, sourceZone, target.ObjectID))
							}
						}
					}
				}
				for _, xValue := range legalXValuesForCost(g, playerID, manaCostPtr(spellDef.ManaCost)) {
					for _, modes := range modeChoicesForSpell(spellDef) {
						targetResult := targetChoicesForSpell(g, playerID, spellDef, modes)
						if targetResult.kind == targetInvalidSpec {
							continue
						}
						for _, targets := range targetResult.choices {
							if e.canCastSpellFaceFromZoneWithKicker(g, playerID, cardID, sourceZone, face, targets, xValue, modes, false) {
								actions = append(actions, actionBuild.castSpell(cardID, sourceZone, face, targets, xValue, modes))
							}
							if sourceZone == zone.Hand && spellHasKicker(spellDef) && e.canCastSpellFaceFromZoneWithKicker(g, playerID, cardID, sourceZone, face, targets, xValue, modes, true) {
								actions = append(actions, actionBuild.castKickedSpell(cardID, sourceZone, face, targets, xValue, modes))
							}
						}
					}
				}
			}
		}
	}
	return actions
}

func (e *Engine) legalCommanderCastActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return nil
	}
	player, ok := playerByID(g, playerID)
	if !ok || player.CommanderInstanceID == 0 || !player.CommandZone.Contains(player.CommanderInstanceID) {
		return nil
	}
	card, ok := g.GetCardInstance(player.CommanderInstanceID)
	if !ok {
		return nil
	}
	var actions []action.Action
	for _, face := range card.Def.LegalCastFaces() {
		spellDef := cardFaceOrDefault(card, face)
		if face == game.FaceFront {
			if _, ok := spellDef.MutateCost(); ok {
				for _, target := range legalMutateTargets(g, playerID, card.Owner, spellDef) {
					if canCastMutateSpell(g, playerID, card.ID, zone.Command, target.ObjectID) {
						actions = append(actions, actionBuild.castMutateSpell(card.ID, zone.Command, target.ObjectID))
					}
				}
			}
		}
		for _, xValue := range legalXValuesForCost(g, playerID, manaCostPtr(spellDef.ManaCost)) {
			for _, modes := range modeChoicesForSpell(spellDef) {
				targetResult := targetChoicesForSpell(g, playerID, spellDef, modes)
				if targetResult.kind == targetInvalidSpec {
					continue
				}
				for _, targets := range targetResult.choices {
					if e.canCastSpellFaceFromZoneWithKicker(g, playerID, card.ID, zone.Command, face, targets, xValue, modes, false) {
						actions = append(actions, actionBuild.castSpell(card.ID, zone.Command, face, targets, xValue, modes))
					}
					if spellHasKicker(spellDef) && e.canCastSpellFaceFromZoneWithKicker(g, playerID, card.ID, zone.Command, face, targets, xValue, modes, true) {
						actions = append(actions, actionBuild.castKickedSpell(card.ID, zone.Command, face, targets, xValue, modes))
					}
				}
			}
		}
	}
	return actions
}

func (e *Engine) applyAction(g *game.Game, playerID game.PlayerID, act action.Action) bool {
	return e.applyActionWithChoices(g, playerID, act, [game.NumPlayers]PlayerAgent{}, nil)
}

func (e *Engine) applyActionWithChoices(g *game.Game, playerID game.PlayerID, act action.Action, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if err := act.Validate(); err != nil {
		return false
	}
	switch act.Kind {
	case action.ActionPass:
		return true
	case action.ActionPlayLand:
		playLand, ok := act.PlayLandPayload()
		return ok && e.applyPlayLandFaceWithChoices(g, playerID, playLand.CardID, playLand.Face, agents, log)
	case action.ActionCastSpell:
		cast, ok := act.CastSpellPayload()
		return ok && e.applyCastSpellWithChoices(g, playerID, cast, agents, log)
	case action.ActionActivateAbility:
		activate, ok := act.ActivateAbilityPayload()
		return ok && e.applyActivateAbilityWithChoices(g, playerID, activate, agents, log)
	case action.ActionSuspendCard:
		suspend, ok := act.SuspendCardPayload()
		return ok && e.applySuspendCard(g, playerID, suspend.CardID, agents, log)
	case action.ActionCastFaceDown:
		faceDown, ok := act.CastFaceDownPayload()
		return ok && e.applyCastFaceDownWithChoices(g, playerID, faceDown, agents, log)
	case action.ActionTurnFaceUp:
		turnFaceUp, ok := act.TurnFaceUpPayload()
		return ok && e.applyTurnFaceUpWithChoices(g, playerID, turnFaceUp.PermanentID, agents, log)
	case action.ActionDeclareAttackers:
		attackers, ok := act.DeclareAttackersPayload()
		return ok && e.applyDeclareAttackers(g, playerID, attackers)
	case action.ActionDeclareBlockers:
		blockers, ok := act.DeclareBlockersPayload()
		return ok && e.applyDeclareBlockers(g, playerID, blockers)
	default:
		return false
	}
}

func (e *Engine) legalActivateAbilityActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return nil
	}

	var actions []action.Action
	for _, permanent := range g.Battlefield {
		if permanent.PhasedOut || effectiveController(g, permanent) != playerID {
			continue
		}
		card, ok := permanentCardDef(g, permanent)
		if !ok {
			continue
		}
		for idx, ability := range permanentEffectiveAbilities(g, permanent) {
			if body, ok := ability.(game.ManaAbility); ok {
				if canActivateManaAbility(g, playerID, permanent, &body, idx) {
					actions = append(actions, actionBuild.activateAbility(permanent.ObjectID, idx, nil, 0))
				}
				continue
			}
			if body, ok := ability.(game.ActivatedAbility); ok {
				for _, xValue := range legalXValuesForCost(g, playerID, manaCostPtr(body.ManaCost)) {
					targetResult := targetChoicesForBodyFromSourceObject(g, playerID, card, permanent.ObjectID, &body)
					if targetResult.kind == targetInvalidSpec {
						continue
					}
					for _, targets := range targetResult.choices {
						if canActivateEquipAbility(g, playerID, permanent, &body, idx, targets, xValue) ||
							canActivateGeneralAbility(g, playerID, permanent, &body, idx, targets, xValue) {
							actions = append(actions, actionBuild.activateAbility(permanent.ObjectID, idx, append([]game.Target(nil), targets...), xValue))
						}
					}
				}
				continue
			}
			body, ok := ability.(game.LoyaltyAbility)
			if !ok {
				continue
			}
			targetResult := targetChoicesForBodyFromSourceObject(g, playerID, card, permanent.ObjectID, &body)
			if targetResult.kind == targetInvalidSpec {
				continue
			}
			for _, targets := range targetResult.choices {
				if canActivateLoyaltyAbility(g, playerID, permanent, &body, idx, targets, 0) {
					actions = append(actions, actionBuild.activateAbility(permanent.ObjectID, idx, append([]game.Target(nil), targets...), 0))
				}
			}
		}
	}
	actions = append(actions, e.legalGraveyardActivateAbilityActions(g, playerID)...)
	return actions
}

func (*Engine) legalGraveyardActivateAbilityActions(g *game.Game, playerID game.PlayerID) []action.Action {
	player, ok := playerByID(g, playerID)
	if !ok {
		return nil
	}
	var actions []action.Action
	for _, cardID := range player.Graveyard.All() {
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		def := cardFaceOrDefault(card, game.FaceFront)
		for i := range def.ActivatedAbilities {
			body := &def.ActivatedAbilities[i]
			idx := def.ActivatedAbilityIndex(i)
			for _, xValue := range legalXValuesForCost(g, playerID, manaCostPtr(body.ManaCost)) {
				targetResult := targetChoicesForBodyFromSourceObject(g, playerID, def, 0, body)
				if targetResult.kind == targetInvalidSpec {
					continue
				}
				for _, targets := range targetResult.choices {
					if canActivateGraveyardAbility(g, playerID, cardID, body, idx, targets, xValue) {
						actions = append(actions, actionBuild.activateAbility(cardID, idx, append([]game.Target(nil), targets...), xValue))
					}
				}
			}
		}
	}
	return actions
}

func (*Engine) legalManaAbilityActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return nil
	}
	var actions []action.Action
	for _, permanent := range g.Battlefield {
		if permanent.PhasedOut || effectiveController(g, permanent) != playerID {
			continue
		}
		for idx, ability := range permanentEffectiveAbilities(g, permanent) {
			body, ok := ability.(game.ManaAbility)
			if ok && canActivateManaAbility(g, playerID, permanent, &body, idx) {
				actions = append(actions, actionBuild.activateAbility(permanent.ObjectID, idx, nil, 0))
			}
		}
	}
	return actions
}

func (*Engine) legalCyclingActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return nil
	}
	player, ok := playerByID(g, playerID)
	if !ok {
		return nil
	}
	var actions []action.Action
	for _, cardID := range player.Hand.All() {
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		frontDef := cardFaceOrDefault(card, game.FaceFront)
		for i := range frontDef.ActivatedAbilities {
			body := &frontDef.ActivatedAbilities[i]
			idx := frontDef.ActivatedAbilityIndex(i)
			if canActivateCyclingAbility(g, playerID, cardID, body, idx, nil, 0) {
				actions = append(actions, actionBuild.activateAbility(cardID, idx, nil, 0))
			}
		}
	}
	return actions
}

func (*Engine) legalNinjutsuActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer || len(unblockedAttackers(g, playerID)) == 0 {
		return nil
	}
	player, ok := playerByID(g, playerID)
	if !ok {
		return nil
	}
	var actions []action.Action
	for _, cardID := range player.Hand.All() {
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		frontDef := cardFaceOrDefault(card, game.FaceFront)
		for i := range frontDef.ActivatedAbilities {
			body := &frontDef.ActivatedAbilities[i]
			idx := frontDef.ActivatedAbilityIndex(i)
			if canActivateNinjutsuAbility(g, playerID, cardID, body, idx, nil, 0) {
				actions = append(actions, actionBuild.activateAbility(cardID, idx, nil, 0))
			}
		}
	}
	return actions
}

func (e *Engine) applyPlayLand(g *game.Game, playerID game.PlayerID, cardID id.ID) bool {
	return e.applyPlayLandFace(g, playerID, cardID, game.FaceFront)
}

func (e *Engine) applyPlayLandFace(g *game.Game, playerID game.PlayerID, cardID id.ID, face game.FaceIndex) bool {
	return e.applyPlayLandFaceWithChoices(g, playerID, cardID, face, [game.NumPlayers]PlayerAgent{}, nil)
}

func (e *Engine) applyPlayLandFaceWithChoices(g *game.Game, playerID game.PlayerID, cardID id.ID, face game.FaceIndex, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if !canPlayAnyLand(g, playerID) {
		return false
	}

	player := g.Players[playerID]
	card, ok := landCardInstanceFace(g, player, cardID, face)
	if !ok {
		return false
	}
	if !player.Hand.Remove(cardID) {
		return false
	}

	if _, ok := createCardPermanentFaceWithChoices(e, g, card, playerID, zone.Hand, face, agents, log); !ok {
		return false
	}
	g.Turn.LandsPlayedThisTurn++
	return true
}

func (e *Engine) applyCastSpell(g *game.Game, playerID game.PlayerID, cast action.CastSpellAction) bool {
	return e.applyCastSpellWithChoices(g, playerID, cast, [game.NumPlayers]PlayerAgent{}, nil)
}

func (e *Engine) applyCastSpellWithChoices(g *game.Game, playerID game.PlayerID, cast action.CastSpellAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if cast.Mutate {
		return e.applyMutateCastWithChoices(g, playerID, cast, agents, log)
	}
	sourceZone := normalizedCastSourceZone(cast)
	if sourceZone == zone.Battlefield {
		return e.applyPreparedCopyWithChoices(g, playerID, cast, agents, log)
	}

	if !e.canCastSpellFaceFromZoneWithKicker(g, playerID, cast.CardID, sourceZone, cast.Face, cast.Targets, cast.XValue, cast.ChosenModes, cast.KickerPaid) {
		return false
	}

	player := g.Players[playerID]
	card, ok := g.GetCardInstance(cast.CardID)
	if !ok {
		return false
	}
	spellDef := cardFaceOrDefault(card, cast.Face)
	completedTargets, ok := e.completeSpellAnnouncementTargets(g, playerID, spellDef, cast.ChosenModes, cast.Targets, agents, log)
	if !ok || !e.canCastSpellFaceFromZoneWithKicker(g, playerID, cast.CardID, sourceZone, cast.Face, completedTargets, cast.XValue, cast.ChosenModes, cast.KickerPaid) {
		return false
	}
	cast.Targets = completedTargets
	targetCounts, ok := spellTargetCounts(g, playerID, spellDef, cast.ChosenModes, cast.Targets)
	if !ok {
		panic("validated spell targets could not be segmented")
	}
	prefs := e.paymentPreferencesForSpellFromZone(g, playerID, card.ID, sourceZone, spellDef, cast.XValue, agents, log)
	additionalCostsPaid, ok := paymentOrch.paySpellCosts(g, payment.SpellRequest{PlayerID: playerID, CardID: card.ID, SourceZone: sourceZone, Card: spellDef, XValue: cast.XValue, KickerPaid: cast.KickerPaid, Prefs: prefs})
	if !ok {
		return false
	}
	if !removeCastSourceCard(g, player, cast.CardID, sourceZone) {
		panic("cast spell disappeared from source zone after validation")
	}
	if sourceZone == zone.Command && player.CommanderInstanceID == cast.CardID {
		player.CommanderCastCount++
	}
	obj := &game.StackObject{
		ID:                  g.IDGen.Next(),
		Kind:                game.StackSpell,
		SourceID:            cast.CardID,
		Face:                cast.Face,
		Controller:          playerID,
		Targets:             append([]game.Target(nil), cast.Targets...),
		TargetCounts:        targetCounts,
		ChosenModes:         append([]int(nil), cast.ChosenModes...),
		XValue:              cast.XValue,
		KickerPaid:          cast.KickerPaid,
		Flashback:           sourceZone == zone.Graveyard && spellDef.HasKeyword(game.Flashback),
		AdditionalCostsPaid: additionalCostsPaid,
		SourceZone:          sourceZone,
	}
	stormCopies := stormCopyCount(g, spellDef)
	pushSpellToStack(g, obj, game.Event{
		SourceID:      cast.CardID,
		StackObjectID: obj.ID,
		Controller:    playerID,
		CardID:        cast.CardID,
		Face:          cast.Face,
		CardTypes:     cardTypes(spellDef),
		Colors:        spellColors(spellDef),
		FromZone:      sourceZone,
		ToZone:        zone.Stack,
	})
	createStormCopies(g, obj, stormCopies)
	e.resolveCascadeForCast(g, obj, spellDef, agents, log)
	return true
}

func (e *Engine) applyMutateCastWithChoices(g *game.Game, playerID game.PlayerID, cast action.CastSpellAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	sourceZone := normalizedCastSourceZone(cast)
	if !canCastMutateSpell(g, playerID, cast.CardID, sourceZone, cast.MutateTargetID) {
		return false
	}
	player := g.Players[playerID]
	card, ok := g.GetCardInstance(cast.CardID)
	if !ok {
		return false
	}
	spellDef := cardFaceOrDefault(card, game.FaceFront)
	mutateCost, ok := spellDef.MutateCost()
	if !ok {
		return false
	}
	alternative := mutateAlternativeCost(mutateCost)
	prefs := e.paymentPreferencesForSpellFromZone(g, playerID, card.ID, sourceZone, spellDef, 0, agents, log)
	additionalCostsPaid, ok := paymentOrch.paySpellCosts(g, payment.SpellRequest{
		PlayerID:    playerID,
		CardID:      card.ID,
		SourceZone:  sourceZone,
		Card:        spellDef,
		Alternative: opt.Val(alternative),
		Prefs:       prefs,
	})
	if !ok {
		return false
	}
	if !removeCastSourceCard(g, player, cast.CardID, sourceZone) {
		panic("mutate spell disappeared from source zone after validation")
	}
	if sourceZone == zone.Command && player.CommanderInstanceID == cast.CardID {
		player.CommanderCastCount++
	}
	obj := &game.StackObject{
		ID:                  g.IDGen.Next(),
		Kind:                game.StackSpell,
		SourceID:            cast.CardID,
		Face:                game.FaceFront,
		Controller:          playerID,
		Targets:             []game.Target{game.PermanentTarget(cast.MutateTargetID)},
		TargetCounts:        []int{1},
		Mutate:              true,
		MutateTargetID:      cast.MutateTargetID,
		AdditionalCostsPaid: additionalCostsPaid,
		SourceZone:          sourceZone,
	}
	pushSpellToStack(g, obj, game.Event{
		SourceID:      cast.CardID,
		StackObjectID: obj.ID,
		Controller:    playerID,
		CardID:        cast.CardID,
		Face:          game.FaceFront,
		CardTypes:     cardTypes(spellDef),
		Colors:        spellColors(spellDef),
		FromZone:      sourceZone,
		ToZone:        zone.Stack,
	})
	return true
}

func canCastMutateSpell(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, targetID id.ID) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return false
	}
	player := g.Players[playerID]
	card, ok := g.GetCardInstance(cardID)
	if !ok || !castSourceContains(player, cardID, sourceZone) {
		return false
	}
	switch sourceZone {
	case zone.Hand:
	case zone.Command:
		if player.CommanderInstanceID != cardID {
			return false
		}
	case zone.Exile:
		if !g.AdventureCards[cardID] {
			return false
		}
	default:
		return false
	}
	spellDef := cardFaceOrDefault(card, game.FaceFront)
	mutateCost, ok := spellDef.MutateCost()
	if !ok ||
		!spellDef.HasType(types.Creature) ||
		!isSupportedSpell(spellDef) ||
		!canCastAtCurrentTiming(g, playerID, spellDef) ||
		!mutateTargetLegal(g, playerID, card.Owner, spellDef, targetID) {
		return false
	}
	return paymentOrch.canPaySpellCosts(g, payment.SpellRequest{
		PlayerID:    playerID,
		CardID:      cardID,
		SourceZone:  sourceZone,
		Card:        spellDef,
		Alternative: opt.Val(mutateAlternativeCost(mutateCost)),
	})
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
		!targetProtectedFromSource(g, playerID, spellDef, game.PermanentTarget(targetID))
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
	if xValue != 0 && !costHasVariableMana(manaCostPtr(spellDef.ManaCost)) {
		return false
	}
	if !modesValidForSpell(spellDef, chosenModes) ||
		!isSupportedSpell(spellDef) ||
		(!spellDef.HasType(types.Instant) && !spellDef.HasType(types.Sorcery)) ||
		!targetsValidForSpell(g, playerID, spellDef, chosenModes, targets) ||
		!canCastAtCurrentTiming(g, playerID, spellDef) {
		return false
	}
	return paymentOrch.canPaySpellCosts(g, payment.SpellRequest{
		PlayerID:   playerID,
		CardID:     sourceID,
		SourceZone: zone.Battlefield,
		Card:       spellDef,
		XValue:     xValue,
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
	completedTargets, ok := e.completeSpellAnnouncementTargets(g, playerID, spellDef, cast.ChosenModes, cast.Targets, agents, log)
	if !ok || !canCastPreparedCopy(g, playerID, permanent, completedTargets, cast.XValue, cast.ChosenModes) {
		return false
	}
	cast.Targets = completedTargets
	targetCounts, ok := spellTargetCounts(g, playerID, spellDef, cast.ChosenModes, cast.Targets)
	if !ok {
		panic("validated prepared spell targets could not be segmented")
	}
	prefs := e.paymentPreferencesForSpellFromZone(g, playerID, sourceID, zone.Battlefield, spellDef, cast.XValue, agents, log)
	additionalCostsPaid, ok := paymentOrch.paySpellCosts(g, payment.SpellRequest{
		PlayerID:   playerID,
		CardID:     sourceID,
		SourceZone: zone.Battlefield,
		Card:       spellDef,
		XValue:     cast.XValue,
		Prefs:      prefs,
	})
	if !ok {
		return false
	}
	permanent.Prepared = false
	obj := &game.StackObject{
		ID:                  g.IDGen.Next(),
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
		AdditionalCostsPaid: additionalCostsPaid,
		SourceZone:          zone.Battlefield,
	}
	stormCopies := stormCopyCount(g, spellDef)
	g.Stack.Push(obj)
	emitTargetEvents(g, obj)
	emitEvent(g, game.Event{
		Kind:          game.EventSpellCast,
		SourceID:      sourceID,
		StackObjectID: obj.ID,
		Controller:    playerID,
		CardID:        permanent.CardInstanceID,
		Face:          game.FaceAlternate,
		PermanentID:   permanent.ObjectID,
		TokenDef:      permanent.TokenDef,
		CardTypes:     cardTypes(spellDef),
		Colors:        spellColors(spellDef),
	})
	createStormCopies(g, obj, stormCopies)
	e.resolveCascadeForCast(g, obj, spellDef, agents, log)
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

func (e *Engine) applyActivateAbility(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction) bool {
	return e.applyActivateAbilityWithChoices(g, playerID, activate, [game.NumPlayers]PlayerAgent{}, nil)
}

func (e *Engine) applyActivateAbilityWithChoices(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if e.applyCyclingAbilityWithChoices(g, playerID, activate, agents, log) {
		return true
	}
	if e.applyNinjutsuAbilityWithChoices(g, playerID, activate, agents, log) {
		return true
	}
	if e.applyGraveyardAbilityWithChoices(g, playerID, activate, agents, log) {
		return true
	}
	permanent, body, ok := activatedAbilitySource(g, playerID, activate.SourceID, activate.AbilityIndex)
	if !ok {
		return false
	}

	if manaBody, ok := body.(game.ManaAbility); ok && canActivateManaAbility(g, playerID, permanent, &manaBody, activate.AbilityIndex) {
		if len(activate.Targets) != 0 || activate.XValue != 0 {
			return false
		}
		prefs := e.paymentPreferencesForCost(g, playerID, manaCostPtr(manaBody.ManaCost), abilityAdditionalCosts(manaBody.AdditionalCosts), agents, log)
		if !paymentOrch.payAbilityCosts(g, payment.AbilityRequest{
			PlayerID:        playerID,
			Source:          permanent,
			ManaCost:        manaBody.ManaCost,
			AdditionalCosts: abilityAdditionalCosts(manaBody.AdditionalCosts),
			XValue:          0,
			Prefs:           prefs,
		}) {
			return false
		}
		obj := &game.StackObject{
			ID:             g.IDGen.Next(),
			Kind:           game.StackActivatedAbility,
			SourceID:       permanent.ObjectID,
			Face:           permanent.Face,
			SourceCardID:   permanent.CardInstanceID,
			SourceTokenDef: permanent.TokenDef,
			AbilityIndex:   activate.AbilityIndex,
			Controller:     playerID,
		}
		if len(manaBody.Content.Modes) > 0 {
			e.resolveAbilityContentWithChoices(g, obj, manaBody.Content, agents, log)
		}
		recordActivatedAbilityUse(g, permanent.ObjectID, activate.AbilityIndex, manaBody.Timing)
		return true
	}

	card, ok := permanentCardDef(g, permanent)
	if !ok {
		return false
	}
	activatedBody, activatedOK := body.(game.ActivatedAbility)
	loyaltyBody, loyaltyOK := body.(game.LoyaltyAbility)
	if !activatedOK && !loyaltyOK {
		return false
	}
	if activatedOK &&
		!canActivateEquipAbility(g, playerID, permanent, &activatedBody, activate.AbilityIndex, activate.Targets, activate.XValue) &&
		!canActivateGeneralAbility(g, playerID, permanent, &activatedBody, activate.AbilityIndex, activate.Targets, activate.XValue) &&
		!loyaltyOK {
		return false
	}
	if loyaltyOK && !canActivateLoyaltyAbility(g, playerID, permanent, &loyaltyBody, activate.AbilityIndex, activate.Targets, activate.XValue) {
		return false
	}
	completedTargets, ok := e.completeAbilityAnnouncementTargets(g, playerID, card, permanent.ObjectID, body, activate.Targets, agents, log)
	if !ok {
		return false
	}
	activate.Targets = completedTargets
	targetCounts, ok := bodyTargetCounts(g, playerID, card, permanent.ObjectID, body, activate.Targets)
	if !ok {
		panic("validated ability targets could not be segmented")
	}
	if activatedOK &&
		!canActivateEquipAbility(g, playerID, permanent, &activatedBody, activate.AbilityIndex, activate.Targets, activate.XValue) &&
		!canActivateGeneralAbility(g, playerID, permanent, &activatedBody, activate.AbilityIndex, activate.Targets, activate.XValue) &&
		!loyaltyOK {
		return false
	}
	if loyaltyOK && !canActivateLoyaltyAbility(g, playerID, permanent, &loyaltyBody, activate.AbilityIndex, activate.Targets, activate.XValue) {
		return false
	}
	sourceCardID := permanent.CardInstanceID
	sourceTokenDef := permanent.TokenDef
	manaCost := opt.V[cost.Mana]{}
	var additionalCosts []cost.Additional
	var alternativeCosts []cost.Alternative
	timing := game.NoTimingRestriction
	if activatedOK {
		manaCost = activatedBody.ManaCost
		additionalCosts = abilityAdditionalCosts(activatedBody.AdditionalCosts)
		alternativeCosts = append([]cost.Alternative(nil), activatedBody.AlternativeCosts...)
		timing = activatedBody.Timing
	}
	var tapExclusions []id.ID
	if hasTapCostOf(additionalCosts) {
		tapExclusions = append(tapExclusions, permanent.ObjectID)
	}
	prefs := e.paymentPreferencesForCost(g, playerID, manaCostPtr(manaCost), additionalCosts, agents, log, tapExclusions...)
	if !paymentOrch.payAbilityCosts(g, payment.AbilityRequest{
		PlayerID:         playerID,
		Source:           permanent,
		ManaCost:         manaCost,
		AdditionalCosts:  additionalCosts,
		AlternativeCosts: alternativeCosts,
		XValue:           activate.XValue,
		Prefs:            prefs,
	}) {
		return false
	}
	if loyaltyOK {
		applyLoyaltyCost(permanent, loyaltyBody.LoyaltyCost)
	}
	obj := &game.StackObject{
		ID:             g.IDGen.Next(),
		Kind:           game.StackActivatedAbility,
		SourceID:       permanent.ObjectID,
		Face:           permanent.Face,
		SourceCardID:   sourceCardID,
		SourceTokenDef: sourceTokenDef,
		AbilityIndex:   activate.AbilityIndex,
		Controller:     playerID,
		Targets:        append([]game.Target(nil), activate.Targets...),
		TargetCounts:   targetCounts,
		XValue:         activate.XValue,
	}
	if activatedOK {
		obj.InlineActivated = &activatedBody
	}
	if loyaltyOK {
		obj.InlineLoyalty = &loyaltyBody
	}
	pushAbilityToStack(g, obj)
	recordActivatedAbilityUse(g, permanent.ObjectID, activate.AbilityIndex, timing)
	if loyaltyOK {
		recordActivatedAbilityUse(g, permanent.ObjectID, -1, game.OncePerTurn)
	}
	return true
}

func (e *Engine) applyGraveyardAbilityWithChoices(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	card, ability, ok := graveyardAbilitySource(g, playerID, activate.SourceID, activate.AbilityIndex)
	if !ok || !canActivateGraveyardAbility(g, playerID, card.ID, &ability, activate.AbilityIndex, activate.Targets, activate.XValue) {
		return false
	}
	sourceZoneVersion := card.ZoneVersion
	def := cardFaceOrDefault(card, game.FaceFront)
	completedTargets, ok := e.completeAbilityAnnouncementTargets(g, playerID, def, 0, &ability, activate.Targets, agents, log)
	if !ok || !canActivateGraveyardAbility(g, playerID, card.ID, &ability, activate.AbilityIndex, completedTargets, activate.XValue) {
		return false
	}
	targetCounts, ok := bodyTargetCounts(g, playerID, def, 0, &ability, completedTargets)
	if !ok {
		panic("validated graveyard ability targets could not be segmented")
	}
	prefs := e.paymentPreferencesForCost(g, playerID, manaCostPtr(ability.ManaCost), abilityAdditionalCosts(ability.AdditionalCosts), agents, log)
	if !paymentOrch.payAbilityCosts(g, payment.AbilityRequest{
		PlayerID:         playerID,
		SourceCardID:     card.ID,
		SourceZone:       zone.Graveyard,
		ManaCost:         ability.ManaCost,
		AdditionalCosts:  abilityAdditionalCosts(ability.AdditionalCosts),
		AlternativeCosts: append([]cost.Alternative(nil), ability.AlternativeCosts...),
		XValue:           activate.XValue,
		Prefs:            prefs,
	}) {
		return false
	}
	obj := &game.StackObject{
		ID:                g.IDGen.Next(),
		Kind:              game.StackActivatedAbility,
		SourceID:          card.ID,
		SourceCardID:      card.ID,
		SourceZone:        zone.Graveyard,
		SourceZoneVersion: sourceZoneVersion,
		AbilityIndex:      activate.AbilityIndex,
		Controller:        playerID,
		Targets:           append([]game.Target(nil), completedTargets...),
		TargetCounts:      targetCounts,
		XValue:            activate.XValue,
	}
	pushAbilityToStack(g, obj)
	recordActivatedAbilityUse(g, card.ID, activate.AbilityIndex, ability.Timing)
	return true
}

func canActivateLoyaltyAbility(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, body *game.LoyaltyAbility, abilityIndex int, targets []game.Target, xValue int) bool {
	_ = abilityIndex
	if body == nil || !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer || permanent.PhasedOut || effectiveController(g, permanent) != playerID {
		return false
	}
	if xValue != 0 || !bodyFunctionsOnBattlefield(body) || !permanentHasType(g, permanent, types.Planeswalker) {
		return false
	}
	if !isSorcerySpeed(g, playerID) || g.ActivatedAbilitiesThisTurn[game.ActivatedAbilityUse{SourceID: permanent.ObjectID, AbilityIndex: -1}] {
		return false
	}
	if body.LoyaltyCost < 0 && permanent.Counters.Get(counter.Loyalty) < -body.LoyaltyCost {
		return false
	}
	if !activationConditionSatisfied(g, playerID, permanent, body.ActivationCondition) {
		return false
	}
	card, ok := permanentCardDef(g, permanent)
	if !ok || !targetsValidForBodyFromSourceObject(g, playerID, card, permanent.ObjectID, body, targets) {
		return false
	}
	return paymentOrch.buildAbilityCostPlan(g, payment.AbilityRequest{PlayerID: playerID, Source: permanent, XValue: xValue})
}

func applyLoyaltyCost(permanent *game.Permanent, loyalty int) {
	if loyalty >= 0 {
		permanent.Counters.Add(counter.Loyalty, loyalty)
		return
	}
	permanent.Counters.Remove(counter.Loyalty, -loyalty)
}

func (e *Engine) applyCyclingAbility(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction) bool {
	return e.applyCyclingAbilityWithChoices(g, playerID, activate, [game.NumPlayers]PlayerAgent{}, nil)
}

func (e *Engine) applyCyclingAbilityWithChoices(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	card, ability, ok := cyclingAbilitySource(g, playerID, activate.SourceID, activate.AbilityIndex)
	if !ok {
		return false
	}
	if !canActivateCyclingAbility(g, playerID, activate.SourceID, &ability, activate.AbilityIndex, activate.Targets, activate.XValue) {
		return false
	}
	prefs := e.paymentPreferencesForCost(g, playerID, manaCostPtr(ability.ManaCost), nil, agents, log)
	if !paymentOrch.payGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: manaCostPtr(ability.ManaCost), XValue: activate.XValue, Prefs: prefs}) {
		return false
	}
	if !discardCardFromHand(g, playerID, card.ID) {
		panic("cycling card disappeared from hand after validation")
	}
	obj := &game.StackObject{
		ID:                  g.IDGen.Next(),
		Kind:                game.StackActivatedAbility,
		SourceID:            card.ID,
		SourceCardID:        card.ID,
		AbilityIndex:        activate.AbilityIndex,
		Controller:          playerID,
		Targets:             append([]game.Target(nil), activate.Targets...),
		XValue:              activate.XValue,
		AdditionalCostsPaid: []string{"Discard this card"},
	}
	pushAbilityToStack(g, obj)
	return true
}

func (e *Engine) applyNinjutsuAbilityWithChoices(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	card, ability, ok := handActivatedAbilitySource(g, playerID, activate.SourceID, activate.AbilityIndex)
	if !ok || !canActivateNinjutsuAbility(g, playerID, activate.SourceID, &ability, activate.AbilityIndex, activate.Targets, activate.XValue) {
		return false
	}
	attacker := chooseNinjutsuAttacker(e, g, playerID, unblockedAttackers(g, playerID), agents, log)
	if attacker == nil {
		return false
	}
	attackTarget, ok := attackTargetForAttacker(g, attacker.ObjectID)
	if !ok || attackerWasBlocked(g, attacker.ObjectID) {
		return false
	}
	prefs := e.paymentPreferencesForCost(g, playerID, manaCostPtr(ability.ManaCost), nil, agents, log)
	if !paymentOrch.payGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: manaCostPtr(ability.ManaCost), Prefs: prefs}) {
		return false
	}
	removePermanentFromCombat(g, attacker.ObjectID)
	if !movePermanentToZone(g, attacker, zone.Hand) {
		panic("Ninjutsu attacker disappeared after validation")
	}
	pushAbilityToStack(g, &game.StackObject{
		ID:                   g.IDGen.Next(),
		Kind:                 game.StackActivatedAbility,
		SourceID:             card.ID,
		SourceCardID:         card.ID,
		SourceZone:           zone.Hand,
		SourceZoneVersion:    card.ZoneVersion,
		AbilityIndex:         activate.AbilityIndex,
		Controller:           playerID,
		Ninjutsu:             true,
		NinjutsuAttackTarget: attackTarget,
		AdditionalCostsPaid:  []string{"Return an unblocked attacker you control to its owner's hand"},
	})
	return true
}

func chooseNinjutsuAttacker(e *Engine, g *game.Game, playerID game.PlayerID, attackers []*game.Permanent, agents [game.NumPlayers]PlayerAgent, log *TurnLog) *game.Permanent {
	if len(attackers) == 0 {
		return nil
	}
	if len(attackers) == 1 {
		return attackers[0]
	}
	options := make([]game.ChoiceOption, 0, len(attackers))
	for i, attacker := range attackers {
		options = append(options, game.ChoiceOption{Index: i, Label: permanentEffectiveName(g, attacker)})
	}
	selected := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind:             game.ChoicePayment,
		Player:           playerID,
		Prompt:           "Choose an unblocked attacker to return",
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}, log)
	if len(selected) != 1 || selected[0] < 0 || selected[0] >= len(attackers) {
		return nil
	}
	return attackers[selected[0]]
}

func unblockedAttackers(g *game.Game, playerID game.PlayerID) []*game.Permanent {
	if g.Combat == nil ||
		g.Turn.Phase != game.PhaseCombat ||
		g.Turn.Step < game.StepDeclareBlockers ||
		g.Turn.Step > game.StepEndOfCombat {
		return nil
	}
	var attackers []*game.Permanent
	for _, attack := range g.Combat.Attackers {
		permanent, ok := permanentByObjectID(g, attack.Attacker)
		if !ok || effectiveController(g, permanent) != playerID || attackerWasBlocked(g, attack.Attacker) {
			continue
		}
		attackers = append(attackers, permanent)
	}
	return attackers
}

func attackTargetForAttacker(g *game.Game, attackerID id.ID) (game.AttackTarget, bool) {
	if g.Combat == nil {
		return game.AttackTarget{}, false
	}
	for _, attack := range g.Combat.Attackers {
		if attack.Attacker == attackerID {
			return attack.Target, true
		}
	}
	return game.AttackTarget{}, false
}

func canActivateNinjutsuAbility(g *game.Game, playerID game.PlayerID, cardID id.ID, body *game.ActivatedAbility, abilityIndex int, targets []game.Target, xValue int) bool {
	if body == nil || !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return false
	}
	if xValue != 0 ||
		abilityIndex < 0 ||
		!game.BodyHasKeyword(body, game.Ninjutsu) ||
		game.BodyFunctionZone(body) != zone.Hand ||
		body.Timing != game.DuringCombat ||
		len(targets) != 0 ||
		len(game.BodyTargets(body)) != 0 ||
		!abilityHasReturnUnblockedAttackerCost(body.AdditionalCosts) ||
		len(unblockedAttackers(g, playerID)) == 0 {
		return false
	}
	_, gotAbility, ok := handActivatedAbilitySource(g, playerID, cardID, abilityIndex)
	if !ok || !game.BodyHasKeyword(gotAbility, game.Ninjutsu) {
		return false
	}
	return paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: manaCostPtr(body.ManaCost)})
}

func abilityHasReturnUnblockedAttackerCost(costs []cost.Additional) bool {
	return len(costs) == 1 &&
		costs[0].Kind == cost.AdditionalReturnUnblockedAttacker &&
		payment.AdditionalCostAmount(costs[0]) == 1
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

func (*Engine) canCastSpellFaceFromZoneWithKicker(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, targets []game.Target, xValue int, chosenModes []int, kickerPaid bool) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return false
	}
	if xValue < 0 {
		return false
	}
	player := g.Players[playerID]
	card, ok := g.GetCardInstance(cardID)
	if !ok || !castSourceContains(player, cardID, sourceZone) {
		return false
	}
	spellDef, ok := cardFaceDef(card, face)
	if !ok || !card.Def.CanChooseCastFace(face) {
		return false
	}
	switch sourceZone {
	case zone.Command:
		if player.CommanderInstanceID != cardID {
			return false
		}
	case zone.Hand:
	case zone.Graveyard:
		if !canCastFromZoneByRuleEffect(g, playerID, cardID, sourceZone, face) {
			return false
		}
	case zone.Exile:
		if !g.AdventureCards[cardID] {
			return false
		}
		if face != game.FaceFront {
			return false
		}
	default:
		return false
	}
	if xValue != 0 && !costHasVariableMana(manaCostPtr(spellDef.ManaCost)) {
		return false
	}
	if !modesValidForSpell(spellDef, chosenModes) || !isSupportedSpell(spellDef) || !targetsValidForSpell(g, playerID, spellDef, chosenModes, targets) {
		return false
	}
	if !canCastAtCurrentTiming(g, playerID, spellDef) {
		return false
	}
	if kickerPaid && !spellHasKicker(spellDef) {
		return false
	}
	if !paymentOrch.canPaySpellCosts(g, payment.SpellRequest{PlayerID: playerID, CardID: card.ID, SourceZone: sourceZone, Card: spellDef, XValue: xValue, KickerPaid: kickerPaid}) {
		return false
	}
	return true
}

func legalCastFacesForZone(g *game.Game, playerID game.PlayerID, card *game.CardInstance, sourceZone zone.Type) []game.FaceIndex {
	if sourceZone == zone.Graveyard {
		var faces []game.FaceIndex
		for _, face := range card.Def.LegalCastFaces() {
			if canCastFromZoneByRuleEffect(g, playerID, card.ID, sourceZone, face) {
				faces = append(faces, face)
			}
		}
		return faces
	}
	return card.Def.LegalCastFaces()
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
	default:
		return false
	}
}

func canAct(g *game.Game, playerID game.PlayerID) bool {
	return isPlayerAlive(g, playerID)
}

func canPlayAnyLand(g *game.Game, playerID game.PlayerID) bool {
	return canAct(g, playerID) &&
		playerID == g.Turn.ActivePlayer &&
		playerID == g.Turn.PriorityPlayer &&
		isSorcerySpeed(g, playerID) &&
		g.Turn.CanPlayLand()
}

func canCastAtCurrentTiming(g *game.Game, playerID game.PlayerID, card *game.CardDef) bool {
	if card.HasType(types.Instant) || card.HasKeyword(game.Flash) {
		return true
	}
	return isSorcerySpeed(g, playerID)
}

func legalXValuesForCost(g *game.Game, playerID game.PlayerID, manaCost *cost.Mana) []int {
	if !costHasVariableMana(manaCost) {
		return []int{0}
	}
	var values []int
	for x := 0; x <= maxLegalXValue; x++ {
		if !paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: manaCost, XValue: x}) {
			break
		}
		values = append(values, x)
	}
	return values
}

func costHasVariableMana(manaCost *cost.Mana) bool {
	if manaCost == nil {
		return false
	}
	for _, symbol := range *manaCost {
		if symbol.Kind == cost.VariableSymbol {
			return true
		}
	}
	return false
}

func activatedAbilitySource(g *game.Game, playerID game.PlayerID, sourceID id.ID, abilityIndex int) (*game.Permanent, game.Ability, bool) {
	if abilityIndex < 0 {
		return nil, nil, false
	}
	permanent, ok := permanentByObjectID(g, sourceID)
	if !ok || permanent.PhasedOut || effectiveController(g, permanent) != playerID {
		return nil, nil, false
	}
	abilities := permanentEffectiveAbilities(g, permanent)
	if abilityIndex >= len(abilities) {
		return nil, nil, false
	}
	return permanent, abilities[abilityIndex], true
}

func cyclingAbilitySource(g *game.Game, playerID game.PlayerID, sourceID id.ID, abilityIndex int) (*game.CardInstance, game.ActivatedAbility, bool) {
	return handActivatedAbilitySource(g, playerID, sourceID, abilityIndex)
}

func handActivatedAbilitySource(g *game.Game, playerID game.PlayerID, sourceID id.ID, abilityIndex int) (*game.CardInstance, game.ActivatedAbility, bool) {
	if abilityIndex < 0 {
		return nil, game.ActivatedAbility{}, false
	}

	player, ok := playerByID(g, playerID)
	if !ok || !player.Hand.Contains(sourceID) {
		return nil, game.ActivatedAbility{}, false
	}
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		return nil, game.ActivatedAbility{}, false
	}
	frontDef := cardFaceOrDefault(card, game.FaceFront)
	body, ok := frontDef.BodyAt(abilityIndex).(game.ActivatedAbility)
	if !ok {
		return nil, game.ActivatedAbility{}, false
	}
	return card, body, true
}

func graveyardAbilitySource(g *game.Game, playerID game.PlayerID, sourceID id.ID, abilityIndex int) (*game.CardInstance, game.ActivatedAbility, bool) {
	if abilityIndex < 0 {
		return nil, game.ActivatedAbility{}, false
	}
	player, ok := playerByID(g, playerID)
	if !ok || !player.Graveyard.Contains(sourceID) {
		return nil, game.ActivatedAbility{}, false
	}
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		return nil, game.ActivatedAbility{}, false
	}
	frontDef := cardFaceOrDefault(card, game.FaceFront)
	body, ok := frontDef.BodyAt(abilityIndex).(game.ActivatedAbility)
	if !ok {
		return nil, game.ActivatedAbility{}, false
	}
	return card, body, true
}

func canActivateEquipAbility(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, body *game.ActivatedAbility, abilityIndex int, targets []game.Target, xValue int) bool {
	if body == nil || !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer || permanent.PhasedOut || effectiveController(g, permanent) != playerID {
		return false
	}
	if xValue != 0 || !bodyFunctionsOnBattlefield(body) || !isEquipmentPermanent(g, permanent) {
		return false
	}
	if !game.BodyHasKeyword(body, game.Equip) && body.Timing != game.SorceryOnly {
		return false
	}
	if !isSorcerySpeed(g, playerID) || abilityHasNonTapAdditionalCosts(body.AdditionalCosts) || activatedAbilityUsedThisTurn(g, permanent.ObjectID, abilityIndex, body.Timing) {
		return false
	}
	if !activationConditionSatisfied(g, playerID, permanent, body.ActivationCondition) {
		return false
	}
	card, ok := permanentCardDef(g, permanent)
	if !ok || !targetsValidForBodyFromSourceObject(g, playerID, card, permanent.ObjectID, body, targets) {
		return false
	}
	if len(targets) != 1 || targets[0].Kind != game.TargetPermanent {
		return false
	}
	target, ok := permanentByObjectID(g, targets[0].PermanentID)
	if !ok || effectiveController(g, target) != playerID || !canAttachPermanent(g, permanent, target) {
		return false
	}
	return paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: manaCostPtr(body.ManaCost)})
}

func canActivateGeneralAbility(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, body *game.ActivatedAbility, abilityIndex int, targets []game.Target, xValue int) bool {
	if body == nil || !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer || permanent.PhasedOut || effectiveController(g, permanent) != playerID {
		return false
	}
	if game.BodyHasKeyword(body, game.Equip) || !bodyFunctionsOnBattlefield(body) {
		return false
	}
	if !activatedAbilityTimingAllows(g, playerID, body.Timing) || activatedAbilityUsedThisTurn(g, permanent.ObjectID, abilityIndex, body.Timing) {
		return false
	}
	if !activationConditionSatisfied(g, playerID, permanent, body.ActivationCondition) {
		return false
	}
	card, ok := permanentCardDef(g, permanent)
	if !ok || !targetsValidForBodyFromSourceObject(g, playerID, card, permanent.ObjectID, body, targets) {
		return false
	}
	return paymentOrch.buildAbilityCostPlan(g, payment.AbilityRequest{
		PlayerID:         playerID,
		Source:           permanent,
		ManaCost:         body.ManaCost,
		AdditionalCosts:  abilityAdditionalCosts(body.AdditionalCosts),
		AlternativeCosts: append([]cost.Alternative(nil), body.AlternativeCosts...),
		XValue:           xValue,
	})
}

func canActivateCyclingAbility(g *game.Game, playerID game.PlayerID, cardID id.ID, body *game.ActivatedAbility, abilityIndex int, targets []game.Target, xValue int) bool {
	if body == nil || !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return false
	}

	if xValue != 0 || abilityIndex < 0 || !game.BodyHasKeyword(body, game.Cycling) {
		return false
	}
	if body.Timing != game.NoTimingRestriction || !abilityHasDiscardThisCardCost(body.AdditionalCosts) {
		return false
	}
	if len(targets) != 0 || len(game.BodyTargets(body)) != 0 {
		return false
	}
	_, gotAbility, ok := cyclingAbilitySource(g, playerID, cardID, abilityIndex)
	if !ok || !game.BodyHasKeyword(gotAbility, game.Cycling) {
		return false
	}
	return paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: manaCostPtr(body.ManaCost)})
}

func canActivateGraveyardAbility(g *game.Game, playerID game.PlayerID, cardID id.ID, body *game.ActivatedAbility, abilityIndex int, targets []game.Target, xValue int) bool {
	if body == nil || !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return false
	}
	if game.BodyFunctionZone(body) != zone.Graveyard {
		return false
	}
	if !activatedAbilityTimingAllows(g, playerID, body.Timing) || activatedAbilityUsedThisTurn(g, cardID, abilityIndex, body.Timing) {
		return false
	}
	card, _, ok := graveyardAbilitySource(g, playerID, cardID, abilityIndex)
	if !ok {
		return false
	}
	def := cardFaceOrDefault(card, game.FaceFront)
	if !targetsValidForBodyFromSourceObject(g, playerID, def, 0, body, targets) {
		return false
	}
	return paymentOrch.buildAbilityCostPlan(g, payment.AbilityRequest{
		PlayerID:         playerID,
		SourceCardID:     cardID,
		SourceZone:       zone.Graveyard,
		ManaCost:         body.ManaCost,
		AdditionalCosts:  abilityAdditionalCosts(body.AdditionalCosts),
		AlternativeCosts: append([]cost.Alternative(nil), body.AlternativeCosts...),
		XValue:           xValue,
	})
}

func canActivateManaAbility(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, body *game.ManaAbility, abilityIndex int) bool {
	if body == nil || !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer || permanent.PhasedOut || effectiveController(g, permanent) != playerID {
		return false
	}
	if !bodyFunctionsOnBattlefield(body) {
		return false
	}
	if len(game.BodyTargets(body)) != 0 || !manaBodyHasAddManaEffect(body) || !manaBodyChoicesAvailable(g, playerID, body) {
		return false
	}
	if !activatedAbilityTimingAllows(g, playerID, body.Timing) ||
		activatedAbilityUsedThisTurn(g, permanent.ObjectID, abilityIndex, body.Timing) {
		return false
	}
	if !activationConditionSatisfied(g, playerID, permanent, body.ActivationCondition) {
		return false
	}
	return paymentOrch.buildAbilityCostPlan(g, payment.AbilityRequest{
		PlayerID:        playerID,
		Source:          permanent,
		ManaCost:        body.ManaCost,
		AdditionalCosts: abilityAdditionalCosts(body.AdditionalCosts),
	})
}

func manaBodyHasAddManaEffect(body *game.ManaAbility) bool {
	if body == nil {
		return false
	}
	if sequence, ok := manaBodyInstructionSequence(body); ok {
		hasAddMana := false
		for i := range sequence {
			if sequence[i].Primitive == nil {
				return false
			}
			switch sequence[i].Primitive.Kind() {
			case game.PrimitiveAddMana:
				hasAddMana = true
			case game.PrimitiveChoose:
				choice, ok := sequence[i].Primitive.(game.Choose)
				if !ok || choice.Choice.Kind != game.ResolutionChoiceMana {
					return false
				}
			default:
				return false
			}
		}
		return hasAddMana
	}
	return false
}

func manaBodyChoicesAvailable(g *game.Game, playerID game.PlayerID, body *game.ManaAbility) bool {
	if body == nil {
		return false
	}
	if sequence, ok := manaBodyInstructionSequence(body); ok {
		for i := range sequence {
			if sequence[i].Primitive == nil {
				return false
			}
			if sequence[i].Primitive.Kind() != game.PrimitiveChoose {
				continue
			}
			primitive, ok := sequence[i].Primitive.(game.Choose)
			if !ok {
				return false
			}
			choicePlayer := resolutionChoicePlayer(playerID, &primitive.Choice)
			_, values := resolutionChoiceOptions(g, choicePlayer, &primitive.Choice)
			if len(values) == 0 {
				return false
			}
		}
		return true
	}
	return true
}

func manaBodyInstructionSequence(body *game.ManaAbility) ([]game.Instruction, bool) {
	if body == nil {
		return nil, false
	}
	if len(body.Content.Modes) == 0 || body.Content.IsModal() {
		return nil, false
	}
	return body.Content.Modes[0].Sequence, true
}

func hasTapCostOf(additionalCosts []cost.Additional) bool {
	for _, addCost := range additionalCosts {
		if addCost.Kind == cost.AdditionalTap {
			return true
		}
	}
	return false
}

func abilityHasNonTapAdditionalCosts(additionalCosts []cost.Additional) bool {
	for _, addCost := range additionalCosts {
		if addCost.Kind != cost.AdditionalTap {
			return true
		}
	}
	return false
}

func abilityHasDiscardThisCardCost(costs []cost.Additional) bool {
	if len(costs) != 1 {
		return false
	}
	addCost := costs[0]
	if addCost.Kind != cost.AdditionalDiscard || payment.AdditionalCostAmount(addCost) != 1 {
		return false
	}
	if addCost.Text != "" {
		return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(addCost.Text)), ".") == "discard this card"
	}
	return addCost.Source == zone.Hand
}

func canTapPermanentForAbility(g *game.Game, permanent *game.Permanent) bool {
	if permanent.Tapped {
		return false
	}
	return !permanentHasType(g, permanent, types.Creature) || !permanent.SummoningSick
}

func tapPermanentForAbility(g *game.Game, permanent *game.Permanent) bool {
	if !canTapPermanentForAbility(g, permanent) {
		return false
	}
	setPermanentTapped(g, permanent, true)
	return true
}

func activatedAbilityTimingAllows(g *game.Game, playerID game.PlayerID, timing game.TimingRestriction) bool {
	switch timing {
	case game.NoTimingRestriction, game.OncePerTurn:
		return true
	case game.SorceryOnly, game.SorceryOncePerTurn:
		return isSorcerySpeed(g, playerID)
	case game.DuringCombat:
		return g.Turn.Phase == game.PhaseCombat
	case game.DuringUpkeep:
		return g.Turn.ActivePlayer == playerID &&
			g.Turn.Phase == game.PhaseBeginning &&
			g.Turn.Step == game.StepUpkeep
	default:
		return false
	}
}

func activatedAbilityUsedThisTurn(g *game.Game, sourceID id.ID, abilityIndex int, timing game.TimingRestriction) bool {
	if !abilityHasOncePerTurnRestriction(timing) {
		return false
	}
	return g.ActivatedAbilitiesThisTurn[game.ActivatedAbilityUse{
		SourceID:     sourceID,
		AbilityIndex: abilityIndex,
	}]
}

func recordActivatedAbilityUse(g *game.Game, sourceID id.ID, abilityIndex int, timing game.TimingRestriction) {
	if !abilityHasOncePerTurnRestriction(timing) {
		return
	}
	if g.ActivatedAbilitiesThisTurn == nil {
		g.ActivatedAbilitiesThisTurn = make(map[game.ActivatedAbilityUse]bool)
	}
	g.ActivatedAbilitiesThisTurn[game.ActivatedAbilityUse{SourceID: sourceID, AbilityIndex: abilityIndex}] = true
}

func abilityHasOncePerTurnRestriction(timing game.TimingRestriction) bool {
	return timing == game.OncePerTurn || timing == game.SorceryOncePerTurn
}

func isSorcerySpeed(g *game.Game, playerID game.PlayerID) bool {
	return playerID == g.Turn.ActivePlayer &&
		g.Turn.IsMainPhase() &&
		g.Turn.Step == game.StepNone &&
		g.Stack.IsEmpty()
}

func landCardInstance(g *game.Game, player *game.Player, cardID id.ID) (*game.CardInstance, bool) {
	return landCardInstanceFace(g, player, cardID, game.FaceFront)
}

func landCardInstanceFace(g *game.Game, player *game.Player, cardID id.ID, face game.FaceIndex) (*game.CardInstance, bool) {
	if !player.Hand.Contains(cardID) {
		return nil, false
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok || !card.Def.CanChooseLandFace(face) {
		return nil, false
	}
	return card, true
}

func entersSummoningSick(card *game.CardDef) bool {
	return !card.HasKeyword(game.Haste)
}

func isSupportedSpell(card *game.CardDef) bool {
	return !card.HasType(types.Land) &&
		(card.IsPermanent() ||
			card.HasType(types.Instant) ||
			card.HasType(types.Sorcery))
}
