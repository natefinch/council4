package rules

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules/payment"
)

const maxLegalXValue = 20

func canPayCost(g *game.Game, playerID game.PlayerID, cost *mana.Cost) bool {
	return paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: cost})
}

func canPayCostWithX(g *game.Game, playerID game.PlayerID, cost *mana.Cost, xValue int) bool {
	return paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: cost, XValue: xValue})
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
	actions = append(actions, e.legalFaceDownCastActions(g, playerID)...)
	actions = append(actions, e.legalCommanderCastActions(g, playerID)...)
	actions = append(actions, e.legalActivateAbilityActions(g, playerID)...)
	actions = append(actions, e.legalCyclingActions(g, playerID)...)
	actions = append(actions, e.legalSuspendActions(g, playerID)...)
	actions = append(actions, e.legalTurnFaceUpActions(g, playerID)...)
	actions = append(actions, actionBuild.pass())
	return actions
}

func normalizedCastSourceZone(cast action.CastSpellAction) game.ZoneType {
	if cast.SourceZone == game.ZoneNone {
		return game.ZoneHand
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
			for _, face := range legalCastFacesForZone(card.Def, sourceZone) {
				spellDef := cardFaceOrDefault(card, face)
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
							if sourceZone == game.ZoneHand && spellHasKicker(spellDef) && e.canCastSpellFaceFromZoneWithKicker(g, playerID, cardID, sourceZone, face, targets, xValue, modes, true) {
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
	for _, face := range legalCastFacesForZone(card.Def, game.ZoneCommand) {
		spellDef := cardFaceOrDefault(card, face)
		for _, xValue := range legalXValuesForCost(g, playerID, manaCostPtr(spellDef.ManaCost)) {
			for _, modes := range modeChoicesForSpell(spellDef) {
				targetResult := targetChoicesForSpell(g, playerID, spellDef, modes)
				if targetResult.kind == targetInvalidSpec {
					continue
				}
				for _, targets := range targetResult.choices {
					if e.canCastSpellFaceFromZoneWithKicker(g, playerID, card.ID, game.ZoneCommand, face, targets, xValue, modes, false) {
						actions = append(actions, actionBuild.castSpell(card.ID, game.ZoneCommand, face, targets, xValue, modes))
					}
					if spellHasKicker(spellDef) && e.canCastSpellFaceFromZoneWithKicker(g, playerID, card.ID, game.ZoneCommand, face, targets, xValue, modes, true) {
						actions = append(actions, actionBuild.castKickedSpell(card.ID, game.ZoneCommand, face, targets, xValue, modes))
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
		for i := range card.Abilities {
			ability := &card.Abilities[i]
			if canActivateManaAbility(g, playerID, permanent, ability, i) {
				actions = append(actions, actionBuild.activateAbility(permanent.ObjectID, i, nil, 0))
				continue
			}
			for _, xValue := range legalXValuesForCost(g, playerID, manaCostPtr(ability.ManaCost)) {
				targetResult := targetChoicesForAbilityFromSourceObject(g, playerID, card, permanent.ObjectID, ability)
				if targetResult.kind == targetInvalidSpec {
					continue
				}
				for _, targets := range targetResult.choices {
					if canActivateEquipAbility(g, playerID, permanent, ability, i, targets, xValue) ||
						canActivateLoyaltyAbility(g, playerID, permanent, ability, i, targets, xValue) ||
						canActivateGeneralAbility(g, playerID, permanent, ability, i, targets, xValue) {
						actions = append(actions, actionBuild.activateAbility(permanent.ObjectID, i, append([]game.Target(nil), targets...), xValue))
					}
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
		for i := range def.Abilities {
			ability := &def.Abilities[i]
			for _, xValue := range legalXValuesForCost(g, playerID, manaCostPtr(ability.ManaCost)) {
				targetResult := targetChoicesForAbilityFromSourceObject(g, playerID, def, 0, ability)
				if targetResult.kind == targetInvalidSpec {
					continue
				}
				for _, targets := range targetResult.choices {
					if canActivateGraveyardAbility(g, playerID, cardID, ability, i, targets, xValue) {
						actions = append(actions, actionBuild.activateAbility(cardID, i, append([]game.Target(nil), targets...), xValue))
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
		card, ok := permanentCardDef(g, permanent)
		if !ok {
			continue
		}
		for i := range card.Abilities {
			if canActivateManaAbility(g, playerID, permanent, &card.Abilities[i], i) {
				actions = append(actions, actionBuild.activateAbility(permanent.ObjectID, i, nil, 0))
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
		for i := range frontDef.Abilities {
			ability := &frontDef.Abilities[i]
			if canActivateCyclingAbility(g, playerID, cardID, ability, i, nil, 0) {
				actions = append(actions, actionBuild.activateAbility(cardID, i, nil, 0))
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

	if _, ok := createCardPermanentFaceWithChoices(e, g, card, playerID, game.ZoneHand, face, agents, log); !ok {
		return false
	}
	g.Turn.LandsPlayedThisTurn++
	return true
}

func (e *Engine) applyCastSpell(g *game.Game, playerID game.PlayerID, cast action.CastSpellAction) bool {
	return e.applyCastSpellWithChoices(g, playerID, cast, [game.NumPlayers]PlayerAgent{}, nil)
}

func (e *Engine) applyCastSpellWithChoices(g *game.Game, playerID game.PlayerID, cast action.CastSpellAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	sourceZone := normalizedCastSourceZone(cast)
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
	prefs := e.paymentPreferencesForSpellFromZone(g, playerID, card.ID, sourceZone, spellDef, cast.XValue, agents, log)
	additionalCostsPaid, ok := paymentOrch.paySpellCosts(g, payment.SpellRequest{PlayerID: playerID, CardID: card.ID, SourceZone: sourceZone, Card: spellDef, XValue: cast.XValue, KickerPaid: cast.KickerPaid, Prefs: prefs})
	if !ok {
		return false
	}
	if !removeCastSourceCard(player, cast.CardID, sourceZone) {
		panic("cast spell disappeared from source zone after validation")
	}
	if sourceZone == game.ZoneCommand && player.CommanderInstanceID == cast.CardID {
		player.CommanderCastCount++
	}
	obj := &game.StackObject{
		ID:                  g.IDGen.Next(),
		Kind:                game.StackSpell,
		SourceID:            cast.CardID,
		Face:                cast.Face,
		Controller:          playerID,
		Targets:             append([]game.Target(nil), cast.Targets...),
		ChosenModes:         append([]int(nil), cast.ChosenModes...),
		XValue:              cast.XValue,
		KickerPaid:          cast.KickerPaid,
		Flashback:           sourceZone == game.ZoneGraveyard && spellDef.HasKeyword(game.Flashback),
		AdditionalCostsPaid: additionalCostsPaid,
	}
	stormCopies := stormCopyCount(g, spellDef)
	pushSpellToStack(g, obj, game.GameEvent{
		SourceID:      cast.CardID,
		StackObjectID: obj.ID,
		Controller:    playerID,
		CardID:        cast.CardID,
		Face:          cast.Face,
		CardTypes:     cardTypes(spellDef),
		FromZone:      sourceZone,
		ToZone:        game.ZoneStack,
	})
	createStormCopies(g, obj, stormCopies)
	e.resolveCascadeForCast(g, obj, spellDef, agents, log)
	return true
}

func (e *Engine) applyActivateAbility(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction) bool {
	return e.applyActivateAbilityWithChoices(g, playerID, activate, [game.NumPlayers]PlayerAgent{}, nil)
}

func (e *Engine) applyActivateAbilityWithChoices(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if e.applyCyclingAbilityWithChoices(g, playerID, activate, agents, log) {
		return true
	}
	if e.applyGraveyardAbilityWithChoices(g, playerID, activate, agents, log) {
		return true
	}
	permanent, ability, ok := activatedAbilitySource(g, playerID, activate.SourceID, activate.AbilityIndex)
	if !ok {
		return false
	}

	if canActivateManaAbility(g, playerID, permanent, ability, activate.AbilityIndex) {
		if len(activate.Targets) != 0 || activate.XValue != 0 {
			return false
		}
		prefs := e.paymentPreferencesForCost(g, playerID, manaCostPtr(ability.ManaCost), abilityAdditionalCosts(ability), agents, log)
		if !paymentOrch.payAbilityCosts(g, payment.AbilityRequest{PlayerID: playerID, Source: permanent, Ability: ability, XValue: 0, Prefs: prefs}) {
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
		for i := range ability.Effects {
			e.resolveEffectWithChoices(g, obj, &ability.Effects[i], agents, log)
		}
		recordActivatedAbilityUse(g, permanent.ObjectID, activate.AbilityIndex, ability)
		return true
	}

	card, ok := permanentCardDef(g, permanent)
	if !ok {
		return false
	}
	if !canActivateEquipAbility(g, playerID, permanent, ability, activate.AbilityIndex, activate.Targets, activate.XValue) &&
		!canActivateLoyaltyAbility(g, playerID, permanent, ability, activate.AbilityIndex, activate.Targets, activate.XValue) &&
		!canActivateGeneralAbility(g, playerID, permanent, ability, activate.AbilityIndex, activate.Targets, activate.XValue) {
		return false
	}
	completedTargets, ok := e.completeAbilityAnnouncementTargets(g, playerID, card, permanent.ObjectID, ability, activate.Targets, agents, log)
	if !ok {
		return false
	}
	activate.Targets = completedTargets
	if !canActivateEquipAbility(g, playerID, permanent, ability, activate.AbilityIndex, activate.Targets, activate.XValue) &&
		!canActivateLoyaltyAbility(g, playerID, permanent, ability, activate.AbilityIndex, activate.Targets, activate.XValue) &&
		!canActivateGeneralAbility(g, playerID, permanent, ability, activate.AbilityIndex, activate.Targets, activate.XValue) {
		return false
	}
	sourceCardID := permanent.CardInstanceID
	sourceTokenDef := permanent.TokenDef
	prefs := e.paymentPreferencesForCost(g, playerID, manaCostPtr(ability.ManaCost), abilityAdditionalCosts(ability), agents, log)
	if !paymentOrch.payAbilityCosts(g, payment.AbilityRequest{PlayerID: playerID, Source: permanent, Ability: ability, XValue: activate.XValue, Prefs: prefs}) {
		return false
	}
	if ability.IsLoyaltyAbility {
		applyLoyaltyCost(permanent, ability.LoyaltyCost)
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
		XValue:         activate.XValue,
	}
	pushAbilityToStack(g, obj)
	recordActivatedAbilityUse(g, permanent.ObjectID, activate.AbilityIndex, ability)
	if ability.IsLoyaltyAbility {
		recordActivatedAbilityUse(g, permanent.ObjectID, -1, &game.AbilityDef{Timing: game.OncePerTurn})
	}
	return true
}

func (e *Engine) applyGraveyardAbilityWithChoices(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	card, ability, ok := graveyardAbilitySource(g, playerID, activate.SourceID, activate.AbilityIndex)
	if !ok || !canActivateGraveyardAbility(g, playerID, card.ID, ability, activate.AbilityIndex, activate.Targets, activate.XValue) {
		return false
	}
	def := cardFaceOrDefault(card, game.FaceFront)
	completedTargets, ok := e.completeAbilityAnnouncementTargets(g, playerID, def, 0, ability, activate.Targets, agents, log)
	if !ok || !canActivateGraveyardAbility(g, playerID, card.ID, ability, activate.AbilityIndex, completedTargets, activate.XValue) {
		return false
	}
	prefs := e.paymentPreferencesForCost(g, playerID, manaCostPtr(ability.ManaCost), abilityAdditionalCosts(ability), agents, log)
	if !paymentOrch.payAbilityCosts(g, payment.AbilityRequest{
		PlayerID:     playerID,
		SourceCardID: card.ID,
		SourceZone:   game.ZoneGraveyard,
		Ability:      ability,
		XValue:       activate.XValue,
		Prefs:        prefs,
	}) {
		return false
	}
	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackActivatedAbility,
		SourceID:     card.ID,
		SourceCardID: card.ID,
		SourceZone:   game.ZoneGraveyard,
		AbilityIndex: activate.AbilityIndex,
		Controller:   playerID,
		Targets:      append([]game.Target(nil), completedTargets...),
		XValue:       activate.XValue,
	}
	pushAbilityToStack(g, obj)
	recordActivatedAbilityUse(g, card.ID, activate.AbilityIndex, ability)
	return true
}

func canActivateLoyaltyAbility(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, ability *game.AbilityDef, abilityIndex int, targets []game.Target, xValue int) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer || permanent.PhasedOut || effectiveController(g, permanent) != playerID || ability == nil {
		return false
	}
	if xValue != 0 || ability.Kind != game.ActivatedAbility || !ability.IsLoyaltyAbility || !abilityFunctionsOnBattlefield(ability) || !permanentHasType(g, permanent, types.Planeswalker) {
		return false
	}
	if !isSorcerySpeed(g, playerID) || activatedAbilityUsedThisTurn(g, permanent.ObjectID, abilityIndex, ability) || g.ActivatedAbilitiesThisTurn[game.ActivatedAbilityUse{SourceID: permanent.ObjectID, AbilityIndex: -1}] {
		return false
	}
	if ability.LoyaltyCost < 0 && permanent.Counters.Get(counter.Loyalty) < -ability.LoyaltyCost {
		return false
	}
	if !activationConditionSatisfied(g, playerID, permanent, ability) {
		return false
	}
	card, ok := permanentCardDef(g, permanent)
	if !ok || !targetsValidForAbilityFromSourceObject(g, playerID, card, permanent.ObjectID, ability, targets) {
		return false
	}
	return paymentOrch.buildAbilityCostPlan(g, payment.AbilityRequest{PlayerID: playerID, Source: permanent, Ability: ability, XValue: xValue})
}

func applyLoyaltyCost(permanent *game.Permanent, cost int) {
	if cost >= 0 {
		permanent.Counters.Add(counter.Loyalty, cost)
		return
	}
	permanent.Counters.Remove(counter.Loyalty, -cost)
}

func (e *Engine) applyCyclingAbility(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction) bool {
	return e.applyCyclingAbilityWithChoices(g, playerID, activate, [game.NumPlayers]PlayerAgent{}, nil)
}

func (e *Engine) applyCyclingAbilityWithChoices(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	card, ability, ok := cyclingAbilitySource(g, playerID, activate.SourceID, activate.AbilityIndex)
	if !ok {
		return false
	}
	if !canActivateCyclingAbility(g, playerID, activate.SourceID, ability, activate.AbilityIndex, activate.Targets, activate.XValue) {
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

func (e *Engine) canCastSpell(g *game.Game, playerID game.PlayerID, cardID id.ID, targets []game.Target, xValue int, chosenModes []int) bool {
	return e.canCastSpellWithKicker(g, playerID, cardID, targets, xValue, chosenModes, false)
}

func (e *Engine) canCastSpellWithKicker(g *game.Game, playerID game.PlayerID, cardID id.ID, targets []game.Target, xValue int, chosenModes []int, kickerPaid bool) bool {
	return e.canCastSpellFaceFromZoneWithKicker(g, playerID, cardID, game.ZoneHand, game.FaceFront, targets, xValue, chosenModes, kickerPaid)
}

func (e *Engine) canCastSpellFromZoneWithKicker(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone game.ZoneType, targets []game.Target, xValue int, chosenModes []int, kickerPaid bool) bool {
	return e.canCastSpellFaceFromZoneWithKicker(g, playerID, cardID, sourceZone, game.FaceFront, targets, xValue, chosenModes, kickerPaid)
}

func (*Engine) canCastSpellFaceFromZoneWithKicker(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone game.ZoneType, face game.FaceIndex, targets []game.Target, xValue int, chosenModes []int, kickerPaid bool) bool {
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
	if sourceZone != game.ZoneHand && sourceZone != game.ZoneCommand && face != game.FaceFront {
		return false
	}
	switch sourceZone {
	case game.ZoneCommand:
		if player.CommanderInstanceID != cardID {
			return false
		}
	case game.ZoneHand:
	case game.ZoneGraveyard:
		if !canCastFromZoneByRuleEffect(g, playerID, cardID, sourceZone) {
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

func legalCastFacesForZone(card *game.CardDef, sourceZone game.ZoneType) []game.FaceIndex {
	if sourceZone != game.ZoneHand && sourceZone != game.ZoneCommand {
		if card.CanChooseCastFace(game.FaceFront) {
			return []game.FaceIndex{game.FaceFront}
		}
		return nil
	}
	return card.LegalCastFaces()
}

func castSourceContains(player *game.Player, cardID id.ID, sourceZone game.ZoneType) bool {
	switch sourceZone {
	case game.ZoneHand:
		return player.Hand.Contains(cardID)
	case game.ZoneCommand:
		return player.CommandZone.Contains(cardID)
	case game.ZoneGraveyard:
		return player.Graveyard.Contains(cardID)
	default:
		return false
	}
}

func castSourceZoneCards(player *game.Player, sourceZone game.ZoneType) []id.ID {
	switch sourceZone {
	case game.ZoneHand:
		return player.Hand.All()
	case game.ZoneGraveyard:
		return player.Graveyard.All()
	default:
		return nil
	}
}

func removeCastSourceCard(player *game.Player, cardID id.ID, sourceZone game.ZoneType) bool {
	switch sourceZone {
	case game.ZoneHand:
		return player.Hand.Remove(cardID)
	case game.ZoneCommand:
		return player.CommandZone.Remove(cardID)
	case game.ZoneGraveyard:
		return player.Graveyard.Remove(cardID)
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

func legalXValuesForCost(g *game.Game, playerID game.PlayerID, cost *mana.Cost) []int {
	if !costHasVariableMana(cost) {
		return []int{0}
	}
	var values []int
	for x := 0; x <= maxLegalXValue; x++ {
		if !paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: cost, XValue: x}) {
			break
		}
		values = append(values, x)
	}
	return values
}

func costHasVariableMana(cost *mana.Cost) bool {
	if cost == nil {
		return false
	}
	for _, symbol := range *cost {
		if symbol.Kind == mana.VariableSymbol {
			return true
		}
	}
	return false
}

func activatedAbilitySource(g *game.Game, playerID game.PlayerID, sourceID id.ID, abilityIndex int) (*game.Permanent, *game.AbilityDef, bool) {
	if abilityIndex < 0 {
		return nil, nil, false
	}
	permanent, ok := permanentByObjectID(g, sourceID)
	if !ok || permanent.PhasedOut || effectiveController(g, permanent) != playerID {
		return nil, nil, false
	}
	card, ok := permanentCardDef(g, permanent)
	if !ok || abilityIndex >= len(card.Abilities) {
		return nil, nil, false
	}
	return permanent, &card.Abilities[abilityIndex], true
}

func cyclingAbilitySource(g *game.Game, playerID game.PlayerID, sourceID id.ID, abilityIndex int) (*game.CardInstance, *game.AbilityDef, bool) {
	if abilityIndex < 0 {
		return nil, nil, false
	}

	player, ok := playerByID(g, playerID)
	if !ok || !player.Hand.Contains(sourceID) {
		return nil, nil, false
	}
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		return nil, nil, false
	}
	frontDef := cardFaceOrDefault(card, game.FaceFront)
	if abilityIndex >= len(frontDef.Abilities) {
		return nil, nil, false
	}
	return card, &frontDef.Abilities[abilityIndex], true
}

func graveyardAbilitySource(g *game.Game, playerID game.PlayerID, sourceID id.ID, abilityIndex int) (*game.CardInstance, *game.AbilityDef, bool) {
	if abilityIndex < 0 {
		return nil, nil, false
	}
	player, ok := playerByID(g, playerID)
	if !ok || !player.Graveyard.Contains(sourceID) {
		return nil, nil, false
	}
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		return nil, nil, false
	}
	frontDef := cardFaceOrDefault(card, game.FaceFront)
	if abilityIndex >= len(frontDef.Abilities) {
		return nil, nil, false
	}
	return card, &frontDef.Abilities[abilityIndex], true
}

func canActivateEquipAbility(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, ability *game.AbilityDef, abilityIndex int, targets []game.Target, xValue int) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer || permanent.PhasedOut || effectiveController(g, permanent) != playerID || ability == nil {
		return false
	}
	if xValue != 0 || ability.Kind != game.ActivatedAbility || ability.IsManaAbility || !abilityFunctionsOnBattlefield(ability) || !isEquipmentPermanent(g, permanent) {
		return false
	}
	if !abilityHasKeyword(ability, game.Equip) && ability.Timing != game.SorceryOnly {
		return false
	}
	if !isSorcerySpeed(g, playerID) || abilityHasNonTapAdditionalCosts(ability) || activatedAbilityUsedThisTurn(g, permanent.ObjectID, abilityIndex, ability) {
		return false
	}
	if !activationConditionSatisfied(g, playerID, permanent, ability) {
		return false
	}
	card, ok := permanentCardDef(g, permanent)
	if !ok || !targetsValidForAbilityFromSourceObject(g, playerID, card, permanent.ObjectID, ability, targets) {
		return false
	}
	if len(targets) != 1 || targets[0].Kind != game.TargetPermanent {
		return false
	}
	target, ok := permanentByObjectID(g, targets[0].PermanentID)
	if !ok || effectiveController(g, target) != playerID || !canAttachPermanent(g, permanent, target) {
		return false
	}
	return paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: manaCostPtr(ability.ManaCost)})
}

func canActivateGeneralAbility(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, ability *game.AbilityDef, abilityIndex int, targets []game.Target, xValue int) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer || permanent.PhasedOut || effectiveController(g, permanent) != playerID || ability == nil {
		return false
	}
	if ability.Kind != game.ActivatedAbility || ability.IsManaAbility || ability.IsLoyaltyAbility || abilityHasKeyword(ability, game.Equip) || !abilityFunctionsOnBattlefield(ability) {
		return false
	}
	if !activatedAbilityTimingAllows(g, playerID, ability) || activatedAbilityUsedThisTurn(g, permanent.ObjectID, abilityIndex, ability) {
		return false
	}
	if !activationConditionSatisfied(g, playerID, permanent, ability) {
		return false
	}
	card, ok := permanentCardDef(g, permanent)
	if !ok || !targetsValidForAbilityFromSourceObject(g, playerID, card, permanent.ObjectID, ability, targets) {
		return false
	}
	return paymentOrch.buildAbilityCostPlan(g, payment.AbilityRequest{PlayerID: playerID, Source: permanent, Ability: ability, XValue: xValue})
}

func canActivateCyclingAbility(g *game.Game, playerID game.PlayerID, cardID id.ID, ability *game.AbilityDef, abilityIndex int, targets []game.Target, xValue int) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer || ability == nil {
		return false
	}

	if xValue != 0 || abilityIndex < 0 || ability.Kind != game.ActivatedAbility || ability.IsManaAbility || !abilityHasKeyword(ability, game.Cycling) {
		return false
	}
	if ability.Timing != game.NoTimingRestriction || !abilityHasDiscardThisCardCost(ability) {
		return false
	}
	if len(targets) != 0 || len(ability.Targets) != 0 {
		return false
	}
	_, gotAbility, ok := cyclingAbilitySource(g, playerID, cardID, abilityIndex)
	if !ok || !abilityHasKeyword(gotAbility, game.Cycling) {
		return false
	}
	return paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: manaCostPtr(ability.ManaCost)})
}

func canActivateGraveyardAbility(g *game.Game, playerID game.PlayerID, cardID id.ID, ability *game.AbilityDef, abilityIndex int, targets []game.Target, xValue int) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer || ability == nil {
		return false
	}
	if ability.Kind != game.ActivatedAbility || ability.IsManaAbility || ability.IsLoyaltyAbility || ability.ZoneOfFunction != game.ZoneGraveyard {
		return false
	}
	if !activatedAbilityTimingAllows(g, playerID, ability) || activatedAbilityUsedThisTurn(g, cardID, abilityIndex, ability) {
		return false
	}
	card, _, ok := graveyardAbilitySource(g, playerID, cardID, abilityIndex)
	if !ok {
		return false
	}
	def := cardFaceOrDefault(card, game.FaceFront)
	if !targetsValidForAbilityFromSourceObject(g, playerID, def, 0, ability, targets) {
		return false
	}
	return paymentOrch.buildAbilityCostPlan(g, payment.AbilityRequest{
		PlayerID:     playerID,
		SourceCardID: cardID,
		SourceZone:   game.ZoneGraveyard,
		Ability:      ability,
		XValue:       xValue,
	})
}

func abilityHasKeyword(ability *game.AbilityDef, keyword game.Keyword) bool {
	if ability == nil {
		return false
	}
	return slices.Contains(ability.Keywords, keyword)
}

func canActivateManaAbility(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, ability *game.AbilityDef, abilityIndex int) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer || permanent.PhasedOut || effectiveController(g, permanent) != playerID || ability == nil {
		return false
	}
	if ability.Kind != game.ActivatedAbility || !ability.IsManaAbility || ability.IsLoyaltyAbility || !abilityFunctionsOnBattlefield(ability) {
		return false
	}
	if len(ability.Targets) != 0 || !manaAbilityHasAddManaEffect(ability) || !manaAbilityChoicesAvailable(g, playerID, ability) {
		return false
	}
	if ability.Timing != game.NoTimingRestriction || activatedAbilityUsedThisTurn(g, permanent.ObjectID, abilityIndex, ability) {
		return false
	}
	if !activationConditionSatisfied(g, playerID, permanent, ability) {
		return false
	}
	if hasTapCost(ability) {
		if !canTapPermanentForAbility(g, permanent) {
			return false
		}
	} else if abilityHasNonTapAdditionalCosts(ability) {
		return false
	}
	return paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: manaCostPtr(ability.ManaCost)})
}

func manaAbilityHasAddManaEffect(ability *game.AbilityDef) bool {
	if ability == nil || len(ability.Effects) == 0 {
		return false
	}
	hasAddMana := false
	for i := range ability.Effects {
		effect := &ability.Effects[i]
		switch effect.Type {
		case game.EffectAddMana:
			hasAddMana = true
		case game.EffectChoose:
			if !effect.Choice.Exists || effect.Choice.Val.Kind != game.ResolutionChoiceColor {
				return false
			}
		default:
			return false
		}
	}
	return hasAddMana
}

func manaAbilityChoicesAvailable(g *game.Game, playerID game.PlayerID, ability *game.AbilityDef) bool {
	for i := range ability.Effects {
		effect := &ability.Effects[i]
		if effect.Type != game.EffectChoose || !effect.Choice.Exists {
			continue
		}
		choice := effect.Choice.Val
		choicePlayer := resolutionChoicePlayer(playerID, &choice)
		_, values := resolutionChoiceOptions(g, choicePlayer, &choice)
		if len(values) == 0 {
			return false
		}
	}
	return true
}

func hasTapCost(ability *game.AbilityDef) bool {
	if ability == nil {
		return false
	}
	for _, cost := range ability.AdditionalCosts {
		if cost.Kind == game.AdditionalCostTap {
			return true
		}
	}
	return false
}

func abilityHasNonTapAdditionalCosts(ability *game.AbilityDef) bool {
	for _, cost := range abilityAdditionalCosts(ability) {
		if cost.Kind != game.AdditionalCostTap {
			return true
		}
	}
	return false
}

func abilityHasDiscardThisCardCost(ability *game.AbilityDef) bool {
	costs := abilityAdditionalCosts(ability)
	if len(costs) != 1 {
		return false
	}
	cost := costs[0]
	if cost.Kind != game.AdditionalCostDiscard || payment.AdditionalCostAmount(cost) != 1 {
		return false
	}
	if cost.Text != "" {
		return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(cost.Text)), ".") == "discard this card"
	}
	return cost.Zone == game.ZoneHand
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

func activatedAbilityTimingAllows(g *game.Game, playerID game.PlayerID, ability *game.AbilityDef) bool {
	if ability == nil {
		return false
	}
	switch ability.Timing {
	case game.NoTimingRestriction, game.OncePerTurn:
		return true
	case game.SorceryOnly, game.SorceryOncePerTurn:
		return isSorcerySpeed(g, playerID)
	case game.DuringCombat:
		return g.Turn.Phase == game.PhaseCombat
	case game.DuringUpkeep:
		return g.Turn.Phase == game.PhaseBeginning && g.Turn.Step == game.StepUpkeep
	default:
		return false
	}
}

func activatedAbilityUsedThisTurn(g *game.Game, sourceID id.ID, abilityIndex int, ability *game.AbilityDef) bool {
	if ability == nil || !abilityHasOncePerTurnRestriction(ability) {
		return false
	}
	return g.ActivatedAbilitiesThisTurn[game.ActivatedAbilityUse{
		SourceID:     sourceID,
		AbilityIndex: abilityIndex,
	}]
}

func recordActivatedAbilityUse(g *game.Game, sourceID id.ID, abilityIndex int, ability *game.AbilityDef) {
	if ability == nil || !abilityHasOncePerTurnRestriction(ability) {
		return
	}
	if g.ActivatedAbilitiesThisTurn == nil {
		g.ActivatedAbilitiesThisTurn = make(map[game.ActivatedAbilityUse]bool)
	}
	g.ActivatedAbilitiesThisTurn[game.ActivatedAbilityUse{SourceID: sourceID, AbilityIndex: abilityIndex}] = true
}

func abilityHasOncePerTurnRestriction(ability *game.AbilityDef) bool {
	return ability.Timing == game.OncePerTurn || ability.Timing == game.SorceryOncePerTurn
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
