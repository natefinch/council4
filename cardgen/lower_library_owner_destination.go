package cardgen

import "github.com/natefinch/council4/cardgen/oracle/compiler"

func libraryCardActionSubject(effect compiler.CompiledEffect, references []compiler.CompiledReference) (compiler.CompiledReference, bool) {
	var subject, owner *compiler.CompiledReference
	for i := range references {
		reference := &references[i]
		if reference.ProducerClauseID <= 0 {
			return compiler.CompiledReference{}, false
		}
		switch reference.Binding {
		case compiler.ReferenceBindingPriorInstructionResult:
			if subject != nil || reference.Kind == compiler.ReferenceThatPlayer {
				return compiler.CompiledReference{}, false
			}
			subject = reference
		case compiler.ReferenceBindingLibraryOwner:
			if owner != nil || reference.Kind != compiler.ReferenceThatPlayer {
				return compiler.CompiledReference{}, false
			}
			owner = reference
		default:
			return compiler.CompiledReference{}, false
		}
	}
	if subject == nil || effect.LibraryOwnerDestination != (owner != nil) {
		return compiler.CompiledReference{}, false
	}
	if owner != nil && (owner.ProducerClauseID != subject.ProducerClauseID ||
		owner.PriorInstruction != subject.PriorInstruction) {
		return compiler.CompiledReference{}, false
	}
	return *subject, true
}
