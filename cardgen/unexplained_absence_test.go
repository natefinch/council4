package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerUnexplainedAbsence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Unexplained Absence",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{3}{W}",
		OracleText: "For each player, exile up to one target nonland permanent that player controls. For each permanent exiled this way, its controller cloaks the top card of their library. (To cloak a card, put it onto the battlefield face down as a 2/2 creature with ward {2}. Turn it face up any time for its mana cost if it's a creature card.)",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 0 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %#v", mode)
	}
	exile, ok := mode.Sequence[0].Primitive.(game.ExileForEachPlayer)
	if !ok ||
		exile.Chooser != game.ControllerReference() ||
		len(exile.Selection.ExcludedTypes) != 1 ||
		exile.Selection.ExcludedTypes[0] != types.Land ||
		exile.LinkedKey == "" {
		t.Fatalf("exile = %#v", mode.Sequence[0].Primitive)
	}
	cloak, ok := mode.Sequence[1].Primitive.(game.ManifestForEachLinked)
	if !ok || !cloak.Cloak || cloak.Dread || cloak.LinkedKey != exile.LinkedKey {
		t.Fatalf("cloak = %#v", mode.Sequence[1].Primitive)
	}
}
