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
	ObjectID  id.ID
	CardID    id.ID
	TokenName string
	TokenDef  *CardDef
	// CopiableDef preserves the object's exact copiable values as it last existed,
	// including copy effects and merged-permanent abilities. It is immutable rules
	// data used when a later effect creates a token copy from last-known information.
	CopiableDef *CardDef
	// ZoneCards identifies the exact physical card objects this permanent became
	// when it left the battlefield. A merged permanent contributes every
	// nontoken component, each with the post-move zone version that distinguishes
	// that incarnation from a card that later leaves and reenters the zone.
	ZoneCards       []ZoneCardSnapshot
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
	BasePower       opt.V[int]
	Toughness       opt.V[int]
	Keywords        []Keyword
	Counters        counter.Set
	EntryChoices    map[ChoiceKey]ResolutionChoiceResult
	RuleEffectKinds []RuleEffectKind
	MarkedDamage    int
	Attachments     []id.ID
	// SaddleContributorIDs preserves the exact creature objects that saddled this
	// permanent this turn so a triggered ability can resolve after its source
	// leaves the battlefield.
	SaddleContributorIDs []id.ID
	AttachedTo           opt.V[id.ID]
	AttachedToPlayer     opt.V[PlayerID]
	ZoneOrderIndex       int
}

// ZoneCardSnapshot identifies one exact card incarnation produced when a
// permanent leaves the battlefield.
type ZoneCardSnapshot struct {
	CardID      id.ID
	ZoneVersion uint64
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
	ObjectID        id.ID
	CardID          id.ID
	CardZoneVersion uint64
	// CorrelatedPlayer records the player whose per-player choice produced this
	// linked object. It lets later per-player instructions recover the specific
	// object chosen by that player without relying on record order.
	CorrelatedPlayer opt.V[PlayerID]
}
