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
