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
	if !effect.Exact {
		t.Fatalf("effect not exact: %#v", effect)
	}
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

func TestCompileDynamicEventPlayerResolutionPayment(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Whenever an opponent casts their first noncreature spell each turn, draw a card unless that player pays {X}, where X is this creature's power.",
		pipelineContext{CardName: "Esper Sentinel"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Trigger.Pattern.PlayerEventOrdinalThisTurn != 1 ||
		len(ability.Trigger.Pattern.CardSelection.ExcludedTypes) != 1 {
		t.Fatalf("trigger pattern = %#v", ability.Trigger.Pattern)
	}
	effect := ability.Content.Effects[0]
	amount := effect.Payment.GenericManaAmount
	if !effect.Exact {
		t.Fatalf("dynamic payment effect not exact: %#v", effect)
	}
	if effect.Payment.Payer != parser.EffectPaymentPayerEventPlayer ||
		!slices.Equal(effect.Payment.ManaCost, cost.Mana{cost.X}) ||
		amount.DynamicKind != DynamicAmountSourcePower ||
		amount.DynamicForm != DynamicAmountWhereX ||
		amount.Multiplier != 1 {
		t.Fatalf("payment = %#v", effect.Payment)
	}
	if len(ability.Content.References) != 3 {
		t.Fatalf("references = %#v", ability.Content.References)
	}
	if len(ability.Content.Conditions) != 1 ||
		ability.Content.Conditions[0].Predicate != ConditionPredicateEventPlayerDoesNotPay {
		t.Fatalf("conditions = %#v; payment = %#v", ability.Content.Conditions, effect.Payment)
	}
}

func TestCompileEventPlayerMayPayFailureConsequence(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Whenever an opponent draws a card, that player may pay {2}. If the player doesn't, you create a Treasure token.",
		pipelineContext{CardName: "Smothering Tithe"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	effect := ability.Content.Effects[0]
	if effect.Payment.Form != parser.EffectPaymentFormMayPayThenIfDoesNot ||
		effect.Payment.Payer != parser.EffectPaymentPayerEventPlayer ||
		!slices.Equal(effect.Payment.ManaCost, cost.Mana{cost.O(2)}) {
		t.Fatalf("payment = %#v", effect.Payment)
	}
	if effect.Optional || len(ability.Content.Conditions) != 1 {
		t.Fatalf("content = %#v", ability.Content)
	}
	condition := ability.Content.Conditions[0]
	if condition.Kind != ConditionIf ||
		condition.Predicate != ConditionPredicateEventPlayerDoesNotPay ||
		condition.NodeID != effect.Payment.FailureConditionNodeID {
		t.Fatalf("condition = %#v", condition)
	}
}
