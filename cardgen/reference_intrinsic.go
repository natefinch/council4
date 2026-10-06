package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
)

func contentSubjectProofsOwned(content compiler.AbilityContent) bool {
	validate := func(reference compiler.CompiledReference) bool {
		switch reference.Binding {
		case compiler.ReferenceBindingUnsupported, compiler.ReferenceBindingAmbiguous,
			compiler.ReferenceBindingPaidCost:
			return true
		default:
			return content.OwnsSubject(reference)
		}
	}
	for _, reference := range content.References {
		if !validate(reference) {
			return false
		}
	}
	for _, condition := range content.Conditions {
		if condition.ObjectReference != nil && !validate(*condition.ObjectReference) {
			return false
		}
	}
	return true
}

func lowerIntrinsicSourceObject(ctx contentCtx) (game.ObjectReference, bool) {
	if !contentSubjectProofsOwned(ctx.content) {
		return game.ObjectReference{}, false
	}
	reference, ok := ctx.content.Source.OriginalObjectSubject()
	if !ok {
		return game.ObjectReference{}, false
	}
	return lowerObjectReference(reference, referenceLoweringContext{AllowSource: true})
}

// These object references are capture adapters, not live permanent subjects.
// capturedCard validates their exact source/event card incarnation at scheduling.
func lowerCapturedCardSubject(reference compiler.CompiledReference, bindings referenceLoweringContext) (game.ObjectReference, bool) {
	if !reference.SubjectSupported() || reference.SubjectDomain() != compiler.ReferenceSubjectCard {
		return game.ObjectReference{}, false
	}
	switch reference.Binding {
	case compiler.ReferenceBindingSource:
		if bindings.AllowSource && reference.SubjectLifetime() == compiler.ReferenceLifetimeCardIncarnation {
			return game.SourceCardPermanentReference(), true
		}
	case compiler.ReferenceBindingEventCard:
		if bindings.AllowEvent {
			return game.EventPermanentReference(), true
		}
	case compiler.ReferenceBindingTarget:
		if bindings.AllowTarget && reference.Occurrence >= 0 {
			return game.TargetCardReference(reference.Occurrence), true
		}
	case compiler.ReferenceBindingPriorInstructionResult:
		return lowerObjectReference(reference, bindings)
	}
	return game.ObjectReference{}, false
}
