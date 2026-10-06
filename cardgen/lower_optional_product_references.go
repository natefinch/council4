package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

func bindOptionalEnteredObjectReference(
	ctx *contentCtx,
	sequence []game.Instruction,
	ranges [][2]int,
	targetIndices map[shared.Span]int,
	targets []game.TargetSpec,
) string {
	if len(ctx.content.Effects) != 1 || len(ctx.content.References) != 1 {
		return ""
	}
	effect := &ctx.content.Effects[0]
	if effect.Context != parser.EffectContextReferencedObject && effect.Context != parser.EffectContextPriorSubject {
		return ""
	}
	reference := ctx.content.References[0]
	if reference.Binding != compiler.ReferenceBindingTarget ||
		reference.Occurrence < 0 || reference.Occurrence >= len(ctx.content.Targets) {
		return ""
	}
	target := ctx.content.Targets[reference.Occurrence]
	if _, permanent := permanentTargetSpec(target); permanent {
		return ""
	}
	targetIndex, mapped := targetIndices[target.Span]
	if !mapped {
		for _, instruction := range sequence {
			if (instruction.PublishOptionalDecision != "" || instruction.OptionalDecisionGate != "") &&
				instruction.Primitive != nil && instruction.Primitive.Kind() == game.PrimitivePutOnBattlefield {
				return "structural — optional entered-object target is not mapped"
			}
		}
		return ""
	}
	if targetIndex < 0 || targetIndex >= len(targets) || targets[targetIndex].Allow&game.TargetAllowCard == 0 {
		return ""
	}
	cardIndex := cardTargetSpecsBefore(targets, targetIndex)
	for ei := len(ranges) - 1; ei >= 0; ei-- {
		span := ranges[ei]
		for ii := span[1] - 1; ii >= span[0]; ii-- {
			if ii < 0 || ii >= len(sequence) {
				return "structural — optional entered-object producer range is unavailable"
			}
			instruction := &sequence[ii]
			if instruction.PublishOptionalDecision == "" && instruction.OptionalDecisionGate == "" {
				continue
			}
			put, ok := instruction.Primitive.(game.PutOnBattlefield)
			if !ok {
				continue
			}
			card, ok := put.Source.CardRef()
			if !ok || card.Kind != game.CardReferenceTarget || card.TargetIndex != cardIndex {
				continue
			}
			reference.Binding = compiler.ReferenceBindingPriorInstructionResult
			reference.PriorInstruction = ei
			producer, key, published := sequencePriorInstructionLink(
				[]compiler.CompiledReference{reference}, sequence, ranges,
			)
			if !published || ctx.priorLinkedKey != "" &&
				(ctx.priorInstruction != producer || ctx.priorLinkedKey != key) {
				return "structural — optional entered-object subject has no actual publication"
			}
			ctx.priorInstruction, ctx.priorLinkedKey = producer, key
			ctx.content.References = []compiler.CompiledReference{reference}
			effect.References = slices.Clone(ctx.content.References)
			if len(effect.SubjectReferences) == 1 &&
				effect.SubjectReferences[0].Binding == compiler.ReferenceBindingTarget {
				effect.SubjectReferences = slices.Clone(ctx.content.References)
			}
			ctx.content.Targets = nil
			return ""
		}
	}
	return ""
}
