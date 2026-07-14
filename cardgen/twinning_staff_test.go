package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerTwinningStaff(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Twinning Staff",
		Layout:   "normal",
		ManaCost: "{3}",
		TypeLine: "Artifact",
		OracleText: "If you would copy a spell one or more times, instead copy it that many times plus an additional time. You may choose new targets for the additional copy.\n" +
			"{7}, {T}: Copy target instant or sorcery spell you control. You may choose new targets for the copy.",
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("replacement abilities = %d, want 1", len(face.ReplacementAbilities))
	}
	replacement := face.ReplacementAbilities[0].Replacement
	if replacement.SpellCopyAddend != 1 ||
		!replacement.SpellCopyAdditionalMayChooseNewTargets {
		t.Fatalf("spell-copy replacement = %#v", replacement)
	}
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	copyEffect, ok := face.ActivatedAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.CopyStackObject)
	if !ok || !copyEffect.MayChooseNewTargets {
		t.Fatalf("activated copy = %#v", face.ActivatedAbilities[0].Content)
	}
}
