package cardgen

import "github.com/natefinch/council4/mtg/game"

func optionalLinkedPublicationsModeled(sequence []game.Instruction) bool {
	optionalProducts := make(map[game.LinkedKey]bool)
	for _, instruction := range sequence {
		if instruction.Primitive == nil {
			continue
		}
		key := game.PublishedLinkedKey(instruction.Primitive)
		if key == "" {
			continue
		}
		if instruction.PublishOptionalDecision != "" || instruction.OptionalDecisionGate != "" {
			if !sequencePublisherInvalidatesBeforeGates(instruction.Primitive) &&
				!(instruction.ClearLinkedBeforeGate && captureMovePublisher(instruction.Primitive)) {
				return false
			}
			optionalProducts[key] = true
			continue
		}
		var object game.ObjectReference
		if apply, ok := instruction.Primitive.(game.ApplyContinuous); ok && apply.Object.Exists {
			object = apply.Object.Val
		} else if modify, ok := instruction.Primitive.(game.ModifyPT); ok {
			object = modify.Object
		}
		if object.Kind() == game.ObjectReferenceLinkedObject &&
			optionalProducts[game.LinkedKey(object.LinkID())] &&
			!sequencePublisherInvalidatesBeforeGates(instruction.Primitive) {
			return false
		}
		optionalProducts[key] = false
	}
	return true
}
