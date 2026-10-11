package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

func lowerFixedStackTargetAction(
	ctx contentCtx,
	targetSpec func(compiler.CompiledTarget) (game.TargetSpec, bool),
	factory func(game.ObjectReference) game.Instruction,
) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 1 || len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 || len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if !effect.Exact || effect.Negated || effect.Context != parser.EffectContextController ||
		effect.Amount.Known || effect.DelayedTiming != 0 || effect.Duration != compiler.DurationNone {
		return game.AbilityContent{}, false
	}
	subject, ok := ctx.content.TargetActionSubject(effect)
	if !ok || subject.SubjectDomain() != compiler.ReferenceSubjectStackObject ||
		subject.SubjectLifetime() != compiler.ReferenceLifetimeStackOccurrence {
		return game.AbilityContent{}, false
	}
	object, ok := lowerContentObjectReference(ctx, subject, referenceLoweringContext{AllowTarget: true})
	if !ok || object.Kind() != game.ObjectReferenceTargetStackObject {
		return game.AbilityContent{}, false
	}
	spec, ok := targetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets:  []game.TargetSpec{spec},
		Sequence: []game.Instruction{factory(object)},
	}.Ability(), true
}

func lowerStackControlAction(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 1 || ctx.content.Effects[0].Kind != compiler.EffectGainControl ||
		ctx.content.Effects[0].Optional || ctx.optional {
		return game.AbilityContent{}, false
	}
	return lowerFixedStackTargetAction(ctx, stackSpellTargetSpec, func(object game.ObjectReference) game.Instruction {
		return game.Instruction{Primitive: game.ChangeStackObjectController{
			Object: object, Controller: game.ControllerReference(),
		}}
	})
}
