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

func TestExactStunCounterPlacementAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Put a stun counter on target creature.",
		"Put a stun counter on target creature an opponent controls.",
		"Put two stun counters on target creature.",
	}
	for _, source := range accepted {
		if !counterPlacementExact(t, source) {
			t.Errorf("counterPlacementExact(%q) = false, want true", source)
		}
	}
}

func TestExactFinalityCounterPlacementFailsClosed(t *testing.T) {
	t.Parallel()
	// Finality counters have no complete runtime semantics, so their placement
	// clause stays inexact and unlowerable.
	rejected := []string{
		"Put a finality counter on target creature.",
		"Put two finality counters on target creature.",
	}
	for _, source := range rejected {
		if counterPlacementExact(t, source) {
			t.Errorf("counterPlacementExact(%q) = true, want false", source)
		}
	}
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

// counterPlacementEffect parses a single counter-placement sentence and returns
// its resolving effect for recipient-shape assertions.
func counterPlacementEffect(t *testing.T, source string) EffectSyntax {
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

// TestExactAttachedCounterPlacementAccepts covers the Aura recipient "enchanted
// creature": the counter is placed on the permanent the source is attached to.
func TestExactAttachedCounterPlacementAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Put a +1/+1 counter on enchanted creature.",
		"Put two -1/-1 counters on enchanted creature.",
		"Put six +1/+1 counters on enchanted creature.",
	}
	for _, source := range accepted {
		effect := counterPlacementEffect(t, source)
		if !effect.CounterRecipientAttached {
			t.Errorf("CounterRecipientAttached(%q) = false, want true", source)
		}
		if !effect.Exact {
			t.Errorf("counterPlacementExact(%q) = false, want true", source)
		}
	}
}

// TestExactAttachedCounterPlacementFailsClosed keeps recipients that are not the
// bare "enchanted creature" out of the attached-recipient form.
func TestExactAttachedCounterPlacementFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		// A trailing selector qualifier is not the bare recipient.
		"Put a +1/+1 counter on enchanted creature with flying.",
		// "enchanted permanent" is a different recipient the runtime is not asked
		// to model here.
		"Put a +1/+1 counter on enchanted permanent.",
	}
	for _, source := range rejected {
		effect := counterPlacementEffect(t, source)
		if effect.CounterRecipientAttached && effect.Exact {
			t.Errorf("counterPlacement(%q) accepted as exact attached recipient, want fail-closed", source)
		}
	}
}
