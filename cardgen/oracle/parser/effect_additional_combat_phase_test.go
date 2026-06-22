package parser

import (
	"testing"
)

// TestParseAdditionalCombatPhaseEffect proves the parser recognizes the
// extra-phase-insertion wording and records which phases are added.
func TestParseAdditionalCombatPhaseEffect(t *testing.T) {
	t.Parallel()
	cases := []struct {
		source string
		main   bool
	}{
		{"After this main phase, there is an additional combat phase followed by an additional main phase.", true},
		{"After this phase, there is an additional combat phase.", false},
		{"After this combat phase, there is an additional combat phase.", false},
	}
	for _, tc := range cases {
		document, diagnostics := Parse(tc.source, Context{InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			t.Fatalf("Parse(%q) diagnostics = %#v", tc.source, diagnostics)
		}
		effects := document.Abilities[0].Sentences[0].Effects
		if len(effects) != 1 {
			t.Fatalf("Parse(%q) effects = %#v, want one", tc.source, effects)
		}
		effect := effects[0]
		if effect.Kind != EffectAdditionalCombatPhase {
			t.Errorf("Parse(%q) kind = %v, want EffectAdditionalCombatPhase", tc.source, effect.Kind)
		}
		if !effect.Exact {
			t.Errorf("Parse(%q) Exact = false, want true", tc.source)
		}
		if !effect.AdditionalCombatPhase {
			t.Errorf("Parse(%q) AdditionalCombatPhase = false, want true", tc.source)
		}
		if effect.AdditionalMainPhase != tc.main {
			t.Errorf("Parse(%q) AdditionalMainPhase = %v, want %v", tc.source, effect.AdditionalMainPhase, tc.main)
		}
	}
}

// TestParseAdditionalCombatPhaseEffectFailsClosed proves unrelated or malformed
// wordings do not match the extra-phase recognizer.
func TestParseAdditionalCombatPhaseEffectFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"After this main phase, there is an additional main phase.",
		"There is an additional combat phase.",
		"After this main phase, there is an additional combat phase followed by an additional turn.",
	} {
		document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			continue
		}
		for _, ability := range document.Abilities {
			for _, sentence := range ability.Sentences {
				for _, effect := range sentence.Effects {
					if effect.Kind == EffectAdditionalCombatPhase {
						t.Errorf("Parse(%q) matched extra-phase effect, want fail-closed", source)
					}
				}
			}
		}
	}
}
