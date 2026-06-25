package agent

import "math/rand/v2"

// Personality tunes how a GenericStrategy plays without changing its rules
// knowledge (ADR 0003; docs/research/card-game-ai-research.md §7.3). Every knob
// is an additive bias whose zero value is neutral, so the zero Personality
// reproduces the plain generic strategy exactly. Positive values push the agent
// further from neutral; the effects are deterministic given a fixed noise source.
type Personality struct {
	// Aggression makes the agent attack more and deploy threats more eagerly.
	Aggression float64
	// RiskTolerance makes the agent hold up less reactive mana, spend
	// interaction more freely, and attack into possible blocks.
	RiskTolerance float64
	// PoliticsWeight makes the agent weight an opponent's overall threat more
	// heavily when choosing whom to attack and target, focusing the table's
	// biggest threat.
	PoliticsWeight float64
	// NoiseMagnitude is the maximum random jitter added to each action score,
	// for behavioural variety. It has no effect without a noise source, and zero
	// keeps scoring deterministic.
	NoiseMagnitude float64

	rng *rand.Rand
}

// Personality knob units. Each scales one knob into a concrete score adjustment;
// they are deliberately coarse and only matter relative to the base weights.
const (
	aggressionAttackUnit = 3.0 // extra attack value per attacker per aggression point
	aggressionDeployUnit = 4.0 // extra creature-deploy value per aggression point
	riskLossUnit         = 3.0 // attack-into-block penalty forgiven per risk point
	riskHoldUpUnit       = 0.5 // hold-up penalty removed per risk point (as a fraction)
	riskCardCostUnit     = 0.5 // reactive card-economy cost removed per risk point
	politicsThreatUnit   = 1.0 // extra opponent-threat weight per politics point

	// aggressionBoardUnit scales up the value an aggressive agent places on
	// board-presence effects (tokens), per aggression point.
	aggressionBoardUnit = 0.15
	// aggressionCardUnit scales down the value an aggressive agent places on card
	// advantage (drawing, and the cost of losing cards), per aggression point, so
	// a controlling (low-aggression) agent values cards relatively more.
	aggressionCardUnit = 0.10
)

// WithNoiseSource returns a copy of the personality that draws decision noise
// from rng. Give each seat its own seeded rng so games stay reproducible and
// parallel simulations never share a source.
func (p Personality) WithNoiseSource(rng *rand.Rand) Personality {
	p.rng = rng
	return p
}

func (p Personality) attackBonus() float64 {
	return p.Aggression * aggressionAttackUnit
}

func (p Personality) deployBonus() float64 {
	return p.Aggression * aggressionDeployUnit
}

func (p Personality) attackLossReduction() float64 {
	return p.RiskTolerance * riskLossUnit
}

// holdUpScale scales the hold-up penalty: higher risk tolerance keeps less mana
// open. It never goes below zero.
func (p Personality) holdUpScale() float64 {
	return clampFloat(1-p.RiskTolerance*riskHoldUpUnit, 0, 1)
}

// cardCostScale scales the card-economy cost of reactive spells: higher risk
// tolerance spends interaction more freely. It never goes below zero.
func (p Personality) cardCostScale() float64 {
	return clampFloat(1-p.RiskTolerance*riskCardCostUnit, 0, 1)
}

func (p Personality) extraThreatWeight() float64 {
	return p.PoliticsWeight * politicsThreatUnit
}

// boardValueScale scales the value of board-presence effects (tokens): an
// aggressive agent weights developing a board more heavily. It never goes below
// zero.
func (p Personality) boardValueScale() float64 {
	return max(0, 1+p.Aggression*aggressionBoardUnit)
}

// cardValueScale scales the value of card advantage (drawing, and the cost of
// losing cards): an aggressive agent cares less about cards, a controlling one
// more. It never goes below zero.
func (p Personality) cardValueScale() float64 {
	return max(0, 1-p.Aggression*aggressionCardUnit)
}

// noise returns the random jitter to add to one action score, or zero when no
// noise is configured. It advances the noise source, so scores depend on the
// order the engine scores actions — deterministic for a fixed seed.
func (p Personality) noise() float64 {
	if p.NoiseMagnitude <= 0 || p.rng == nil {
		return 0
	}
	return (p.rng.Float64()*2 - 1) * p.NoiseMagnitude
}

func clampFloat(value, low, high float64) float64 {
	return max(low, min(value, high))
}
