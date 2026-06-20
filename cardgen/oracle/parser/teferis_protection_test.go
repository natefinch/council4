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

func TestParseSourceNameExilePreservesAbilityShell(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source  string
		context Context
		want    AbilityKind
	}{
		"spell": {
			source:  "Exile Test Spell.",
			context: Context{InstantOrSorcery: true, CardName: "Test Spell"},
			want:    AbilitySpell,
		},
		"activated": {
			source:  "{T}: Exile Test Relic.",
			context: Context{CardName: "Test Relic"},
			want:    AbilityActivated,
		},
		"triggered": {
			source:  "When Test Relic enters, draw a card. Exile Test Relic.",
			context: Context{CardName: "Test Relic"},
			want:    AbilityTriggered,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, test.context)
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %+v", diagnostics)
			}
			if len(document.Abilities) != 1 || document.Abilities[0].Kind != test.want {
				t.Fatalf("abilities = %+v, want one %s ability", document.Abilities, test.want)
			}
			var exileEffects []EffectSyntax
			for _, sentence := range document.Abilities[0].Sentences {
				for _, effect := range sentence.Effects {
					if effect.Kind == EffectExile {
						exileEffects = append(exileEffects, effect)
					}
				}
			}
			if len(exileEffects) != 1 ||
				!exileEffects[0].Exact ||
				len(exileEffects[0].References) != 1 ||
				exileEffects[0].References[0].Kind != ReferenceSelfName {
				t.Fatalf("exile effects = %+v, want exact self-name exile", exileEffects)
			}
		})
	}
}
