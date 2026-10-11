package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

func libraryCardCharacteristicConsumer(effect compiler.CompiledEffect, effects []compiler.CompiledEffect) bool {
	_, ok := libraryCardCharacteristicReference(effect, effects)
	return ok
}

func libraryCardCharacteristicReference(effect compiler.CompiledEffect, effects []compiler.CompiledEffect) (compiler.CompiledReference, bool) {
	amount := effect.Amount
	if amount.ReferenceNodeID < 0 || amount.Known || amount.Multiplier != 1 ||
		amount.Addend != 0 || len(amount.Operands) != 0 || effect.Duration != compiler.DurationNone ||
		amount.DynamicKind != compiler.DynamicAmountSourcePower && amount.DynamicKind != compiler.DynamicAmountSourceToughness {
		return compiler.CompiledReference{}, false
	}
	var subject compiler.CompiledReference
	found := false
	for _, reference := range effect.References {
		if reference.NodeID != amount.ReferenceNodeID {
			continue
		}
		if found || !sequenceLibraryCardReference(reference, effects) {
			return compiler.CompiledReference{}, false
		}
		subject, found = reference, true
	}
	return subject, found
}

func observedLibraryCardAmount(ctx contentCtx, amount compiler.CompiledAmount) (game.DynamicAmount, bool) {
	if ctx.observedCharacteristicKey == "" ||
		amount.Known || amount.Multiplier != 1 || amount.Addend != 0 ||
		len(amount.Operands) != 0 || len(ctx.content.References) != 1 ||
		amount.DynamicKind != compiler.DynamicAmountSourcePower &&
			amount.DynamicKind != compiler.DynamicAmountSourceToughness {
		return game.DynamicAmount{}, false
	}
	reference := ctx.content.References[0]
	if reference.NodeID != amount.ReferenceNodeID || !reference.SupportsUse(compiler.ReferenceUseCharacteristic) ||
		(reference.Binding != compiler.ReferenceBindingTarget &&
			(reference.ProducerClauseID <= 0 || reference.Binding != compiler.ReferenceBindingPriorInstructionResult ||
				reference.PriorInstruction != ctx.priorInstruction || ctx.priorLinkedKey == "")) {
		return game.DynamicAmount{}, false
	}
	return game.DynamicAmount{
		Kind: game.DynamicAmountPreviousEffectResult, ResultKey: ctx.observedCharacteristicKey, Multiplier: 1,
	}, true
}

func applyLibraryCardCharacteristicGate(sequence []game.Instruction, key game.ResultKey) bool {
	if key == "" {
		return true
	}
	for _, instruction := range sequence {
		if instruction.ResultGate.Exists {
			return false
		}
	}
	for i := range sequence {
		sequence[i].ResultGate = opt.Val(game.InstructionResultGate{Key: key, AmountAvailable: true})
	}
	return true
}

func sequenceLibraryCardCharacteristic(
	effect compiler.CompiledEffect,
	effects []compiler.CompiledEffect,
	sequence []game.Instruction,
	ranges [][2]int,
) (game.ResultKey, bool) {
	subject, ok := libraryCardCharacteristicReference(effect, effects)
	if !ok {
		return "", false
	}
	index := subject.PriorInstruction
	if index >= len(ranges) || ranges[index][1]-ranges[index][0] != 1 {
		return "", false
	}
	instructionIndex := ranges[index][0]
	if instructionIndex < 0 || instructionIndex >= len(sequence) {
		return "", false
	}
	instruction := sequence[instructionIndex]
	if instruction.Primitive == nil {
		return "", false
	}
	var outputs game.LibraryCardCharacteristics
	switch instruction.Primitive.Kind() {
	case game.PrimitiveLookAtLibraryTop:
		look, ok := instruction.Primitive.(game.LookAtLibraryTop)
		if !ok {
			return "", false
		}
		outputs = look.PublishCharacteristics
	case game.PrimitiveReveal:
		reveal, ok := instruction.Primitive.(game.Reveal)
		if !ok || reveal.Card.Kind != game.CardReferenceNone || reveal.Amount.IsDynamic() || reveal.Amount.Value() != 1 {
			return "", false
		}
		outputs = reveal.PublishCharacteristics
	default:
		return "", false
	}
	property := "power"
	existing := outputs.Power
	if effect.Amount.DynamicKind == compiler.DynamicAmountSourceToughness {
		property, existing = "toughness", outputs.Toughness
	}
	key := game.ResultKey(fmt.Sprintf("sequence-effect-%d-%s", index, property))
	if existing != "" && existing != key {
		return "", false
	}
	if _, published := trySetInstructionPublishLinked(&instruction, sequenceProductKey(index)); !published {
		return "", false
	}
	if property == "power" {
		outputs.Power = key
	} else {
		outputs.Toughness = key
	}
	switch instruction.Primitive.Kind() {
	case game.PrimitiveLookAtLibraryTop:
		primitive, ok := instruction.Primitive.(game.LookAtLibraryTop)
		if !ok {
			return "", false
		}
		primitive.PublishCharacteristics = outputs
		instruction.Primitive = primitive
	case game.PrimitiveReveal:
		primitive, ok := instruction.Primitive.(game.Reveal)
		if !ok {
			return "", false
		}
		primitive.PublishCharacteristics = outputs
		instruction.Primitive = primitive
	default:
		return "", false
	}
	if !instruction.LocalProducts.HasLink(sequenceProductKey(index)) {
		instruction.LocalProducts.Links = append(instruction.LocalProducts.Links, sequenceProductKey(index))
	}
	if !instruction.LocalProducts.HasResult(key) {
		instruction.LocalProducts.Results = append(instruction.LocalProducts.Results, key)
	}
	sequence[instructionIndex] = instruction
	return key, true
}
