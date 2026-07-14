package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestLowerDevotionLookWinTrigger verifies that Thassa's Oracle's exact enters
// trigger body lowers to a two-instruction sequence: a Dig that looks at the top
// X (devotion to blue) cards, keeps up to one on top of the library, and bottoms
// the rest, followed by a PlayerWinsGame gated on the controller's library size
// being at most that same live devotion.
func TestLowerDevotionLookWinTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Oracle",
		Layout:   "normal",
		TypeLine: "Creature — Merfolk Wizard",
		OracleText: "When this creature enters, look at the top X cards of your library, " +
			"where X is your devotion to blue. Put up to one of them on top of your library " +
			"and the rest on the bottom of your library in a random order. " +
			"If X is greater than or equal to the number of cards in your library, you win the game. " +
			"(Each {U} in the mana costs of permanents you control counts toward your devotion to blue.)",
		Power:     new("1"),
		Toughness: new("3"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if trigger.Trigger.Pattern.Source != game.TriggerSourceSelf ||
		trigger.Trigger.Pattern.Event != game.EventPermanentEnteredBattlefield {
		t.Fatalf("trigger pattern = %#v, want self enters-the-battlefield", trigger.Trigger.Pattern)
	}
	seq := trigger.Content.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(seq))
	}

	dig, ok := seq[0].Primitive.(game.Dig)
	if !ok {
		t.Fatalf("instruction[0] = %#v, want Dig", seq[0].Primitive)
	}
	if dig.Destination != zone.Library {
		t.Errorf("Dig.Destination = %v, want zone.Library", dig.Destination)
	}
	if !dig.TakeUpTo {
		t.Error("Dig.TakeUpTo = false, want true (up to one)")
	}
	if dig.Reveal {
		t.Error("Dig.Reveal = true, want false (a library-top dig does not reveal)")
	}
	if dig.Remainder != game.DigRemainderLibraryBottom {
		t.Errorf("Dig.Remainder = %v, want DigRemainderLibraryBottom", dig.Remainder)
	}
	lookAmount := dig.Look.DynamicAmount()
	if !dig.Look.IsDynamic() || !lookAmount.Exists || lookAmount.Val.Kind != game.DynamicAmountDevotion ||
		len(lookAmount.Val.Colors) != 1 || lookAmount.Val.Colors[0] != color.Blue {
		t.Fatalf("Dig.Look = %#v, want dynamic devotion to blue", dig.Look)
	}
	if dig.Take.IsDynamic() || dig.Take.Value() != 1 {
		t.Errorf("Dig.Take = %#v, want fixed 1", dig.Take)
	}

	win, ok := seq[1].Primitive.(game.PlayerWinsGame)
	if !ok {
		t.Fatalf("instruction[1] = %#v, want PlayerWinsGame", seq[1].Primitive)
	}
	if win.Player.Kind() == game.PlayerReferenceNone {
		t.Error("PlayerWinsGame.Player = none, want the controller reference")
	}
	if !seq[1].Condition.Exists || !seq[1].Condition.Val.Condition.Exists {
		t.Fatalf("win instruction condition = %#v, want an aggregate condition", seq[1].Condition)
	}
	aggregates := seq[1].Condition.Val.Condition.Val.Aggregates
	if len(aggregates) != 1 {
		t.Fatalf("win condition aggregates = %#v, want one comparison", aggregates)
	}
	agg := aggregates[0]
	if agg.Aggregate != game.AggregateControllerLibrarySize || agg.Op != compare.LessOrEqual {
		t.Errorf("win comparison = %#v, want controller library size <= threshold", agg)
	}
	if !agg.ValueAmount.Exists || agg.ValueAmount.Val.Kind != game.DynamicAmountDevotion ||
		len(agg.ValueAmount.Val.Colors) != 1 || agg.ValueAmount.Val.Colors[0] != color.Blue {
		t.Fatalf("win threshold = %#v, want dynamic devotion to blue", agg.ValueAmount)
	}
}
