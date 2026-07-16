package game

import (
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/opt"
)

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

	// AggregateEventPlayerHandSize is the number of cards in the triggering
	// player's hand ("if that player has two or fewer cards in hand"). It reads
	// the player recorded on the triggering step event, so it resolves only in a
	// phase/step trigger's intervening-if context (each opponent's or each
	// player's upkeep) and fails closed elsewhere.
	AggregateEventPlayerHandSize

	// AggregateAnyOpponentDamageTakenThisTurn is the largest total damage dealt
	// to a single non-eliminated opponent of the context controller during the
	// current turn. An "an opponent was dealt N or more damage this turn"
	// existential gate (Spinerock Knoll) is satisfied exactly when this maximum
	// is at least N.
	AggregateAnyOpponentDamageTakenThisTurn

	// AggregateMinPlayerLibrarySize is the smallest library size among all
	// non-eliminated players. A "a library has N or fewer cards in it" existential
	// gate (Shelldock Isle) is satisfied exactly when this minimum is at most N.
	AggregateMinPlayerLibrarySize

	// AggregateAnyOpponentLifeLostThisTurn is the largest total life lost by a
	// single non-eliminated opponent of the context controller during the current
	// turn. Unlike AggregateAnyOpponentDamageTakenThisTurn, it counts every kind
	// of life loss — combat and noncombat damage (CR 120.3), life paid as a cost,
	// and direct "loses life" effects — because all of these are emitted as
	// per-player life-loss events. An "an opponent lost N or more life this turn"
	// existential gate (Bloodchief Ascension) is satisfied exactly when this
	// maximum is at least N.
	AggregateAnyOpponentLifeLostThisTurn

	// AggregateAttackersInBatchAttackedController is the number of attackers in
	// the triggering attack batch (the EventAttackerDeclared events sharing the
	// trigger event's simultaneous batch) that were declared attacking the
	// context controller as a player directly. Attacks on another player, on any
	// planeswalker (including the controller's own), or on a battle are excluded,
	// so it counts only direct player-attacks on the controller. It reads the
	// declared batch from the triggering event rather than live combat, so it is
	// stable across source-controller changes and creatures leaving combat, and
	// it backs the "if none of those creatures attacked you" gate (Firemane
	// Commando) when compared as "at most zero". It resolves only in an attacker-
	// declared trigger's intervening-if context and fails closed elsewhere.
	AggregateAttackersInBatchAttackedController

	// AggregateEventSpellManaFromCreaturesSpentToCast is the amount of the mana
	// spent to cast a spell-cast trigger's triggering spell that was produced by
	// creature permanents ("if three or more mana from creatures was spent to
	// cast it", Inga and Esika). Each mana carries "from a creature" provenance
	// fixed when it was produced, so it counts creature mana even after that
	// creature changes type or leaves. It resolves only in a spell-cast trigger's
	// intervening-if context, where the triggering event records the creature
	// mana spent; it fails closed elsewhere.
	AggregateEventSpellManaFromCreaturesSpentToCast

	// AggregateControllerDevotion is the context controller's devotion to the
	// colors listed in the comparison's Colors field: the number of mana symbols
	// of those colors among the mana costs of the permanents that player controls
	// (CR 700.5). A hybrid or Phyrexian symbol counts once when it matches any
	// listed color, so multicolor devotion counts each qualifying symbol a single
	// time. It backs the Theros Gods' "as long as your devotion to <color(s)> is
	// less than N" type-changing static, compared as "less than N". The evaluated
	// colors travel as typed data on Colors rather than being derived from any
	// card text.
	AggregateControllerDevotion
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

	// ValueAmount, when present, supplies the comparison threshold as a
	// resolution-time dynamic amount (CR 608.2c) instead of the fixed Value. It
	// backs comparisons between two live quantities, such as Thassa's Oracle's
	// "if X is greater than or equal to the number of cards in your library",
	// where Aggregate is AggregateControllerLibrarySize, Op is LessOrEqual, and
	// ValueAmount is the controller's devotion (X). It is evaluated against the
	// condition's resolving stack object and controller; an amount that cannot be
	// evaluated (for example, a stackless condition context) fails the comparison
	// closed. When absent, the fixed Value is used.
	ValueAmount opt.V[DynamicAmount]

	// Colors lists the colors evaluated by color-parameterized aggregates. It is
	// consulted only by AggregateControllerDevotion, where it names the colors
	// whose mana symbols are counted; other aggregates ignore it. The colors
	// travel as typed data so devotion comparisons carry no card text.
	Colors []color.Color
}
