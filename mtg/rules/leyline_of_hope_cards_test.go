package rules

import (
	"testing"

	cards "github.com/natefinch/council4/mtg/cards/l"
	"github.com/natefinch/council4/mtg/game"
)

// TestLeylineOfHopeAddsOneToLifeGain proves the real card's "If you would gain
// life, you gain that much life plus 1 instead." replacement adds one to each
// life-gain event for the controller only.
func TestLeylineOfHopeAddsOneToLifeGain(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, cards.LeylineOfHope())

	if got := gainLife(g, game.Player1, 3); got != 4 {
		t.Fatalf("controller gainLife(3) = %d, want 4 (3 plus 1)", got)
	}
	if got := gainLife(g, game.Player1, 1); got != 2 {
		t.Fatalf("controller gainLife(1) = %d, want 2 (1 plus 1)", got)
	}
	if got := gainLife(g, game.Player2, 3); got != 3 {
		t.Fatalf("opponent gainLife(3) = %d, want 3 (unaffected)", got)
	}
}

// TestLeylineOfHopeAnthemActivatesAtSevenAboveStarting proves the conditional
// anthem "As long as you have at least 7 life more than your starting life
// total, creatures you control get +2/+2." turns on only once the controller is
// seven or more life above their starting total.
func TestLeylineOfHopeAnthemActivatesAtSevenAboveStarting(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, cards.LeylineOfHope())
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	start := g.Players[game.Player1].StartingLife

	// One short of the threshold: no buff.
	g.Players[game.Player1].Life = start + 6
	if got := effectivePower(g, creature); got != 2 {
		t.Fatalf("creature power at +6 life = %d, want 2 (anthem inactive)", got)
	}

	// Exactly at the threshold: +2/+2.
	g.Players[game.Player1].Life = start + 7
	if got := effectivePower(g, creature); got != 4 {
		t.Fatalf("creature power at +7 life = %d, want 4 (anthem active)", got)
	}
	toughness, ok := effectiveToughness(g, creature)
	if !ok || toughness != 4 {
		t.Fatalf("creature toughness at +7 life = %d (ok=%v), want 4", toughness, ok)
	}
}

// TestLeylineOfHopeAnthemOnlyBuffsControllerCreatures proves the anthem is
// scoped to the controller: an opponent's creature is never buffed, even when
// the controller is well above their starting life total.
func TestLeylineOfHopeAnthemOnlyBuffsControllerCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, cards.LeylineOfHope())
	opponentCreature := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Players[game.Player1].Life = g.Players[game.Player1].StartingLife + 20

	if got := effectivePower(g, opponentCreature); got != 2 {
		t.Fatalf("opponent creature power = %d, want 2 (anthem is controller-scoped)", got)
	}
}
