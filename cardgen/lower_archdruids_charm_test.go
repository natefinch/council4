package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLowerArchdruidsCharm(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Archdruid's Charm",
		Layout:   "normal",
		TypeLine: "Instant",
		ManaCost: "{G}{G}{G}",
		OracleText: "Choose one —\n" +
			"• Search your library for a creature or land card and reveal it. Put it onto the battlefield tapped if it's a land card. Otherwise, put it into your hand. Then shuffle.\n" +
			"• Put a +1/+1 counter on target creature you control. It deals damage equal to its power to target creature you don't control.\n" +
			"• Exile target artifact or enchantment.",
	})
	if !face.SpellAbility.Exists || len(face.SpellAbility.Val.Modes) != 3 {
		t.Fatalf("spell ability = %#v, want three modes", face.SpellAbility)
	}
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	if len(sequence) != 3 {
		t.Fatalf("mode 1 sequence = %#v, want search, conditional place, then shuffle", sequence)
	}
	search, ok := sequence[0].Primitive.(game.Search)
	if !ok ||
		!search.Spec.RevealOnly ||
		!search.Spec.Reveal ||
		len(search.Spec.Filter.RequiredTypesAny) != 2 {
		t.Fatalf("search = %#v", sequence[0].Primitive)
	}
	if search.Spec.Filter.RequiredTypesAny[0] != types.Creature ||
		search.Spec.Filter.RequiredTypesAny[1] != types.Land {
		t.Fatalf("search filter = %#v", search.Spec.Filter)
	}
	place, ok := sequence[1].Primitive.(game.ConditionalDestinationPlace)
	if !ok ||
		!place.ThenMandatory ||
		!place.EntryTapped ||
		place.Else != zone.Hand {
		t.Fatalf("conditional place = %#v", sequence[1].Primitive)
	}
	shuffle, ok := sequence[2].Primitive.(game.ShuffleLibrary)
	if !ok || shuffle.Player != game.ControllerReference() {
		t.Fatalf("instruction 2 = %#v, want controller library shuffle", sequence[2].Primitive)
	}
}
