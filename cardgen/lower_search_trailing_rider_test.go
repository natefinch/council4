package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerSearchToHandLoseLifeRider proves the Grim Tutor shape — a
// search-to-hand sequence followed by "You lose 3 life." — lowers to a Search
// instruction followed by a fixed controller LoseLife rider, rather than being
// rejected because the shuffle is no longer the final effect.
func TestLowerSearchToHandLoseLifeRider(t *testing.T) {
	t.Parallel()
	seq := lowerSpellSequence(t, "Grim Tutor Test",
		"Search your library for a card, put that card into your hand, then shuffle. You lose 3 life.")
	if len(seq) != 2 {
		t.Fatalf("sequence = %#v, want [Search, LoseLife]", seq)
	}
	if _, ok := seq[0].Primitive.(game.Search); !ok {
		t.Fatalf("sequence[0] = %#v, want game.Search", seq[0].Primitive)
	}
	loss, ok := seq[1].Primitive.(game.LoseLife)
	if !ok {
		t.Fatalf("sequence[1] = %#v, want game.LoseLife", seq[1].Primitive)
	}
	if got := loss.Amount.Value(); got != 3 {
		t.Errorf("LoseLife amount = %d, want 3", got)
	}
}

// TestLowerSearchToHandGainLifeRider proves the Environmental Sciences shape —
// the same search-to-hand sequence followed by "You gain 2 life." — lowers to a
// Search followed by a fixed controller GainLife rider.
func TestLowerSearchToHandGainLifeRider(t *testing.T) {
	t.Parallel()
	seq := lowerSpellSequence(t, "Environmental Sciences Test",
		"Search your library for a basic land card, reveal it, put it into your hand, then shuffle. You gain 2 life.")
	if len(seq) != 2 {
		t.Fatalf("sequence = %#v, want [Search, GainLife]", seq)
	}
	if _, ok := seq[0].Primitive.(game.Search); !ok {
		t.Fatalf("sequence[0] = %#v, want game.Search", seq[0].Primitive)
	}
	gain, ok := seq[1].Primitive.(game.GainLife)
	if !ok {
		t.Fatalf("sequence[1] = %#v, want game.GainLife", seq[1].Primitive)
	}
	if got := gain.Amount.Value(); got != 2 {
		t.Errorf("GainLife amount = %d, want 2", got)
	}
}
