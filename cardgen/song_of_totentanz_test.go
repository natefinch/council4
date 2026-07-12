package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerVariableRatTokensWithCantBlock(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Rat Song",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{X}{R}",
		OracleText: "Create X 1/1 black Rat creature tokens with \"This token can't block.\"",
	})
	create, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.CreateToken)
	if !ok || !create.Amount.IsDynamic() {
		t.Fatalf("create = %#v, want variable token count", face.SpellAbility.Val)
	}

}

func TestLowerControlledCreaturesGainHaste(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Haste Song",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{R}",
		OracleText: "Creatures you control gain haste until end of turn.",
	})
	if len(face.SpellAbility.Val.Modes[0].Sequence) != 1 {
		t.Fatalf("spell = %#v, want one haste instruction", face.SpellAbility.Val)
	}
}

func TestLowerSongOfTotentanz(t *testing.T) {
	t.Parallel()
	const oracleText = "Create X 1/1 black Rat creature tokens with \"This token can't block.\" Creatures you control gain haste until end of turn."
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Song of Totentanz",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{X}{R}",
		OracleText: oracleText,
	})
	if len(face.SpellAbility.Val.Modes[0].Sequence) != 2 {
		t.Fatalf("spell = %#v, want create then haste", face.SpellAbility.Val)
	}
}
