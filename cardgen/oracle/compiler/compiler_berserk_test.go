package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestCompileBerserkSemanticsPositionBlind(t *testing.T) {
	t.Parallel()
	document, parseDiagnostics := parser.Parse(
		"Cast this spell only before the combat damage step.\n"+
			"Target creature gains trample and gets +X/+0 until end of turn, where X is its power. "+
			"At the beginning of the next end step, destroy that creature if it attacked this turn.",
		parser.Context{InstantOrSorcery: true},
	)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", parseDiagnostics)
	}
	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compiler diagnostics = %#v", diagnostics)
	}
	if !compilation.Abilities[0].CastOnlyBeforeCombatDamageStep {
		t.Fatal("compiled cast restriction is absent")
	}
	var found bool
	for _, condition := range compilation.Abilities[1].Content.Conditions {
		if condition.Predicate == ConditionPredicateObjectAttackedThisTurn {
			found = true
			if condition.ObjectBinding != ReferenceBindingTarget {
				t.Fatalf("object binding = %v, want target", condition.ObjectBinding)
			}
		}
	}
	if !found {
		t.Fatalf("conditions = %#v, want object-attacked-this-turn", compilation.Abilities[1].Content.Conditions)
	}
}
