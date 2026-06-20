package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

const brainstormOracleText = "Draw three cards, then put two cards from your hand on top of your library in any order."

func TestLowerBrainstormSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Brainstorm",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: brainstormOracleText,
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want draw then move", mode.Sequence)
	}
	draw, ok := mode.Sequence[0].Primitive.(game.Draw)
	if !ok || draw.Amount.Value() != 3 || draw.Player.Kind() != game.PlayerReferenceController {
		t.Fatalf("draw = %#v, want controller draw three", mode.Sequence[0].Primitive)
	}
	move, ok := mode.Sequence[1].Primitive.(game.MoveCard)
	if !ok ||
		move.Player.Kind() != game.PlayerReferenceController ||
		move.Amount.Value() != 2 ||
		move.FromZone != zone.Hand ||
		move.Destination != zone.Library ||
		move.DestinationBottom {
		t.Fatalf("move = %#v, want choose two hand cards for library top", mode.Sequence[1].Primitive)
	}
}

func TestGenerateBrainstormExecutableSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Brainstorm",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: brainstormOracleText,
	}, "b")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.Draw",
		"game.MoveCard",
		"game.Fixed(2)",
		"zone.Hand",
		"zone.Library",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

func TestBrainstormCategoryFailsClosedOutsideEnvelope(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Draw three cards, then put two cards from your hand on the bottom of your library in any order.",
		"Draw three cards, then put two cards from your hand on top of your library in a random order.",
		"Draw three cards, then put two cards from an opponent's hand on top of your library in any order.",
		"Draw three cards, then put two cards from your hand on top of your library.",
		"Draw three cards, then put two cards from your hand on top of your library in the same order.",
		"Draw three cards, then put X cards from your hand on top of your library in any order.",
		"Draw three cards, then reveal two cards from your hand, then put them on top of your library in any order.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			faces, _ := lowerExecutableFaces(&ScryfallCard{
				Name:       "Unsupported Brainstorm Variant",
				Layout:     "normal",
				TypeLine:   "Instant",
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
