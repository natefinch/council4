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
	// EventClash is the captured clash fact of a triggered ability whose
	// trigger event was a clash ("whenever you clash"; CR 701.30).
	EventClash opt.V[ClashEvent]
}

// ClashEvent is one player's completed clash as captured by a clash trigger:
// Player clashed with Opponent, and Won reports whether Player won (CR
// 701.30d). It stays fixed for the trigger's lifetime, including copies.
type ClashEvent struct {
	Player   game.PlayerID
	Opponent game.PlayerID
	Won      bool
}

// Outcome reports an operation's actual success and copied result snapshots.
// Scalar receipts are deliberately not exposed as universal mutation counts:
// legacy handlers give those scalars operation-specific meanings.
type Outcome struct {
	Available bool
	Succeeded bool
	// Objects carry public characteristics, including last-known power,
	// toughness and mana value for permanents that left the battlefield.
	Objects []game.ObjectSnapshot
	// Cards contains exact local handles for cards observed by Reveal.
	Cards []FoundCard
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
	// Cards authorizes further placements of actual public graveyard/exile
	// arrivals in this invocation. Hidden-zone redirects do not mint handles.
	Cards []FoundCard
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

// FoundCard is an invocation-local card incarnation obtained through an Engine
// operation: a revealed search, public target, or actual public arrival.
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

// ClashReveal is one clashing player's public reveal and placement (CR
// 701.30a, 701.30c). Card is unset when that player's library was empty, so
// nothing was revealed; Card.ManaValue is the card's mana value in the library
// (a split card's halves combined, CR 709.4b). Bottom reports that the player
// put the revealed card on the bottom of their library rather than the top.
type ClashReveal struct {
	Player game.PlayerID
	Card   opt.V[game.ChoiceCardInfo]
	Bottom bool
}

// ClashResult reports one completed clash with an opponent (CR 701.30b-d).
// Available is false when the controller or a living opponent was not
// available, in which case nobody clashed. Reveals are in APNAP order. Winner
// is set only when one player revealed a card whose mana value is greater than
// every other card revealed in the clash (CR 701.30d).
type ClashResult struct {
	Available bool
	Opponent  game.PlayerID
	Reveals   []ClashReveal
	Winner    opt.V[game.PlayerID]
}

// Won reports whether player won the clash.
func (r ClashResult) Won(player game.PlayerID) bool {
	return r.Winner.Exists && r.Winner.Val == player
}

// CardPlacement assigns minted card handles to a simultaneous destination.
type CardPlacement struct {
	Cards         []FoundCard
	Destination   zone.Type
	LibraryBottom bool
	// Battlefield options apply to one simultaneous placement of all Cards.
	// Without Recipient, each card enters under its owner's control.
	Recipient         opt.V[game.PlayerReference]
	ContinuousEffects []game.ContinuousEffect
	EntryCounters     []game.CounterPlacement
	EntryTapped       bool
}

// View resolves invocation metadata and read-only references.
type View interface {
	Source() Source
	Targets() []game.Target
	TargetCards() ([]FoundCard, error)
	ActivePlayer() (game.PlayerID, error)
	Player(game.PlayerReference) (game.PlayerID, bool, error)
	Players(game.PlayerGroupReference) ([]game.PlayerID, error)
	Permanents(game.GroupReference) ([]game.ObjectSnapshot, error)
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
	// ControlNextTurn schedules control using captured identities, not agents.
	ControlNextTurn(game.PlayerReference) (bool, error)
	// Clash makes the source's controller clash with an opponent (CR 701.30).
	// An unset opponent is chosen by the controller among living opponents; a
	// set reference names that opponent.
	Clash(opponent opt.V[game.PlayerReference]) (ClashResult, error)
}
