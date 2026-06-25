package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

// firstDiscardTriggerPattern matches "Whenever you discard one or more cards for
// the first time each turn" (Rielle, the Everwise): a batch-aware discard
// trigger gated on the first discard occurrence of the turn.
func firstDiscardTriggerPattern() *game.TriggerPattern {
	return &game.TriggerPattern{
		Event:                      game.EventCardDiscarded,
		Player:                     game.TriggerPlayerYou,
		PlayerEventOrdinalThisTurn: 1,
		OneOrMore:                  true,
	}
}

func discardableCard(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: name, Types: []types.Card{types.Sorcery}}}
}

// TestDiscardBatchOrdinalThreadsOccurrences confirms cards discarded together in
// one simultaneous batch share occurrence ordinal 1, a later single discard the
// same turn is occurrence 2, and the count resets on a new turn.
func TestDiscardBatchOrdinalThreadsOccurrences(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	batchID := g.IDGen.Next()
	first := addCardToHand(g, game.Player1, discardableCard("Batch One"))
	second := addCardToHand(g, game.Player1, discardableCard("Batch Two"))
	if !discardCardFromHandInBatch(g, game.Player1, first, batchID) ||
		!discardCardFromHandInBatch(g, game.Player1, second, batchID) {
		t.Fatal("batch discard failed")
	}
	assertDiscardOrdinal(t, g, first, 1)
	assertDiscardOrdinal(t, g, second, 1)

	later := addCardToHand(g, game.Player1, discardableCard("Single"))
	if !discardCardFromHand(g, game.Player1, later) {
		t.Fatal("single discard failed")
	}
	assertDiscardOrdinal(t, g, later, 2)

	g.Turn.TurnNumber++
	markCurrentTurnEventStart(g)
	nextTurn := addCardToHand(g, game.Player1, discardableCard("Next Turn"))
	if !discardCardFromHand(g, game.Player1, nextTurn) {
		t.Fatal("next-turn discard failed")
	}
	assertDiscardOrdinal(t, g, nextTurn, 1)
}

// TestFirstDiscardEachTurnTriggerGatesOnFirstOccurrence confirms the
// first-discard-each-turn trigger fires once for the first discard of the turn
// and not for a later discard the same turn.
func TestFirstDiscardEachTurnTriggerGatesOnFirstOccurrence(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, firstDiscardTriggerPattern(), discardTriggerInstructions(), nil)

	batchID := g.IDGen.Next()
	first := addCardToHand(g, game.Player1, discardableCard("First A"))
	second := addCardToHand(g, game.Player1, discardableCard("First B"))
	if !discardCardFromHandInBatch(g, game.Player1, first, batchID) ||
		!discardCardFromHandInBatch(g, game.Player1, second, batchID) {
		t.Fatal("batch discard failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("first-discard trigger did not fire for the first discard of the turn")
	}
	if depth := g.Stack.Size(); depth != 1 {
		t.Fatalf("first discard put %d abilities on the stack, want 1 (batch fires once)", depth)
	}

	later := addCardToHand(g, game.Player1, discardableCard("Second Occurrence"))
	if !discardCardFromHand(g, game.Player1, later) {
		t.Fatal("single discard failed")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("first-discard trigger fired for a later discard the same turn")
	}
}

func assertDiscardOrdinal(t *testing.T, g *game.Game, cardID id.ID, want int) {
	t.Helper()
	for _, event := range g.Events {
		if event.Kind == game.EventCardDiscarded && event.CardID == cardID {
			if event.PlayerEventOrdinalThisTurn != want {
				t.Fatalf("discard ordinal for %v = %d, want %d", cardID, event.PlayerEventOrdinalThisTurn, want)
			}
			return
		}
	}
	t.Fatalf("no discard event found for card %v", cardID)
}
