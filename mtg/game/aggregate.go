package game

import "github.com/natefinch/council4/mtg/game/compare"

// AggregateKind identifies a player- or board-derived quantity that a
// Condition can compare against a threshold. It replaces the family of
// dedicated comparator fields (AtLeast/AtMost/Exactly) that previously encoded
// both the quantity and the comparator direction separately. A single quantity
// is described once by an AggregateKind; the comparator direction is carried by
// compare.Op, so adding support for "at least", "at most", and "exactly" for a
// quantity requires no new Condition fields.
type AggregateKind int

const (
	// AggregateNone is the zero value and identifies no quantity.
	AggregateNone AggregateKind = iota

	// AggregateControllerLife is the context controller's current life total.
	AggregateControllerLife

	// AggregateControllerLifeAboveStarting is the context controller's current
	// life total minus their starting life total ("N life more than your
	// starting life total").
	AggregateControllerLifeAboveStarting

	// AggregateControllerHandSize is the number of cards in the context
	// controller's hand.
	AggregateControllerHandSize

	// AggregateControllerLibrarySize is the number of cards in the context
	// controller's library.
	AggregateControllerLibrarySize

	// AggregateControllerGraveyardCardCount is the number of cards in the
	// context controller's graveyard.
	AggregateControllerGraveyardCardCount

	// AggregateControllerGraveyardCardTypeCount is the number of distinct card
	// types among cards in the context controller's graveyard (delirium).
	AggregateControllerGraveyardCardTypeCount

	// AggregateControllerBasicLandTypeCount is the number of distinct basic land
	// types among lands the context controller controls (domain).
	AggregateControllerBasicLandTypeCount

	// AggregateControllerCreaturePowerDiversity is the number of distinct power
	// values among creatures the context controller controls (coven).
	AggregateControllerCreaturePowerDiversity

	// AggregateOpponentCount is the number of non-eliminated opponents of the
	// context controller.
	AggregateOpponentCount

	// AggregateAttackersAttackingController is the number of attackers attacking
	// the context controller or a planeswalker they control.
	AggregateAttackersAttackingController

	// AggregateControllerGainedLifeThisTurn is the total life the context
	// controller has gained so far this turn.
	AggregateControllerGainedLifeThisTurn

	// AggregateSpellX is the resolving spell's chosen value of {X}.
	AggregateSpellX

	// AggregateControllerGraveyardPermanentCardCount is the number of permanent
	// cards in the context controller's graveyard (descend).
	AggregateControllerGraveyardPermanentCardCount

	// AggregateControllerGraveyardManaValueCount is the number of distinct mana
	// values among cards in the context controller's graveyard.
	AggregateControllerGraveyardManaValueCount

	// AggregateAnyOpponentGraveyardCardCount is the largest number of cards in a
	// single opponent's graveyard among the context controller's opponents. A
	// "an opponent has N or more cards in their graveyard" existential gate is
	// satisfied exactly when this maximum is at least N.
	AggregateAnyOpponentGraveyardCardCount

	// AggregateEventSpellManaSpentToCast is the total amount of mana spent to
	// cast a spell-cast trigger's triggering spell ("if no mana was spent to
	// cast it", "if at least four mana was spent to cast it"). It resolves only
	// in a spell-cast trigger's intervening-if context, where the triggering
	// event records the mana spent; it fails closed elsewhere.
	AggregateEventSpellManaSpentToCast
)

// AggregateComparison compares a player- or board-derived quantity against a
// threshold using a typed comparator. Conditions hold a slice of these so
// multiple quantities can be ANDed without one Condition field per quantity.
type AggregateComparison struct {
	// Aggregate identifies which quantity to evaluate.
	Aggregate AggregateKind

	// Op is the comparator direction applied to the evaluated quantity.
	Op compare.Op

	// Value is the threshold the quantity is compared against.
	Value int
}
