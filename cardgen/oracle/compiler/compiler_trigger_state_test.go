package compiler

import (
	"slices"
	"testing"
)

func TestCompileStateTriggerControlsNoSubtype(t *testing.T) {
	t.Parallel()
	source := "When you control no Islands, sacrifice this creature."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	ability := compilation.Abilities[0]
	if ability.Trigger == nil || ability.Trigger.Kind != TriggerState {
		t.Fatalf("trigger = %#v, want state trigger", ability.Trigger)
	}
	state := ability.Trigger.Pattern.StateCondition
	if state == nil {
		t.Fatal("state condition = nil")
	}
	if state.Predicate != ConditionPredicateControllerControls {
		t.Fatalf("predicate = %q, want controller controls", state.Predicate)
	}
	// "control no X" lowers to a negated "control at least one X".
	if !state.Negated || state.Threshold != 1 {
		t.Fatalf("negated = %v, threshold = %d, want negated at-least-1", state.Negated, state.Threshold)
	}
	if !slices.Contains(state.Selection.SubtypesAny, "Island") {
		t.Fatalf("selection = %#v, want Island subtype", state.Selection)
	}
	if len(ability.Content.Effects) != 1 || ability.Content.Effects[0].Kind != EffectSacrifice {
		t.Fatalf("effects = %#v, want single sacrifice", ability.Content.Effects)
	}
}

func TestCompileStateTriggerControlsNoOtherCreatures(t *testing.T) {
	t.Parallel()
	source := "When you control no other creatures, sacrifice this creature."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	ability := compilation.Abilities[0]
	if ability.Trigger == nil || ability.Trigger.Kind != TriggerState {
		t.Fatalf("trigger = %#v, want state trigger", ability.Trigger)
	}
	state := ability.Trigger.Pattern.StateCondition
	if state == nil || !state.Negated || state.Threshold != 1 {
		t.Fatalf("state condition = %#v, want negated at-least-1", state)
	}
	if !state.Selection.ExcludeSource {
		t.Fatalf("selection = %#v, want ExcludeSource for \"other\"", state.Selection)
	}
}
