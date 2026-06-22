package game

import (
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
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
	DynamicValueAllGraveyardsSize
	DynamicValueCreatureCardsInAllGraveyards
	DynamicValueCardTypesAmongAllGraveyards
)

// DynamicValue is data for a characteristic-defining numeric value.
type DynamicValue struct {
	Kind  DynamicValueKind
	Value int
	// Offset is a fixed integer added to the resolved count, modeling
	// "<count> plus N" characteristic-defining toughness (CR 208.2, Tarmogoyf's
	// "its toughness is equal to that number plus 1").
	Offset int
}

// CopyableValues records the copiable printed/effective values copied in layer
// 1 (CR 707, CR 613). Optional fields mean "leave that value absent.".
type CopyableValues struct {
	Name             string
	Colors           []color.Color
	Supertypes       []types.Super
	Types            []types.Card
	Subtypes         []types.Sub
	Power            opt.V[PT]
	Toughness        opt.V[PT]
	DynamicPower     opt.V[DynamicValue]
	DynamicToughness opt.V[DynamicValue]
	Abilities        []Ability
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
	AffectedSource   bool
	Group            GroupReference

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
	// AddSubtypeFromEntryChoice adds the subtype recorded under this key on the
	// effect's source permanent. A missing source, choice, or subtype result has
	// no effect.
	AddSubtypeFromEntryChoice ChoiceKey

	SetColors    []color.Color
	AddColors    []color.Color
	RemoveColors []color.Color
	// SetColorless makes the affected object colorless (its color set becomes
	// empty) at LayerColor. It is the explicit "becomes colorless" set, distinct
	// from a no-op empty SetColors.
	SetColorless bool

	AddKeywords    []Keyword
	RemoveKeywords []Keyword
	AddAbilities   []Ability

	// RemoveAllAbilities removes every ability and keyword the affected object
	// has from effects with an earlier timestamp ("loses all abilities"). It
	// applies at LayerAbility before this effect's own ability additions.
	RemoveAllAbilities bool

	SetPower              opt.V[PT]
	SetToughness          opt.V[PT]
	PowerDelta            int
	ToughnessDelta        int
	PowerDeltaDynamic     opt.V[DynamicAmount]
	ToughnessDeltaDynamic opt.V[DynamicAmount]
	// DoublePower and DoubleToughness add each affected permanent's own current
	// power/toughness (the value running through earlier layers and earlier 7c
	// effects) back into itself at LayerPowerToughnessModify, doubling that
	// characteristic (CR 107.16, Unnatural Growth). They are independent of the
	// fixed and dynamic deltas above and apply after them.
	DoublePower     bool
	DoubleToughness bool
}
