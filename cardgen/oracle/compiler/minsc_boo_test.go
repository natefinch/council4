package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

const minscBooOracle = `When Minsc & Boo enters and at the beginning of your upkeep, you may create Boo, a legendary 1/1 red Hamster creature token with trample and haste.
+1: Put three +1/+1 counters on up to one target creature with trample or haste.
−2: Sacrifice a creature. When you do, Minsc & Boo deals X damage to any target, where X is that creature's power. If the sacrificed creature was a Hamster, draw X cards.
Minsc & Boo, Timeless Heroes can be your commander.`

func TestCompileMinscBooTypedBindings(t *testing.T) {
	t.Parallel()
	document, parseDiagnostics := parser.Parse(minscBooOracle, parser.Context{
		CardName:     "Minsc & Boo, Timeless Heroes",
		Legendary:    true,
		Planeswalker: true,
	})
	if len(parseDiagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", parseDiagnostics)
	}
	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(compilation.Abilities) != 5 || !compilation.Abilities[4].CanBeCommander {
		t.Fatalf("compiled abilities = %#v", compilation.Abilities)
	}
	plusTarget := compilation.Abilities[2].Content.Targets[0]
	if !plusTarget.Exact || len(plusTarget.Selector.Alternatives) != 2 {
		t.Fatalf("+1 target = %#v", plusTarget)
	}
	minus := compilation.Abilities[3].Content
	nodeID := minus.Effects[1].Amount.ReferenceNodeID
	found := false
	for i := range minus.References {
		reference := minus.References[i]
		if reference.NodeID == nodeID {
			found = reference.Binding == ReferenceBindingPriorInstructionResult &&
				reference.PriorInstruction == 0
		}
	}
	if !found {
		t.Fatalf("damage amount is not bound to the sacrifice: %#v", minus.References)
	}
	if !minus.Effects[2].Amount.VariableX ||
		minus.Effects[2].Amount.DynamicKind != DynamicAmountNone {
		t.Fatalf("draw amount = %#v, want shared bare X", minus.Effects[2].Amount)
	}
	if len(minus.Conditions) != 2 ||
		!minus.Conditions[0].Reflexive ||
		minus.Conditions[1].Predicate != ConditionPredicateObjectMatches ||
		minus.Conditions[1].ObjectBinding != ReferenceBindingPriorInstructionResult {
		t.Fatalf("-2 conditions = %#v", minus.Conditions)
	}
}
