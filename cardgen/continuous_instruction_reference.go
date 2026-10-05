package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
)

func continuousInstructionReferenceObject(
	ctx contentCtx,
	reference compiler.CompiledReference,
	effect *compiler.CompiledEffect,
	sourceAsCard bool,
) (game.ObjectReference, bool) {
	if reference.Binding == compiler.ReferenceBindingPriorInstructionResult {
		return lowerObjectReference(reference, referenceLoweringContext{
			PriorInstruction: ctx.priorInstruction,
			PriorLinkedKey:   ctx.priorLinkedKey,
		})
	}
	return continuousReferenceObject(reference, effect, sourceAsCard, ctx.enclosingKind == compiler.AbilitySpell)
}
