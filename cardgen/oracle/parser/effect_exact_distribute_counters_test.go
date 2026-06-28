package parser

import "testing"

func distributeCountersEffectSyntax(t *testing.T, source string) EffectSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 || effects[0].Kind != EffectPut {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0]
}

// TestDistributeCountersAccepts covers the recognized distribute-counters forms:
// the enumerated "one or two" and "one, two, or three" cardinalities, the
// unbounded "any number of" form, a controller-restricted "you control" target,
// a variable X total, and a non-+1/+1 counter kind. Each round-trips byte-exact
// and sets the DistributeCounters flag the lowering gates on.
func TestDistributeCountersAccepts(t *testing.T) {
	t.Parallel()
	sources := []string{
		"Distribute two +1/+1 counters among one or two target creatures.",
		"Distribute three +1/+1 counters among one, two, or three target creatures.",
		"Distribute four +1/+1 counters among any number of target creatures.",
		"Distribute three +1/+1 counters among one, two, or three target creatures you control.",
		"Distribute X +1/+1 counters among any number of target creatures.",
		"Distribute two -1/-1 counters among one or two target creatures.",
	}
	for _, source := range sources {
		effect := distributeCountersEffectSyntax(t, source)
		if !effect.DistributeCounters {
			t.Errorf("DistributeCounters(%q) = false, want true", source)
		}
		if !effect.Exact {
			t.Errorf("Exact(%q) = false, want true", source)
		}
	}
}

// TestDistributeCountersFailsClosed covers wordings the distribute-counters
// round-trip does not represent: a non-creature target noun, a wider Oxford-comma
// cardinality the cardinality phrase does not model, and a dynamic total. Each
// keeps DistributeCounters unset so the lowering never approximates it.
func TestDistributeCountersFailsClosed(t *testing.T) {
	t.Parallel()
	sources := []string{
		"Distribute three +1/+1 counters among one, two, three, or four target creatures.",
		"Distribute two +1/+1 counters among one or two target artifacts.",
		"Distribute two +1/+1 counters among one or two target creatures an opponent controls.",
	}
	for _, source := range sources {
		document, _ := Parse(source, Context{InstantOrSorcery: true})
		for _, ability := range document.Abilities {
			for _, sentence := range ability.Sentences {
				for _, effect := range sentence.Effects {
					if effect.DistributeCounters {
						t.Errorf("DistributeCounters(%q) = true, want false", source)
					}
				}
			}
		}
	}
}
