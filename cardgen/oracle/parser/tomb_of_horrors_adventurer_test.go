package parser

import "testing"

func TestParseConditionalMultiCopyWithPluralRetargetRider(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Whenever you cast your second spell each turn, copy it. "+
			"If you've completed a dungeon, copy that spell twice instead. "+
			"You may choose new targets for the copies.",
		Context{CardName: "Tomb of Horrors Adventurer"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	var effects []EffectSyntax
	for i := range ability.Sentences {
		effects = append(effects, ability.Sentences[i].Effects...)
	}
	if len(effects) != 2 {
		t.Fatalf("effects = %#v", effects)
	}
	base := effects[0]
	if base.Kind != EffectCopyStackObject ||
		!base.Exact ||
		base.Amount.Known ||
		!base.CopyMayChooseNewTargets {
		t.Fatalf("base copy = %#v", base)
	}
	escalated := effects[1]
	if escalated.Kind != EffectCopyStackObject ||
		!escalated.Exact ||
		!escalated.Amount.Known ||
		escalated.Amount.Value != 2 ||
		escalated.Replacement.Kind != EffectReplacementInstead ||
		!escalated.CopyMayChooseNewTargets {
		t.Fatalf("escalated copy = %#v", escalated)
	}
	if len(ability.ConditionClauses) != 1 ||
		ability.ConditionClauses[0].Predicate != ConditionPredicateControllerCompletedADungeon {
		t.Fatalf("conditions = %#v", ability.ConditionClauses)
	}
}
