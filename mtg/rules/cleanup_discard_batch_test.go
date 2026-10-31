package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

// Source review: Rielle, the Everwise IKO 203 printing
// 5f03c944-1929-4cb2-a373-d57eefa29ed1 (oracle 6f981e70-5828-49ac-8a72-cf1cc09be991);
// https://api.scryfall.com/cards/5f03c944-1929-4cb2-a373-d57eefa29ed1/rulings
// returned three rulings dated 2020-04-17, including: "If the first time you
// discard cards during your turn is due to having too many cards in hand during
// your cleanup step, you'll draw that many cards." CR 514.1 (the cleanup discard
// to maximum hand size is one turn-based action), CR 701.9a (discard) and
// CR 603.2c (one trigger per event) from the 2025-11-14 official copy.
func rielleCleanupScenario(t *testing.T, handSize int) (*game.Game, *Engine, []id.ID) {
	t.Helper()
	def := cards.NewDefaultRegistry().Lookup("Rielle, the Everwise")
	if def == nil {
		t.Fatal("production Rielle, the Everwise is missing")
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, def)
	for i := range 5 {
		addCardToLibrary(g, game.Player1, discardableCard(fmt.Sprintf("Library %d", i)))
	}
	hand := make([]id.ID, handSize)
	for i := range hand {
		hand[i] = addCardToHand(g, game.Player1, discardableCard(fmt.Sprintf("Hand %d", i)))
	}
	g.Turn.ActivePlayer = game.Player1
	return g, NewEngine(nil), hand
}

func cleanupDiscardEvents(g *game.Game) []game.Event {
	var discards []game.Event
	for _, event := range g.Events {
		if event.Kind == game.EventCardDiscarded {
			discards = append(discards, event)
		}
	}
	return discards
}

// TestCleanupDiscardToHandSizeIsOneSimultaneousBatch proves the cleanup
// turn-based discard is a single discard event (CR 514.1): every discarded card
// shares one nonzero SimultaneousID and the first-discard ordinal of the turn.
func TestCleanupDiscardToHandSizeIsOneSimultaneousBatch(t *testing.T) {
	g, engine, _ := rielleCleanupScenario(t, maximumHandSize+3)

	engine.runCleanupStep(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	discards := cleanupDiscardEvents(g)
	if len(discards) != 3 {
		t.Fatalf("cleanup discarded %d cards, want 3", len(discards))
	}
	batch := discards[0].SimultaneousID
	if batch == 0 {
		t.Fatal("cleanup discard has SimultaneousID 0, want one shared batch")
	}
	for _, event := range discards {
		if event.SimultaneousID != batch {
			t.Fatalf("cleanup discard SimultaneousID = %v, want shared %v", event.SimultaneousID, batch)
		}
		if event.PlayerEventOrdinalThisTurn != 1 {
			t.Fatalf("cleanup discard ordinal = %d, want 1", event.PlayerEventOrdinalThisTurn)
		}
	}
	if got := triggerEventCardCount(g, discards[0]); got != 3 {
		t.Fatalf("cleanup discard batch card count = %d, want 3", got)
	}
}

// TestRielleTheEverwiseCleanupDiscardDrawsThatMany proves the production Rielle
// triggers once for a multi-card cleanup discard and draws that many cards, per
// its cleanup-step ruling.
func TestRielleTheEverwiseCleanupDiscardDrawsThatMany(t *testing.T) {
	g, engine, _ := rielleCleanupScenario(t, maximumHandSize+3)
	agents := [game.NumPlayers]PlayerAgent{}

	engine.runCleanupStep(g, agents, &TurnLog{})
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, nil) {
		t.Fatal("Rielle did not trigger for the cleanup discard")
	}
	if depth := g.Stack.Size(); depth != 1 {
		t.Fatalf("cleanup discard put %d Rielle triggers on the stack, want 1", depth)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	player := g.Players[game.Player1]
	if got, want := player.Hand.Size(), maximumHandSize+3; got != want {
		t.Fatalf("hand size after Rielle = %d, want %d (drew 3)", got, want)
	}
	if got := player.Library.Size(); got != 2 {
		t.Fatalf("library size after Rielle = %d, want 2", got)
	}

	// The second cleanup step's discard is the second occurrence this turn
	// (CR 514.3a), so Rielle does not trigger again.
	engine.runCleanupStep(g, agents, &TurnLog{})
	if engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, nil) {
		t.Fatal("Rielle triggered again for a later cleanup discard the same turn")
	}
	if got := player.Hand.Size(); got != maximumHandSize {
		t.Fatalf("hand size after second cleanup = %d, want %d", got, maximumHandSize)
	}
}

// TestGroupDiscardIsOneSimultaneousBatchPerPlayer proves a player-group
// "discards N cards" effect batches each affected player's discards: Player1's
// two chosen discards share one nonzero batch, so Rielle triggers once and draws
// two.
func TestGroupDiscardIsOneSimultaneousBatchPerPlayer(t *testing.T) {
	g, engine, _ := rielleCleanupScenario(t, 4)
	addEffectSpellToStack(g, game.Player1, game.Discard{
		Amount:      game.Fixed(2),
		PlayerGroup: game.AllPlayersReference(),
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	discards := cleanupDiscardEvents(g)
	if len(discards) != 2 || discards[0].SimultaneousID == 0 || discards[0].SimultaneousID != discards[1].SimultaneousID {
		t.Fatalf("group discard produced %d discards, want two sharing one nonzero batch", len(discards))
	}
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
		t.Fatalf("Rielle trigger count = %d, want 1", g.Stack.Size())
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 4 {
		t.Fatalf("hand size after Rielle = %d, want 4 (drew 2)", got)
	}
}
