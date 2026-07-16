package parser

import "testing"

const berserkOracleText = "Cast this spell only before the combat damage step.\n" +
	"Target creature gains trample and gets +X/+0 until end of turn, where X is its power. " +
	"At the beginning of the next end step, destroy that creature if it attacked this turn."

func TestParseBerserkCastRestrictionAndDelayedCondition(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(berserkOracleText, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 2 {
		t.Fatalf("abilities = %d, want 2: %#v", len(document.Abilities), document.Abilities)
	}
	if !document.Abilities[0].CastOnlyBeforeCombatDamageStep {
		t.Fatal("before-combat-damage cast restriction was not recognized")
	}
	body := document.Abilities[1]
	if len(body.ConditionClauses) != 1 {
		t.Fatalf("condition clauses = %#v, want one attacked-this-turn condition", body.ConditionClauses)
	}
	condition := body.ConditionClauses[0]
	if condition.Predicate != ConditionPredicateObjectAttackedThisTurn ||
		condition.ObjectBinding != ConditionObjectBindingTarget {
		t.Fatalf("condition = %#v, want target object attacked this turn", condition)
	}
}

func TestParseBeforeCombatDamageCastRestrictionFailsClosed(t *testing.T) {
	t.Parallel()
	document, _ := Parse(
		"Cast this spell only before a combat damage step.",
		Context{InstantOrSorcery: true},
	)
	if document.Abilities[0].CastOnlyBeforeCombatDamageStep {
		t.Fatal("near-match cast restriction was recognized")
	}
}
