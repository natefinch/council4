package compiler

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompileTurnFaceDownThenBecomeCharacteristics(t *testing.T) {
	compilation, diagnostics := compileSource(
		"Turn target creature face down. It's a 2/2 Cyberman artifact creature.",
		pipelineContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 || len(compilation.Abilities) != 1 {
		t.Fatalf("compilation = %#v diagnostics = %#v", compilation, diagnostics)
	}
	content := compilation.Abilities[0].Content
	if len(content.Targets) != 1 ||
		content.Targets[0].Selector.Kind != SelectorCreature ||
		len(content.Effects) != 2 {
		t.Fatalf("content = %#v", content)
	}
	if content.Effects[0].Kind != EffectTurnFaceDown {
		t.Fatalf("turn effect = %#v", content.Effects[0])
	}
	become := content.Effects[1]
	if become.Kind != EffectPolymorph ||
		!become.PolymorphPermanent ||
		become.PolymorphLosesAllAbilities ||
		become.PolymorphBasePower != 2 ||
		become.PolymorphBaseToughness != 2 ||
		!slices.Equal(become.PolymorphTypes, []types.Card{types.Artifact, types.Creature}) ||
		!slices.Equal(become.PolymorphSubtypes, []types.Sub{types.Cyberman}) {
		t.Fatalf("become payload = kind %v permanent %v loses %v pt %d/%d types %v subtypes %v",
			become.Kind, become.PolymorphPermanent, become.PolymorphLosesAllAbilities,
			become.PolymorphBasePower, become.PolymorphBaseToughness,
			become.PolymorphTypes, become.PolymorphSubtypes)
	}
	if len(become.References) != 1 ||
		become.References[0].Binding != ReferenceBindingTarget ||
		become.References[0].Occurrence != 0 {
		t.Fatalf("become references = %#v", become.References)
	}
}
