package cardprogram

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
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
	// AlternativeCost is the declared alternative selected for the resolving
	// spell, including copies. Zero means no declared alternative was selected.
	AlternativeCost cost.AlternativeIndex
	// EventPlayer is a captured historical identity, even if that player left.
	EventPlayer opt.V[game.PlayerID]
	// EventClash is the captured clash fact of a triggered ability whose
	// trigger event was a clash ("whenever you clash"; CR 701.30).
	EventClash opt.V[ClashEvent]
	// SneakPaid reports whether the resolving spell's sneak cost was paid
	// (CR 702.190a); a copy of a sneaked spell keeps it (CR 707.10). For an
	// activated or triggered ability it reports whether the ability's source
	// permanent, or its last known information once it left, is the permanent a
	// sneaked spell became (CR 702.190b): "if this creature's sneak cost was
	// paid". SneakPaidThisTurn additionally requires that it entered this turn:
	// "if its sneak cost was paid this turn".
	SneakPaid         bool
	SneakPaidThisTurn bool
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
// Power and Toughness are the card's values in the zone where the handle was
// minted, including counters on it. They are unset for a card without them or
// whose value comes from a characteristic-defining ability the Engine cannot
// yet evaluate outside the battlefield (#3836). Zone is the zone the identified
// incarnation was in when the operation reported it: a found card's origin, or
// a moved card's actual destination after replacement effects. Owner is the
// card's owner.
type FoundCard struct {
	Identity     game.ZoneCardSnapshot
	Info         game.ChoiceCardInfo
	LibraryEpoch uint64
	Power        opt.V[int]
	Toughness    opt.V[int]
	Zone         zone.Type
	Owner        game.PlayerID
}

// UntilResult reports one "<act> cards until a matching card" process.
// Available is false when the player was unavailable, so nothing happened.
// Hit is the matching card, unset when the library ran out first; Others holds
// every other card the process acted on, in order. Each operation documents
// whether its handles authorize later placements.
type UntilResult struct {
	Available bool
	Hit       opt.V[FoundCard]
	Others    []FoundCard
}

// SearchResult distinguishes an unavailable searcher from an empty search.
type SearchResult struct {
	Available bool
	Cards     []FoundCard
}

// SameNameQuantity is how many cards a same-name search finds.
type SameNameQuantity int

const (
	// SameNameAll finds "all cards with the same name": every match in the
	// public graveyard is found, while matches in the hidden hand and library
	// may be left behind (CR 701.23b; Eradicate and Quash rulings).
	SameNameAll SameNameQuantity = iota
	// SameNameAnyNumber finds "any number of cards", none required, the
	// graveyard included (Surgical Extraction ruling).
	SameNameAnyNumber
	// SameNameUpTo finds "up to Max cards" across all three zones.
	SameNameUpTo
	// SameNameCards finds "Max cards" ("a card named ..."): it must find as
	// many as the graveyard's matches allow, up to Max, and may find the rest
	// in the hand or library or not at all (CR 701.23b).
	SameNameCards
)

// SameName is the reference a same-name search compares against. Set exactly
// one field. Two objects have the same name when they share at least one name;
// an object with no name, such as a face-down permanent, matches nothing (CR
// 201.2a), but the zones are still searched.
type SameName struct {
	// Snapshot is a permanent's last known information, such as the object
	// Execute reports exiling (CR 608.2h): the name it had as it left.
	Snapshot opt.V[game.ObjectSnapshot]
	// Spell is a target stack object reference: the countered spell's
	// captured cast-face name, or its current name if it is still on the stack.
	Spell opt.V[game.ObjectReference]
	// Card is a card in a zone other than the stack: every name it has there,
	// both halves of a split card (CR 709.4) but only a double-faced card's
	// front face (CR 712.6a).
	Card opt.V[FoundCard]
	// Name is a chosen card name.
	Name string
}

// SameNameSearch searches one player's graveyard, hand and library for cards
// with the same name as a reference. The source's controller searches. Max is
// the count for SameNameUpTo and SameNameCards and must be zero otherwise.
type SameNameSearch struct {
	Player   game.PlayerReference
	Name     SameName
	Quantity SameNameQuantity
	Max      int
}

// SameNameResult reports a same-name search. Available is false when the
// searched player was unavailable, so nothing was searched. Cards are the found
// cards, unrevealed (CR 701.23e), with Zone set to their graveyard, hand or
// library origin; library handles stop authorizing placement once that library
// is shuffled, so move them first. Cards that an opponent's search control
// exiled during the search (Opposition Agent) are not returned.
type SameNameResult struct {
	Available bool
	Player    game.PlayerID
	Cards     []FoundCard
}

// HandReveal has one player reveal their whole hand (CR 701.20a), after which
// the source's controller chooses up to Max revealed cards matching Filter (a
// zero Filter matches every card). Unless Optional, the controller must choose
// Max matching cards, or every matching card when fewer match (CR 608.2c-d);
// Optional ("you may choose") allows choosing none. Max zero only reveals.
// All chooses every matching card without asking ("exile all ..."); it
// excludes Max and Optional.
//
// Graveyard, when set, also offers the revealed player's graveyard cards
// matching it in the same choice: "a nonland card from it or a card from their
// graveyard" chooses one card from either zone, and with All "all <filter>
// cards from that player's hand and graveyard" takes every match in both.
//
// Chooser, when set, names the player who makes the choice instead of the
// source's controller ("That player exiles a card from it"); the filter is
// still read relative to the source's controller. An unavailable chooser
// chooses nothing. KeepRevealed keeps the hand revealed until the program
// finishes or a later reveal finds the hand empty; every later RevealHand of
// that hand in the same program continues the one reveal, showing the hand as
// it now is and announcing only cards not already revealed, so a program can
// repeat choices from one revealed hand ("Repeat this process until all cards
// in that hand have been exiled") and then return cards to the emptied hand
// without revealing them.
type HandReveal struct {
	Player       game.PlayerReference
	Filter       game.Selection
	Graveyard    opt.V[game.Selection]
	Max          int
	Optional     bool
	All          bool
	Chooser      opt.V[game.PlayerReference]
	KeepRevealed bool
}

// HandRevealResult reports one hand reveal. Available is false when the
// player was unavailable, so nothing was revealed. Revealed holds every card
// that was in the revealed hand, Matching the revealed cards that matched the
// filter, GraveyardMatching the graveyard cards offered by HandReveal.Graveyard,
// and Chosen the chooser's chosen cards, hand cards first, each as a handle
// whose Zone is its hand or graveyard origin, for one later MoveCards placement
// of them all together; the cards stay where they are until moved (CR
// 701.20b), and a handle stops authorizing placement once its card changes
// zones.
type HandRevealResult struct {
	Available         bool
	Player            game.PlayerID
	Revealed          []FoundCard
	Matching          []FoundCard
	GraveyardMatching []FoundCard
	Chosen            []FoundCard
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
	// PublishLinkedProducer remembers every card of an exile placement that
	// actually reaches exile in the declared producer channel of the resolving
	// ability's original source permanent (CR 607.2a "the exiled card", "cards
	// exiled with this"), even if that permanent has already left the
	// battlefield. The resolving ability body must declare the producer. A
	// later linked ability reads the cards with LinkedExiledCards.
	PublishLinkedProducer game.LinkedProducerKey
	// UntilSourceLeaves exiles hand cards "until <this permanent> leaves the
	// battlefield" (CR 610.3): immediately after the resolving ability's source
	// permanent leaves, each such card still in exile returns to its owner's
	// hand, its previous zone. If the source already left before this
	// placement, those cards don't move (CR 610.3a-b). It requires an exile
	// placement of hand handles made by an activated or triggered ability.
	UntilSourceLeaves bool
	// CastWhileExiled lets the resolving spell's or ability's controller cast
	// each card of an exile placement that actually reaches exile face up, for
	// as long as that card remains exiled ("You may cast that card for as long
	// as it remains exiled.", CR 611.2b), even after the source leaves the
	// battlefield. Ordinary timing and costs apply (CR 601.3). The permission
	// ends when the card leaves exile, including by being cast, so it allows
	// one cast. It requires an exile placement.
	CastWhileExiled bool
}

// CardDelayedTrigger schedules a fixed-timing delayed triggered ability (CR
// 603.7) bound to one card handle minted by this invocation, normally a
// MoveCards result ("At the beginning of the next end step, ... return it to
// its owner's hand."). The trigger captures that exact card incarnation: its
// content acts on it with game.CardReferenceCaptured, which finds nothing once
// the card has changed zones (CR 400.7, 603.7c). Trigger must not set an event
// pattern or its own capture references; it may set
// InterveningIfCapturedCardNotCast ("if you haven't cast the card", CR 603.4).
type CardDelayedTrigger struct {
	Card    FoundCard
	Trigger game.DelayedTriggerDef
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
	// RevealUntil reveals cards from the top of one player's library until a
	// card matches Until (CR 701.20), then puts every revealed card, the hit
	// included, on the bottom of that library in a random order. It requires a
	// single Player and Destination zone.Library. The reordered cards became
	// new objects (CR 701.20d), so the handles carry what was revealed but
	// authorize no placement, and earlier handles to those cards are revoked.
	RevealUntil(game.RevealUntil) (UntilResult, error)
	// SearchSameName searches one player's graveyard, hand and library for
	// cards with the same name as a reference (CR 701.23). Move the returned
	// handles with MoveCards as one simultaneous placement, then shuffle with
	// Execute(game.ShuffleLibrary).
	SearchSameName(SameNameSearch) (SameNameResult, error)
	// RevealHand has one player reveal their hand to every player, as a
	// public announcement of each card's characteristics (not its identity,
	// CR 406.3), and the source's controller choose among the revealed cards
	// (CR 701.20a), optionally together with that player's graveyard cards.
	// Move the chosen handles with MoveCards ("exile that card").
	RevealHand(HandReveal) (HandRevealResult, error)
	// LinkedExiledCards returns the cards the resolving ability's original
	// source permanent exiled through producer (CardPlacement or primitive
	// PublishLinkedProducer) that are still in exile as the same objects (CR
	// 607.2a), with their current characteristics there. The handles authorize
	// one later MoveCards placement ("return the exiled card to its owner's
	// hand"). A card that left exile is a new object and is not returned (CR
	// 400.7). A spell has no source permanent and finds nothing.
	LinkedExiledCards(producer game.LinkedProducerKey) ([]FoundCard, error)
	// DelayCardTrigger schedules a delayed triggered ability bound to one card
	// handle. It reports false when nothing was scheduled.
	DelayCardTrigger(CardDelayedTrigger) (bool, error)
}
