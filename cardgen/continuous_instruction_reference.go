package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
)

func continuousInstructionReferenceObject(
	ctx contentCtx,
	reference compiler.CompiledReference,
	effect *compiler.CompiledEffect,
) (game.ObjectReference, bool) {
	if !ctx.content.OwnsSubject(reference) {
		return game.ObjectReference{}, false
	}
	if reference.Binding == compiler.ReferenceBindingPriorInstructionResult {
		return lowerObjectReference(reference, referenceLoweringContext{
			PriorInstruction: ctx.priorInstruction,
			PriorLinkedKey:   ctx.priorLinkedKey,
		})
	}
	return continuousReferenceObject(reference, effect)
}
