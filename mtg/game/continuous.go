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
	DynamicValueControllerCreatureCardsInGraveyard
	DynamicValueControllerInstantOrSorceryCardsInGraveyard
	DynamicValueControllerLandCardsInGraveyard
	DynamicValueControllerCardTypesInGraveyard
	DynamicValueControllerPermanentCardsInGraveyard
	DynamicValueControllerSubtypeCount
	DynamicValueControllerBasicLandTypeCount
	DynamicValueControllerLifeTotal
	DynamicValueAllPlayersHandSize
	DynamicValueControllerColorPermanentCount
	// DynamicValueControllerCardsDrawnThisTurn is the number of cards the
	// controller has drawn so far this turn ("its power is equal to the number
	// of cards you've drawn this turn", Duelist of the Mind), counted from the
	// turn's EventCardDrawn events for that player (CR 608.2c).
	DynamicValueControllerCardsDrawnThisTurn
	// DynamicValueSourceLinkedExileCount is the number of cards currently in
	// exile that were published under the source permanent's LinkedKey.
	DynamicValueSourceLinkedExileCount
)

// DynamicValue is data for a characteristic-defining numeric value.
type DynamicValue struct {
	Kind  DynamicValueKind
	Value int
	// Offset is a fixed integer added to the resolved count, modeling
	// "<count> plus N" characteristic-defining toughness (CR 208.2, Tarmogoyf's
	// "its toughness is equal to that number plus 1").
	Offset int
	// Subtype selects the subtype counted by DynamicValueControllerSubtypeCount
	// ("the number of Swamps you control", "the number of Goblins you
	// control"). It is unused by every other kind.
	Subtype types.Sub
	// Color selects the color of permanents counted by
	// DynamicValueControllerColorPermanentCount ("the number of red permanents
	// you control"). It is unused by every other kind.
	Color color.Color
	// LinkedKey identifies the source-scoped linked-exile pool counted by
	// DynamicValueSourceLinkedExileCount.
	LinkedKey LinkedKey
	// LinkedObjectScoped keys the linked pool by the source's current object
	// identity rather than its stable card identity.
	LinkedObjectScoped bool
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

	// ExpiresForRef binds ExpiresFor to a player resolved at application time,
	// used by DurationForAsLongAsPlayerIsMonarch to bind the duration to the
	// triggering player who became the monarch ("... for as long as they're the
	// monarch."). The runtime resolves it once when the effect is created and
	// stores the result in ExpiresFor; it is the ExpiresFor analogue of
	// NewControllerRef.
	ExpiresForRef opt.V[PlayerReference]

	AffectedObjectID id.ID
	AffectedSource   bool
	Group            GroupReference

	Layer ContinuousLayer

	CopyValues opt.V[CopyableValues]

	NewController opt.V[PlayerID]

	// NewControllerRef binds a LayerControl effect's new controller to a player
	// resolved at application time (the give-control forms whose new controller
	// is a chosen target player, e.g. "Target player gains control of target
	// permanent you control."). The runtime resolves it once when the effect is
	// created and stores the result in NewController; it is mutually exclusive
	// with the NewController sentinel.
	NewControllerRef opt.V[PlayerReference]

	// NewControllerIsMonarch binds a LayerControl effect's new controller to
	// whoever currently holds the monarch designation, re-evaluated every time
	// control is computed ("The monarch controls enchanted creature.", Fealty to
	// the Realm). Unlike NewController/NewControllerRef, which fix the controller
	// once, this flag makes control follow the crown: while a monarch exists the
	// affected object is controlled by the monarch, and when no player is the
	// monarch the effect makes no change and the object keeps its normal
	// controller (CR 720). It requires the control layer and is mutually
	// exclusive with NewController and NewControllerRef.
	NewControllerIsMonarch bool

	TextFrom string
	TextTo   string

	// SetName replaces the affected object's name at LayerText ("becomes a ...
	// creature named Fenric", CR 613.1c). An empty value leaves the name
	// unchanged. The legend rule and name-matching predicates read the resulting
	// effective name.
	SetName string
	// SetNameFromSourceChoice replaces the affected object's name with the card
	// name stored under this key on the effect's source permanent.
	SetNameFromSourceChoice ChoiceKey

	SetSupertypes    []types.Super
	AddSupertypes    []types.Super
	RemoveSupertypes []types.Super

	SetTypes    []types.Card
	AddTypes    []types.Card
	RemoveTypes []types.Card

	SetSubtypes    []types.Sub
	AddSubtypes    []types.Sub
	RemoveSubtypes []types.Sub
	// AddEveryCreatureType adds every creature subtype to the affected object at
	// LayerType ("[group/this creature] is/are every creature type", Maskwood
	// Nexus, Mistform Ultimus). It mirrors the Changeling keyword's type-layer
	// expansion (CR 702.73) without enumerating the full subtype list in data.
	AddEveryCreatureType bool
	// AddEveryBasicLandType adds the five basic land subtypes (Plains, Island,
	// Swamp, Mountain, Forest) to the affected object at LayerType ("[group/this
	// land] is/are every basic land type", Dryad of the Ilysian Grove, Prismatic
	// Omen). Each added basic subtype also confers its intrinsic mana ability
	// (CR 305.6) through the same path as an explicitly added basic land type.
	AddEveryBasicLandType bool
	// AddSubtypeFromEntryChoice adds the subtype recorded under this key on the
	// effect's source permanent. A missing source, choice, or subtype result has
	// no effect.
	AddSubtypeFromEntryChoice ChoiceKey
	// SetSubtypeFromSourceChoice replaces subtypes belonging to
	// SetSubtypeChoiceType with the subtype stored under this key on the effect's
	// source permanent, preserving subtypes from other card-type families.
	SetSubtypeFromSourceChoice ChoiceKey
	SetSubtypeChoiceType       types.Card

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

	SetPower     opt.V[PT]
	SetToughness opt.V[PT]
	// SetPowerDynamic and SetToughnessDynamic set base power/toughness to a
	// rules-derived amount evaluated as the effect resolves at
	// LayerPowerToughnessSet (CR 613.4b, CR 608.2c). They back the one-shot
	// dynamic base-P/T set such as Mirror Entity's "creatures you control have
	// base power and toughness X/X until end of turn", where X is the cost paid.
	// The amount is locked when the effect resolves (see snapshotContinuousX),
	// folding into the fixed SetPower/SetToughness so the layer reads a frozen
	// value; an amount that survives unfrozen is evaluated per layer pass like a
	// group count. SetPower/SetToughness take precedence when both are present.
	SetPowerDynamic       opt.V[DynamicAmount]
	SetToughnessDynamic   opt.V[DynamicAmount]
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
