package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerTearAsunder(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Tear Asunder",
		Layout:   "normal",
		ManaCost: "{1}{G}",
		TypeLine: "Instant",
		Colors:   []string{"G"},
		OracleText: "Kicker {1}{B} (You may pay an additional {1}{B} as you cast this spell.)\n" +
			"Exile target artifact or enchantment. If this spell was kicked, exile target nonland permanent instead.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 2 {
		t.Fatalf("targets = %#v, want base and kicked target specs", mode.Targets)
	}
	baseTarget := mode.Targets[0]
	if baseTarget.Gate != game.TargetGateSpellNotKicked ||
		!baseTarget.Selection.Exists ||
		!slices.Equal(baseTarget.Selection.Val.RequiredTypesAny, []types.Card{types.Artifact, types.Enchantment}) {
		t.Fatalf("base target = %#v", baseTarget)
	}
	kickedTarget := mode.Targets[1]
	if kickedTarget.Gate != game.TargetGateSpellKicked ||
		!kickedTarget.Selection.Exists ||
		!slices.Equal(kickedTarget.Selection.Val.ExcludedTypes, []types.Card{types.Land}) {
		t.Fatalf("kicked target = %#v", kickedTarget)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want complementary exile instructions", mode.Sequence)
	}
	for i, want := range []game.ObjectReference{
		game.TargetPermanentReference(0),
		game.TargetPermanentReference(1),
	} {
		exile, ok := mode.Sequence[i].Primitive.(game.Exile)
		if !ok || exile.Object != want {
			t.Fatalf("sequence[%d] = %#v, want exile of target %d", i, mode.Sequence[i].Primitive, i)
		}
		if !mode.Sequence[i].Condition.Exists ||
			!mode.Sequence[i].Condition.Val.Condition.Exists ||
			!mode.Sequence[i].Condition.Val.Condition.Val.SpellWasKicked ||
			mode.Sequence[i].Condition.Val.Condition.Val.Negate != (i == 0) {
			t.Fatalf("sequence[%d] condition = %#v", i, mode.Sequence[i].Condition)
		}
	}
}
