package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerUrzasRuinousBlast(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Urza's Ruinous Blast",
		Layout:     "normal",
		TypeLine:   "Legendary Sorcery",
		ManaCost:   "{4}{W}",
		OracleText: "(You may cast a legendary sorcery only if you control a legendary creature or planeswalker.)\nExile all nonland permanents that aren't legendary.",
	})
	exile, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.Exile)
	if !ok || !exile.Group.Valid() {
		t.Fatalf("exile = %#v", face.SpellAbility.Val)
	}
	selection := exile.Group.Selection()
	if len(selection.ExcludedTypes) != 1 ||
		selection.ExcludedTypes[0] != types.Land ||
		selection.ExcludedSupertype != types.Legendary {
		t.Fatalf("selection = %#v", selection)
	}
}
