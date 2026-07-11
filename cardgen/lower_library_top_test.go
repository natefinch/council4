package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestLowerSenseisDiviningTopEndToEnd verifies the anchor card compiles to a
// faithful runtime plan: a {1} ability that reorders the top three cards of the
// library, and a {T} ability that draws a card then puts the source artifact on
// top of its owner's library.
func TestLowerSenseisDiviningTopEndToEnd(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Sensei's Divining Top",
		Layout:   "normal",
		TypeLine: "Artifact",
		OracleText: "{1}: Look at the top three cards of your library, then put them back in any order.\n" +
			"{T}: Draw a card, then put this artifact on top of its owner's library.",
	})
	if len(face.ActivatedAbilities) != 2 {
		t.Fatalf("activated abilities = %d, want 2", len(face.ActivatedAbilities))
	}

	reorderSeq := face.ActivatedAbilities[0].Content.Modes[0].Sequence
	if len(reorderSeq) != 1 {
		t.Fatalf("reorder sequence = %#v, want one instruction", reorderSeq)
	}
	reorder, ok := reorderSeq[0].Primitive.(game.ReorderLibraryTop)
	if !ok || reorder.Amount.Value() != 3 || reorder.Player.Kind() != game.PlayerReferenceController {
		t.Fatalf("reorder = %#v", reorderSeq[0].Primitive)
	}

	drawSeq := face.ActivatedAbilities[1].Content.Modes[0].Sequence
	if len(drawSeq) != 2 {
		t.Fatalf("draw sequence = %#v, want two instructions", drawSeq)
	}
	if _, ok := drawSeq[0].Primitive.(game.Draw); !ok {
		t.Fatalf("first instruction = %#v, want game.Draw", drawSeq[0].Primitive)
	}
	put, ok := drawSeq[1].Primitive.(game.PutPermanentOnLibrary)
	if !ok || put.Bottom || put.Object.Kind() != game.ObjectReferenceSourcePermanent {
		t.Fatalf("put = %#v", drawSeq[1].Primitive)
	}
}

// TestLowerStandaloneReorderLibraryTop verifies a lone "Look at the top N cards
// … put them back in any order." spell (Index) lowers to a single
// ReorderLibraryTop instruction.
func TestLowerStandaloneReorderLibraryTop(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Index",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Look at the top five cards of your library, then put them back in any order.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want one instruction", mode.Sequence)
	}
	reorder, ok := mode.Sequence[0].Primitive.(game.ReorderLibraryTop)
	if !ok || reorder.Amount.Value() != 5 || reorder.Player.Kind() != game.PlayerReferenceController {
		t.Fatalf("reorder = %#v", mode.Sequence[0].Primitive)
	}
}

// TestLowerPutSourceOnLibraryBottom verifies the bottom-of-library wording sets
// the PutPermanentOnLibrary Bottom flag.
func TestLowerPutSourceOnLibraryBottom(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Sink",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{T}: Put this artifact on the bottom of its owner's library.",
	})
	seq := face.ActivatedAbilities[0].Content.Modes[0].Sequence
	put, ok := seq[len(seq)-1].Primitive.(game.PutPermanentOnLibrary)
	if !ok || !put.Bottom || put.Object.Kind() != game.ObjectReferenceSourcePermanent {
		t.Fatalf("put = %#v", seq[len(seq)-1].Primitive)
	}
}

// TestLowerPutSourceFromGraveyardOnLibrary verifies the graveyard-recursion form
// "put this card from your graveyard on top of your library" (Champion of Stray
// Souls, Gate Colossus) lowers as graveyard recursion — the ability functions
// from the graveyard and the source card moves graveyard -> library — rather
// than as the battlefield self-tuck. The graveyard source zone alone
// distinguishes it from Sensei's Divining Top's battlefield "put this artifact on
// top of its owner's library".
func TestLowerPutSourceFromGraveyardOnLibrary(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		wantBottom bool
	}{
		{
			name:       "top",
			oracleText: "{5}{B}{B}: Put this card from your graveyard on top of your library.",
		},
		{
			name:       "bottom",
			oracleText: "{5}{B}{B}: Put this card from your graveyard on the bottom of your library.",
			wantBottom: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Graveyard Tuck",
				Layout:     "normal",
				TypeLine:   "Creature — Skeleton Warrior",
				OracleText: test.oracleText,
				Power:      new("4"),
				Toughness:  new("4"),
			})
			if len(face.ActivatedAbilities) != 1 {
				t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
			}
			ability := face.ActivatedAbilities[0]
			if ability.ZoneOfFunction != zone.Graveyard {
				t.Fatalf("zone of function = %v, want Graveyard (recursion functions from the graveyard)", ability.ZoneOfFunction)
			}
			seq := ability.Content.Modes[0].Sequence
			if len(seq) != 1 {
				t.Fatalf("sequence = %#v, want one instruction", seq)
			}
			move, ok := seq[0].Primitive.(game.MoveCard)
			if !ok {
				t.Fatalf("primitive = %#v, want game.MoveCard (graveyard -> library), not a battlefield tuck", seq[0].Primitive)
			}
			if move.Card.Kind != game.CardReferenceSource ||
				move.FromZone != zone.Graveyard ||
				move.Destination != zone.Library ||
				move.DestinationBottom != test.wantBottom {
				t.Fatalf("move = %#v, want source card graveyard -> library (bottom=%v)", move, test.wantBottom)
			}
		})
	}
}
