package compiler

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
)

func TestCompileTheOneRingTypedSemantics(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"When The One Ring enters, if you cast it, you gain protection from everything until your next turn.\n"+
			"At the beginning of your upkeep, you lose 1 life for each burden counter on The One Ring.\n"+
			"{T}: Put a burden counter on The One Ring, then draw a card for each burden counter on The One Ring.",
		pipelineContext{CardName: "The One Ring"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(compilation.Abilities) != 3 {
		t.Fatalf("abilities = %d, want 3", len(compilation.Abilities))
	}

	enter := compilation.Abilities[0]
	if enter.Trigger == nil || enter.Trigger.Condition == nil ||
		enter.Trigger.Condition.Predicate != ConditionPredicateEventSubjectWasCastByController {
		t.Fatalf("enter condition = %#v, want cast by controller", enter.Trigger)
	}
	var protection *CompiledEffect
	for i := range enter.Content.Effects {
		if enter.Content.Effects[i].Kind == EffectGain {
			protection = &enter.Content.Effects[i]
		}
	}
	if protection == nil || !protection.Exact ||
		protection.Duration != DurationUntilYourNextTurn ||
		len(enter.Content.Keywords) != 1 ||
		!enter.Content.Keywords[0].Protection.Everything {
		t.Fatalf("protection = %#v, want exact protection from everything", protection)
	}

	for _, ability := range compilation.Abilities[1:] {
		for _, effect := range ability.Content.Effects {
			if effect.Amount.DynamicKind == DynamicAmountNone {
				continue
			}
			if effect.Amount.DynamicKind != DynamicAmountSourceCounterCount ||
				effect.Amount.CounterKind != counter.Burden ||
				effect.Amount.ReferenceSpan.Start.Offset == 0 {
				t.Fatalf("amount = %#v, want burden counters on source", effect.Amount)
			}
		}
	}
}
