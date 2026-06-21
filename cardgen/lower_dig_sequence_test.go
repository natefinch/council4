package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerDigSpellSequence verifies the impulse "dig" sequence ("Look at the
// top N cards of your library. Put M of them into your hand and the rest into
// your graveyard.") lowers to a single Dig primitive that looks at N, takes M,
// and sends the remainder to the controller's graveyard.
func TestLowerDigSpellSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Dig",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Look at the top three cards of your library. Put one of them into your hand and the rest into your graveyard.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 0 || len(mode.Sequence) != 1 {
		t.Fatalf("mode = %+v, want no targets and one instruction", mode)
	}
	dig, ok := mode.Sequence[0].Primitive.(game.Dig)
	if !ok {
		t.Fatalf("primitive = %T, want game.Dig", mode.Sequence[0].Primitive)
	}
	if dig.Look != game.Fixed(3) || dig.Take != game.Fixed(1) {
		t.Fatalf("dig = %+v, want Look 3 Take 1", dig)
	}
	if dig.Remainder != game.DigRemainderGraveyard {
		t.Fatalf("dig remainder = %v, want graveyard", dig.Remainder)
	}
	if dig.Player != game.ControllerReference() {
		t.Fatalf("dig player = %+v, want controller", dig.Player)
	}
}

// TestLowerDigTakeTwoSequence verifies a dig that takes two cards and uses the
// bare "Put two of them" / "the other" wordings still lowers with the correct
// look and take counts.
func TestLowerDigTakeTwoSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Dig Two",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Look at the top three cards of your library. Put two of them into your hand and the rest into your graveyard.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	dig, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.Dig)
	if !ok {
		t.Fatalf("primitive = %T, want game.Dig", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
	if dig.Look != game.Fixed(3) || dig.Take != game.Fixed(2) {
		t.Fatalf("dig = %+v, want Look 3 Take 2", dig)
	}
}

// TestLowerDigTriggeredSequence verifies the dig sequence also lowers inside a
// triggered ability, the common printing on creatures.
func TestLowerDigTriggeredSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Dig Creature",
		Layout:     "normal",
		TypeLine:   "Creature — Bird",
		OracleText: "When Test Dig Creature enters, look at the top two cards of your library. Put one of them into your hand and the other into your graveyard.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	dig, ok := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.Dig)
	if !ok {
		t.Fatalf("primitive = %T, want game.Dig", face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive)
	}
	if dig.Look != game.Fixed(2) || dig.Take != game.Fixed(1) {
		t.Fatalf("dig = %+v, want Look 2 Take 1", dig)
	}
}

// TestLowerDigBottomRemainderSequence verifies the library-bottom remainder
// forms ("the rest on the bottom of your library in any order / in a random
// order") lower to a Dig primitive whose remainder is the library bottom.
func TestLowerDigBottomRemainderSequence(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Look at the top four cards of your library. Put one of them into your hand and the rest on the bottom of your library in any order.",
		"Look at the top four cards of your library. Put one of them into your hand and the rest on the bottom of your library in a random order.",
		"Look at the top two cards of your library. Put one of them into your hand and the other on the bottom of your library.",
	} {
		face := lowerSingleFace(t, &ScryfallCard{
			Name:       "Test Dig Bottom",
			Layout:     "normal",
			TypeLine:   "Sorcery",
			OracleText: text,
		})
		if !face.SpellAbility.Exists {
			t.Fatalf("OracleText %q did not lower a spell ability", text)
		}
		dig, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.Dig)
		if !ok {
			t.Fatalf("OracleText %q primitive = %T, want game.Dig", text, face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
		}
		if dig.Remainder != game.DigRemainderLibraryBottom {
			t.Fatalf("OracleText %q dig remainder = %v, want library bottom", text, dig.Remainder)
		}
	}
}

// TestLowerDigFailsClosed verifies dig shapes the Dig primitive does not model
// stay unsupported: an unrecognized remainder destination, a variable looked-at
// count, and a degenerate look count that does not exceed the take count.
func TestLowerDigFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		"Look at the top four cards of your library. Put one of them into your hand and the rest on top of your library in any order.",
		"Look at the top X cards of your library. Put one of them into your hand and the rest into your graveyard.",
		"Look at the top one cards of your library. Put one of them into your hand and the rest into your graveyard.",
	}
	for _, text := range rejected {
		faces, _ := lowerExecutableFaces(&ScryfallCard{
			Name:       "Test Reject",
			Layout:     "normal",
			TypeLine:   "Sorcery",
			OracleText: text,
		})
		for _, face := range faces {
			if face.SpellAbility.Exists {
				t.Errorf("OracleText %q lowered a spell ability, want fail closed", text)
			}
		}
	}
}
