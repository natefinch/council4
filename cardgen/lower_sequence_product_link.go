package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
)

// This file generalizes the linking half of CR 607 ("Linked Abilities") for
// ordered sequences: when one clause's Oracle text causes an object to be
// created, exiled, milled, sacrificed, or otherwise produced, and a later
// clause refers back to "it"/"that <noun>"/"them" naming that same object, the
// two clauses are linked regardless of which verbs are involved. The compiler
// already recognizes this structurally (ReferenceBindingPriorInstructionResult
// + PriorInstruction), so the remaining gap was that every producer/consumer
// verb pair needed its own hand-written function to publish and re-resolve the
// link (createdTokenLinkKey, milledCardsLinkKey, sacrificedCreatureLinkKey, ...
// each with a matching bespoke sequence lowerer). sequencePriorInstructionLink
// closes that gap for the single-instruction case: it publishes the antecedent
// instruction's result under a canonical per-effect key and hands the
// consuming clause the same key through contentCtx, so any single-effect
// lowerer that already resolves its object through lowerObjectReference/
// lowerCardReference gains this capability for free.

// sequencePriorInstructionLink reports the antecedent effect index and link key
// a clause's own references need to resolve a ReferenceBindingPriorInstructionResult
// binding, publishing that antecedent's single instruction under a fresh
// canonical key if it is not already linked. ranges holds each effect's
// instruction span within sequence, recorded as each clause lowers in effect
// order; sequence holds the instructions lowered so far.
//
// It only handles a single-instruction antecedent (the common case: a create,
// exile, mill, or sacrifice clause that produces exactly one instruction). A
// multi-instruction antecedent (for example a multi-target clause) is left
// unlinked — callers still fail closed exactly as before, so this is strictly
// additive. It also fails closed rather than overwrite an antecedent
// instruction that already publishes a different key, so it never conflicts
// with an existing hand-written linking path.
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
		candidateKey := game.LinkedKey(fmt.Sprintf("sequence-effect-%d-product", j))
		linked, published := trySetInstructionPublishLinked(&sequence[instructionIndex], candidateKey)
		if !published {
			return 0, "", false
		}
		return j, linked, true
	}
	return 0, "", false
}

// trySetInstructionPublishLinked sets instr's PublishLinked field to key,
// reporting the key actually in effect afterward. It reports false (fail
// closed) for a primitive kind with no PublishLinked field, and reuses the
// existing key rather than overwriting it when the primitive already publishes
// one — succeeding with that key when it matches (a second consumer of the same
// antecedent), failing when it names something else (an existing hand-written
// linking path already claimed this instruction for a different key).
//
// Add a case here whenever a new producing primitive needs to participate in
// generic prior-instruction linking; every game.* primitive with a
// PublishLinked field is a candidate.
func trySetInstructionPublishLinked(instr *game.Instruction, key game.LinkedKey) (game.LinkedKey, bool) {
	switch primitive := instr.Primitive.(type) {
	case game.CreateToken:
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
