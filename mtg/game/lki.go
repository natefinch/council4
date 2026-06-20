package game

import (
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ObjectSnapshot is last-known information for an object that changed zones.
type ObjectSnapshot struct {
	ObjectID        id.ID
	CardID          id.ID
	TokenName       string
	TokenDef        *CardDef
	Face            FaceIndex
	FaceDown        bool
	FaceDownFace    FaceIndex
	FaceDownKind    FaceDownKind
	MergedCards     []MergedCard
	Name            string
	Owner           PlayerID
	Controller      PlayerID
	FromZone        zone.Type
	Tapped          bool
	Attacking       bool
	Blocking        bool
	Colors          []color.Color
	Supertypes      []types.Super
	Types           []types.Card
	Subtypes        []types.Sub
	Power           opt.V[int]
	Toughness       opt.V[int]
	Keywords        []Keyword
	Counters        counter.Set
	EntryChoices    map[ChoiceKey]ResolutionChoiceResult
	RuleEffectKinds []RuleEffectKind
	MarkedDamage    int
	Attachments     []id.ID
	AttachedTo      opt.V[id.ID]
	ZoneOrderIndex  int
}

// LinkedObjectKey identifies objects exiled or otherwise tracked by one linked
// ability pair on one source.
type LinkedObjectKey struct {
	SourceID id.ID
	LinkID   string
}

// LinkedObjectRef records an object/card pair or a card-only reference tracked
// by a linked ability.
type LinkedObjectRef struct {
	ObjectID id.ID
	CardID   id.ID
}
