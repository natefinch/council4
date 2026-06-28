package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// isSequentialReanimationTypeColorGrantEffect reports whether effect is the
// permanent referenced-object type-and-color grant a reanimation rider applies
// to the creature an earlier clause in the same sequence returned to the
// battlefield ("That creature is a black Zombie in addition to its other colors
// and types." — Rise from the Grave; "It's a Phyrexian in addition to its other
// types." — Portal to Phyrexia). "That creature"/"It" binds to that earlier
// permanent. The grant adds colors, card types, and/or creature subtypes for the
// permanent's lifetime on the battlefield, so the duration is absent rather than
// until end of turn (the targeted Liquimetal form sets BecomeTypeUntilEndOfTurn
// and is lowered through the ordinary target path).
func isSequentialReanimationTypeColorGrantEffect(effect *compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectBecomeType &&
		!effect.Negated &&
		!effect.Optional &&
		!effect.BecomeTypeUntilEndOfTurn &&
		effect.Context == parser.EffectContextReferencedObject &&
		(len(effect.BecomeTypeAddTypes) != 0 ||
			len(effect.BecomeTypeAddSubtypes) != 0 ||
			len(effect.BecomeTypeAddColors) != 0) &&
		referencesBindTo(effect.References, compiler.ReferenceBindingTarget, 0)
}

// lowerSequentialReanimationTypeColorGrant lowers a permanent "That creature is
// a <color> <type> in addition to its other [colors and] types." clause that
// adds colors, card types, and/or creature subtypes to the permanent an earlier
// clause in the same sequence returned to the battlefield (Rise from the Grave,
// Liliana, Death's Majesty, Portal to Phyrexia). "That creature"/"It" binds to
// that earlier permanent, which (for a reanimation) is a freshly created object a
// plain target-permanent reference cannot resolve, so the lowering reuses the
// linked key under which an earlier clause already recorded the permanent, or
// rewrites the immediately-prior instruction to publish it. The grant is an
// ApplyContinuous bound to that linked object, applying a LayerColor color
// addition and a LayerType type/subtype addition for the permanent's lifetime on
// the battlefield. It returns the (possibly rewritten) prior publishing primitive
// and the grant content, or false to fail closed so the caller lowers the clause
// normally.
func lowerSequentialReanimationTypeColorGrant(
	effectIndex int,
	ctx contentCtx,
	sequence []game.Instruction,
) (game.Primitive, game.AbilityContent, bool) {
	if effectIndex == 0 ||
		len(sequence) != effectIndex ||
		len(ctx.content.Effects) != 1 ||
		ctx.optional ||
		!isSequentialReanimationTypeColorGrantEffect(&ctx.content.Effects[0]) {
		return nil, game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	consumed := ctx
	consumed.content.References = nil
	consumed.content.Targets = nil
	if consumed.content.Unconsumed() {
		return nil, game.AbilityContent{}, false
	}
	key, publisher, ok := reuseOrPublishLinkedPermanent(effectIndex, sequence)
	if !ok {
		return nil, game.AbilityContent{}, false
	}
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
		TargetLinkedKey: key,
	})
	if !ok {
		return nil, game.AbilityContent{}, false
	}
	var continuousEffects []game.ContinuousEffect
	if len(effect.BecomeTypeAddColors) != 0 {
		continuousEffects = append(continuousEffects, game.ContinuousEffect{
			Layer:     game.LayerColor,
			AddColors: append([]color.Color(nil), effect.BecomeTypeAddColors...),
		})
	}
	if len(effect.BecomeTypeAddTypes) != 0 || len(effect.BecomeTypeAddSubtypes) != 0 {
		continuousEffects = append(continuousEffects, game.ContinuousEffect{
			Layer:       game.LayerType,
			AddTypes:    append([]types.Card(nil), effect.BecomeTypeAddTypes...),
			AddSubtypes: append([]types.Sub(nil), effect.BecomeTypeAddSubtypes...),
		})
	}
	grant := game.ApplyContinuous{
		Object:            opt.Val(object),
		ContinuousEffects: continuousEffects,
		Duration:          game.DurationPermanent,
	}
	return publisher, game.Mode{Sequence: []game.Instruction{{Primitive: grant}}}.Ability(), true
}
