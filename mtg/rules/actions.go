package rules

import (
	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/rules/payment"
)

const maxLegalXValue = 20

func canPayCost(g *game.Game, playerID game.PlayerID, manaCost *cost.Mana) bool {
	return paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: manaCost})
}

func canPayCostWithX(g *game.Game, playerID game.PlayerID, manaCost *cost.Mana, xValue int) bool {
	return paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: manaCost, XValue: xValue})
}

func stackManaValue(card *game.CardDef, xValue int) int {
	if card == nil || !card.ManaCost.Exists {
		return 0
	}
	total := 0
	for _, symbol := range card.ManaCost.Val {
		if symbol.Kind == cost.VariableSymbol {
			total += xValue
			continue
		}
		total += cost.Mana{symbol}.ManaValue()
	}
	return total
}

func (e *Engine) legalActions(g *game.Game, playerID game.PlayerID) []action.Action {
	// Generating legal actions is a pure read that repeatedly evaluates every
	// permanent, so a static-source frame lets the rules layer scan the
	// battlefield for static-ability sources once instead of per permanent.
	g.BeginStaticSourceFrame()
	defer g.EndStaticSourceFrame()
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
				for _, xValue := range legalXValuesForCostAndAdditional(g, playerID, manaCostPtr(body.ManaCost), body.AdditionalCosts) {
					for _, modes := range modeChoicesForBody(&body) {
						targetResult := targetChoicesForBodyFromSourceObjectWithModes(g, playerID, card, permanent.ObjectID, &body, modes)
						if targetResult.kind == targetInvalidSpec {
							continue
						}
						for choiceIndex, targets := range targetResult.choices {
							if canActivateEquipAbilityWithModes(g, playerID, permanent, &body, idx, targets, xValue, modes) ||
								canActivateGeneralAbilityWithModes(g, playerID, permanent, &body, idx, targets, xValue, modes) {
								actions = append(actions, actionBuild.activateAbilityWithModes(permanent.ObjectID, idx, append([]game.Target(nil), targets...), targetResult.targetCounts[choiceIndex], xValue, modes))
							}
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
			for choiceIndex, targets := range targetResult.choices {
				if canActivateLoyaltyAbility(g, playerID, permanent, &body, idx, targets, 0) {
					actions = append(actions, actionBuild.activateAbilityWithModes(permanent.ObjectID, idx, append([]game.Target(nil), targets...), targetResult.targetCounts[choiceIndex], 0, nil))
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
			for _, xValue := range legalXValuesForCostAndAdditional(g, playerID, manaCostPtr(body.ManaCost), body.AdditionalCosts) {
				for _, modes := range modeChoicesForBody(body) {
					targetResult := targetChoicesForBodyFromSourceObjectWithModes(g, playerID, def, 0, body, modes)
					if targetResult.kind == targetInvalidSpec {
						continue
					}
					for choiceIndex, targets := range targetResult.choices {
						if canActivateGraveyardAbilityWithModes(g, playerID, cardID, body, idx, targets, xValue, modes) {
							actions = append(actions, actionBuild.activateAbilityWithModes(cardID, idx, append([]game.Target(nil), targets...), targetResult.targetCounts[choiceIndex], xValue, modes))
						}
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
		effectiveAbilities := effectiveHandActivatedAbilities(g, playerID, card)
		for i := range effectiveAbilities {
			effective := &effectiveAbilities[i]
			if canActivateCyclingAbility(g, playerID, cardID, &effective.body, effective.index, nil, 0) {
				actions = append(actions, actionBuild.activateAbility(cardID, effective.index, nil, 0))
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
