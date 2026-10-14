package cardprogram

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source is immutable invocation metadata, not a pointer to a live object.
type Source struct {
	CardID     id.ID
	ObjectID   id.ID
	StackID    id.ID
	Controller game.PlayerID
	X          int
}

// Outcome reports an operation's actual success and copied result snapshots.
// Scalar receipts are deliberately not exposed as universal mutation counts:
// legacy handlers give those scalars operation-specific meanings.
type Outcome struct {
	Available bool
	Succeeded bool
	Objects   []game.ObjectSnapshot
}

// Departure is one attempted destruction's correlated, frozen outcome.
// Destination is actual, including replacement redirection, not the requested zone.
// It is meaningful only when Moved is true. Arrivals contains physical card
// incarnations; departed tokens are instead identified by Before.ObjectID.
type Departure struct {
	Before      game.ObjectSnapshot
	Attempted   bool
	Moved       bool
	Destination zone.Type
	Arrivals    []game.ZoneCardSnapshot
}

// Destruction distinguishes a known empty batch from an unavailable referent.
type Destruction struct {
	Available bool
	Members   []Departure
}

// TokenRecipient contributes a quantity to one simultaneous token-creation event.
type TokenRecipient struct {
	Player game.PlayerID
	Count  int
}

// Context is the small authoring seam implemented by the Engine.
// Execute uses the same rules operations as declarative Instructions. The
// context is valid only during its synchronous invocation; pending work must
// retain versioned references and game-owned data, not this interface.
type Context interface {
	Source() Source
	Targets() []game.Target
	Execute(game.Primitive) (Outcome, error)
	Ask(game.ChoiceRequest) ([]int, error)
	Destroy(game.Destroy) (Destruction, error)
	CreateTokens(*game.CardDef, ...TokenRecipient) ([]game.ObjectSnapshot, error)
}
