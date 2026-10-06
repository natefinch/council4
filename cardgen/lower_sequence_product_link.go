package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Ordered sequence consumers share a canonical object publication with the
// exact producer identified by the compiler. Resolving permanent consumers use
// CreateToken and single-source PutOnBattlefield; fixed-phase capture additionally
// admits proven single-card exile publishers. Singular library observations
// publish actual card identities. A publication field alone is not an actual-result proof.

// sequencePriorInstructionLink reports the antecedent effect index and link key
// a clause's own references need to resolve a ReferenceBindingPriorInstructionResult
// binding, publishing that antecedent's single instruction under a fresh
// canonical key if it is not already linked. ranges holds each effect's
// instruction span within sequence, recorded as each clause lowers in effect
// order; sequence holds the instructions lowered so far.
//
// Unmodeled optional, multi-instruction and competing antecedents fail closed.
// Scoped optional entered-object publishers invalidate before their gates;
// optional observations publish only actual accepted cards. An
// incompatible existing publication is never overwritten.
func sequencePriorInstructionLink(
	references []compiler.CompiledReference,
	sequence []game.Instruction,
	ranges [][2]int,
) (priorEffect int, key game.LinkedKey, ok bool) {
	return sequencePriorInstructionPublication(references, sequence, ranges, false)
}

func sequencePriorInstructionPublication(
	references []compiler.CompiledReference,
	sequence []game.Instruction,
	ranges [][2]int,
	capture bool,
) (priorEffect int, key game.LinkedKey, ok bool) {
	for _, reference := range references {
		if reference.Binding != compiler.ReferenceBindingPriorInstructionResult &&
			reference.Binding != compiler.ReferenceBindingLibraryOwner {
			continue
		}
		j := reference.PriorInstruction
		if j < 0 || j >= len(ranges) {
			return 0, "", false
		}
		span := ranges[j]
		if span[0] < 0 || span[1] > len(sequence) || span[1] <= span[0] ||
			(!capture && span[1]-span[0] != 1) {
			return 0, "", false
		}
		for _, other := range references {
			if (other.Binding == compiler.ReferenceBindingPriorInstructionResult ||
				other.Binding == compiler.ReferenceBindingLibraryOwner) && other.PriorInstruction != j {
				return 0, "", false
			}
		}
		candidateKey := sequenceProductKey(j)
		instructionIndex := -1
		for i := span[0]; i < span[1]; i++ {
			candidate := sequence[i]
			if !capture && candidate.Primitive != nil &&
				captureMovePublisher(candidate.Primitive) {
				return 0, "", false
			}
			key := sequenceProductKey(j)
			if capture {
				candidate.Optional = false
				if existing := game.PublishedLinkedKey(candidate.Primitive); existing != "" {
					if captureMovePublisher(candidate.Primitive) && !candidate.ClearLinkedBeforeGate {
						return 0, "", false
					}
					key = existing
				}
			}
			if _, published := trySetInstructionPublishLinked(&candidate, key); !published {
				if capture && (game.PublishedLinkedKey(candidate.Primitive) != "" ||
					!captureExpansionInstructionCompatible(candidate.Primitive)) {
					return 0, "", false
				}
				continue
			}
			if instructionIndex >= 0 {
				return 0, "", false
			}
			instructionIndex = i
			candidateKey = key
		}
		if instructionIndex < 0 {
			return 0, "", false
		}
		candidate := sequence[instructionIndex]
		if capture {
			candidate.Optional = false
		}
		linked, published := trySetInstructionPublishLinked(&candidate, candidateKey)
		if !published {
			return 0, "", false
		}
		sequence[instructionIndex].Primitive = candidate.Primitive
		if capture && captureMovePublisher(candidate.Primitive) {
			sequence[instructionIndex].ClearLinkedBeforeGate = true
		}
		if reference.ProducerClauseID > 0 && !sequence[instructionIndex].LocalProducts.HasLink(linked) {
			sequence[instructionIndex].LocalProducts.Links = append(sequence[instructionIndex].LocalProducts.Links, linked)
		}
		return j, linked, true
	}
	return 0, "", false
}

func sequenceProductKey(index int) game.LinkedKey {
	return game.LinkedKey(fmt.Sprintf("sequence-effect-%d-product", index))
}

func sequenceLibraryCardReference(reference compiler.CompiledReference, effects []compiler.CompiledEffect) bool {
	index := reference.PriorInstruction
	return reference.Binding == compiler.ReferenceBindingPriorInstructionResult &&
		reference.ProducerClauseID > 0 && index >= 0 && index < len(effects) &&
		effects[index].ClauseID == reference.ProducerClauseID &&
		compiler.SingularLibraryCardProducer(effects[index])
}

// These actual entered-object publishers invalidate their links before envelope
// gates, so a skipped publication cannot leave an older incarnation available.
func sequencePublisherInvalidatesBeforeGates(primitive game.Primitive) bool {
	return primitive != nil &&
		(primitive.Kind() == game.PrimitivePutOnBattlefield || primitive.Kind() == game.PrimitiveCreateToken)
}

func captureExpansionInstructionCompatible(primitive game.Primitive) bool {
	if primitive == nil {
		return false
	}
	switch primitive.Kind() {
	case game.PrimitiveDraw, game.PrimitiveGainLife, game.PrimitiveLoseLife,
		game.PrimitiveAddCounter, game.PrimitiveModifyPT, game.PrimitiveApplyContinuous,
		game.PrimitiveTap, game.PrimitiveUntap:
		return true
	default:
		return false
	}
}

func captureMovePublisher(primitive game.Primitive) bool {
	return primitive != nil && (primitive.Kind() == game.PrimitiveMovePermanent ||
		primitive.Kind() == game.PrimitiveMoveTopOfLibrary || primitive.Kind() == game.PrimitiveMoveCard)
}

// trySetInstructionPublishLinked sets instr's PublishLinked field to key,
// reporting the key actually in effect afterward. It reports false (fail
// closed) for a primitive kind with no PublishLinked field, and reuses the
// existing key rather than overwriting it when the primitive already publishes
// one — succeeding with that key when it matches (a second consumer of the same
// antecedent), failing when it names something else (an existing hand-written
// linking path already claimed this instruction for a different key).
//
// A PublishLinked field alone is not proof that a primitive publishes the
// correct actual result.
func trySetInstructionPublishLinked(instr *game.Instruction, key game.LinkedKey) (game.LinkedKey, bool) {
	if instr.Primitive == nil ||
		instr.Optional && instr.Primitive.Kind() != game.PrimitiveLookAtLibraryTop &&
			instr.Primitive.Kind() != game.PrimitiveReveal &&
			(instr.PublishOptionalDecision == "" || !sequencePublisherInvalidatesBeforeGates(instr.Primitive)) {
		return "", false
	}
	switch instr.Primitive.Kind() {
	case game.PrimitiveLookAtLibraryTop:
		primitive := instr.Primitive.(game.LookAtLibraryTop)
		if primitive.PublishLinked != "" {
			return primitive.PublishLinked, primitive.PublishLinked == key
		}
		primitive.PublishLinked = key
		instr.Primitive = primitive
		return key, true
	case game.PrimitiveReveal:
		primitive := instr.Primitive.(game.Reveal)
		if primitive.Card.Kind == game.CardReferenceNone &&
			(primitive.Amount.IsDynamic() || primitive.Amount.Value() != 1) {
			return "", false
		}
		if primitive.PublishLinked != "" {
			return primitive.PublishLinked, primitive.PublishLinked == key
		}
		primitive.PublishLinked = key
		instr.Primitive = primitive
		return key, true
	case game.PrimitiveCreateToken:
		primitive, ok := instr.Primitive.(game.CreateToken)
		if !ok {
			return "", false
		}
		if primitive.PublishLinked != "" {
			return primitive.PublishLinked, primitive.PublishLinked == key
		}
		primitive.PublishLinked = key
		instr.Primitive = primitive
		return key, true
	case game.PrimitivePutOnBattlefield:
		primitive, ok := instr.Primitive.(game.PutOnBattlefield)
		if !ok || len(primitive.Sources) != 0 {
			return "", false
		}
		if _, card := primitive.Source.CardRef(); !card {
			if _, linked := primitive.Source.LinkedKey(); !linked {
				return "", false
			}
		}
		if primitive.PublishLinked != "" {
			return primitive.PublishLinked, primitive.PublishLinked == key
		}
		primitive.PublishLinked = key
		instr.Primitive = primitive
		return key, true
	case game.PrimitiveMoveTopOfLibrary:
		primitive, ok := instr.Primitive.(game.MoveTopOfLibrary)
		if !ok || primitive.Destination != zone.Exile ||
			primitive.PlayerGroup.Kind != game.PlayerGroupReferenceNone ||
			primitive.Amount.IsDynamic() || primitive.Amount.Value() != 1 {
			return "", false
		}
		if primitive.PublishLinked != "" {
			return primitive.PublishLinked, primitive.PublishLinked == key
		}
		primitive.PublishLinked = key
		instr.Primitive = primitive
		return key, true
	case game.PrimitiveMovePermanent:
		primitive, ok := instr.Primitive.(game.MovePermanent)
		if !ok || primitive.Destination != zone.Exile || primitive.ControlledChoice ||
			primitive.Group.Domain() != 0 || primitive.Object.Kind() == game.ObjectReferenceNone {
			return "", false
		}
		switch primitive.Object.Kind() {
		case game.ObjectReferenceTargetPermanent, game.ObjectReferenceSourcePermanent,
			game.ObjectReferenceEventPermanent, game.ObjectReferenceEventRelatedPermanent,
			game.ObjectReferenceLinkedObject, game.ObjectReferenceCapturedObject:
		default:
			return "", false
		}
		if primitive.PublishLinked != "" {
			return primitive.PublishLinked, primitive.PublishLinked == key
		}
		primitive.PublishLinked = key
		instr.Primitive = primitive
		return key, true
	case game.PrimitiveMoveCard:
		primitive, ok := instr.Primitive.(game.MoveCard)
		if !ok || primitive.Destination != zone.Exile || primitive.Card.Kind == game.CardReferenceNone ||
			primitive.Player.Kind() != game.PlayerReferenceNone ||
			primitive.PlayerGroup.Kind != game.PlayerGroupReferenceNone ||
			primitive.IncludeEventPermanentComponents || primitive.PublishLinkedObjectScoped {
			return "", false
		}
		if primitive.PublishLinked != "" {
			return primitive.PublishLinked, primitive.PublishLinked == key
		}
		primitive.PublishLinked = key
		primitive.ReplacePublishedLinked = true
		instr.Primitive = primitive
		return key, true
	default:
		return "", false
	}
}
