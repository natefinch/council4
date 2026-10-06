package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// The compiler already chose the entered subject. This adapter checks its
// concrete optional publisher and target mapping without rebinding the noun.
func bindOptionalEnteredObjectReference(
	ctx *contentCtx,
	sequence []game.Instruction,
	ranges [][2]int,
	targetIndices map[shared.Span]int,
	targets []game.TargetSpec,
	producerTargets []compiler.CompiledTarget,
) string {
	if len(ctx.content.References) != 1 {
		return ""
	}
	reference := ctx.content.References[0]
	if reference.Binding != compiler.ReferenceBindingPriorInstructionResult {
		return ""
	}
	occurrence, fromTarget := reference.ProducerTargetOccurrence()
	if !fromTarget {
		return ""
	}
	index := reference.PriorInstruction
	if index < 0 || index >= len(ranges) {
		return "structural — entered-object producer range is unavailable"
	}
	span := ranges[index]
	if span[0] < 0 || span[1] > len(sequence) || span[1]-span[0] != 1 {
		return "structural — entered-object producer has no singular instruction"
	}
	instruction := sequence[span[0]]
	if instruction.PublishOptionalDecision == "" && instruction.OptionalDecisionGate == "" {
		return ""
	}
	if occurrence < 0 || occurrence >= len(producerTargets) ||
		!reference.ProducerTargetMatches(producerTargets[occurrence]) {
		return "structural — entered-object producer target occurrence is unavailable"
	}
	put, entered := instruction.Primitive.(game.PutOnBattlefield)
	if !entered {
		return "structural — entered-object subject has no actual entry publisher"
	}
	card, cardSource := put.Source.CardRef()
	target := producerTargets[occurrence]
	targetIndex, mapped := targetIndices[target.Span]
	if !mapped || targetIndex < 0 || targetIndex >= len(targets) ||
		targets[targetIndex].Allow&game.TargetAllowCard == 0 ||
		!cardSource || card.Kind != game.CardReferenceTarget ||
		card.TargetIndex != cardTargetSpecsBefore(targets, targetIndex) {
		return "structural — optional entered-object target is not mapped"
	}
	producer, key, published := sequencePriorInstructionLink([]compiler.CompiledReference{reference}, sequence, ranges)
	if !published || ctx.priorLinkedKey != "" &&
		(ctx.priorInstruction != producer || ctx.priorLinkedKey != key) {
		return "structural — optional entered-object subject has no actual publication"
	}
	ctx.priorInstruction, ctx.priorLinkedKey = producer, key
	ctx.content.Targets = nil
	return ""
}
