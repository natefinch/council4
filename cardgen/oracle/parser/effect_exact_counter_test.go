package parser

import "testing"

// counterPlacementExact parses a single counter-placement sentence and reports
// whether its resolving effect round-tripped to an exact, lowerable production.
func counterPlacementExact(t *testing.T, source string) bool {
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
	return effects[0].Exact
}

func TestExactMultiTargetCounterPlacementAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Put a +1/+1 counter on each of up to two target creatures.",
		"Put a +1/+1 counter on each of up to three target creatures.",
		"Put a +1/+1 counter on each of up to two target creatures you control.",
		"Put a +1/+1 counter on each of up to two other target creatures.",
		"Put a +1/+1 counter on each of up to two other target creatures you control.",
		"Put a -1/-1 counter on each of up to two target creatures.",
	}
	for _, source := range accepted {
		if !counterPlacementExact(t, source) {
			t.Errorf("counterPlacementExact(%q) = false, want true", source)
		}
	}
}

func TestExactMultiTargetCounterPlacementFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		// The "each of" coordinator is required for the multi-target object.
		"Put a +1/+1 counter on up to two target creatures.",
		// Subtype-restricted plural targets are not a plain permanent noun.
		"Put a +1/+1 counter on each of up to two target Merfolk.",
		// An unbounded cardinality has no exact count word.
		"Put a +1/+1 counter on each of any number of target creatures.",
	}
	for _, source := range rejected {
		if counterPlacementExact(t, source) {
			t.Errorf("counterPlacementExact(%q) = true, want false", source)
		}
	}
}
