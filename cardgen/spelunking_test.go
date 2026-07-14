package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerSpelunking(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Spelunking",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		ManaCost:   "{2}{G}",
		OracleText: "When this enchantment enters, draw a card, then you may put a land card from your hand onto the battlefield. If you put a Cave onto the battlefield this way, you gain 4 life.\nLands you control enter untapped.",
	})
	if len(face.TriggeredAbilities) != 1 || len(face.ReplacementAbilities) != 1 {
		t.Fatalf("face = %#v", face)
	}
	sequence := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(sequence) != 3 {
		t.Fatalf("sequence = %#v", sequence)
	}
	if _, ok := sequence[0].Primitive.(game.Draw); !ok {
		t.Fatalf("draw = %#v", sequence[0])
	}
	put, ok := sequence[1].Primitive.(game.ChooseFromZone)
	if !ok ||
		!sequence[1].Optional ||
		put.Riders.PublishLinked == "" ||
		len(put.Filter.RequiredTypes) != 1 ||
		put.Filter.RequiredTypes[0] != types.Land {
		t.Fatalf("put = %#v", sequence[1])
	}
	gain, ok := sequence[2].Primitive.(game.GainLife)
	if !ok ||
		gain.Amount.Value() != 4 ||
		!sequence[2].ResultGate.Exists ||
		!sequence[2].Condition.Exists {
		t.Fatalf("gain = %#v", sequence[2])
	}
	replacement := face.ReplacementAbilities[0].Replacement
	if !replacement.EntersUntapped ||
		!replacement.EntersUntappedOthers ||
		replacement.EntersTappedSelection == nil ||
		len(replacement.EntersTappedSelection.RequiredTypesAny) != 1 ||
		replacement.EntersTappedSelection.RequiredTypesAny[0] != types.Land {
		t.Fatalf("replacement = %#v", replacement)
	}
}
