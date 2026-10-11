package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

func TestLibraryCardCharacteristicPlannerRequiresTypedExactProducer(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*compiler.CompiledEffect, *compiler.CompiledEffect, *game.Instruction, *[][2]int)
		valid  bool
	}{
		{"owned power", func(*compiler.CompiledEffect, *compiler.CompiledEffect, *game.Instruction, *[][2]int) {}, true},
		{"owned toughness", func(effect *compiler.CompiledEffect, _ *compiler.CompiledEffect, _ *game.Instruction, _ *[][2]int) {
			effect.Amount.DynamicKind = compiler.DynamicAmountSourceToughness
		}, true},
		{"wrong node", func(effect *compiler.CompiledEffect, _ *compiler.CompiledEffect, _ *game.Instruction, _ *[][2]int) {
			effect.Amount.ReferenceNodeID++
		}, false},
		{"wrong clause", func(effect *compiler.CompiledEffect, _ *compiler.CompiledEffect, _ *game.Instruction, _ *[][2]int) {
			effect.References[0].ProducerClauseID++
		}, false},
		{"source subject", func(effect *compiler.CompiledEffect, _ *compiler.CompiledEffect, _ *game.Instruction, _ *[][2]int) {
			effect.References[0].Binding = compiler.ReferenceBindingSource
		}, false},
		{"event subject", func(effect *compiler.CompiledEffect, _ *compiler.CompiledEffect, _ *game.Instruction, _ *[][2]int) {
			effect.References[0].Binding = compiler.ReferenceBindingEventCard
		}, false},
		{"plural producer", func(_ *compiler.CompiledEffect, producer *compiler.CompiledEffect, _ *game.Instruction, _ *[][2]int) {
			producer.Amount.Value = 2
		}, false},
		{"expanded producer", func(_ *compiler.CompiledEffect, _ *compiler.CompiledEffect, _ *game.Instruction, ranges *[][2]int) {
			*ranges = [][2]int{{0, 2}}
		}, false},
		{"missing primitive", func(_ *compiler.CompiledEffect, _ *compiler.CompiledEffect, instruction *game.Instruction, _ *[][2]int) {
			instruction.Primitive = nil
		}, false},
		{"conflicting scalar", func(_ *compiler.CompiledEffect, _ *compiler.CompiledEffect, instruction *game.Instruction, _ *[][2]int) {
			instruction.Primitive = game.LookAtLibraryTop{Player: game.ControllerReference(),
				PublishCharacteristics: game.LibraryCardCharacteristics{Power: "other"}}
		}, false},
		{"modified characteristic", func(effect *compiler.CompiledEffect, _ *compiler.CompiledEffect, _ *game.Instruction, _ *[][2]int) {
			effect.Amount.Addend = 1
		}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			producer := compiler.CompiledEffect{Kind: compiler.EffectLookAtLibraryTop, Exact: true, ClauseID: 3,
				Context: parser.EffectContextController, CardSource: parser.EffectCardSourceTopOfPlayerLibrary,
				Amount: compiler.CompiledAmount{Known: true, Value: 1}}
			effect := compiler.CompiledEffect{Exact: true,
				Amount: compiler.CompiledAmount{DynamicKind: compiler.DynamicAmountSourcePower, ReferenceNodeID: 19, Multiplier: 1},
				References: []compiler.CompiledReference{{NodeID: 19, ProducerClauseID: 3,
					Binding: compiler.ReferenceBindingPriorInstructionResult, PriorInstruction: 0}}}
			instruction := game.Instruction{Primitive: game.LookAtLibraryTop{Player: game.ControllerReference()}}
			ranges := [][2]int{{0, 1}}
			test.mutate(&effect, &producer, &instruction, &ranges)
			before := instruction
			sequence := []game.Instruction{instruction}
			key, valid := sequenceLibraryCardCharacteristic(effect, []compiler.CompiledEffect{producer}, sequence, ranges)
			if valid != test.valid {
				t.Fatalf("accepted=%t, want %t", valid, test.valid)
			}
			if !valid {
				if before.Primitive == nil {
					if sequence[0].Primitive != nil {
						t.Fatal("failed proof installed a primitive")
					}
					return
				}
				if game.PublishedLinkedKey(sequence[0].Primitive) != game.PublishedLinkedKey(before.Primitive) ||
					len(sequence[0].LocalProducts.Results) != 0 {
					t.Fatal("failed proof mutated the producer")
				}
				return
			}
			if !sequence[0].LocalProducts.HasResult(key) || !sequence[0].LocalProducts.HasLink(sequenceProductKey(0)) {
				t.Fatal("exact scalar and object did not receive independent local declarations")
			}
			if err := game.ValidateInstructionSequence(sequence); err != nil {
				t.Fatal(err)
			}
		})
	}
}
