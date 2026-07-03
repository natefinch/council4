package parser

import (
	"strings"
	"testing"
)

// TestParseNonControllerNegativeResolvingGate covers the parser recognition of
// the non-controller negative resolving gate — a non-controller optional action
// ("target opponent may sacrifice ...") whose "If they don't, ..." failure
// branch resolves for the controller (Rakdos, Patron of Chaos). The recognizer
// types the gate as ConditionPredicatePriorInstructionNotAccepted, leaves the
// action and consequence as their own effects, and clears the spurious negation
// the trailing "don't" left on the consequence effect. All three player-subject
// spellings ("if they don't", "if the player doesn't", "if that player
// doesn't") name the same branch.
func TestParseNonControllerNegativeResolvingGate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		body string
	}{
		{
			name: "if they don't",
			body: "At the beginning of your end step, target opponent may sacrifice two nonland, nontoken permanents of their choice. If they don't, you draw two cards.",
		},
		{
			name: "if the player doesn't",
			body: "At the beginning of your end step, target opponent may sacrifice a creature of their choice. If the player doesn't, you draw a card.",
		},
		{
			name: "if that player doesn't",
			body: "At the beginning of your end step, target opponent may sacrifice a creature of their choice. If that player doesn't, you draw a card.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.body, Context{CardName: "Rakdos, Patron of Chaos"})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if len(document.Abilities) != 1 {
				t.Fatalf("abilities = %#v", document.Abilities)
			}
			ability := document.Abilities[0]
			clauses := ability.ConditionClauses
			if len(clauses) != 1 || clauses[0].Predicate != ConditionPredicatePriorInstructionNotAccepted {
				t.Fatalf("condition clauses = %#v, want one PriorInstructionNotAccepted", clauses)
			}
			action := ability.Sentences[0].Effects[0]
			if action.Kind != EffectSacrifice || !action.Optional || action.Context != EffectContextTarget {
				t.Fatalf("action effect = %#v, want optional target sacrifice", action)
			}
			consequence := ability.Sentences[1].Effects[0]
			if consequence.Kind != EffectDraw || consequence.Negated {
				t.Fatalf("consequence effect = %#v, want non-negated draw", consequence)
			}
		})
	}
}

// TestParseControllerNegativeGateStaysUnrecognized confirms the recognizer is
// non-controller only: a controller "you may sacrifice ..." offer is the
// affirmative optional-flow family the shared planner owns, so the recognizer
// must not append its non-controller PriorInstructionNotAccepted clause for it.
func TestParseControllerNegativeGateStaysUnrecognized(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"At the beginning of your end step, you may sacrifice a creature. If you don't, you draw a card.",
		Context{CardName: "X"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, clause := range document.Abilities[0].ConditionClauses {
		action := document.Abilities[0].Sentences[0].Effects[0]
		if clause.Predicate == ConditionPredicatePriorInstructionNotAccepted &&
			action.Context != EffectContextController {
			t.Fatalf("controller offer typed by non-controller recognizer: %#v", clause)
		}
	}
}

// TestParseNonControllerOptionalEdictExactness covers the exact reconstruction of
// the offered edict, including the "nonland, nontoken permanents" filter whose
// excluded card type prints before the "nontoken" qualifier (the reverse of the
// single-qualifier order), and the "may sacrifice" verb of the optional offer.
func TestParseNonControllerOptionalEdictExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		body  string
		exact bool
	}{
		{
			name:  "optional nonland nontoken plural edict",
			body:  "Target opponent may sacrifice two nonland, nontoken permanents of their choice.",
			exact: true,
		},
		{
			name:  "optional single creature edict",
			body:  "Target opponent may sacrifice a creature of their choice.",
			exact: true,
		},
		{
			name:  "mandatory nonland nontoken plural edict",
			body:  "Target opponent sacrifices two nonland, nontoken permanents of their choice.",
			exact: true,
		},
		{
			name:  "color disjunction edict",
			body:  "Target opponent sacrifices a green or white creature of their choice.",
			exact: true,
		},
		{
			name:  "single color edict",
			body:  "Target opponent sacrifices a green creature of their choice.",
			exact: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.body, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 || effects[0].Kind != EffectSacrifice {
				t.Fatalf("effects = %#v, want one sacrifice", effects)
			}
			if effects[0].Exact != test.exact {
				t.Fatalf("effect Exact = %v, want %v", effects[0].Exact, test.exact)
			}
			if len(effects[0].Targets) != 1 || !effects[0].Targets[0].Exact ||
				!strings.EqualFold(effects[0].Targets[0].Text, "target opponent") {
				t.Fatalf("target = %#v, want exact \"target opponent\"", effects[0].Targets)
			}
		})
	}
}
