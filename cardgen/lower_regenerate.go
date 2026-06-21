package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// lowerRegenerateSpell lowers the regeneration family into a single Regenerate
// instruction that sets up a regeneration shield on one permanent. It supports
// three recipients, each reusing the existing regeneration-shield runtime:
//
//   - a chosen target permanent ("Regenerate target creature.", with any
//     supported target restriction such as "target creature you control" or
//     "target artifact"), including the multi-target forms;
//   - the ability's own source ("Regenerate this creature." / "Regenerate this
//     permanent." / "Regenerate <CardName>."), which lowers to the source
//     permanent reference and needs no target; and
//   - the permanent the source Aura or Equipment is attached to ("Regenerate
//     enchanted creature." / "Regenerate equipped creature."), which lowers to
//     the source attached-permanent reference.
//
// Any other regenerate shape — multiple effects, a negated or non-controller
// effect, conditional or modal content, or an unrepresentable recipient — fails
// closed with the shared unsupported-regenerate diagnostic.
func lowerRegenerateSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	if len(ctx.content.Targets) > 0 {
		return lowerFixedPermanentTargetSpell(ctx, "Regenerate", func(object game.ObjectReference) game.Primitive {
			return game.Regenerate{Object: object}
		})
	}
	if object, ok := lowerSourceRegenerateObject(ctx); ok {
		return game.Mode{
			Sequence: []game.Instruction{{Primitive: game.Regenerate{Object: object}}},
		}.Ability(), nil
	}
	return game.AbilityContent{}, contentDiagnostic(
		ctx,
		"unsupported regenerate spell",
		"the executable source backend supports only exact regenerate of one target, source, or attached permanent",
	)
}

// lowerSourceRegenerateObject resolves the non-target regenerate recipient: the
// ability's own source ("Regenerate this creature." / "Regenerate <CardName>.")
// or the permanent the source is attached to ("Regenerate enchanted creature.").
// It requires a single exact controller effect with no conditional or modal
// content and fails closed for every other shape.
func lowerSourceRegenerateObject(ctx contentCtx) (game.ObjectReference, bool) {
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		ctx.optional {
		return game.ObjectReference{}, false
	}
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.Context != parser.EffectContextController {
		return game.ObjectReference{}, false
	}
	if effect.RegenerateAttached {
		if len(ctx.content.References) != 0 {
			return game.ObjectReference{}, false
		}
		object := game.SourceAttachedPermanentReference()
		return object, len(object.Validate()) == 0
	}
	if len(ctx.content.References) != 1 {
		return game.ObjectReference{}, false
	}
	return lowerObjectReference(ctx.content.References[0], referenceLoweringContext{AllowSource: true})
}
