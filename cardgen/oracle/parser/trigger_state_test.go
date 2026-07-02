package parser

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

// parseStateTriggerFromSource parses a single-ability card and returns the
// recognized state-trigger clause, failing if parsing produced diagnostics or a
// non state-trigger ability.
func parseStateTriggerFromSource(t *testing.T, source, cardName string) *StateTriggerClause {
	t.Helper()
	document, diagnostics := Parse(source, Context{CardName: cardName})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %d", len(document.Abilities))
	}
	if document.Abilities[0].Trigger == nil {
		t.Fatal("trigger = nil")
	}
	return document.Abilities[0].Trigger.State
}

func conditionSelectionHasSubtype(selection ConditionSelection, sub types.Sub) bool {
	return slices.Contains(selection.SubtypesAny, sub)
}

func TestStateTriggerRecognizesControllerControlsNoSubtype(t *testing.T) {
	t.Parallel()
	state := parseStateTriggerFromSource(t, "When you control no Islands, sacrifice this creature.", "Sea Serpent")
	if state == nil {
		t.Fatal("state trigger clause = nil")
	}
	cond := state.Condition
	if cond.Predicate != ConditionPredicateControls {
		t.Fatalf("predicate = %q, want %q", cond.Predicate, ConditionPredicateControls)
	}
	if cond.Scope != ConditionControlScopeController {
		t.Fatalf("scope = %q, want controller", cond.Scope)
	}
	if cond.Comparison != ConditionComparisonAtMost || cond.CompareValue != 0 {
		t.Fatalf("comparison = %q, compareValue = %d, want at-most 0", cond.Comparison, cond.CompareValue)
	}
	if !conditionSelectionHasSubtype(cond.Selection, types.Island) {
		t.Fatalf("selection = %#v, want Island subtype", cond.Selection)
	}
}

func TestStateTriggerRecognizesOtherSubtypesAndCardName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		source  string
		card    string
		subtype types.Sub
	}{
		{"swamps", "When you control no Swamps, sacrifice this creature.", "Barbarian Outcast", types.Swamp},
		{"forests", "When you control no Forests, sacrifice this creature.", "Gorilla Pack", types.Forest},
		{"card-name sacrifice", "When you control no Islands, sacrifice Skeleton Ship.", "Skeleton Ship", types.Island},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			state := parseStateTriggerFromSource(t, test.source, test.card)
			if state == nil {
				t.Fatal("state trigger clause = nil")
			}
			if !conditionSelectionHasSubtype(state.Condition.Selection, test.subtype) {
				t.Fatalf("selection = %#v, want %q", state.Condition.Selection, test.subtype)
			}
		})
	}
}

// TestStateTriggerFailsClosedOnNonZeroThreshold confirms the recognizer only
// claims the "control no X" shape; a positive threshold ("control a X") must not
// be miscompiled as a state trigger.
func TestStateTriggerFailsClosedOnNonZeroThreshold(t *testing.T) {
	t.Parallel()
	document, _ := Parse("When you control an Island, sacrifice this creature.", Context{CardName: "Probe"})
	if len(document.Abilities) == 1 && document.Abilities[0].Trigger != nil && document.Abilities[0].Trigger.State != nil {
		t.Fatal("recognizer claimed a positive-threshold controls clause as a state trigger")
	}
}
