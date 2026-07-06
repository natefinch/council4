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

// Life-cost weights for scoring an activated ability's life payment (see
// lifePaymentValue). They keep a life cost cheap while the agent is healthy and
// make it climb steeply as death approaches, so the agent stops paying life for
// marginal value instead of activating a life-cost ability until it dies.
const (
	// lowLifeThreshold is the remaining life below which each point of life is
	// weighted more heavily, so an activation's life cost rises sharply once the
	// agent is genuinely endangered.
	lowLifeThreshold = 10.0
	// prohibitiveActivationCost is a cost no modeled effect can outweigh, used
	// for a life payment that would drop the agent to 0 or less (a lethal cost
	// the agent must never pay for an activated ability).
	prohibitiveActivationCost = 1e6
)

// scoreActivateAbility scores activating an ability as the value of what it does
// minus the value of what it spends, both in the threatScoreUnit currency, so
// the agent activates an ability only when it is worth its cost given the board.
// This replaces the agent's blanket activate preference and stops it from
// "randomly" sacrificing, discarding, or paying life for a marginal effect: a
// sacrifice cost is valued at the threat of the creatures it would consume, so
// paying three useless 1/1s to remove a real threat is favored while feeding
// good creatures for a small effect is not, and a life cost grows toward
// prohibitive as the agent nears death (see lifePaymentValue) so it never pays
// life for a do-nothing ability until it dies. Abilities the observation cannot
// resolve to a scorable summary (hand- or graveyard-activated) keep the routine
// activate score.
func scoreActivateAbility(obs rules.PlayerObservation, act action.Action, personality Personality) float64 {
	if obs.IsManaAbilityActivation(act) {
		// Activating a mana ability standalone only floats mana that empties at end
		// of step; the payment system taps mana sources as it pays for spells and
		// abilities. Scoring it above passing would loop forever on a mana-neutral
		// ability that pays for itself (Skyshroud Elf, "{1}: Add {R} or {W}").
		return scoreNoOpActivation
	}
	if obs.RepeatedFreeActivation(act) {
		// A free ability re-activated this turn almost never changes anything the
		// first activation did not; scoring it above passing would loop forever
		// (equip {0}, a tapped-out "{X}" ability at X=0). See scoreNoOpActivation.
		return scoreNoOpActivation
	}
	ability, ok := obs.ScorableActivatedAbility(act)
	if !ok {
		return scoreActivate
	}
	targets := activationTargets(act)
	dynamicEstimate := dynamicEstimateFor(activationXValue(act))
	return activationEffectValue(obs, targets, ability.Effect, personality, dynamicEstimate) -
		activationCostValue(obs, ability.Costs, dynamicEstimate)
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
// own cards, board, and life total: sacrificing other permanents (valued by the
// threat of the weakest creatures that would be sacrificed), discarding,
// exiling, and paying life (CR 118.4). It does not value sacrificing the
// ability's own source or tapping, so fetchlands and ordinary mana abilities are
// not penalized. dynamicEstimate is the magnitude assumed for a pay-X-life cost
// whose amount is the announced X rather than a fixed number.
func activationCostValue(obs rules.PlayerObservation, costs []cost.Additional, dynamicEstimate float64) float64 {
	var value float64
	for i := range costs {
		c := costs[i]
		switch c.Kind {
		case cost.AdditionalSacrifice:
			value += weakestCreaturesValue(obs, additionalAmount(c))
		case cost.AdditionalDiscard, cost.AdditionalExile:
			value += float64(additionalAmount(c)) * scoreCardValue
		case cost.AdditionalPayLife:
			value += lifePaymentValue(obs, payLifeAmount(c, dynamicEstimate))
		default:
		}
	}
	return value
}

// payLifeAmount is the life an AdditionalPayLife cost spends. A cost whose amount
// is the announced X or a rules-derived value is not statically known, so it
// uses the same conservative dynamic estimate the effect side uses; a fixed cost
// uses its printed amount.
func payLifeAmount(c cost.Additional, dynamicEstimate float64) int {
	if c.AmountFromX || c.AmountDynamic != cost.AdditionalDynamicAmountNone {
		return int(dynamicEstimate)
	}
	return additionalAmount(c)
}

// lifePaymentValue prices paying n life as an activation cost. Life is cheap
// while the agent is healthy (roughly scoreLifeValue per point) and grows
// steeply as its remaining life falls below lowLifeThreshold, so the agent
// stops spending life for marginal value as it nears death. A payment that would
// leave the agent at 0 or less is prohibitive: a player with 0 or less life
// loses the game (CR 704.5a), so the agent never pays lethal life for an
// activated ability's effect, no matter how it is valued.
func lifePaymentValue(obs rules.PlayerObservation, n int) float64 {
	if n <= 0 {
		return 0
	}
	remaining := obs.Life(obs.Player) - n
	if remaining <= 0 {
		return prohibitiveActivationCost
	}
	value := float64(n) * scoreLifeValue
	if float64(remaining) < lowLifeThreshold {
		value *= lowLifeThreshold / float64(remaining)
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
