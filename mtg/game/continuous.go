package game

import (
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DynamicValueKind describes a small rules-derived integer used for
// characteristic-defining abilities such as "* is the number of cards in your
// hand." More cases can be added as card implementations need them.
type DynamicValueKind int

// Dynamic value kind values identify supported derived integers.
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
// 1 (CR 707, CR 613). Optional fields mean "leave that value absent.".
type CopyableValues struct {
	Name             string
	Colors           []mana.Color
	Supertypes       []types.Super
	Types            []types.Card
	Subtypes         []types.Sub
	Power            opt.V[PT]
	Toughness        opt.V[PT]
	DynamicPower     opt.V[DynamicValue]
	DynamicToughness opt.V[DynamicValue]
	Abilities        []AbilityDef
	OracleText       string
}

// ContinuousLayer identifies the layer or sublayer where a continuous effect
// applies. The numeric values intentionally follow CR 613 layer order.
type ContinuousLayer int

// Continuous layer values follow CR 613 layer order.
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

// Timestamp records relative effect ordering for the layer system.
type Timestamp uint64

// ContinuousEffect is a rules-data representation of a runtime continuous
// effect. mtg/rules owns interpretation and expiry.
type ContinuousEffect struct {
	ID             id.ID
	SourceObjectID id.ID
	SourceCardID   id.ID
	Controller     PlayerID
	Timestamp      Timestamp
	DependsOn      []id.ID
	Duration       EffectDuration
	CreatedTurn    int
	ExpiresFor     PlayerID

	AffectedObjectID id.ID
	Selector         EffectSelector

	Layer ContinuousLayer

	CopyValues opt.V[CopyableValues]

	NewController opt.V[PlayerID]

	TextFrom string
	TextTo   string

	SetSupertypes    []types.Super
	AddSupertypes    []types.Super
	RemoveSupertypes []types.Super

	SetTypes    []types.Card
	AddTypes    []types.Card
	RemoveTypes []types.Card

	SetSubtypes    []types.Sub
	AddSubtypes    []types.Sub
	RemoveSubtypes []types.Sub

	SetColors    []mana.Color
	AddColors    []mana.Color
	RemoveColors []mana.Color

	AddKeywords    []Keyword
	RemoveKeywords []Keyword
	AddAbilities   []AbilityDef

	SetPower       opt.V[PT]
	SetToughness   opt.V[PT]
	PowerDelta     int
	ToughnessDelta int
}
