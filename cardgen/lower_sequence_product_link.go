package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Ordered sequence consumers share a canonical object publication with the
// exact producer identified by the compiler. Currently the supported producers
// are CreateToken and single-source PutOnBattlefield; other primitive kinds need
// their own actual-result publication proof before participating.

// sequencePriorInstructionLink reports the antecedent effect index and link key
// a clause's own references need to resolve a ReferenceBindingPriorInstructionResult
// binding, publishing that antecedent's single instruction under a fresh
// canonical key if it is not already linked. ranges holds each effect's
// instruction span within sequence, recorded as each clause lowers in effect
// order; sequence holds the instructions lowered so far.
//
// Optional, multi-instruction and competing antecedents fail closed. An
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
		if reference.Binding != compiler.ReferenceBindingPriorInstructionResult {
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
			if other.Binding == compiler.ReferenceBindingPriorInstructionResult && other.PriorInstruction != j {
				return 0, "", false
			}
		}
		candidateKey := sequenceProductKey(j)
		instructionIndex := -1
		for i := span[0]; i < span[1]; i++ {
			candidate := sequence[i]
			if !capture && candidate.Primitive != nil &&
				(candidate.Primitive.Kind() == game.PrimitiveMovePermanent || candidate.Primitive.Kind() == game.PrimitiveMoveTopOfLibrary) {
				return 0, "", false
			}
			key := sequenceProductKey(j)
			if capture {
				candidate.Optional = false
				if existing := game.PublishedLinkedKey(candidate.Primitive); existing != "" {
					key = existing
				}
			}
			if _, published := trySetInstructionPublishLinked(&candidate, key); !published {
				if capture && game.PublishedLinkedKey(candidate.Primitive) != "" {
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
		return j, linked, true
	}
	return 0, "", false
}

func sequenceProductKey(index int) game.LinkedKey {
	return game.LinkedKey(fmt.Sprintf("sequence-effect-%d-product", index))
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
	if instr.Optional || instr.Primitive == nil {
		return "", false
	}
	switch instr.Primitive.Kind() {
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
		if primitive.PublishLinked != "" {
			return primitive.PublishLinked, primitive.PublishLinked == key
		}
		primitive.PublishLinked = key
		instr.Primitive = primitive
		return key, true
	default:
		return "", false
	}
}
