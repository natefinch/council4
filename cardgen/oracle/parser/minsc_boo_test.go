package parser

import "testing"

const minscBooOracle = `When Minsc & Boo enters and at the beginning of your upkeep, you may create Boo, a legendary 1/1 red Hamster creature token with trample and haste.
+1: Put three +1/+1 counters on up to one target creature with trample or haste.
−2: Sacrifice a creature. When you do, Minsc & Boo deals X damage to any target, where X is that creature's power. If the sacrificed creature was a Hamster, draw X cards.
Minsc & Boo, Timeless Heroes can be your commander.`

func TestParseMinscBooComposableMechanics(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(minscBooOracle, Context{
		CardName:     "Minsc & Boo, Timeless Heroes",
		Legendary:    true,
		Planeswalker: true,
	})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 5 {
		t.Fatalf("abilities = %d, want 5 after ETB/upkeep split", len(document.Abilities))
	}
	if !document.Abilities[4].CanBeCommander {
		t.Fatal("commander permission not recognized")
	}
	for _, index := range []int{0, 1} {
		if document.Abilities[index].Trigger == nil {
			t.Fatalf("ability %d is not a trigger", index)
		}
	}
	plus := document.Abilities[2]
	target := plus.Sentences[0].Targets[0]
	if !target.Exact || target.Cardinality.Min != 0 || target.Cardinality.Max != 1 ||
		target.Selection.Kind != SelectionCreature ||
		len(target.Selection.Alternatives) != 2 ||
		target.Selection.Alternatives[0].Keyword != KeywordTrample ||
		target.Selection.Alternatives[1].Keyword != KeywordHaste {
		t.Fatalf("+1 target = %#v", target)
	}
	minus := document.Abilities[3]
	if len(minus.ConditionClauses) != 2 ||
		!minus.ConditionClauses[0].Reflexive ||
		minus.ConditionClauses[1].Predicate != ConditionPredicateObjectMatches ||
		minus.ConditionClauses[1].ObjectBinding != ConditionObjectBindingPriorInstructionResult {
		t.Fatalf("-2 conditions = %#v", minus.ConditionClauses)
	}
}
