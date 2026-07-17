package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

func lowerMustBeBlockedSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported must-be-blocked effect",
			"the executable source backend supports only an exact creature target or back-reference that must be blocked this combat if able",
		)
	}
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.Duration != compiler.DurationThisCombat ||
		ctx.optional ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return unsupported()
	}

	var object game.ObjectReference
	switch {
	case effect.Context == parser.EffectContextTarget &&
		len(ctx.content.Targets) == 1 &&
		len(ctx.content.References) == 0 &&
		creatureTargetSubject(ctx.content.Targets[0]):
		object = game.TargetPermanentReference(0)
	case effect.Context == parser.EffectContextReferencedObject &&
		len(ctx.content.Targets) == 0 &&
		len(ctx.content.References) == 1:
		var ok bool
		object, ok = lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
			AllowSource: true,
			AllowEvent:  true,
			AllowTarget: true,
		})
		if !ok {
			return unsupported()
		}
	default:
		return unsupported()
	}

	mode := game.Mode{Sequence: []game.Instruction{{
		Primitive: game.ApplyRule{
			Object:      opt.Val(object),
			RuleEffects: []game.RuleEffect{{Kind: game.RuleEffectMustBeBlocked}},
			Duration:    game.DurationUntilEndOfCombat,
		},
	}}}
	if len(ctx.content.Targets) == 1 {
		target, ok := permanentTargetSpecWithCardinality(ctx.content.Targets[0])
		if !ok || target.MinTargets != 1 || target.MaxTargets != 1 {
			return unsupported()
		}
		mode.Targets = []game.TargetSpec{target}
	}
	return mode.Ability(), nil
}
