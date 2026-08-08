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

// TestLowerSawtoothLoonBottomNoOrderClause guards the two generalizations
// this family's development added to HandLibraryPutSyntax: the bottom
// library position (previously top-only) and the complete omission of the
// "in any order" qualifier (previously required verbatim), with a real
// card that uses both at once: "When this creature enters, draw two cards,
// then put two cards from your hand on the bottom of your library."
// (Sawtooth Loon). CR 401.4 grants the player-chosen-order permission by
// default whenever multiple cards go to the same library position at once,
// so the omitted qualifier is not a narrower shape than Brainstorm's -- see
// HandLibraryPutSyntax's doc comment. This also confirms the shared
// sequence lowerer is reachable from a triggered ability's content, not
// only a spell's.
func TestLowerSawtoothLoonBottomNoOrderClause(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Sawtooth Loon",
		Layout:     "normal",
		TypeLine:   "Creature — Bird",
		ManaCost:   "{2}{U}",
		OracleText: "Flying\nWhen this creature enters, draw two cards, then put two cards from your hand on the bottom of your library.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	var mode game.Mode
	found := false
	for _, ability := range face.TriggeredAbilities {
		for _, m := range ability.Content.Modes {
			if len(m.Sequence) == 2 {
				mode = m
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("triggered abilities = %#v, want a two-instruction mode", face.TriggeredAbilities)
	}
	draw, ok := mode.Sequence[0].Primitive.(game.Draw)
	if !ok || draw.Amount.Value() != 2 {
		t.Fatalf("draw = %#v, want controller draw two", mode.Sequence[0].Primitive)
	}
	move, ok := mode.Sequence[1].Primitive.(game.MoveCard)
	if !ok ||
		move.Amount.Value() != 2 ||
		move.FromZone != zone.Hand ||
		move.Destination != zone.Library ||
		!move.DestinationBottom {
		t.Fatalf("move = %#v, want choose two hand cards for library bottom", mode.Sequence[1].Primitive)
	}
}

// TestLowerConchHornTopNoOrderSingleCard guards the omitted-order-qualifier
// generalization for the top position with a real activated ability (Conch
// Horn: "{1}, {T}, Sacrifice this artifact: Draw two cards, then put a card
// from your hand on top of your library."), confirming the shared sequence
// lowerer is reachable from an activated ability's content as well.
func TestLowerConchHornTopNoOrderSingleCard(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Conch Horn",
		Layout:     "normal",
		TypeLine:   "Artifact",
		ManaCost:   "{3}",
		OracleText: "{1}, {T}, Sacrifice this artifact: Draw two cards, then put a card from your hand on top of your library.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %#v, want 1", face.ActivatedAbilities)
	}
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want draw then move", mode.Sequence)
	}
	move, ok := mode.Sequence[1].Primitive.(game.MoveCard)
	if !ok ||
		move.Amount.Value() != 1 ||
		move.FromZone != zone.Hand ||
		move.Destination != zone.Library ||
		move.DestinationBottom {
		t.Fatalf("move = %#v, want choose one hand card for library top", mode.Sequence[1].Primitive)
	}
}

func TestBrainstormCategoryFailsClosedOutsideEnvelope(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		// "In a random order" and "in the same order" name a genuinely
		// different ordering rule than the player's free choice CR 401.4
		// grants by default for multiple cards placed in the same library
		// position at once -- see HandLibraryPutSyntax's doc comment.
		"Draw three cards, then put two cards from your hand on top of your library in a random order.",
		"Draw three cards, then put two cards from an opponent's hand on top of your library in any order.",
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
