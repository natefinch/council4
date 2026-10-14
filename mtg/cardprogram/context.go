package cardprogram

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source is immutable invocation metadata, not a pointer to a live object.
type Source struct {
	CardID     id.ID
	ObjectID   id.ID
	StackID    id.ID
	Controller game.PlayerID
	X          int
	// EventPlayer is a captured historical identity, even if that player left.
	EventPlayer opt.V[game.PlayerID]
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

// Copies reports the exact stack objects created by one copy operation.
type Copies struct {
	Available bool
	StackIDs  []id.ID
}

// Exchange records an all-or-nothing permanent-control exchange.
type Exchange struct {
	Available bool
	Exchanged bool
	Before    []game.ObjectSnapshot
	After     []game.ObjectSnapshot
}

// FoundCard is a revealed card incarnation that a program may subsequently move.
type FoundCard struct {
	Identity     game.ZoneCardSnapshot
	Info         game.ChoiceCardInfo
	LibraryEpoch uint64
}

// SearchResult distinguishes an unavailable searcher from an empty search.
type SearchResult struct {
	Available bool
	Cards     []FoundCard
}

// CardPlacement assigns revealed cards to one destination in a simultaneous move.
type CardPlacement struct {
	Cards       []FoundCard
	Destination zone.Type
}

// View resolves invocation metadata and read-only references.
type View interface {
	Source() Source
	Targets() []game.Target
	Player(game.PlayerReference) (game.PlayerID, bool, error)
	Players(game.PlayerGroupReference) ([]game.PlayerID, error)
}

// Context is the small authoring seam implemented by the Engine.
// Execute uses the same rules operations as declarative Instructions. The
// context is valid only during its synchronous invocation; pending work must
// retain versioned references and game-owned data, not this interface.
type Context interface {
	View() View
	Execute(game.Primitive) (Outcome, error)
	Ask(game.ChoiceRequest) ([]int, error)
	Destroy(game.Destroy) (Destruction, error)
	CreateTokens(*game.CardDef, ...TokenRecipient) ([]game.ObjectSnapshot, error)
	Copy(game.CopyStackObject, game.PlayerID) (Copies, error)
	Exchange(game.ObjectReference, game.ObjectReference) (Exchange, error)
	Search(game.Search) (SearchResult, error)
	MoveCards(...CardPlacement) ([]FoundCard, error)
}
