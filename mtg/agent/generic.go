package agent

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules"
)

// Action-scoring weights for GenericStrategy. They encode a generic "good stuff"
// preference ordering (see docs/research/COMMANDER-AGENT-PLAYBOOK.md §6):
// develop mana, deploy threats and interaction, apply pressure, and never prefer
// passing over a productive play. They are deliberately coarse; threat, combat,
// mana-sequencing, and stack heuristics are refined by later strategy work.
const (
	scorePass        = 0.0
	scorePlayLand    = 100.0
	scoreCastBase    = 50.0
	scoreCastPerMana = 2.0
	scoreActivate    = 20.0
	scoreKeywordPlay = 10.0
	scoreCreature    = 15.0
	scoreAttackBase  = 15.0
	scoreAttackEach  = 2.0
	scoreBlockBase   = 12.0
	scoreBlockEach   = 2.0

	// scoreRemovalPerPower rewards interaction aimed at an opponent's permanent,
	// scaled by that permanent's effective power so the biggest threat is
	// preferred. scoreSelfTargetPenalty discourages aiming a spell at the
	// agent's own permanents, a cheap prune of obviously bad targeting.
	scoreRemovalPerPower   = 3.0
	scoreSelfTargetPenalty = 40.0
)

// GenericStrategy is a generic rule-based Commander strategy. It scores legal
// actions by a weighted preference for developing mana, deploying threats and
// interaction, and pressuring opponents, so an Agent using it plays sensibly
// without archetype-specific knowledge.
//
// Non-action choices use the BaselineStrategy behaviour until the dedicated
// choice heuristics replace it.
type GenericStrategy struct {
	BaselineStrategy
}

// ScoreAction implements Strategy. It is a pure function of the observation and
// action, so an Agent using it is deterministic.
func (GenericStrategy) ScoreAction(obs rules.PlayerObservation, act action.Action) float64 {
	switch act.Kind {
	case action.ActionPass:
		return scorePass
	case action.ActionPlayLand:
		return scorePlayLand
	case action.ActionCastSpell:
		return scoreCastSpell(obs, act)
	case action.ActionActivateAbility:
		return scoreActivate
	case action.ActionDeclareAttackers:
		return scoreDeclareAttackers(act)
	case action.ActionDeclareBlockers:
		return scoreDeclareBlockers(act)
	default:
		// Other productive actions (face-down casts, suspend, turn face up,
		// activated abilities without payloads) rank above passing.
		return scoreKeywordPlay
	}
}

func scoreCastSpell(obs rules.PlayerObservation, act action.Action) float64 {
	cast, ok := act.CastSpellPayload()
	if !ok {
		return scoreCastBase
	}
	score := scoreCastBase
	if card, found := handCard(obs, cast.CardID); found {
		score += float64(card.ManaValue) * scoreCastPerMana
		if isCreature(card) {
			score += scoreCreature
		}
	}
	score += targetingScore(obs, cast.Targets)
	return score
}

// targetingScore rewards aiming a spell at opponents' permanents (scaled by the
// target's effective power, so the biggest threat is preferred) and penalises
// aiming it at the agent's own permanents.
func targetingScore(obs rules.PlayerObservation, targets []game.Target) float64 {
	var score float64
	for i := range targets {
		target := targets[i]
		if target.Kind != game.TargetPermanent {
			continue
		}
		permanent, ok := permanentByID(obs, target.PermanentID)
		if !ok {
			continue
		}
		if permanent.Controller == obs.Player {
			score -= scoreSelfTargetPenalty
			continue
		}
		power := max(0, permanent.Power)
		score += scoreRemovalPerPower * float64(1+power)
	}
	return score
}

func scoreDeclareAttackers(act action.Action) float64 {
	declare, ok := act.DeclareAttackersPayload()
	if !ok {
		return scorePass
	}
	if len(declare.Attackers) == 0 {
		return scorePass
	}
	return scoreAttackBase + scoreAttackEach*float64(len(declare.Attackers))
}

func scoreDeclareBlockers(act action.Action) float64 {
	declare, ok := act.DeclareBlockersPayload()
	if !ok {
		return scorePass
	}
	if len(declare.Blockers) == 0 {
		return scorePass
	}
	return scoreBlockBase + scoreBlockEach*float64(len(declare.Blockers))
}

func handCard(obs rules.PlayerObservation, cardID id.ID) (rules.CardView, bool) {
	hand := obs.Hand()
	for i := range hand {
		if hand[i].CardInstanceID == cardID {
			return hand[i], true
		}
	}
	return rules.CardView{}, false
}

func permanentByID(obs rules.PlayerObservation, objectID id.ID) (rules.PermanentView, bool) {
	battlefield := obs.Battlefield()
	for i := range battlefield {
		if battlefield[i].ObjectID == objectID {
			return battlefield[i], true
		}
	}
	return rules.PermanentView{}, false
}

// isCreature reports whether the card view is a creature, used to weight board
// presence when deploying threats.
func isCreature(card rules.CardView) bool {
	return slices.Contains(card.Types, types.Creature)
}
