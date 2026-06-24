package agent

import (
	"slices"

	"github.com/natefinch/council4/mtg/eval"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/rules"
)

// dynamicAmountEstimate is the conservative magnitude assumed for an effect
// whose amount is derived from game state ("for each ...") and so is not known
// statically and has no announced {X}. One keeps a dynamic effect from
// dominating the score while still letting it register; an ability with an
// announced X uses that X instead (see dynamicEstimateFor).
const dynamicAmountEstimate = 1

// scoreActivateAbility scores activating an ability as the value of what it does
// minus the value of what it spends, both in the threatScoreUnit currency, so
// the agent activates an ability only when it is worth its cost given the board.
// This replaces the agent's blanket activate preference and stops it from
// "randomly" sacrificing or discarding for a marginal effect: a sacrifice cost
// is valued at the threat of the creatures it would consume, so paying three
// useless 1/1s to remove a real threat is favored while feeding good creatures
// for a small effect is not. Abilities the observation cannot resolve to a
// scorable summary (hand- or graveyard-activated) keep the routine activate
// score.
func scoreActivateAbility(obs rules.PlayerObservation, act action.Action, personality Personality) float64 {
	ability, ok := obs.ScorableActivatedAbility(act)
	if !ok {
		return scoreActivate
	}
	targets := activationTargets(act)
	dynamicEstimate := dynamicEstimateFor(activationXValue(act))
	return activationEffectValue(obs, targets, ability.Effect, personality, dynamicEstimate) -
		activationCostValue(obs, ability.Costs)
}

func activationTargets(act action.Action) []game.Target {
	payload, ok := act.ActivateAbilityPayload()
	if !ok {
		return nil
	}
	return payload.Targets
}

func activationXValue(act action.Action) int {
	payload, ok := act.ActivateAbilityPayload()
	if !ok {
		return 0
	}
	return payload.XValue
}

// dynamicEstimateFor picks the magnitude to assume for a dynamic ({X} or
// "for each") effect: the announced X when the agent paid one, since an X
// ability's dynamic amounts are dominated by X, and otherwise the conservative
// floor.
func dynamicEstimateFor(xValue int) float64 {
	if xValue > dynamicAmountEstimate {
		return float64(xValue)
	}
	return dynamicAmountEstimate
}

// activationEffectValue sums the value of an ability's modeled effects. Targeted
// removal, damage, and tapping are valued once by the chosen targets' threat
// (reusing targetingScore), while card, life, mana, token, tutor, and counter
// effects are valued by magnitude. Effects whose audience is not clearly the
// controller are left unvalued so the scorer never credits the agent for an
// opponent's gain. dynamicEstimate is the magnitude assumed for atoms whose
// amount is not statically known.
func activationEffectValue(obs rules.PlayerObservation, targets []game.Target, atoms []eval.EffectAtom, personality Personality, dynamicEstimate float64) float64 {
	var value float64
	targetingScored := false
	for i := range atoms {
		atom := atoms[i]
		switch atom.Kind {
		case eval.EffectPermanentRemoved, eval.EffectDamageDealt, eval.EffectPermanentTapped:
			if !targetingScored {
				value += targetingScore(obs, targets, personality)
				targetingScored = true
			}
		case eval.EffectCardsDrawn:
			value += controllerValue(atom, scoreCardValue, dynamicEstimate) * personality.cardValueScale()
		case eval.EffectCardsLost:
			value -= controllerValue(atom, scoreCardValue, dynamicEstimate) * personality.cardValueScale()
		case eval.EffectLifeGained:
			value += controllerValue(atom, scoreLifeValue, dynamicEstimate)
		case eval.EffectLifeLost:
			value -= controllerValue(atom, scoreLifeValue, dynamicEstimate)
		case eval.EffectManaAdded:
			value += atomMagnitude(atom, dynamicEstimate) * scoreManaValue
		case eval.EffectTokenCreated:
			value += atomMagnitude(atom, dynamicEstimate) * scoreTokenValue * personality.boardValueScale()
		case eval.EffectCardTutored:
			value += atomMagnitude(atom, dynamicEstimate) * scoreTutorValue
		case eval.EffectLandRamp:
			value += atomMagnitude(atom, dynamicEstimate) * scoreRampLand
		case eval.EffectCounterAdded:
			value += atomMagnitude(atom, dynamicEstimate) * scoreCounterValue
		default:
		}
	}
	return value
}

// controllerValue values an atom only when it clearly affects the ability's
// controller, returning zero otherwise. It is used for card and life effects,
// where crediting the wrong player would invert the sign.
func controllerValue(atom eval.EffectAtom, unit, dynamicEstimate float64) float64 {
	if atom.Affected != eval.AffectedYou {
		return 0
	}
	return atomMagnitude(atom, dynamicEstimate) * unit
}

func atomMagnitude(atom eval.EffectAtom, dynamicEstimate float64) float64 {
	if atom.IsDynamic {
		return dynamicEstimate
	}
	if atom.Amount <= 0 {
		return 1
	}
	return float64(atom.Amount)
}

// activationCostValue values the resources an ability spends from the agent's
// own cards and board: sacrificing other permanents (valued by the threat of the
// weakest creatures that would be sacrificed), discarding, and exiling. It does
// not value sacrificing the ability's own source, paying life, or tapping, so
// fetchlands and ordinary mana abilities are not penalized.
func activationCostValue(obs rules.PlayerObservation, costs []cost.Additional) float64 {
	var value float64
	for i := range costs {
		c := costs[i]
		switch c.Kind {
		case cost.AdditionalSacrifice:
			value += weakestCreaturesValue(obs, additionalAmount(c))
		case cost.AdditionalDiscard, cost.AdditionalExile:
			value += float64(additionalAmount(c)) * scoreCardValue
		default:
		}
	}
	return value
}

func additionalAmount(c cost.Additional) int {
	if c.Amount <= 0 {
		return 1
	}
	return c.Amount
}

// weakestCreaturesValue values sacrificing n creatures as the agent would: it
// sacrifices its least threatening creatures first (see loseLeastValuable), so
// the cost is the summed threat of the n weakest creatures the agent controls,
// scaled into the targetingScore currency. With fewer than n creatures it values
// only those available, matching what could actually be paid.
func weakestCreaturesValue(obs rules.PlayerObservation, n int) float64 {
	if n <= 0 {
		return 0
	}
	battlefield := obs.Battlefield()
	var threats []float64
	for i := range battlefield {
		permanent := battlefield[i]
		if permanent.Controller != obs.Player || !isCreaturePermanent(permanent) {
			continue
		}
		threats = append(threats, permanentThreat(permanent))
	}
	slices.Sort(threats)
	var sum float64
	for i := 0; i < n && i < len(threats); i++ {
		sum += threats[i]
	}
	return threatScoreUnit * sum
}
