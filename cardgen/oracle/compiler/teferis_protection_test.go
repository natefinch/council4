package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestCompileTeferisProtectionEffectsTextBlind(t *testing.T) {
	t.Parallel()
	document, diagnostics := parser.Parse(
		"Until your next turn, your life total can't change and you gain protection from everything. All permanents you control phase out. Exile this spell.",
		parser.Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("parser diagnostics = %+v", diagnostics)
	}
	for abilityIndex := range document.Abilities {
		for sentenceIndex := range document.Abilities[abilityIndex].Sentences {
			sentence := &document.Abilities[abilityIndex].Sentences[sentenceIndex]
			for effectIndex := range sentence.Effects {
				sentence.Effects[effectIndex].Text = "not semantic input"
			}
		}
	}
	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compiler diagnostics = %+v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	want := []EffectKind{
		EffectLifeTotalCantChange,
		EffectProtectionFromEverything,
		EffectPhaseOut,
		EffectExile,
	}
	if len(effects) != len(want) {
		t.Fatalf("effects = %+v, want %d", effects, len(want))
	}
	for i := range want {
		if effects[i].Kind != want[i] {
			t.Fatalf("effects[%d].Kind = %v, want %v", i, effects[i].Kind, want[i])
		}
	}
}

func TestCompileSourceNameExilePreservesAbilityShell(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source  string
		context pipelineContext
		want    AbilityKind
	}{
		"spell": {
			source:  "Exile Test Spell.",
			context: pipelineContext{InstantOrSorcery: true, CardName: "Test Spell"},
			want:    AbilitySpell,
		},
		"activated": {
			source:  "{T}: Exile Test Relic.",
			context: pipelineContext{CardName: "Test Relic"},
			want:    AbilityActivated,
		},
		"triggered": {
			source:  "When Test Relic enters, draw a card. Exile Test Relic.",
			context: pipelineContext{CardName: "Test Relic"},
			want:    AbilityTriggered,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(test.source, test.context)
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %+v", diagnostics)
			}
			if len(compilation.Abilities) != 1 || compilation.Abilities[0].Kind != test.want {
				t.Fatalf("abilities = %+v, want one %s ability", compilation.Abilities, test.want)
			}
			var exileEffects []CompiledEffect
			for _, effect := range compilation.Abilities[0].Content.Effects {
				if effect.Kind == EffectExile {
					exileEffects = append(exileEffects, effect)
				}
			}
			if len(exileEffects) != 1 ||
				!exileEffects[0].Exact ||
				len(exileEffects[0].References) != 1 ||
				exileEffects[0].References[0].Kind != ReferenceSelfName ||
				exileEffects[0].References[0].Binding != ReferenceBindingSource {
				t.Fatalf("exile effects = %+v, want exact source-bound exile", exileEffects)
			}
		})
	}
}
