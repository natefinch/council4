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

// TestParseAdditionalBeginningPhaseEffect proves the parser recognizes the
// extra-beginning-phase insertion wording (Sphinx of the Second Sun, Cyclonus,
// Cybertronian Fighter) and records it as an additional beginning phase rather
// than a combat phase. It also proves a leading "If you do," condition clause is
// stripped, the reflexive form Cyclonus prints on its back face.
func TestParseAdditionalBeginningPhaseEffect(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"There is an additional beginning phase after this phase.",
		"If you do, there is an additional beginning phase after this phase.",
	} {
		document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
		}
		effects := document.Abilities[0].Sentences[0].Effects
		if len(effects) != 1 {
			t.Fatalf("Parse(%q) effects = %#v, want one", source, effects)
		}
		effect := effects[0]
		if effect.Kind != EffectAdditionalCombatPhase {
			t.Errorf("Parse(%q) kind = %v, want EffectAdditionalCombatPhase", source, effect.Kind)
		}
		if !effect.Exact {
			t.Errorf("Parse(%q) Exact = false, want true", source)
		}
		if !effect.AdditionalBeginningPhase {
			t.Errorf("Parse(%q) AdditionalBeginningPhase = false, want true", source)
		}
		if effect.AdditionalCombatPhase {
			t.Errorf("Parse(%q) AdditionalCombatPhase = true, want false", source)
		}
		if effect.AdditionalMainPhase {
			t.Errorf("Parse(%q) AdditionalMainPhase = true, want false", source)
		}
	}
}

// TestParseAdditionalBeginningPhaseEffectFailsClosed proves wordings that only
// exist for combat phases (leading "after this phase" order, "followed by an
// additional main phase" tail) do not match the beginning-phase recognizer.
func TestParseAdditionalBeginningPhaseEffectFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"After this phase, there is an additional beginning phase.",
		"There is an additional beginning phase after this phase followed by an additional main phase.",
	} {
		document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			continue
		}
		for _, ability := range document.Abilities {
			for _, sentence := range ability.Sentences {
				for _, effect := range sentence.Effects {
					if effect.Kind == EffectAdditionalCombatPhase && effect.AdditionalBeginningPhase {
						t.Errorf("Parse(%q) matched extra-beginning-phase effect, want fail-closed", source)
					}
				}
			}
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
