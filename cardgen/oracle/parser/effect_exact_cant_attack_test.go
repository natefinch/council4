package parser

import "testing"

// singleCombatRequirementEffect parses source and returns its single resolving effect,
// asserting the parser recognized exactly one ability, sentence, and effect of
// the wanted kind.
func singleCombatRequirementEffect(t *testing.T, source string, kind EffectKind) EffectSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 || effects[0].Kind != kind {
		t.Fatalf("Parse(%q) effects = %#v, want one %s", source, effects, kind)
	}
	return effects[0]
}

func TestExactCantAttackThisTurnAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Target creature can't attack this turn.",
		"Target creature an opponent controls can't attack this turn.",
		"Up to two target creatures can't attack this turn.",
	}
	for _, source := range accepted {
		effect := singleCombatRequirementEffect(t, source, EffectCantAttack)
		if !effect.Exact {
			t.Errorf("cantAttackEffect(%q).Exact = false, want true", source)
		}
		if effect.Context != EffectContextTarget {
			t.Errorf("cantAttackEffect(%q).Context = %s, want EffectContextTarget", source, effect.Context)
		}
		if effect.Duration != EffectDurationThisTurn {
			t.Errorf("cantAttackEffect(%q).Duration = %s, want EffectDurationThisTurn", source, effect.Duration)
		}
		if len(effect.Targets) != 1 {
			t.Errorf("cantAttackEffect(%q) targets = %d, want 1", source, len(effect.Targets))
		}
	}
}

func TestExactCantAttackOrBlockThisTurnAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Target creature can't attack or block this turn.",
		"Target creature an opponent controls can't attack or block this turn.",
	}
	for _, source := range accepted {
		effect := singleCombatRequirementEffect(t, source, EffectCantAttackOrBlock)
		if !effect.Exact {
			t.Errorf("cantAttackOrBlockEffect(%q).Exact = false, want true", source)
		}
		if effect.Context != EffectContextTarget {
			t.Errorf("cantAttackOrBlockEffect(%q).Context = %s, want EffectContextTarget", source, effect.Context)
		}
		if effect.Duration != EffectDurationThisTurn {
			t.Errorf("cantAttackOrBlockEffect(%q).Duration = %s, want EffectDurationThisTurn", source, effect.Duration)
		}
		if len(effect.Targets) != 1 {
			t.Errorf("cantAttackOrBlockEffect(%q) targets = %d, want 1", source, len(effect.Targets))
		}
	}
}

func TestExactTargetMustAttackThisTurnAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Target creature attacks this turn if able.",
		"Target creature an opponent controls attacks this turn if able.",
	}
	for _, source := range accepted {
		effect := singleCombatRequirementEffect(t, source, EffectMustAttack)
		if !effect.Exact {
			t.Errorf("targetMustAttackEffect(%q).Exact = false, want true", source)
		}
		if effect.Context != EffectContextTarget {
			t.Errorf("targetMustAttackEffect(%q).Context = %s, want EffectContextTarget", source, effect.Context)
		}
		if len(effect.Targets) != 1 {
			t.Errorf("targetMustAttackEffect(%q) targets = %d, want 1", source, len(effect.Targets))
		}
	}
}

func TestExactCantAttackFamilyFailsClosed(t *testing.T) {
	t.Parallel()
	// Each wording deviates from an exact temporary single-target combat
	// requirement or restriction, so its round-trip must not reach an exact,
	// lowerable production: continuous static prohibitions with no duration, the
	// inverse can't-be-blocked operation, and the directed "attacks <player>"
	// must-attack form whose chosen-player redirection is not yet lowered.
	rejected := map[string]EffectKind{
		"Creatures can't attack.":                                    EffectCantAttack,
		"Target creature can't attack.":                              EffectCantAttack,
		"Target creature can't be blocked this turn.":                EffectCantAttackOrBlock,
		"Target creature attacks target opponent this turn if able.": EffectMustAttack,
		"Target creature can't attack you this turn.":                EffectCantAttack,
	}
	for source, kind := range rejected {
		document, _ := Parse(source, Context{})
		if len(document.Abilities) == 0 {
			continue
		}
		for _, sentence := range document.Abilities[0].Sentences {
			for _, effect := range sentence.Effects {
				if effect.Kind == kind && effect.Exact {
					t.Errorf("Parse(%q) produced an exact %s, want fail closed", source, kind)
				}
			}
		}
	}
}
