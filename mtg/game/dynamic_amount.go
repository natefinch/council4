package game

import (
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/zone"
)

// DynamicAmountKind identifies a rules-derived integer for effect resolution.
// Variable values such as X and "equal to" quantities are determined as the
// resolving instruction applies unless the card text says otherwise
// (CR 107.3, CR 608.2c).
type DynamicAmountKind int

// Dynamic amount kind values identify supported effect-resolution quantities.
const (
	DynamicAmountNone DynamicAmountKind = iota
	DynamicAmountConstant
	DynamicAmountX
	DynamicAmountTargetPower
	DynamicAmountTargetToughness
	DynamicAmountTargetManaValue
	DynamicAmountTargetCounters
	DynamicAmountControllerLife
	DynamicAmountControllerHandSize
	DynamicAmountControllerGraveyardSize
	DynamicAmountControllerBasicLandTypeCount
	DynamicAmountCountSelector
	DynamicAmountCountCardsInZone
	DynamicAmountPreviousEffectResult
	DynamicAmountOpponentCount
	DynamicAmountEventDamage
	DynamicAmountPreviousEffectExcessDamage
	DynamicAmountObjectPower
	// DynamicAmountEventCardCount is the number of cards drawn or discarded in
	// the triggering event batch (CR 122, CR 700.4). It scales draw and discard
	// triggers such as "for each card discarded this way" by reading the
	// simultaneous batch recorded on the resolving ability's TriggerEvent.
	DynamicAmountEventCardCount
	// DynamicAmountObjectToughness is the toughness of a referenced object,
	// read from the live permanent or, once it has left the battlefield, from
	// its last-known snapshot (CR 608.2h). It backs "gain/lose life equal to
	// its toughness" riders whose subject was exiled, destroyed, or bounced by
	// an earlier effect in the same resolution.
	DynamicAmountObjectToughness
	// DynamicAmountObjectManaValue is the mana value of a referenced object,
	// read from the live permanent's printed mana cost or, once it has left the
	// battlefield, from its last-known snapshot (CR 202.3, CR 608.2h). Unlike
	// DynamicAmountTargetManaValue, which reads only the live target permanent,
	// this kind falls back to last-known information so a "gain/lose life equal
	// to its mana value" rider reads the same object after a zone move. It backs
	// destroy-then-life staples (Feed the Swarm, Divine Offering) and linked
	// graveyard-return riders (Reanimate). Added last so existing kinds keep
	// their wire values.
	DynamicAmountObjectManaValue
	// DynamicAmountChosenNumber reads a prior ResolutionChoiceNumber result
	// published under ResultKey.
	DynamicAmountChosenNumber
	// DynamicAmountObjectCounters is the number of counters of CounterKind on a
	// referenced object, using last-known information after it leaves play.
	DynamicAmountObjectCounters
	// DynamicAmountCapturedTargetManaValue reads the mana value captured for an
	// enclosing effect's stack-object target when a delayed trigger was created.
	DynamicAmountCapturedTargetManaValue
	// DynamicAmountGreatestPowerInGroup is the greatest power among the
	// permanents of Group, evaluated as the effect resolves (CR 608.2c). It
	// backs "equal to the greatest power among <group>" amounts (Garruk, Primal
	// Hunter; Fungal Sprouting); an empty group yields zero. Added last so
	// existing kinds keep their wire values.
	DynamicAmountGreatestPowerInGroup
	// DynamicAmountGreatestToughnessInGroup is the greatest toughness among the
	// permanents of Group, the toughness sibling of
	// DynamicAmountGreatestPowerInGroup. An empty group yields zero.
	DynamicAmountGreatestToughnessInGroup
	// DynamicAmountGreatestManaValueInGroup is the greatest mana value among the
	// permanents of Group, the mana-value sibling of
	// DynamicAmountGreatestPowerInGroup. An empty group yields zero.
	DynamicAmountGreatestManaValueInGroup
	// DynamicAmountDevotion is the controller's devotion to Colors, the number
	// of mana symbols of those colors among the mana costs of permanents the
	// controller controls (CR 700.5). A hybrid or Phyrexian symbol counts toward
	// every color it contains, and a symbol counts once for a multi-color
	// devotion when it matches any listed color. Added last so existing kinds
	// keep their wire values.
	DynamicAmountDevotion
	// DynamicAmountSpellsCastThisTurn is the number of spells the controller has
	// cast this turn, counted from the turn's spell-cast events (CR 608.2c). It
	// backs the storm-counter family such as Aetherflux Reservoir's "you gain 1
	// life for each spell you've cast this turn." Added last so existing kinds
	// keep their wire values.
	DynamicAmountSpellsCastThisTurn
	// DynamicAmountEventLifeChange is the amount of life gained or lost by the
	// event that triggered the resolving ability ("that much life" in "Whenever
	// you gain life, target opponent loses that much life."). It reads the
	// triggering event's life quantity from the resolving ability's
	// TriggerEvent, backing the life-drain mirror family (Sanguine Bond,
	// Exquisite Blood). Added last so existing kinds keep their wire values.
	DynamicAmountEventLifeChange
	// DynamicAmountTotalPowerInGroup is the sum of power among the permanents of
	// Group, evaluated as the effect resolves (CR 608.2c). It backs "the total
	// power of <group>" amounts (Ghalta, Primal Hunger's cost reduction); an
	// empty group yields zero. DynamicAmountTotalToughnessInGroup is the
	// toughness sibling. Added last so existing kinds keep their wire values.
	DynamicAmountTotalPowerInGroup
	DynamicAmountTotalToughnessInGroup
	// DynamicAmountColorCountInGroup is the number of distinct colors among the
	// permanents of Group, evaluated as the effect resolves (CR 608.2c). A
	// permanent contributes each of its colors (CR 105.2, CR 202.2); colorless
	// permanents contribute none. It backs "the number of colors among <group>"
	// amounts such as Faeburrow Elder's "+1/+1 for each color among permanents
	// you control"; an empty or fully colorless group yields zero. Added last so
	// existing kinds keep their wire values.
	DynamicAmountColorCountInGroup
	// DynamicAmountSharedCreatureTypeCountInGroup is the number of permanents of
	// Group, other than the permanent the amount is evaluated for, that share at
	// least one creature type with it (CR 700.4, CR 608.2c). Unlike every other
	// group amount it yields a different value per affected permanent, so it is
	// evaluated against the permanent a continuous power/toughness modification
	// applies to rather than the effect's source. A Changeling has every creature
	// type, so it shares with any other creature that has at least one creature
	// type. It backs the shared-creature-type anthem family (Coat of Arms: "Each
	// creature gets +1/+1 for each other creature on the battlefield that shares a
	// creature type with it"). Added last so existing kinds keep their wire
	// values.
	DynamicAmountSharedCreatureTypeCountInGroup
	// DynamicAmountSourceCardPower is the power of the resolving ability's source
	// card, read from the card instance in whatever zone it occupies (CR 702.94).
	// It backs Scavenge, whose exile-from-graveyard cost moves the source card to
	// exile before the ability resolves, so the count reads the card's printed
	// power directly rather than a battlefield permanent. Added last so existing
	// kinds keep their wire values.
	DynamicAmountSourceCardPower
	// DynamicAmountBlockingCreaturesBeyondFirst is the number of creatures
	// blocking the resolving ability's source permanent beyond the first, read
	// from the current combat's block declarations as the ability resolves
	// (CR 509.1, CR 702.23). It is zero when the source is blocked by one or no
	// creatures, or outside combat. It backs Rampage N, whose Multiplier scales
	// the count by the printed N. Added last so existing kinds keep their wire
	// values.
	DynamicAmountBlockingCreaturesBeyondFirst
	// DynamicAmountLifeLostThisTurn is the total life Player has lost so far this
	// turn, summed from the turn's EventLifeLost amounts (CR 608.2c). Damage to
	// the player counts, because dealing damage to a player causes that player to
	// lose that much life (CR 120.3), which the rules emit as a life-loss event.
	// It backs Children of Korlis's "gain life equal to the life you've lost this
	// turn." DynamicAmountLifeGainedThisTurn is the life-gained sibling, summed
	// from the turn's EventLifeGained amounts. Both read the resolving ability's
	// controller. Added last so existing kinds keep their wire values.
	DynamicAmountLifeLostThisTurn
	DynamicAmountLifeGainedThisTurn
)

// DynamicAmount describes an effect amount determined as the effect resolves
// (CR 608.2c), separate from characteristic-defining P/T values in layers.
type DynamicAmount struct {
	Kind DynamicAmountKind

	Constant   int
	Multiplier int

	CounterKind counter.Kind
	Group       GroupReference
	Object      ObjectReference
	Player      *PlayerReference
	CardZone    zone.Type
	Selection   *Selection
	ResultKey   ResultKey
	// Colors lists the colors counted by a DynamicAmountDevotion amount; empty
	// for every other kind.
	Colors []color.Color
	// ColorFrom, when set on a DynamicAmountDevotion amount, names a published
	// ResolutionChoiceMana result whose chosen color is the single devotion
	// color, overriding Colors. It backs "Add an amount of mana of that color
	// equal to your devotion to that color." (Nykthos, Shrine to Nyx), where the
	// devotion color is the color chosen as the ability resolves rather than a
	// fixed color printed in the amount.
	ColorFrom ChoiceKey
}
