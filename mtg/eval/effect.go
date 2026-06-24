// Package eval reduces an engine ability to a small, value-oriented summary an
// agent (and reports) can score, instead of the ~100 execution primitives the
// rules engine resolves. The engine's primitives describe how to mutate game
// state; the eval types describe what an effect is worth, so strategy code never
// needs to know the primitive surface.
//
// The summary is a deliberately coarse heuristic aid: it captures the
// value-dominant majority of effects and degrades to a neutral classification
// for anything it does not yet model, never guessing a wrong value sign.
package eval

import "github.com/natefinch/council4/mtg/game/cost"

// EffectKind classifies what one consequence of an effect does in value terms,
// collapsing the engine's resolution primitives into a small vocabulary.
type EffectKind int

// EffectKind values name the value-relevant consequences an agent scores.
const (
	// EffectNeutral is the safe default for a consequence with no clear value
	// contribution, or one the translator does not yet model.
	EffectNeutral EffectKind = iota
	// EffectCardsDrawn is a player drawing cards (card advantage).
	EffectCardsDrawn
	// EffectCardsLost is a player discarding or milling cards from a zone they
	// own (card disadvantage).
	EffectCardsLost
	// EffectLifeGained is a player gaining life.
	EffectLifeGained
	// EffectLifeLost is a player losing or paying life as an effect.
	EffectLifeLost
	// EffectDamageDealt is damage dealt to a player or permanent.
	EffectDamageDealt
	// EffectPermanentRemoved is a permanent destroyed, exiled, or bounced.
	EffectPermanentRemoved
	// EffectPermanentTapped is a permanent tapped (tempo).
	EffectPermanentTapped
	// EffectManaAdded is mana added to a pool (ramp/fixing).
	EffectManaAdded
	// EffectTokenCreated is one or more tokens created (board presence).
	EffectTokenCreated
	// EffectCounterAdded is counters placed on a permanent.
	EffectCounterAdded
	// EffectCardTutored is a library search moving cards to hand or battlefield.
	EffectCardTutored
)

// Affected identifies who or what a consequence affects, which carries its value
// sign: drawing cards or gaining life is good for AffectedYou and bad for an
// opponent. The translator sets it only when the engine reference makes the
// audience unambiguous, leaving AffectedUnknown otherwise so a scorer never
// infers a wrong sign.
type Affected int

// Affected values name the audiences a scorer distinguishes.
const (
	// AffectedUnknown means the audience could not be resolved unambiguously.
	AffectedUnknown Affected = iota
	// AffectedYou is the ability's controller — the agent itself.
	AffectedYou
	// AffectedEachOpponent is every opponent of the controller.
	AffectedEachOpponent
	// AffectedTarget is a chosen target; the scorer resolves it against the
	// action's targets (a removed target's value is its threat, for example).
	AffectedTarget
)

// EffectAtom is one value-relevant consequence of an effect.
type EffectAtom struct {
	Kind EffectKind
	// Amount is the magnitude (cards, damage, mana, life, counters). It is 0 when
	// IsDynamic is true or the kind carries no amount.
	Amount int
	// IsDynamic reports that the amount is {X}, "for each ...", or otherwise
	// derived from game state at resolution, so a scorer must estimate it rather
	// than trust Amount.
	IsDynamic bool
	Affected  Affected
}

// ScorableAbility is an ability reduced to the cost and effect terms an agent
// scores. Costs reuse the engine's enum-based additional-cost vocabulary; Effect
// is the value-oriented summary produced by ScorableEffect.
type ScorableAbility struct {
	Costs  []cost.Additional
	Effect []EffectAtom
}
