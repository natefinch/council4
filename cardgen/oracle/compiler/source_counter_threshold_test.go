package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/counter"
)

// TestCompileSourceCounterThresholdRemoveThoseCounters confirms the compiler
// carries Prize Pig's "Whenever you gain life, put that many ribbon counters on
// this creature. Then if there are three or more ribbon counters on this
// creature, remove those counters and untap it." through to a compiled ability
// whose tail removal keeps the RemoveThoseCounters marker and whose gate condition
// compiles to a source object-match requiring three or more ribbon counters.
func TestCompileSourceCounterThresholdRemoveThoseCounters(t *testing.T) {
	t.Parallel()
	document, diagnostics := parser.Parse(
		"Whenever you gain life, put that many ribbon counters on this creature. Then if there are three or more ribbon counters on this creature, remove those counters and untap it.",
		parser.Context{CardName: "Prize Pig"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", diagnostics)
	}
	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]

	var remove *CompiledEffect
	for i := range ability.Content.Effects {
		if ability.Content.Effects[i].Kind == EffectRemoveCounter {
			remove = &ability.Content.Effects[i]
			break
		}
	}
	if remove == nil {
		t.Fatalf("no remove-counter effect; effects = %#v", ability.Content.Effects)
	}
	if !remove.RemoveThoseCounters {
		t.Fatal("remove effect RemoveThoseCounters = false, want true")
	}

	if len(ability.Content.Conditions) != 1 {
		t.Fatalf("conditions = %d, want 1", len(ability.Content.Conditions))
	}
	condition := ability.Content.Conditions[0]
	if condition.Predicate != ConditionPredicateObjectMatches {
		t.Fatalf("condition predicate = %v, want object-matches", condition.Predicate)
	}
	if condition.ObjectBinding != ReferenceBindingSource {
		t.Fatalf("condition object binding = %v, want source", condition.ObjectBinding)
	}
	if !condition.Selection.CounterKindKnown ||
		condition.Selection.CounterKind != counter.Ribbon ||
		condition.Selection.CounterCountAtLeast != 3 {
		t.Fatalf("condition selection = %#v, want ribbon count >= 3", condition.Selection)
	}
}
