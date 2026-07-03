package parser

import "testing"

// TestParseNonControllerMayHaveActionGate covers the parser recognition of the
// non-controller "may have" causative gate — a non-controller player decides
// ("target opponent may have you create ...", "defending player may have you draw
// ...") whether a caused action happens, with a resolving consequence gated on
// that decision. The recognizer retags the caused action from the sentence
// chooser to its true actor (the controller for "may have you ...") and, for the
// negative branch ("If they don't, ..."), appends the
// ConditionPredicatePriorInstructionNotAccepted clause while clearing the spurious
// negation the trailing "don't" left on the consequence.
func TestParseNonControllerMayHaveActionGate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		body           string
		cardName       string
		haveContext    EffectContextKind
		actionKind     EffectKind
		actionContext  EffectContextKind
		wantPredicate  ConditionPredicateKind
		consequenceNeg bool
	}{
		{
			name:          "negative target-opponent create",
			body:          "When this creature enters, target opponent may have you create two Lander tokens. If they don't, put two +1/+1 counters on this creature.",
			cardName:      "Terrapact Intimidator",
			haveContext:   EffectContextTarget,
			actionKind:    EffectCreate,
			actionContext: EffectContextController,
			wantPredicate: ConditionPredicatePriorInstructionNotAccepted,
		},
		{
			name:          "affirmative defending-player draw",
			body:          "Whenever this creature attacks, defending player may have you draw a card. If they do, put a +1/+1 counter on this creature.",
			cardName:      "X",
			haveContext:   EffectContextDefendingPlayer,
			actionKind:    EffectDraw,
			actionContext: EffectContextController,
			wantPredicate: ConditionPredicatePriorInstructionAccepted,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.body, Context{CardName: test.cardName})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if len(document.Abilities) != 1 {
				t.Fatalf("abilities = %#v", document.Abilities)
			}
			ability := document.Abilities[0]
			if len(ability.Sentences[0].Effects) != 2 {
				t.Fatalf("action sentence effects = %#v, want [have, action]", ability.Sentences[0].Effects)
			}
			have := ability.Sentences[0].Effects[0]
			if have.Kind != EffectGrantKeyword || !have.Optional || have.Context != test.haveContext {
				t.Fatalf("have effect = %#v, want optional grant with context %v", have, test.haveContext)
			}
			action := ability.Sentences[0].Effects[1]
			if action.Kind != test.actionKind || action.Context != test.actionContext {
				t.Fatalf("action effect = %#v, want kind %v context %v", action, test.actionKind, test.actionContext)
			}
			clauses := ability.ConditionClauses
			if len(clauses) != 1 || clauses[0].Predicate != test.wantPredicate {
				t.Fatalf("condition clauses = %#v, want one %v", clauses, test.wantPredicate)
			}
			consequence := ability.Sentences[1].Effects[0]
			if consequence.Negated != test.consequenceNeg {
				t.Fatalf("consequence negated = %v, want %v", consequence.Negated, test.consequenceNeg)
			}
		})
	}
}

// TestParseControllerMayHaveGateStaysNatural confirms the recognizer leaves the
// controller "may have" gate untouched: "you may have target player lose 3 life.
// If you do, ..." (Ob Nixilis, the Fallen) already types with a controller-context
// "have" grant and a naturally recognized PriorInstructionAccepted link, so the
// recognizer must not retag the caused action (whose actor is the target player)
// or re-derive the affirmative clause.
func TestParseControllerMayHaveGateStaysNatural(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Whenever a land you control enters, you may have target player lose 3 life. If you do, put three +1/+1 counters on this creature.",
		Context{CardName: "Ob Nixilis, the Fallen"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	have := ability.Sentences[0].Effects[0]
	if have.Kind != EffectGrantKeyword || !have.Optional || have.Context != EffectContextController {
		t.Fatalf("have effect = %#v, want optional grant with controller context", have)
	}
	action := ability.Sentences[0].Effects[1]
	if action.Kind != EffectLose || action.Context != EffectContextTarget {
		t.Fatalf("action effect = %#v, want target-context life loss", action)
	}
	clauses := ability.ConditionClauses
	if len(clauses) != 1 || clauses[0].Predicate != ConditionPredicatePriorInstructionAccepted {
		t.Fatalf("condition clauses = %#v, want one PriorInstructionAccepted", clauses)
	}
}
