package rules

import (
	"testing"

	cardn "github.com/natefinch/council4/mtg/cards/n"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// necropotenceCardDef loads the registered Necropotence card definition and
// asserts the three-line Oracle text lowered to exactly the expected shape: one
// "Skip your draw step." static ability, one "Pay 1 life:" activated ability, and
// one "Whenever you discard a card, ..." triggered ability. Sourcing behavior
// from the real generated definition proves the curated card — not a hand-written
// stand-in — drives the runtime.
func necropotenceCardDef(t *testing.T) *game.CardDef {
	t.Helper()
	def := cardn.Necropotence()
	if got := len(def.StaticAbilities); got != 1 {
		t.Fatalf("Necropotence has %d static abilities, want 1", got)
	}
	if got := len(def.ActivatedAbilities); got != 1 {
		t.Fatalf("Necropotence has %d activated abilities, want 1", got)
	}
	if got := len(def.TriggeredAbilities); got != 1 {
		t.Fatalf("Necropotence has %d triggered abilities, want 1", got)
	}
	return def
}

// necropotenceDiscardTrigger returns the registered card's discard-linked
// triggered ability and asserts it is the non-optional "Whenever you discard a
// card" trigger whose body moves the exact discarded card from the graveyard to
// exile.
func necropotenceDiscardTrigger(t *testing.T) game.TriggeredAbility {
	t.Helper()
	trigger := necropotenceCardDef(t).TriggeredAbilities[0]
	if trigger.Trigger.Pattern.Event != game.EventCardDiscarded {
		t.Fatalf("discard trigger event = %v, want EventCardDiscarded", trigger.Trigger.Pattern.Event)
	}
	if trigger.Trigger.Pattern.Player != game.TriggerPlayerYou {
		t.Fatalf("discard trigger player = %v, want TriggerPlayerYou", trigger.Trigger.Pattern.Player)
	}
	if trigger.Optional {
		t.Fatal("discard trigger must be mandatory, not optional")
	}
	seq := trigger.Content.Modes[0].Sequence
	if len(seq) != 1 {
		t.Fatalf("discard trigger body has %d instructions, want 1", len(seq))
	}
	move, ok := seq[0].Primitive.(game.MoveCard)
	if !ok {
		t.Fatalf("discard trigger body is %T, want game.MoveCard", seq[0].Primitive)
	}
	if move.Card.Kind != game.CardReferenceEvent {
		t.Fatalf("discard trigger moves card reference %v, want CardReferenceEvent (the exact discarded card)", move.Card.Kind)
	}
	if move.FromZone != zone.Graveyard || move.Destination != zone.Exile {
		t.Fatalf("discard trigger moves %v->%v, want Graveyard->Exile", move.FromZone, move.Destination)
	}
	return trigger
}

// addNecropotence puts the real Necropotence enchantment onto the battlefield
// under the given controller so its registered discard trigger is live.
func addNecropotence(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, cardn.Necropotence())
}

func TestNecropotenceDiscardTriggerExilesDiscardedCardFromGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	necropotenceDiscardTrigger(t) // structural assertions on the real definition
	addNecropotence(g, game.Player1)
	cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Discarded Card", Types: []types.Card{types.Creature}}})

	if !discardCardFromHand(g, game.Player1, cardID) {
		t.Fatal("discardCardFromHand() = false, want true")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("discarded card is not in the graveyard before the trigger resolves")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("discard trigger was not put on the stack")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want 1 (one discard trigger)", got)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("discarded card remained in the graveyard; the trigger must exile it")
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("discarded card was not exiled from the graveyard")
	}
}

// TestNecropotenceDiscardTriggerFiresOncePerSimultaneouslyDiscardedCard proves a
// multi-card discard batch (CR 701.9a) yields one independent trigger per
// discarded card, and each resolves to exile its own card — the trigger observes
// the exact discarded card, not a single batch event.
func TestNecropotenceDiscardTriggerFiresOncePerSimultaneouslyDiscardedCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addNecropotence(g, game.Player1)
	const discarded = 3
	cardIDs := make([]id.ID, 0, discarded)
	for range discarded {
		cardIDs = append(cardIDs, addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Batch Card", Types: []types.Card{types.Creature}}}))
	}

	simultaneousID := g.IDGen.Next()
	for _, cardID := range cardIDs {
		if !discardCardFromHandInBatch(g, game.Player1, cardID, simultaneousID) {
			t.Fatalf("discardCardFromHandInBatch(%v) = false, want true", cardID)
		}
	}

	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("discard triggers were not put on the stack")
	}
	if got := g.Stack.Size(); got != discarded {
		t.Fatalf("stack size = %d, want %d (one trigger per discarded card)", got, discarded)
	}

	for !g.Stack.IsEmpty() {
		engine.resolveTopOfStack(g, &TurnLog{})
	}
	for _, cardID := range cardIDs {
		if !g.Players[game.Player1].Exile.Contains(cardID) {
			t.Fatalf("discarded card %v was not exiled", cardID)
		}
	}
	if got := g.Players[game.Player1].Graveyard.Size(); got != 0 {
		t.Fatalf("graveyard size = %d, want 0 (all discarded cards exiled)", got)
	}
	if got := g.Players[game.Player1].Exile.Size(); got != discarded {
		t.Fatalf("exile size = %d, want %d", got, discarded)
	}
}

// TestNecropotenceDiscardTriggerNoOpsWhenCardLeftGraveyard proves the exile is a
// non-targeted move that fails closed: if the discarded card has already left the
// graveyard when the trigger resolves, the trigger does nothing rather than
// pulling the card out of its new zone.
func TestNecropotenceDiscardTriggerNoOpsWhenCardLeftGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addNecropotence(g, game.Player1)
	cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Discarded Card", Types: []types.Card{types.Creature}}})

	if !discardCardFromHand(g, game.Player1, cardID) {
		t.Fatal("discardCardFromHand() = false, want true")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("discard trigger was not put on the stack")
	}
	// Another effect moves the card from the graveyard to hand before the trigger
	// resolves.
	g.Players[game.Player1].Graveyard.Remove(cardID)
	g.Players[game.Player1].Hand.Add(cardID)

	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("card that had left the graveyard was disturbed; the trigger must no-op")
	}
	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("card was exiled from a zone other than the graveyard; the exile must be graveyard-only")
	}
}

// TestNecropotenceDiscardTriggerResolvesAfterSourceLeaves proves the exile
// resolves from the independent trigger on the stack even after Necropotence has
// left the battlefield: a triggered ability exists independently of its source
// once it triggers (CR 603.3d/CR 112.7a).
func TestNecropotenceDiscardTriggerResolvesAfterSourceLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addNecropotence(g, game.Player1)
	cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Discarded Card", Types: []types.Card{types.Creature}}})

	if !discardCardFromHand(g, game.Player1, cardID) {
		t.Fatal("discardCardFromHand() = false, want true")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("discard trigger was not put on the stack")
	}
	// Necropotence leaves the battlefield after the trigger is on the stack.
	removePermanentFromBattlefield(g, source.ObjectID)

	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("discard trigger did not exile the card after its source left the battlefield")
	}
}

// TestNecropotenceDiscardTriggerIgnoresOpponentDiscard proves the "you" in
// "Whenever you discard a card" keys to the controller: an opponent discarding a
// card does not trigger the controller's Necropotence.
func TestNecropotenceDiscardTriggerIgnoresOpponentDiscard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addNecropotence(g, game.Player1)
	cardID := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Opponent Card", Types: []types.Card{types.Creature}}})

	if !discardCardFromHand(g, game.Player2, cardID) {
		t.Fatal("discardCardFromHand() = false, want true")
	}
	if engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("opponent's discard triggered the controller's Necropotence")
	}
	if !g.Players[game.Player2].Graveyard.Contains(cardID) {
		t.Fatal("opponent's discarded card should remain in their graveyard")
	}
}
