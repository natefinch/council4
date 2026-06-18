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
// passing over a productive play. They are deliberately coarse; the threat,
// combat, mana-sequencing (see mana.go), and stack heuristics refine the raw
// action scores.
const (
	scorePass        = 0.0
	scorePlayLand    = 100.0
	scoreCastBase    = 50.0
	scoreCastPerMana = 2.0
	scoreActivate    = 20.0
	scoreKeywordPlay = 10.0
	scoreCreature    = 15.0

	// scoreSelfTargetPenalty discourages aiming a spell at the agent's own
	// permanents or face, a cheap prune of obviously bad targeting. Interaction
	// aimed at opponents is rewarded by the threat model (see threat.go), so the
	// biggest threat is preferred.
	scoreSelfTargetPenalty = 40.0
)

// GenericStrategy is a generic rule-based Commander strategy. It scores legal
// actions by a weighted preference for developing mana, deploying threats and
// interaction, and pressuring opponents, so an Agent using it plays sensibly
// without archetype-specific knowledge.
//
// Non-action choices use the choice heuristics in choices.go; mana sequencing,
// combat, and stack interaction refine the raw action scores.
type GenericStrategy struct {
	BaselineStrategy

	// Profile is the optional once-per-match analysis of the agent's own deck
	// (see AnalyzeDeck). It is nil when no deck analysis was supplied. The
	// generic action scoring does not depend on it; it is exposed so deck-aware
	// strategy tuning can consult the deck's archetype, curve, and power band.
	Profile *DeckProfile

	// Personality tunes how the strategy plays (aggression, risk tolerance,
	// politics weight, decision noise). Its zero value is neutral, so a
	// GenericStrategy with no personality set plays the plain generic strategy.
	Personality Personality
}

// ScoreAction implements Strategy. With a neutral Personality it is a pure
// function of the observation and action; a non-zero NoiseMagnitude adds seeded
// jitter, which stays deterministic for a fixed noise source and scoring order.
func (s GenericStrategy) ScoreAction(obs rules.PlayerObservation, act action.Action) float64 {
	return s.baseScore(obs, act) + s.Personality.noise()
}

func (s GenericStrategy) baseScore(obs rules.PlayerObservation, act action.Action) float64 {
	switch act.Kind {
	case action.ActionPass:
		return scorePass
	case action.ActionPlayLand:
		return scoreLandPlay(obs, act)
	case action.ActionCastSpell:
		return scoreCastSpell(obs, act, s.Personality)
	case action.ActionActivateAbility:
		return scoreActivate
	case action.ActionDeclareAttackers:
		return scoreAttackDeclarations(obs, act, s.Personality)
	case action.ActionDeclareBlockers:
		return scoreBlockDeclarations(obs, act)
	default:
		// Other productive actions (face-down casts, suspend, turn face up,
		// activated abilities without payloads) rank above passing.
		return scoreKeywordPlay
	}
}

func scoreCastSpell(obs rules.PlayerObservation, act action.Action, personality Personality) float64 {
	cast, ok := act.CastSpellPayload()
	if !ok {
		return scoreCastBase
	}
	card, found := handCard(obs, cast.CardID)
	if found {
		if score, reactive := reactiveSpellScore(obs, card, cast, personality); reactive {
			return score
		}
	}
	score := scoreCastBase
	if found {
		score += float64(card.ManaValue) * scoreCastPerMana
		if isCreature(card) {
			score += scoreCreature + personality.deployBonus()
		}
		score -= holdUpPenalty(obs, card, personality)
	}
	score += targetingScore(obs, cast.Targets, personality)
	return score
}

// targetingScore rewards aiming a spell at the most dangerous opponent
// permanents and players (using the threat model, so the biggest threat is
// preferred and a near-dead player is not kingmade) and penalises aiming it at
// the agent's own permanents or face.
func targetingScore(obs rules.PlayerObservation, targets []game.Target, personality Personality) float64 {
	var score float64
	var model *ThreatModel
	for i := range targets {
		target := targets[i]
		switch target.Kind {
		case game.TargetPermanent:
			permanent, ok := permanentByID(obs, target.PermanentID)
			if !ok {
				continue
			}
			if permanent.Controller == obs.Player {
				score -= scoreSelfTargetPenalty
				continue
			}
			score += threatScoreUnit * permanentThreat(permanent)
		case game.TargetPlayer:
			if target.PlayerID == obs.Player {
				score -= scoreSelfTargetPenalty
				continue
			}
			if model == nil {
				built := NewThreatModel(obs)
				model = &built
			}
			score += (threatScoreUnit + personality.extraThreatWeight()) * model.PlayerThreat(target.PlayerID)
		default:
		}
	}
	return score
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
