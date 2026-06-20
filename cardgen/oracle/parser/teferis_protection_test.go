package parser

import "testing"

func TestParseTeferisProtectionTypedEffects(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Until your next turn, your life total can't change and you gain protection from everything. All permanents you control phase out. Exile this spell.",
		Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
	var effects []EffectSyntax
	for _, sentence := range document.Abilities[0].Sentences {
		effects = append(effects, sentence.Effects...)
	}
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
		if effects[i].Kind != want[i] || !effects[i].Exact {
			t.Fatalf("effects[%d] = %+v, want exact %v", i, effects[i], want[i])
		}
	}

	if effects[0].Duration != EffectDurationUntilYourNextTurn ||
		effects[1].Duration != EffectDurationUntilYourNextTurn {
		t.Fatalf("durations = %v/%v, want until your next turn", effects[0].Duration, effects[1].Duration)
	}
}

func TestSourceSpellExileFailsClosedForOtherThisObject(t *testing.T) {
	t.Parallel()
	document, _ := Parse("Exile this creature.", Context{InstantOrSorcery: true})
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 || effects[0].Exact {
		t.Fatalf("effects = %+v, want inexact unsupported variant", effects)
	}
}
