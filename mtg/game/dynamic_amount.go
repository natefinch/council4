package game

import (
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
}
