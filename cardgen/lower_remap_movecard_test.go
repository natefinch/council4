package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestRebaseMoveCardPlayerZoneGroupTargetReference proves that the player-zone
// group form of MoveCard ("Exile target player's graveyard.") has its
// target-bearing Player reference rebased by the accumulated target offset when
// it appears in a multi-clause sequence. Before the fix, rebaseTargetedPrimitive
// ignored Player and silently retained the clause-local target index 0.
func TestRebaseMoveCardPlayerZoneGroupTargetReference(t *testing.T) {
	t.Parallel()
	primitive, ok := rebaseTargetedPrimitive(game.MoveCard{
		Player:      game.TargetPlayerReference(0),
		FromZone:    zone.Graveyard,
		Destination: zone.Exile,
	}, 3, 0)
	if !ok {
		t.Fatal("MoveCard player-zone target was not rebased")
	}
	move, ok := primitive.(game.MoveCard)
	if !ok {
		t.Fatalf("rebased primitive = %+v", primitive)
	}
	if move.Player != game.TargetPlayerReference(3) {
		t.Fatalf("rebased Player = %+v, want TargetPlayerReference(3)", move.Player)
	}
	if move.Card.Kind != game.CardReferenceNone {
		t.Fatalf("rebased Card = %+v, want unset", move.Card)
	}
	if move.FromZone != zone.Graveyard || move.Destination != zone.Exile {
		t.Fatalf("rebased zones = %+v -> %+v", move.FromZone, move.Destination)
	}
}

// TestRebaseMoveCardSingleCardTargetReference proves the single-card form still
// rebases its Card slot by cardOffset (not the player offset) and leaves Player
// unset, preserving the pre-existing behavior.
func TestRebaseMoveCardSingleCardTargetReference(t *testing.T) {
	t.Parallel()
	primitive, ok := rebaseTargetedPrimitive(game.MoveCard{
		Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 0},
		FromZone:    zone.Graveyard,
		Destination: zone.Exile,
	}, 5, 2)
	if !ok {
		t.Fatal("MoveCard single-card target was not rebased")
	}
	move, ok := primitive.(game.MoveCard)
	if !ok {
		t.Fatalf("rebased primitive = %+v", primitive)
	}
	if move.Card.Kind != game.CardReferenceTarget || move.Card.TargetIndex != 2 {
		t.Fatalf("rebased Card = %+v, want TargetIndex 2", move.Card)
	}
	if move.Player.Kind() != game.PlayerReferenceNone {
		t.Fatalf("rebased Player = %+v, want unset", move.Player)
	}
}

// TestRebaseMoveCardPlayerZoneGroupFailsClosed proves an unrebasable Player
// reference closes the sequence (returns false) rather than emitting a MoveCard
// whose Player still points at a clause-local index.
func TestRebaseMoveCardPlayerZoneGroupFailsClosed(t *testing.T) {
	t.Parallel()
	// An object-controller Player reference embeds an object reference; an empty
	// (invalid) object cannot be rebased, so the primitive must fail closed
	// rather than emit a MoveCard with a clause-local index.
	_, ok := rebaseTargetedPrimitive(game.MoveCard{
		Player:      game.ObjectControllerReference(game.ObjectReference{}),
		FromZone:    zone.Graveyard,
		Destination: zone.Exile,
	}, 3, 0)
	if ok {
		t.Fatal("MoveCard with unrebasable Player reference did not fail closed")
	}
}
