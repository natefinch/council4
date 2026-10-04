package compiler

import (
	"testing"
)

func TestCompileCounterTaxConditionOwnership(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Counter target spell unless its controller pays {3}. You gain 2 life.",
		"If you control a creature, counter target spell unless its controller pays {3}. You gain 2 life.",
		"Counter target spell unless its controller pays {1} for each card in your graveyard. You gain 2 life.",
		"Whenever an opponent casts a spell, counter that spell unless its controller pays {X}. You gain 2 life.",
		"Choose one —\n• Counter target spell unless its controller pays {3}. You gain 2 life.\n• Draw a card.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(text, pipelineContext{CardName: "Tax Probe", InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("compile = %#v", diagnostics)
			}
			content := compilation.Abilities[0].Content
			if len(content.Modes) != 0 {
				content = content.Modes[0].Content
			}
			owner := content.Effects[0].Payment.FailureConditionNodeID
			matches := 0
			for _, condition := range content.Conditions {
				if condition.Predicate == ConditionPredicateTargetControllerDoesNotPay {
					matches++
					if condition.NodeID != owner {
						t.Fatalf("condition node = %d, payment owner = %d", condition.NodeID, owner)
					}
				}
			}
			if matches != 1 {
				t.Fatalf("tax conditions = %d, want one", matches)
			}
		})
	}
}
