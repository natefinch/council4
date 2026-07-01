package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerTrailingBackReferenceExile lowers a spell paragraph that exiles the
// permanent an earlier paragraph already acted on ("... exile that creature."),
// optionally gated by an "if" condition. It is the trailing clause of a
// two-paragraph removal spell such as Dispatch ("Tap target creature.
// Metalcraft — If you control three or more artifacts, exile that creature."):
// the first paragraph targets and taps a creature, and this paragraph exiles
// that same creature when the condition holds.
//
// The paragraph carries no target of its own; its single ReferenceThatObject
// back-references the permanent the leading paragraph targeted. Because the
// antecedent target lives in a different paragraph, the compiler leaves the
// reference unbound, so this lowerer maps it to target index 0 and emits a
// targetless resolving sequence. The spell-face combiner folds that sequence
// onto the leading paragraph (which owns target 0) via appendSequentialSpellParagraph,
// at which point TargetPermanentReference(0) resolves to the tapped creature.
//
// It returns ok=false for any shape it does not fully consume: a fresh target,
// a non-exile or non-controller effect, more than one back-reference, or a
// condition that is not a supported effect-gate condition.
func lowerTrailingBackReferenceExile(ctx contentCtx) (game.AbilityContent, bool) {
	// lowerContent calls this only from its len(Effects)==1 block, so a different
	// effect count is a dispatch bug rather than an unsupported card.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf("lowerTrailingBackReferenceExile: reached with %d effects; lowerContent dispatches here only for single-effect content", len(ctx.content.Effects)))
	}
	if ctx.enclosingKind != compiler.AbilitySpell ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		ctx.optional {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectExile ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.Context != parser.EffectContextController ||
		effect.Duration != compiler.DurationNone {
		return game.AbilityContent{}, false
	}
	if len(ctx.content.References) != 1 ||
		ctx.content.References[0].Kind != compiler.ReferenceThatObject {
		return game.AbilityContent{}, false
	}
	instruction := game.Instruction{
		Primitive: game.Exile{Object: game.TargetPermanentReference(0)},
	}
	switch len(ctx.content.Conditions) {
	case 0:
	case 1:
		gate, ok := lowerCondition(ctx.content.Conditions[0], conditionContextEffectGate)
		if !ok {
			return game.AbilityContent{}, false
		}
		instruction.Condition = opt.Val(game.EffectCondition{Condition: opt.Val(gate)})
	default:
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{instruction}}.Ability(), true
}
