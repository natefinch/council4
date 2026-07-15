package parser

import "testing"

// fightEffectFrom parses source and returns the single EffectFight it produces,
// searching every sentence's effects.
func fightEffectFrom(t *testing.T, source string) EffectSyntax {
	t.Helper()
	document, _ := Parse(source, Context{})
	for _, ability := range document.Abilities {
		for _, sentence := range ability.Sentences {
			for _, effect := range sentence.Effects {
				if effect.Kind == EffectFight {
					return effect
				}
			}
		}
	}
	t.Fatalf("no EffectFight parsed from %q", source)
	return EffectSyntax{}
}

// TestCorrelatedDistributiveFightRecognized proves the parser flags the
// distributive one-to-one fight form "Each of those <A> fights a different one of
// those <B>." (Ezuri's Predation) with CorrelatedDistributiveFight, the marker the
// backend uses to pair the created-token group with the counted-permanent group.
func TestCorrelatedDistributiveFightRecognized(t *testing.T) {
	t.Parallel()
	fight := fightEffectFrom(t,
		"For each creature your opponents control, create a 4/4 green Phyrexian Beast creature token. Each of those tokens fights a different one of those creatures.")
	if !fight.CorrelatedDistributiveFight {
		t.Fatal("CorrelatedDistributiveFight = false, want true for the distributive fight clause")
	}
}

// TestCorrelatedDistributiveFightRejectsPlainFight proves a plain single-fight
// clause is not flagged, so it keeps lowering through the ordinary two-object
// fight path rather than the correlated-group path.
func TestCorrelatedDistributiveFightRejectsPlainFight(t *testing.T) {
	t.Parallel()
	fight := fightEffectFrom(t, "Target creature you control fights target creature you don't control.")
	if fight.CorrelatedDistributiveFight {
		t.Fatal("CorrelatedDistributiveFight = true, want false for a plain two-target fight")
	}
}

// TestCorrelatedDistributiveFightRejectsNonDistributiveThose proves the flag
// requires both the distributive "each of those" subject and the "a different one
// of those" object; a "those tokens fight target creature" wording, lacking the
// distributive object, is not flagged.
func TestCorrelatedDistributiveFightRejectsNonDistributiveThose(t *testing.T) {
	t.Parallel()
	document, _ := Parse("Those creatures fight target creature you control.", Context{})
	for _, ability := range document.Abilities {
		for _, sentence := range ability.Sentences {
			for _, effect := range sentence.Effects {
				if effect.Kind == EffectFight && effect.CorrelatedDistributiveFight {
					t.Fatal("CorrelatedDistributiveFight = true, want false without the distributive object")
				}
			}
		}
	}
}
