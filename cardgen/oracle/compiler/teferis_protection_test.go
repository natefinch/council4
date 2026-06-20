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
