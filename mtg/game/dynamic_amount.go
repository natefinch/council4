package game

import "github.com/natefinch/council4/mtg/game/counter"

// DynamicAmountKind identifies a rules-derived integer for effect resolution.
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

// DynamicAmount describes an effect amount determined as the effect resolves.
type DynamicAmount struct {
	Kind DynamicAmountKind

	Constant   int
	Multiplier int

	TargetIndex int
	CounterKind counter.Kind
	Selector    EffectSelector
	LinkID      string
}
