package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

const ponderOracleText = "Look at the top three cards of your library, then put them back in any order. You may shuffle.\nDraw a card."

func TestLowerPonderSequenceWithoutCardNameDependency(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Contemplate",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: ponderOracleText,
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 3 {
		t.Fatalf("sequence = %#v, want reorder, optional shuffle, draw", mode.Sequence)
	}
	reorder, ok := mode.Sequence[0].Primitive.(game.ReorderLibraryTop)
	if !ok || reorder.Player.Kind() != game.PlayerReferenceController || reorder.Amount.Value() != 3 {
		t.Fatalf("reorder = %#v, want controller top three", mode.Sequence[0].Primitive)
	}
	shuffle, ok := mode.Sequence[1].Primitive.(game.ShuffleLibrary)
	if !ok || shuffle.Player.Kind() != game.PlayerReferenceController || !mode.Sequence[1].Optional {
		t.Fatalf("shuffle = %#v, want optional controller shuffle", mode.Sequence[1])
	}
	draw, ok := mode.Sequence[2].Primitive.(game.Draw)
	if !ok || draw.Player.Kind() != game.PlayerReferenceController || draw.Amount.Value() != 1 {
		t.Fatalf("draw = %#v, want controller draw one", mode.Sequence[2].Primitive)
	}
}

func TestLowerPonderSequenceInTriggeredAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Reflective Wizard",
		Layout:     "normal",
		TypeLine:   "Creature — Human Wizard",
		OracleText: "When this creature enters, look at the top three cards of your library, then put them back in any order. You may shuffle. Draw a card.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %#v, want one", face.TriggeredAbilities)
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 3 {
		t.Fatalf("sequence = %#v, want reorder, optional shuffle, draw", mode.Sequence)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.ReorderLibraryTop); !ok {
		t.Fatalf("first primitive = %T, want ReorderLibraryTop", mode.Sequence[0].Primitive)
	}
	if _, ok := mode.Sequence[1].Primitive.(game.ShuffleLibrary); !ok || !mode.Sequence[1].Optional {
		t.Fatalf("second instruction = %#v, want optional ShuffleLibrary", mode.Sequence[1])
	}
	if _, ok := mode.Sequence[2].Primitive.(game.Draw); !ok {
		t.Fatalf("third primitive = %T, want Draw", mode.Sequence[2].Primitive)
	}
}

func TestGeneratePonderExecutableSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Contemplate",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: ponderOracleText,
	}, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.ReorderLibraryTop",
		"game.ShuffleLibrary",
		"Optional: true",
		"game.Draw",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

func TestPonderCategoryFailsClosedOutsideExactEnvelope(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Look at the top three cards of your library, then put them back in any order. You may shuffle.",
		"Look at the top three cards of your library, then put them back in a random order. You may shuffle.\nDraw a card.",
		"Look at the top three cards of your library, then put them back in any order. Shuffle.\nDraw a card.",
		"Look at the top three cards of an opponent's library, then put them back in any order. You may shuffle.\nDraw a card.",
		"Look at the top X cards of your library, then put them back in any order. You may shuffle.\nDraw a card.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			faces, _ := lowerExecutableFaces(&ScryfallCard{
				Name:       "Unsupported Contemplation",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: oracleText,
			})
			for i := range faces {
				if faces[i].SpellAbility.Exists {
					t.Fatalf("%q unexpectedly lowered: %#v", oracleText, faces[i].SpellAbility.Val)
				}
			}
		})
	}
}

func TestPonderTriggerWithoutDrawFailsClosed(t *testing.T) {
	t.Parallel()
	faces, _ := lowerExecutableFaces(&ScryfallCard{
		Name:       "Incomplete Reflection",
		Layout:     "normal",
		TypeLine:   "Creature — Human Wizard",
		OracleText: "When this creature enters, look at the top three cards of your library, then put them back in any order. You may shuffle.",
	})
	for i := range faces {
		if len(faces[i].TriggeredAbilities) != 0 {
			t.Fatalf("draw-less trigger unexpectedly lowered: %#v", faces[i].TriggeredAbilities)
		}
	}
}
