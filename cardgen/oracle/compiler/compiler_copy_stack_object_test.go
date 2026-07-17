package compiler

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
)

func TestCompileCopyStackObjectColorException(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Copy target instant or sorcery spell, except that the copy is red. You may choose new targets for the copy.",
		pipelineContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effect := compilation.Abilities[0].Content.Effects[0]
	if effect.Kind != EffectCopyStackObject || !effect.CopyMayChooseNewTargets {
		t.Fatalf("copy effect = %#v", effect)
	}
	if effect.CopyCharacteristics == nil ||
		!slices.Equal(effect.CopyCharacteristics.SetColors, []color.Color{color.Red}) {
		t.Fatalf("copy characteristics = %#v, want red", effect.CopyCharacteristics)
	}
}
