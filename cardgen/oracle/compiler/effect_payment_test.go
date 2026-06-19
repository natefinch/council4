package compiler

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/cost"
)

func TestCompileEventPlayerResolutionPayment(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Whenever an opponent casts a spell, you may draw a card unless that player pays {2}.",
		pipelineContext{CardName: "Tax Study"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	effect := ability.Content.Effects[0]
	if effect.Payment.Payer != parser.EffectPaymentPayerEventPlayer ||
		!slices.Equal(effect.Payment.ManaCost, cost.Mana{cost.O(2)}) {
		t.Fatalf("payment = %#v", effect.Payment)
	}
	if len(ability.Content.Conditions) != 1 ||
		ability.Content.Conditions[0].Predicate != ConditionPredicateEventPlayerDoesNotPay {
		t.Fatalf("conditions = %#v", ability.Content.Conditions)
	}
	if len(ability.Content.References) != 1 ||
		ability.Content.References[0].Binding != ReferenceBindingEventPlayer {
		t.Fatalf("references = %#v", ability.Content.References)
	}
}
