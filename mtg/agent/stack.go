package agent

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules"
)

// Stack-interaction weights for GenericStrategy (see
// docs/research/COMMANDER-AGENT-PLAYBOOK.md §9). Reactive spells (counters and
// instant-speed removal) are scored by interaction value minus a card-economy
// cost, so a low-value answer scores below Pass and the agent holds it for a
// better target rather than wasting it.
const (
	// scoreCounterPerMana values countering a spell by its mana value, the best
	// public proxy for how impactful resolving it would be.
	scoreCounterPerMana = 8.0
	// scoreCounterThreatType rewards countering a spell that would resolve into a
	// lasting threat (a creature or planeswalker) over an ephemeral one.
	scoreCounterThreatType = 20.0
	// scoreCounterCardCost is the card-economy cost of spending a counter. A
	// spell must clear it to be worth countering, so cheap spells are let go.
	scoreCounterCardCost = 35.0
	// scoreCounterOwnSpell makes the agent never counter its own spell.
	scoreCounterOwnSpell = 100.0

	// scoreRemovalCardCost is the card-economy cost of spending instant-speed
	// removal. A target's threat must clear it, so removal is held for a target
	// worth a card rather than spent on a small creature.
	scoreRemovalCardCost = 12.0

	// scoreHoldRemoval discourages firing instant-speed removal on the agent's
	// own turn, nudging it to hold the answer for an opponent's turn where it can
	// react with more information (playbook §9.3). It is modest so a large enough
	// threat is still removed immediately.
	scoreHoldRemoval = 15.0

	// scoreOwnInstantValue is the modest value of an instant aimed at the agent's
	// own permanent — typically a combat trick, protection, or pump. The coarse
	// model cannot tell a beneficial instant from a destructive one, so this
	// value is paid for any own-target instant. It is bounded behaviourally
	// rather than semantically: at 6.0 it sits below every productive action
	// (land, cast, activate) and below removing any worthwhile enemy target, so
	// the agent only ever fires an instant at its own board when nothing better
	// than Pass exists, while still being able to use real tricks.
	scoreOwnInstantValue = 6.0
)

// reactiveSpellScore scores a reactive interaction spell — a counter (it targets
// a stack object) or instant-speed removal (an instant aimed only at
// permanents) — reporting ok=false for any other cast so the caller uses the
// proactive scorer. Unlike a proactive cast it has no base floor, so a low-value
// answer scores below Pass and the agent holds it.
func reactiveSpellScore(obs rules.PlayerObservation, card rules.CardView, cast action.CastSpellAction, personality Personality) (float64, bool) {
	if hasStackTarget(cast.Targets) {
		return counterScore(obs, cast.Targets, personality), true
	}
	if isInstant(card) && permanentTargetsOnly(cast.Targets) {
		return removalScore(obs, cast.Targets, personality), true
	}
	return 0, false
}

// counterScore values countering each opposing spell the cast targets by its
// impact (mana value plus a bonus for lasting threats) minus the card-economy
// cost of the counter, and strongly penalises countering the agent's own spell.
// Risk tolerance lowers the card-economy cost, so the agent counters more freely.
func counterScore(obs rules.PlayerObservation, targets []game.Target, personality Personality) float64 {
	stack := obs.Stack()
	cardCost := scoreCounterCardCost * personality.cardCostScale()
	var score float64
	for i := range targets {
		if targets[i].Kind != game.TargetStackObject {
			continue
		}
		object, ok := stackObjectByID(stack, targets[i].StackObjectID)
		if !ok {
			continue
		}
		if object.Controller == obs.Player {
			score -= scoreCounterOwnSpell
			continue
		}
		score += spellInteractValue(object) - cardCost
	}
	return score
}

// spellInteractValue is how much resolving the spell would matter to the agent,
// approximated by its mana value with a bonus for spells that leave a lasting
// permanent threat.
func spellInteractValue(object rules.StackObjectView) float64 {
	value := scoreCounterPerMana * float64(object.ManaValue)
	if isThreatSpell(object) {
		value += scoreCounterThreatType
	}
	return value
}

// removalScore values aiming an instant-speed permanent spell at each target. An
// opposing permanent is valued by its threat minus the card-economy cost, so a
// target below the cost yields a negative score and the agent holds the spell
// for a worthier one. Risk tolerance lowers the card-economy cost, so the agent
// spends removal more freely. The agent's own permanent yields a modest positive
// value, treating the cast as a beneficial trick that stays castable but ranks
// below removing a real threat.
func removalScore(obs rules.PlayerObservation, targets []game.Target, personality Personality) float64 {
	cardCost := scoreRemovalCardCost * personality.cardCostScale()
	var score float64
	targetsOpponent := false
	for i := range targets {
		permanent, ok := permanentByID(obs, targets[i].PermanentID)
		if !ok {
			continue
		}
		if permanent.Controller == obs.Player {
			score += scoreOwnInstantValue
			continue
		}
		targetsOpponent = true
		score += threatScoreUnit*permanentThreat(permanent) - cardCost
	}
	if targetsOpponent {
		score -= holdRemovalTiming(obs, personality)
	}
	return score
}

// holdRemovalTiming discourages spending instant-speed removal on the agent's own
// turn. An instant can be held for an opponent's turn — ideally their end step,
// after they have committed their mana — where it answers what they actually did
// with the most information and the least waste (playbook §9.3). The penalty
// applies only on the agent's own turn and is eased by risk tolerance (a
// risk-tolerant agent fires removal more freely); it is modest, so a big enough
// threat is still worth removing immediately rather than held.
func holdRemovalTiming(obs rules.PlayerObservation, personality Personality) float64 {
	if obs.Turn.ActivePlayer != obs.Player {
		return 0
	}
	return scoreHoldRemoval * personality.cardCostScale()
}

func isThreatSpell(object rules.StackObjectView) bool {
	return slices.Contains(object.Types, types.Creature) ||
		slices.Contains(object.Types, types.Planeswalker)
}

func hasStackTarget(targets []game.Target) bool {
	for i := range targets {
		if targets[i].Kind == game.TargetStackObject {
			return true
		}
	}
	return false
}

// permanentTargetsOnly reports whether the cast targets at least one object and
// every target is a permanent, the shape of single- or multi-target removal and
// of own-permanent combat tricks.
func permanentTargetsOnly(targets []game.Target) bool {
	if len(targets) == 0 {
		return false
	}
	for i := range targets {
		if targets[i].Kind != game.TargetPermanent {
			return false
		}
	}
	return true
}

func isInstant(card rules.CardView) bool {
	return slices.Contains(card.Types, types.Instant)
}

func stackObjectByID(stack []rules.StackObjectView, objectID id.ID) (rules.StackObjectView, bool) {
	for i := range stack {
		if stack[i].ID == objectID {
			return stack[i], true
		}
	}
	return rules.StackObjectView{}, false
}
