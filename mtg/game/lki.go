package game

import (
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

// ObjectSnapshot is last-known information for an object that changed zones.
type ObjectSnapshot struct {
	ObjectID       id.ID
	CardID         id.ID
	TokenName      string
	TokenDef       *CardDef
	Face           FaceIndex
	Name           string
	Owner          PlayerID
	Controller     PlayerID
	FromZone       ZoneType
	Colors         []mana.Color
	Supertypes     []Supertype
	Types          []CardType
	Subtypes       []string
	Power          opt.V[int]
	Toughness      opt.V[int]
	Counters       counter.Set
	MarkedDamage   int
	Attachments    []id.ID
	AttachedTo     opt.V[id.ID]
	ZoneOrderIndex int
}

// LinkedObjectKey identifies objects exiled or otherwise tracked by one linked
// ability pair on one source.
type LinkedObjectKey struct {
	SourceID id.ID
	LinkID   string
}

// LinkedObjectRef records the object/card pair tracked by a linked ability.
type LinkedObjectRef struct {
	ObjectID id.ID
	CardID   id.ID
}
