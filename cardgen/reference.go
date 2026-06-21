package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
)

// referenceLoweringContext states which bindings exist at the lowering seam.
// PriorInstruction and PriorLinkedKey must identify the same published result.
type referenceLoweringContext struct {
	AllowSource      bool
	AllowTarget      bool
	AllowEvent       bool
	SourceCardObject bool
	PriorInstruction int
	PriorLinkedKey   game.LinkedKey
	TargetLinkedKey  game.LinkedKey
}

func lowerObjectReference(reference compiler.CompiledReference, ctx referenceLoweringContext) (game.ObjectReference, bool) {
	var result game.ObjectReference
	switch reference.Binding {
	case compiler.ReferenceBindingSource:
		if !ctx.AllowSource {
			return game.ObjectReference{}, false
		}
		if ctx.SourceCardObject {
			result = game.SourceCardPermanentReference()
		} else {
			result = game.SourcePermanentReference()
		}
	case compiler.ReferenceBindingTarget:
		switch {
		case ctx.TargetLinkedKey != "":
			result = game.LinkedObjectReference(string(ctx.TargetLinkedKey))
		case !ctx.AllowTarget || reference.Occurrence < 0:
			return game.ObjectReference{}, false
		default:
			result = game.TargetPermanentReference(reference.Occurrence)
		}
	case compiler.ReferenceBindingEventPermanent:
		if !ctx.AllowEvent {
			return game.ObjectReference{}, false
		}
		result = game.EventPermanentReference()
	case compiler.ReferenceBindingEventStackObject:
		if !ctx.AllowEvent {
			return game.ObjectReference{}, false
		}
		result = game.EventStackObjectReference()
	case compiler.ReferenceBindingPriorInstructionResult:
		if ctx.PriorLinkedKey == "" || reference.PriorInstruction != ctx.PriorInstruction {
			return game.ObjectReference{}, false
		}
		result = game.LinkedObjectReference(string(ctx.PriorLinkedKey))
	default:
		return game.ObjectReference{}, false
	}
	return result, len(result.Validate()) == 0
}

func lowerCardReference(reference compiler.CompiledReference, ctx referenceLoweringContext) (game.CardReference, bool) {
	switch reference.Binding {
	case compiler.ReferenceBindingSource:
		if !ctx.AllowSource {
			return game.CardReference{}, false
		}
		return game.CardReference{Kind: game.CardReferenceSource}, true
	case compiler.ReferenceBindingTarget:
		if ctx.TargetLinkedKey != "" {
			return game.CardReference{Kind: game.CardReferenceLinked, LinkID: string(ctx.TargetLinkedKey)}, true
		}
		if !ctx.AllowTarget || reference.Occurrence < 0 {
			return game.CardReference{}, false
		}
		return game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: reference.Occurrence}, true
	case compiler.ReferenceBindingEventPermanent, compiler.ReferenceBindingEventCard:
		if !ctx.AllowEvent {
			return game.CardReference{}, false
		}
		return game.CardReference{Kind: game.CardReferenceEvent}, true
	case compiler.ReferenceBindingPriorInstructionResult:
		if ctx.PriorLinkedKey == "" || reference.PriorInstruction != ctx.PriorInstruction {
			return game.CardReference{}, false
		}
		return game.CardReference{Kind: game.CardReferenceLinked, LinkID: string(ctx.PriorLinkedKey)}, true
	default:
		return game.CardReference{}, false
	}
}

// lowerPlayerReference maps a CompiledReference to a game.PlayerReference.
// It handles EventPlayer → EventPlayerReference() and Source → ControllerReference()
// bindings. AllowEvent must be set for EventPlayer; AllowSource for Source.
func lowerPlayerReference(reference compiler.CompiledReference, ctx referenceLoweringContext) (game.PlayerReference, bool) {
	switch reference.Binding {
	case compiler.ReferenceBindingEventPlayer:
		if !ctx.AllowEvent {
			return game.PlayerReference{}, false
		}
		return game.EventPlayerReference(), true
	case compiler.ReferenceBindingSource:
		if !ctx.AllowSource {
			return game.PlayerReference{}, false
		}
		return game.ControllerReference(), true
	default:
		return game.PlayerReference{}, false
	}
}

func referencesBindTo(
	references []compiler.CompiledReference,
	binding compiler.ReferenceBinding,
	occurrence int,
) bool {
	if len(references) == 0 {
		return false
	}
	for _, reference := range references {
		if reference.Binding != binding {
			return false
		}
		switch binding {
		case compiler.ReferenceBindingTarget:
			if reference.Occurrence != occurrence {
				return false
			}
		case compiler.ReferenceBindingPriorInstructionResult:
			if reference.PriorInstruction != occurrence {
				return false
			}
		default:
		}
	}
	return true
}

func referencesContainKind(references []compiler.CompiledReference, kind compiler.ReferenceKind) bool {
	for _, reference := range references {
		if reference.Kind == kind {
			return true
		}
	}
	return false
}
