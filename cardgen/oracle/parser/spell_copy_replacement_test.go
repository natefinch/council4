package parser

import "testing"

func TestParseAdditionalSpellCopyReplacement(t *testing.T) {
	t.Parallel()
	source := "If you would copy a spell one or more times, instead copy it that many times plus an additional time. You may choose new targets for the additional copy."
	document, diagnostics := Parse(source, Context{CardName: "Twinning Staff"})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	ability := document.Abilities[0]
	if ability.Kind != AbilityReplacement ||
		len(ability.Sentences) != 2 ||
		len(ability.Sentences[0].Effects) != 1 ||
		len(ability.Sentences[1].Effects) != 0 {
		t.Fatalf("ability = %#v", ability)
	}
	if len(ability.ConditionClauses) != 1 ||
		ability.ConditionClauses[0].Predicate != ConditionPredicateSpellCopyUnderController {
		t.Fatalf("conditions = %#v", ability.ConditionClauses)
	}
	effect := ability.Sentences[0].Effects[0]
	if !effect.Exact ||
		effect.Kind != EffectCopyStackObject ||
		effect.Replacement.Kind != EffectReplacementPlusAdditional ||
		effect.Replacement.Amount != 1 ||
		!effect.CopyMayChooseNewTargets {
		t.Fatalf("copy effect = %#v", effect)
	}
}
