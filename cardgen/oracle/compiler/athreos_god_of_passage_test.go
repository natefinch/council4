package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestCompileTargetOpponentPayLifeUnlessEventCardReturn(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Whenever another creature you own dies, return it to your hand unless target opponent pays 3 life.",
		pipelineContext{CardName: "Athreos, God of Passage"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if len(ability.Content.Effects) != 1 ||
		len(ability.Content.Targets) != 1 ||
		len(ability.Content.Conditions) != 1 {
		t.Fatalf("compiled content = %#v", ability.Content)
	}
	effect := ability.Content.Effects[0]
	if effect.Kind != EffectReturn ||
		effect.Payment.Form != parser.EffectPaymentFormUnless ||
		effect.Payment.Payer != parser.EffectPaymentPayerTargetPlayer ||
		effect.Payment.AdditionalCost == nil {
		t.Fatalf("compiled effect = %#v", effect)
	}
	if !effect.Exact {
		t.Fatalf("compiled return is not exact: targets=%#v payment=%#v refs=%#v", effect.Targets, effect.Payment, effect.References)
	}
	if ability.Content.Targets[0].Selector.Kind != SelectorOpponent {
		t.Fatalf("compiled target = %#v", ability.Content.Targets[0])
	}
	if ability.Content.Conditions[0].Predicate != ConditionPredicateTargetControllerDoesNotPay {
		t.Fatalf("compiled condition = %#v", ability.Content.Conditions[0])
	}
}
