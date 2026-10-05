package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
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
// Unmodeled optional, multi-instruction and competing antecedents fail closed.
// A scoped optional entered-object publisher invalidates before its gates. An
// incompatible existing publication is never overwritten.
func sequencePriorInstructionLink(
	references []compiler.CompiledReference,
	sequence []game.Instruction,
	ranges [][2]int,
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
		if span[1]-span[0] != 1 {
			return 0, "", false
		}
		instructionIndex := span[0]
		if instructionIndex < 0 || instructionIndex >= len(sequence) {
			return 0, "", false
		}
		for _, other := range references {
			if other.Binding == compiler.ReferenceBindingPriorInstructionResult && other.PriorInstruction != j {
				return 0, "", false
			}
		}
		candidateKey := sequenceProductKey(j)
		linked, published := trySetInstructionPublishLinked(&sequence[instructionIndex], candidateKey)
		if !published {
			return 0, "", false
		}
		return j, linked, true
	}
	return 0, "", false
}

func sequenceProductKey(index int) game.LinkedKey {
	return game.LinkedKey(fmt.Sprintf("sequence-effect-%d-product", index))
}

// These actual entered-object publishers invalidate their links before envelope
// gates, so a skipped publication cannot leave an older incarnation available.
func sequencePublisherInvalidatesBeforeGates(primitive game.Primitive) bool {
	return primitive != nil &&
		(primitive.Kind() == game.PrimitivePutOnBattlefield || primitive.Kind() == game.PrimitiveCreateToken)
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
		instr.Optional && (instr.PublishOptionalDecision == "" || !sequencePublisherInvalidatesBeforeGates(instr.Primitive)) {
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
	default:
		return "", false
	}
}
