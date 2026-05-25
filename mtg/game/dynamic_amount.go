package game

import "github.com/natefinch/council4/mtg/game/counter"

// DynamicAmountKind identifies a rules-derived integer for effect resolution.
// Variable values such as X and "equal to" quantities are determined as the
// resolving instruction applies unless the card text says otherwise
// (CR 107.3, CR 608.2c).
type DynamicAmountKind int

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
	DynamicAmountCountSelector
	DynamicAmountPreviousEffectResult
)

// DynamicAmount describes an effect amount determined as the effect resolves
// (CR 608.2c), separate from characteristic-defining P/T values in layers.
type DynamicAmount struct {
	Kind DynamicAmountKind

	Constant   int
	Multiplier int

	TargetIndex int
	CounterKind counter.Kind
	Selector    EffectSelector
	LinkID      string
}
