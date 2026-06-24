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

	// Effect- and cost-value units for scoring an activated ability by what it
	// does and what it spends (see activation.go). They are expressed in the same
	// currency as targetingScore, where threatScoreUnit (3) × a permanent's
	// threat values removing it: a 1/1 (threat 2) is worth 6, a 3/3 (threat 4) is
	// worth 12, a 5/5 (threat 6) is worth 18. The per-kind values are calibrated
	// against that scale and justified individually below.
	//
	// scoreCardValue: a card is worth about a mid-sized creature (a 3/3 ≈ 12),
	// reflecting that card advantage is roughly one average permanent.
	scoreCardValue = 12.0
	// scoreLifeValue: one life point is minor; eight life ≈ two thirds of a card.
	scoreLifeValue = 1.0
	// scoreManaValue: mana added by a non-mana ability (the rare ability that
	// adds mana yet still targets, so it is not a mana ability handled by the
	// payment system) is incidental — unspent mana empties — so it is valued well
	// below a card.
	scoreManaValue = 3.0
	// scoreTokenValue: a token is board presence worth roughly a small creature.
	// The IR carries the token count but not its size, so this is a deliberately
	// modest per-token estimate rather than a precise creature value.
	scoreTokenValue = 8.0
	// scoreTutorValue: tutoring a card to hand is a drawn card the agent chooses,
	// so it is worth slightly more than a raw draw.
	scoreTutorValue = 14.0
	// scoreCounterValue: a +1/+1 counter is about one point of power, which the
	// threat model values at roughly targetingScore of a single power point.
	scoreCounterValue = 4.0

	// Ramp incentives bias the agent toward accelerating its mana early. A mana
	// source (rock or dork) is prized over a land-fetch because it also fixes and
	// survives; both bonuses decay linearly to zero once the agent already has
	// scoreRampDecayFrom untapped mana sources, so ramp is preferred on the early
	// turns where it compounds and ignored once the mana base is developed.
	scoreRampSource    = 30.0
	scoreRampLand      = 20.0
	scoreRampDecayFrom = 6.0

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
		return scoreActivateAbility(obs, act, s.Personality)
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
		score += rampBonus(obs, card)
		score -= holdUpPenalty(obs, card, personality)
	}
	score += targetingScore(obs, cast.Targets, personality)
	return score
}

// rampBonus rewards casting ramp — a mana rock or dork, or a spell that puts a
// land onto the battlefield — early, when the extra mana most accelerates the
// agent's development. The bonus fades as the agent's available mana grows, so a
// ramp spell is prized on turn two and ignorable late, and a mana source is
// valued more than a land-fetch because it also fixes and survives.
func rampBonus(obs rules.PlayerObservation, card rules.CardView) float64 {
	if !card.ProducesMana && !card.RampsLand {
		return 0
	}
	bonus := scoreRampLand
	if card.ProducesMana {
		bonus = scoreRampSource
	}
	available := availableManaSources(obs)
	if decay := scoreRampDecayFrom - available; decay > 0 {
		return bonus * float64(decay) / scoreRampDecayFrom
	}
	return 0
}

// availableManaSources counts the untapped mana sources the agent controls, a
// coarse proxy for how developed its mana is, used to fade the ramp bonus.
func availableManaSources(obs rules.PlayerObservation) int {
	battlefield := obs.Battlefield()
	count := 0
	for i := range battlefield {
		permanent := battlefield[i]
		if permanent.Controller == obs.Player && permanent.ProducesMana && !permanent.Tapped {
			count++
		}
	}
	return count
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
