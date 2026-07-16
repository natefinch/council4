package rules

import (
	"testing"

	cardsb "github.com/natefinch/council4/mtg/cards/b"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestBoltBendFerociousReductionEvaluatesBoardStateAtCastTime proves Bolt Bend's
// "This spell costs {3} less to cast if you control a creature with power 4 or
// greater." runs through the existing self cost-modifier machinery: the flat {3}
// reduction is granted at cost determination exactly when the caster controls a
// power-4-or-greater creature, and is withheld otherwise. It is controller
// scoped ("you control"), reads the creature's power against the 4 threshold,
// and never scales beyond the printed {3} generic.
func TestBoltBendFerociousReductionEvaluatesBoardStateAtCastTime(t *testing.T) {
	t.Parallel()

	reductionWith := func(setup func(g *game.Game)) int {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		if setup != nil {
			setup(g)
		}
		return spellGenericReductionFromZone(g, cardsb.BoltBend(), zone.Hand)
	}

	t.Run("power exactly four grants the reduction", func(t *testing.T) {
		t.Parallel()
		got := reductionWith(func(g *game.Game) { addCreatureWithPower(g, game.Player1, 4) })
		if got != 3 {
			t.Fatalf("reduction with a power-4 creature = %d, want 3", got)
		}
	})

	t.Run("power above the threshold does not scale the reduction", func(t *testing.T) {
		t.Parallel()
		got := reductionWith(func(g *game.Game) { addCreatureWithPower(g, game.Player1, 6) })
		if got != 3 {
			t.Fatalf("reduction with a power-6 creature = %d, want a flat 3", got)
		}
	})

	t.Run("power below the threshold grants nothing", func(t *testing.T) {
		t.Parallel()
		got := reductionWith(func(g *game.Game) { addCreatureWithPower(g, game.Player1, 3) })
		if got != 0 {
			t.Fatalf("reduction with a power-3 creature = %d, want 0", got)
		}
	})

	t.Run("no creatures grants nothing", func(t *testing.T) {
		t.Parallel()
		if got := reductionWith(nil); got != 0 {
			t.Fatalf("reduction with an empty board = %d, want 0", got)
		}
	})

	t.Run("opponent's big creature does not count", func(t *testing.T) {
		t.Parallel()
		// "if you control a creature ..." is controller scoped: an opponent's
		// power-6 creature must not discount the caster's spell.
		got := reductionWith(func(g *game.Game) { addCreatureWithPower(g, game.Player2, 6) })
		if got != 0 {
			t.Fatalf("reduction from an opponent's power-6 creature = %d, want 0", got)
		}
	})
}
