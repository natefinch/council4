package rules

import (
	"reflect"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/rules/payment"
)

const maxLegalXValue = 20

// maxLegalMultikickCount bounds the number of times a Multikicker cost may be
// enumerated as paid (CR 702.32), matching the X-value enumeration cap so the
// action space stays finite.
const maxLegalMultikickCount = 20

// appendMultikickedCastActions enumerates Multikicker casts that pay the kicker
// cost one or more times. Target choices are generated for each concrete count
// so kicker-scaled target specs offer exactly the legal cardinality.
func (e *Engine) appendMultikickedCastActions(g *game.Game, playerID game.PlayerID, actions []action.Action, actionBuild actionBuilderType, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, spellDef *game.CardDef, xValue int, modes []int) []action.Action {
	for count := 1; count <= maxLegalMultikickCount; count++ {
		result := targetChoicesForSpellWithKickerCount(g, playerID, spellDef, modes, game.CastBranch{Kicked: true}, count)
		if result.kind == targetInvalidSpec || result.kind == targetNoLegalChoices {
			break
		}
		for _, targets := range result.choices {
			if e.canCastSpellFaceFromZoneWithMultikick(g, playerID, cardID, sourceZone, face, targets, xValue, modes, count) {
				actions = append(actions, actionBuild.castMultikickedSpell(cardID, sourceZone, face, targets, xValue, modes, count))
			}
		}
	}
	return actions
}

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
	actions = append(actions, e.legalPlotActions(g, playerID)...)
	actions = append(actions, e.legalForetellActions(g, playerID)...)
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
		for _, xValue := range legalXValuesForCostAndAdditional(g, playerID, manaCostPtr(spellDef.ManaCost), spellDef.AdditionalCosts) {
			for _, modes := range modeChoicesForSpellAt(g, playerID, spellDef) {
				targetResult := targetChoicesForSpell(g, playerID, spellDef, modes, game.CastBranch{})
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
			if _, ok := landCardInstanceFaceFromZone(g, player, cardID, zone.Hand, face); ok {
				actions = append(actions, actionBuild.playLandFromZone(cardID, zone.Hand, face))
			}
		}
	}
	for _, cardID := range player.Exile.All() {
		if !canPlayLandFromZoneByRuleEffect(g, playerID, cardID, zone.Exile) {
			continue
		}
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		for _, face := range card.Def.FaceIndexes() {
			if _, ok := landCardInstanceFaceFromZone(g, player, cardID, zone.Exile, face); ok {
				actions = append(actions, actionBuild.playLandFromZone(cardID, zone.Exile, face))
			}
		}
	}
	for _, cardID := range foreignExileCastableCards(g, playerID) {
		if !canPlayLandFromZoneByRuleEffect(g, playerID, cardID, zone.Exile) {
			continue
		}
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		sourcePlayer := castSourcePlayer(g, player, cardID, zone.Exile)
		for _, face := range card.Def.FaceIndexes() {
			if _, ok := landCardInstanceFaceFromZone(g, sourcePlayer, cardID, zone.Exile, face); ok {
				actions = append(actions, actionBuild.playLandFromZone(cardID, zone.Exile, face))
			}
		}
	}
	for _, cardID := range player.Graveyard.All() {
		if !canPlayLandFromZoneByRuleEffect(g, playerID, cardID, zone.Graveyard) {
			continue
		}
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		for _, face := range card.Def.FaceIndexes() {
			if _, ok := landCardInstanceFaceFromZone(g, player, cardID, zone.Graveyard, face); ok {
				actions = append(actions, actionBuild.playLandFromZone(cardID, zone.Graveyard, face))
			}
		}
	}
	if topID, ok := player.Library.Top(); ok && canPlayLandFromZoneByRuleEffect(g, playerID, topID, zone.Library) {
		if card, ok := g.GetCardInstance(topID); ok {
			for _, face := range card.Def.FaceIndexes() {
				if _, ok := landCardInstanceFaceFromZone(g, player, topID, zone.Library, face); ok {
					actions = append(actions, actionBuild.playLandFromZone(topID, zone.Library, face))
				}
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
		cards := castSourceZoneCards(player, sourceZone)
		if sourceZone == zone.Exile {
			cards = append(cards, foreignExileCastableCards(g, playerID)...)
		}
		for _, cardID := range cards {
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
				for _, xValue := range legalXValuesForCostAndAdditional(g, playerID, manaCostPtr(spellDef.ManaCost), spellDef.AdditionalCosts) {
					for _, modes := range modeChoicesForSpellAt(g, playerID, spellDef) {
						plainResult := targetChoicesForSpell(g, playerID, spellDef, modes, game.CastBranch{})
						if plainResult.kind != targetInvalidSpec {
							for _, targets := range plainResult.choices {
								if e.canCastSpellFaceFromZoneWithKicker(g, playerID, cardID, sourceZone, face, targets, xValue, modes, false) {
									actions = append(actions, actionBuild.castSpell(cardID, sourceZone, face, targets, xValue, modes))
								}
							}
						}
						if spellHasGift(spellDef) {
							giftResult := targetChoicesForSpell(g, playerID, spellDef, modes, game.CastBranch{GiftPromised: true})
							if giftResult.kind != targetInvalidSpec {
								for _, targets := range giftResult.choices {
									if e.canCastGiftSpellFaceFromZone(g, playerID, cardID, sourceZone, face, targets, xValue, modes) {
										for _, opponent := range aliveOpponents(g, playerID) {
											actions = append(actions, actionBuild.castGiftSpell(cardID, sourceZone, face, targets, xValue, modes, opponent))
										}
									}
								}
							}
						}
						if spellHasBargain(spellDef) {
							bargainResult := targetChoicesForSpell(g, playerID, spellDef, modes, game.CastBranch{Bargained: true})
							if bargainResult.kind != targetInvalidSpec {
								for _, targets := range bargainResult.choices {
									if e.canCastBargainedSpellFaceFromZone(g, playerID, cardID, sourceZone, face, targets, xValue, modes) {
										actions = append(actions, actionBuild.castBargainedSpell(cardID, sourceZone, face, targets, xValue, modes))
									}
								}
							}
						}
						if spellHasOffspring(spellDef) {
							offspringResult := targetChoicesForSpell(g, playerID, spellDef, modes, game.CastBranch{Offspring: true})
							if offspringResult.kind != targetInvalidSpec {
								for _, targets := range offspringResult.choices {
									if e.canCastOffspringSpellFaceFromZone(g, playerID, cardID, sourceZone, face, targets, xValue, modes) {
										actions = append(actions, actionBuild.castOffspringSpell(cardID, sourceZone, face, targets, xValue, modes))
									}
								}
							}
						}
						if spellHasBestow(spellDef) {
							bestowResult := targetChoicesForSpell(g, playerID, spellDef, modes, game.CastBranch{Bestowed: true})
							if bestowResult.kind != targetInvalidSpec {
								for _, targets := range bestowResult.choices {
									if e.canCastBestowSpellFaceFromZone(g, playerID, cardID, sourceZone, face, targets, xValue, modes) {
										actions = append(actions, actionBuild.castBestowSpell(cardID, sourceZone, face, targets, xValue, modes))
									}
								}
							}
						}
					}
					if spellHasMultikicker(spellDef) {
						for _, modes := range modeChoicesForSpellAtBranch(g, playerID, spellDef, game.CastBranch{Kicked: true}) {
							actions = e.appendMultikickedCastActions(g, playerID, actions, actionBuild, cardID, sourceZone, face, spellDef, xValue, modes)
						}
					} else if spellHasKicker(spellDef) {
						for _, modes := range modeChoicesForSpellAtBranch(g, playerID, spellDef, game.CastBranch{Kicked: true}) {
							kickedResult := targetChoicesForSpell(g, playerID, spellDef, modes, game.CastBranch{Kicked: true})
							if kickedResult.kind != targetInvalidSpec {
								for _, targets := range kickedResult.choices {
									if e.canCastSpellFaceFromZoneWithKicker(g, playerID, cardID, sourceZone, face, targets, xValue, modes, true) {
										actions = append(actions, actionBuild.castKickedSpell(cardID, sourceZone, face, targets, xValue, modes))
									}
								}
							}
						}
					}
				}
				if spellDef.Overload.Exists {
					overloadedDef := overloadSpellDef(spellDef)
					overloadCost := spellDef.Overload.Val.Cost
					for _, xValue := range legalXValuesForCostAndAdditional(g, playerID, &overloadCost, spellDef.AdditionalCosts) {
						for _, modes := range modeChoicesForSpellAt(g, playerID, overloadedDef) {
							if e.canCastOverloadedSpellFaceFromZoneWithOptions(g, playerID, cardID, sourceZone, face, xValue, modes, false) {
								actions = append(actions, actionBuild.castOverloadedSpell(cardID, sourceZone, face, xValue, modes, false))
							}
						}
						if sourceZone == zone.Hand && spellHasKicker(spellDef) {
							for _, modes := range modeChoicesForSpellAtBranch(g, playerID, overloadedDef, game.CastBranch{Kicked: true}) {
								if e.canCastOverloadedSpellFaceFromZoneWithOptions(g, playerID, cardID, sourceZone, face, xValue, modes, true) {
									actions = append(actions, actionBuild.castOverloadedSpell(cardID, sourceZone, face, xValue, modes, true))
								}
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
		for _, xValue := range legalXValuesForCostAndAdditional(g, playerID, manaCostPtr(spellDef.ManaCost), spellDef.AdditionalCosts) {
			for _, modes := range modeChoicesForSpellAt(g, playerID, spellDef) {
				plainResult := targetChoicesForSpell(g, playerID, spellDef, modes, game.CastBranch{})
				if plainResult.kind != targetInvalidSpec {
					for _, targets := range plainResult.choices {
						if e.canCastSpellFaceFromZoneWithKicker(g, playerID, card.ID, zone.Command, face, targets, xValue, modes, false) {
							actions = append(actions, actionBuild.castSpell(card.ID, zone.Command, face, targets, xValue, modes))
						}
					}
				}
				if spellHasGift(spellDef) {
					giftResult := targetChoicesForSpell(g, playerID, spellDef, modes, game.CastBranch{GiftPromised: true})
					if giftResult.kind != targetInvalidSpec {
						for _, targets := range giftResult.choices {
							if e.canCastGiftSpellFaceFromZone(g, playerID, card.ID, zone.Command, face, targets, xValue, modes) {
								for _, opponent := range aliveOpponents(g, playerID) {
									actions = append(actions, actionBuild.castGiftSpell(card.ID, zone.Command, face, targets, xValue, modes, opponent))
								}
							}
						}
					}
				}
				if spellHasBargain(spellDef) {
					bargainResult := targetChoicesForSpell(g, playerID, spellDef, modes, game.CastBranch{Bargained: true})
					if bargainResult.kind != targetInvalidSpec {
						for _, targets := range bargainResult.choices {
							if e.canCastBargainedSpellFaceFromZone(g, playerID, card.ID, zone.Command, face, targets, xValue, modes) {
								actions = append(actions, actionBuild.castBargainedSpell(card.ID, zone.Command, face, targets, xValue, modes))
							}
						}
					}
				}
				if spellHasOffspring(spellDef) {
					offspringResult := targetChoicesForSpell(g, playerID, spellDef, modes, game.CastBranch{Offspring: true})
					if offspringResult.kind != targetInvalidSpec {
						for _, targets := range offspringResult.choices {
							if e.canCastOffspringSpellFaceFromZone(g, playerID, card.ID, zone.Command, face, targets, xValue, modes) {
								actions = append(actions, actionBuild.castOffspringSpell(card.ID, zone.Command, face, targets, xValue, modes))
							}
						}
					}
				}
			}
			if spellHasMultikicker(spellDef) {
				for _, modes := range modeChoicesForSpellAtBranch(g, playerID, spellDef, game.CastBranch{Kicked: true}) {
					actions = e.appendMultikickedCastActions(g, playerID, actions, actionBuild, card.ID, zone.Command, face, spellDef, xValue, modes)
				}
			} else if spellHasKicker(spellDef) {
				for _, modes := range modeChoicesForSpellAtBranch(g, playerID, spellDef, game.CastBranch{Kicked: true}) {
					kickedResult := targetChoicesForSpell(g, playerID, spellDef, modes, game.CastBranch{Kicked: true})
					if kickedResult.kind != targetInvalidSpec {
						for _, targets := range kickedResult.choices {
							if e.canCastSpellFaceFromZoneWithKicker(g, playerID, card.ID, zone.Command, face, targets, xValue, modes, true) {
								actions = append(actions, actionBuild.castKickedSpell(card.ID, zone.Command, face, targets, xValue, modes))
							}
						}
					}
				}
			}
		}
		if spellDef.Overload.Exists {
			overloadedDef := overloadSpellDef(spellDef)
			overloadCost := spellDef.Overload.Val.Cost
			for _, xValue := range legalXValuesForCostAndAdditional(g, playerID, &overloadCost, spellDef.AdditionalCosts) {
				for _, modes := range modeChoicesForSpellAt(g, playerID, overloadedDef) {
					if e.canCastOverloadedSpellFaceFromZoneWithOptions(g, playerID, card.ID, zone.Command, face, xValue, modes, false) {
						actions = append(actions, actionBuild.castOverloadedSpell(card.ID, zone.Command, face, xValue, modes, false))
					}
				}
				if spellHasKicker(spellDef) {
					for _, modes := range modeChoicesForSpellAtBranch(g, playerID, overloadedDef, game.CastBranch{Kicked: true}) {
						if e.canCastOverloadedSpellFaceFromZoneWithOptions(g, playerID, card.ID, zone.Command, face, xValue, modes, true) {
							actions = append(actions, actionBuild.castOverloadedSpell(card.ID, zone.Command, face, xValue, modes, true))
						}
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
		return ok && e.applyPlayLandFaceFromZoneWithChoices(g, playerID, playLand.CardID, playLand.SourceZone, playLand.Face, agents, log)
	case action.ActionCastSpell:
		cast, ok := act.CastSpellPayload()
		return ok && e.applyCastSpellWithChoices(g, playerID, cast, agents, log)
	case action.ActionActivateAbility:
		activate, ok := act.ActivateAbilityPayload()
		if !ok || !e.applyActivateAbilityWithChoices(g, playerID, activate, agents, log) {
			return false
		}
		if g.AbilityActivationsThisTurn == nil {
			g.AbilityActivationsThisTurn = make(map[game.ActivatedAbilityUse]int)
		}
		g.AbilityActivationsThisTurn[game.ActivatedAbilityUse{SourceID: activate.SourceID, AbilityIndex: activate.AbilityIndex}]++
		return true
	case action.ActionSuspendCard:
		suspend, ok := act.SuspendCardPayload()
		return ok && e.applySuspendCard(g, playerID, suspend.CardID, agents, log)
	case action.ActionPlotCard:
		plot, ok := act.PlotCardPayload()
		return ok && e.applyPlotCard(g, playerID, plot.CardID, agents, log)
	case action.ActionForetellCard:
		foretell, ok := act.ForetellCardPayload()
		return ok && e.applyForetellCard(g, playerID, foretell.CardID, agents, log)
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
		var seenManualMana []*game.ManaAbility
		for idx, ability := range permanentEffectiveAbilitiesView(g, permanent) {
			if body, ok := ability.(*game.ManaAbility); ok {
				if !payment.IsAutomaticManaAbility(body) &&
					!containsEquivalentManaAbility(seenManualMana, body) &&
					canActivateManaAbility(g, playerID, permanent, body, idx) {
					actions = append(actions, actionBuild.activateAbility(permanent.ObjectID, idx, nil, 0))
					seenManualMana = append(seenManualMana, body)
				}
				continue
			}
			if body, ok := ability.(*game.ActivatedAbility); ok {
				sourceCard, _ := g.GetCardInstance(permanent.CardInstanceID)
				effectiveCost := effectiveActivatedAbilityCost(g, playerID, sourceCard, body)
				for _, xValue := range legalXValuesForCostAndAdditional(g, playerID, manaCostPtr(effectiveCost), body.AdditionalCosts) {
					for _, modes := range modeChoicesForBody(body) {
						targetResult := targetChoicesForBodyFromSourceObjectWithModes(g, playerID, card, permanent.ObjectID, game.Event{}, body, modes)
						if targetResult.kind == targetInvalidSpec {
							continue
						}
						for choiceIndex, targets := range targetResult.choices {
							if canActivateEquipAbilityWithModes(g, playerID, permanent, body, idx, targets, xValue, modes) ||
								canActivateGeneralAbilityWithModes(g, playerID, permanent, body, idx, targets, xValue, modes) {
								actions = append(actions, actionBuild.activateAbilityWithModes(permanent.ObjectID, idx, append([]game.Target(nil), targets...), targetResult.targetCounts[choiceIndex], xValue, modes))
							}
						}
					}
				}
				continue
			}
			body, ok := ability.(*game.LoyaltyAbility)
			if !ok {
				continue
			}
			targetResult := targetChoicesForBodyFromSourceObject(g, playerID, card, permanent.ObjectID, body)
			if targetResult.kind == targetInvalidSpec {
				continue
			}
			for choiceIndex, targets := range targetResult.choices {
				if canActivateLoyaltyAbility(g, playerID, permanent, body, idx, targets, 0) {
					actions = append(actions, actionBuild.activateAbilityWithModes(permanent.ObjectID, idx, append([]game.Target(nil), targets...), targetResult.targetCounts[choiceIndex], 0, nil))
				}
			}
		}
	}
	actions = append(actions, e.legalHandActivateAbilityActions(g, playerID)...)
	actions = append(actions, e.legalGraveyardActivateAbilityActions(g, playerID)...)
	actions = append(actions, e.legalHandManaAbilityActions(g, playerID)...)
	return actions
}

func (*Engine) legalHandActivateAbilityActions(g *game.Game, playerID game.PlayerID) []action.Action {
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
		def := cardFaceOrDefault(card, game.FaceFront)
		effectiveAbilities := effectiveHandActivatedAbilities(g, playerID, card)
		for i := range effectiveAbilities {
			indexed := &effectiveAbilities[i]
			body := &indexed.body
			for _, xValue := range legalXValuesForCostAndAdditional(g, playerID, manaCostPtr(body.ManaCost), body.AdditionalCosts) {
				for _, modes := range modeChoicesForBody(body) {
					targetResult := targetChoicesForBodyFromSourceObjectWithModes(g, playerID, def, 0, game.Event{}, body, modes)
					if targetResult.kind == targetInvalidSpec {
						continue
					}
					for choiceIndex, targets := range targetResult.choices {
						if canActivateHandAbilityWithModes(g, playerID, cardID, body, indexed.index, targets, xValue, modes) {
							actions = append(actions, actionBuild.activateAbilityWithModes(cardID, indexed.index, append([]game.Target(nil), targets...), targetResult.targetCounts[choiceIndex], xValue, modes))
						}
					}
				}
			}
		}
	}
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
					targetResult := targetChoicesForBodyFromSourceObjectWithModes(g, playerID, def, 0, game.Event{}, body, modes)
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

func (e *Engine) legalManaAbilityActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return nil
	}
	var actions []action.Action
	for _, permanent := range g.Battlefield {
		if permanent.PhasedOut || effectiveController(g, permanent) != playerID {
			continue
		}
		var seenManualMana []*game.ManaAbility
		for idx, ability := range permanentEffectiveAbilitiesView(g, permanent) {
			body, ok := ability.(*game.ManaAbility)
			if ok && !payment.IsAutomaticManaAbility(body) &&
				!containsEquivalentManaAbility(seenManualMana, body) &&
				canActivateManaAbility(g, playerID, permanent, body, idx) {
				actions = append(actions, actionBuild.activateAbility(permanent.ObjectID, idx, nil, 0))
				seenManualMana = append(seenManualMana, body)
			}
		}
	}
	actions = append(actions, e.legalHandManaAbilityActions(g, playerID)...)

	return actions
}

// legalHandManaAbilityActions enumerates mana abilities printed on cards in the
// player's hand whose cost is exiling that card from hand (Simian/Elvish Spirit
// Guide). Such abilities are activated for mana like any other mana ability but
// function from the hand rather than the battlefield.
func (*Engine) legalHandManaAbilityActions(g *game.Game, playerID game.PlayerID) []action.Action {
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
		def := cardFaceOrDefault(card, game.FaceFront)
		for i := range def.ManaAbilities {
			idx := def.ManaAbilityIndex(i)
			if canActivateHandManaAbility(g, playerID, cardID, &def.ManaAbilities[i], idx) {
				actions = append(actions, actionBuild.activateAbility(cardID, idx, nil, 0))
			}
		}
	}
	return actions
}

func containsEquivalentManaAbility(abilities []*game.ManaAbility, candidate *game.ManaAbility) bool {
	for _, ability := range abilities {
		if reflect.DeepEqual(ability, candidate) {
			return true
		}
	}
	return false
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
