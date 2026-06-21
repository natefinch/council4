package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerWinGameEnterTrigger verifies that the exact controller effect "you
// win the game" lowers to a single PlayerWinsGame instruction scoped to the
// ability's controller when it is the body of an enters-the-battlefield trigger.
func TestLowerWinGameEnterTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Oracle",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "When this enchantment enters, you win the game.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	seq := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(seq) != 1 {
		t.Fatalf("sequence length = %d, want 1", len(seq))
	}
	win, ok := seq[0].Primitive.(game.PlayerWinsGame)
	if !ok {
		t.Fatalf("instruction[0] = %#v, want PlayerWinsGame", seq[0].Primitive)
	}
	if win.Player.Kind() == game.PlayerReferenceNone {
		t.Error("PlayerWinsGame.Player = none, want the controller reference")
	}
}

// TestLowerWinGameUpkeepInterveningIf verifies that the very common "At the
// beginning of your upkeep, if <state>, you win the game." alternate win
// condition (Felidar Sovereign, Test of Endurance, Revel in Riches) lowers end
// to end, pairing the supported upkeep trigger and intervening-if condition with
// the new win effect.
func TestLowerWinGameUpkeepInterveningIf(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Sovereign",
		Layout:     "normal",
		TypeLine:   "Creature — Cat Beast",
		OracleText: "At the beginning of your upkeep, if you have 40 or more life, you win the game.",
		Power:      new("4"),
		Toughness:  new("6"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	seq := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(seq) != 1 {
		t.Fatalf("sequence length = %d, want 1", len(seq))
	}
	if _, ok := seq[0].Primitive.(game.PlayerWinsGame); !ok {
		t.Fatalf("instruction[0] = %#v, want PlayerWinsGame", seq[0].Primitive)
	}
}
