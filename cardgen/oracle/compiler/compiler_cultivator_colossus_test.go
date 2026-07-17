package compiler

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestCompileCultivatorColossusTypedProcess(t *testing.T) {
	t.Parallel()
	oracle := "Trample\n" +
		"Cultivator Colossus's power and toughness are each equal to the number of lands you control.\n" +
		"When this creature enters, you may put a land card from your hand onto the battlefield tapped. If you do, draw a card and repeat this process."
	compilation, diagnostics := compileSource(oracle, pipelineContext{CardName: "Cultivator Colossus"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	var repeat *CompiledEffect
	for i := range compilation.Abilities {
		for j := range compilation.Abilities[i].Content.Effects {
			effect := &compilation.Abilities[i].Content.Effects[j]
			if effect.Kind == EffectRepeatProcess {
				repeat = effect
			}
		}
	}
	if repeat == nil || !repeat.Exact || !repeat.RepeatUntilFailure || len(repeat.RepeatBody) != 2 {
		t.Fatalf("repeat = %#v", repeat)
	}
	if repeat.RepeatBody[0].Kind != EffectPut || !repeat.RepeatBody[0].Optional ||
		repeat.RepeatBody[1].Kind != EffectDraw {
		t.Fatalf("body = %#v", repeat.RepeatBody)
	}
	foundCondition := false
	for i := range compilation.Abilities {
		for _, condition := range compilation.Abilities[i].Content.Conditions {
			if condition.Predicate == ConditionPredicatePriorInstructionAccepted {
				foundCondition = true
			}
		}
	}
	if !foundCondition {
		t.Fatal("compiled process lost its prior-instruction success condition")
	}
	if repeat.RepeatBody[0].Order.End == 0 || repeat.RepeatBody[1].Order.End == 0 {
		t.Fatalf("repeat body source order was not preserved: %#v", repeat.RepeatBody)
	}
	foundCDA := false
	for i := range compilation.Abilities {
		if compilation.Abilities[i].Static == nil {
			continue
		}
		for _, declaration := range compilation.Abilities[i].Static.Declarations {
			if declaration.CharacteristicPT != nil {
				foundCDA = declaration.CharacteristicPT.Value == game.DynamicValueControllerLandCount &&
					declaration.CharacteristicPT.SetsPower &&
					declaration.CharacteristicPT.SetsToughness
			}
		}
	}
	if !foundCDA {
		t.Fatal("compiled static declarations lost the controller-land-count CDA")
	}
}
