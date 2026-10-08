package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
)

func departureCardCharacteristicReference(effect compiler.CompiledEffect, effects []compiler.CompiledEffect) (compiler.CompiledReference, int, bool) {
	amount := effect.Amount
	if amount.Known || amount.Multiplier != 1 || amount.Addend != 0 || len(amount.Operands) != 0 ||
		(amount.DynamicKind != compiler.DynamicAmountSourcePower && amount.DynamicKind != compiler.DynamicAmountSourceToughness) {
		return compiler.CompiledReference{}, 0, false
	}
	for _, reference := range effect.References {
		if reference.NodeID != amount.ReferenceNodeID {
			continue
		}
		if index, ok := reference.DepartureCharacteristicProducer(effects, effect); ok {
			return reference, index, true
		}
	}
	return compiler.CompiledReference{}, 0, false
}

func sequenceDepartureCardCharacteristic(effect compiler.CompiledEffect, effects []compiler.CompiledEffect, sequence []game.Instruction, ranges [][2]int) (game.ResultKey, bool) {
	subject, index, ok := departureCardCharacteristicReference(effect, effects)
	if !ok || index >= len(ranges) || ranges[index][1]-ranges[index][0] != 1 {
		return "", false
	}
	instructionIndex := ranges[index][0]
	if instructionIndex < 0 || instructionIndex >= len(sequence) ||
		sequence[instructionIndex].Primitive.Kind() != game.PrimitiveMoveCard {
		return "", false
	}
	move := sequence[instructionIndex].Primitive.(game.MoveCard)
	occurrence, proved := subject.OriginalTargetOccurrence()
	if !proved || move.Card.Kind != game.CardReferenceTarget || move.Card.TargetIndex != occurrence {
		return "", false
	}
	property := "power"
	if effect.Amount.DynamicKind == compiler.DynamicAmountSourceToughness {
		property = "toughness"
	}
	key := game.ResultKey(fmt.Sprintf("sequence-effect-%d-departure-%s", index, property))
	if property == "power" {
		if move.PublishDepartureCharacteristics.Power != "" && move.PublishDepartureCharacteristics.Power != key {
			return "", false
		}
		move.PublishDepartureCharacteristics.Power = key
	} else {
		if move.PublishDepartureCharacteristics.Toughness != "" && move.PublishDepartureCharacteristics.Toughness != key {
			return "", false
		}
		move.PublishDepartureCharacteristics.Toughness = key
	}
	sequence[instructionIndex].Primitive = move
	if !sequence[instructionIndex].LocalProducts.HasResult(key) {
		sequence[instructionIndex].LocalProducts.Results = append(sequence[instructionIndex].LocalProducts.Results, key)
	}
	return key, true
}
