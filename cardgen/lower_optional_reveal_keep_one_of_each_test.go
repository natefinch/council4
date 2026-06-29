package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestLowerOptionalRevealKeepOneOfEach verifies the inclusive one-of-each reveal
// keep ("Reveal the top N cards. You may put a [type-A] card and/or a [type-B]
// card from among them into your hand. Put the rest into your graveyard.") lowers
// to one Mill of N publishing the milled cards, then one optional return-to-hand
// per named type, each capped at one matching milled card.
func TestLowerOptionalRevealKeepOneOfEach(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Reveal Keep One Of Each",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Reveal the top five cards of your library. You may put a creature card and/or an enchantment card from among them into your hand. Put the rest into your graveyard.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	seq := face.SpellAbility.Val.Modes[0].Sequence
	if len(seq) != 3 {
		t.Fatalf("sequence = %d instructions, want 3 (mill + two keeps)", len(seq))
	}
	mill, ok := seq[0].Primitive.(game.Mill)
	if !ok || mill.Amount != game.Fixed(5) || mill.Player != game.ControllerReference() {
		t.Fatalf("sequence[0] = %#v, want Mill 5 by controller", seq[0].Primitive)
	}
	if mill.PublishLinked == "" {
		t.Fatal("mill must publish its milled cards for the linked keeps")
	}
	wantTypes := []types.Card{types.Creature, types.Enchantment}
	for i, want := range wantTypes {
		keep, ok := seq[i+1].Primitive.(game.ChooseFromZone)
		if !ok {
			t.Fatalf("sequence[%d] = %#v, want ChooseFromZone", i+1, seq[i+1].Primitive)
		}
		if !seq[i+1].Optional {
			t.Fatalf("keep %d not optional", i)
		}
		if keep.SourceZone != zone.Graveyard || keep.Destination.Zone != zone.Hand {
			t.Fatalf("keep %d zones = %v -> %v, want graveyard -> hand", i, keep.SourceZone, keep.Destination.Zone)
		}
		if keep.Quantity != game.Fixed(1) || keep.Riders.FromLinked != mill.PublishLinked {
			t.Fatalf("keep %d = %#v, want one card from milled set", i, keep)
		}
		if len(keep.Filter.RequiredTypes) != 1 || keep.Filter.RequiredTypes[0] != want {
			t.Fatalf("keep %d filter = %v, want [%v]", i, keep.Filter.RequiredTypes, want)
		}
	}
}

// TestLowerOptionalRevealKeepOneOfEachBottomFailsClosed verifies a one-of-each
// keep whose remainder bottoms the library rather than graveyards it fails
// closed: the reveal/mill graveyard equivalence requires a graveyard remainder.
func TestLowerOptionalRevealKeepOneOfEachBottomFailsClosed(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Reveal Keep Bottom",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Reveal the top five cards of your library. You may put a creature card and/or a land card from among them into your hand. Put the rest on the bottom of your library.",
	})
}
