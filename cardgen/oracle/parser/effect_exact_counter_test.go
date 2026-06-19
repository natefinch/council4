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

// TestExactCounterPlacementControllerKeywordOrderingAccepts covers single-target
// recipients whose controller clause precedes a "with"/"without" keyword or a
// numeric "with power/toughness" qualifier, matching the canonical Oracle word
// order ("target creature you control without flying").
func TestExactCounterPlacementControllerKeywordOrderingAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Put a +1/+1 counter on target creature you control without flying.",
		"Put a +1/+1 counter on target creature you control with flying.",
		"Put a +1/+1 counter on target creature you don't control with flying.",
		"Put a +1/+1 counter on target creature an opponent controls with flying.",
		"Put a +1/+1 counter on target creature you control with power 2.",
		"Put a +1/+1 counter on another target creature you control without flying.",
	}
	for _, source := range accepted {
		if !counterPlacementExact(t, source) {
			t.Errorf("counterPlacementExact(%q) = false, want true", source)
		}
	}
}

// TestExactCounterPlacementGroupControllerKeywordOrderingAccepts covers group
// recipients whose controller clause precedes a keyword qualifier ("each
// creature you control with flying"), the dominant Oracle ordering.
func TestExactCounterPlacementGroupControllerKeywordOrderingAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Put a +1/+1 counter on each creature you control with flying.",
		"Put a +1/+1 counter on each creature you control without flying.",
		"Put a +1/+1 counter on each creature you control with menace.",
	}
	for _, source := range accepted {
		if !counterPlacementExact(t, source) {
			t.Errorf("counterPlacementExact(%q) = false, want true", source)
		}
	}
}
