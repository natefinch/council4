package game

import (
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

// DynamicValueKind describes a small rules-derived integer used for
// characteristic-defining abilities such as "* is the number of cards in your
// hand." More cases can be added as card implementations need them.
type DynamicValueKind int

const (
	DynamicValueNone DynamicValueKind = iota
	DynamicValueConstant
	DynamicValueControllerHandSize
	DynamicValueControllerGraveyardSize
	DynamicValueControllerCreatureCount
	DynamicValueControllerLandCount
	DynamicValueControllerArtifactCount
	DynamicValueAllBattlefieldCreatureCount
)

// DynamicValue is data for a characteristic-defining numeric value.
type DynamicValue struct {
	Kind  DynamicValueKind
	Value int
}

// CopyableValues records the copiable printed/effective values copied in layer
// 1 (CR 707, CR 613). Nil pointer fields mean "leave that value absent."
type CopyableValues struct {
	Name             string
	Colors           []mana.Color
	Supertypes       []Supertype
	Types            []CardType
	Subtypes         []string
	Power            *PT
	Toughness        *PT
	DynamicPower     *DynamicValue
	DynamicToughness *DynamicValue
	Abilities        []AbilityDef
	OracleText       string
}

// ContinuousLayer identifies the layer or sublayer where a continuous effect
// applies. The numeric values intentionally follow CR 613 layer order.
type ContinuousLayer int

const (
	LayerCopy ContinuousLayer = iota + 1
	LayerControl
	LayerText
	LayerType
	LayerColor
	LayerAbility
	LayerPowerToughnessCDA
	LayerPowerToughnessSet
	LayerPowerToughnessModify
	LayerPowerToughnessSwitch
)

// ContinuousEffect is a rules-data representation of a runtime continuous
// effect. mtg/rules owns interpretation and expiry.
type ContinuousEffect struct {
	ID             id.ID
	SourceObjectID id.ID
	SourceCardID   id.ID
	Controller     PlayerID
	Timestamp      int64
	DependsOn      []id.ID
	Duration       EffectDuration
	CreatedTurn    int
	ExpiresFor     PlayerID

	AffectedObjectID id.ID
	Selector         EffectSelector

	Layer ContinuousLayer

	CopyValues *CopyableValues

	NewController *PlayerID

	TextFrom string
	TextTo   string

	SetSupertypes    []Supertype
	AddSupertypes    []Supertype
	RemoveSupertypes []Supertype

	SetTypes    []CardType
	AddTypes    []CardType
	RemoveTypes []CardType

	SetSubtypes    []string
	AddSubtypes    []string
	RemoveSubtypes []string

	SetColors    []mana.Color
	AddColors    []mana.Color
	RemoveColors []mana.Color

	AddKeywords    []Keyword
	RemoveKeywords []Keyword
	AddAbilities   []AbilityDef

	SetPower       *PT
	SetToughness   *PT
	PowerDelta     int
	ToughnessDelta int
}
