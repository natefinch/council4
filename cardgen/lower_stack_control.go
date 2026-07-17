package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// lowerGainControlStackSpellSequence lowers "Gain control of target spell. You
// may choose new targets for it." as an ordered controller change followed by an
// optional retarget of the same stack object.
func lowerGainControlStackSpellSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 1 {
		return game.AbilityContent{}, false
	}
	gain := &ctx.content.Effects[0]
	retarget := &ctx.content.Effects[1]
	if gain.Kind != compiler.EffectGainControl ||
		!isExactMandatoryEffect(gain) ||
		gain.Context != parser.EffectContextController ||
		gain.Duration != compiler.DurationNone ||
		len(gain.Targets) != 1 ||
		len(gain.References) != 0 {
		return game.AbilityContent{}, false
	}
	if retarget.Kind != compiler.EffectChooseNewTargets ||
		!retarget.Exact ||
		retarget.Negated ||
		!retarget.Optional ||
		retarget.Context != parser.EffectContextController ||
		retarget.Duration != compiler.DurationNone ||
		retarget.DelayedTiming != 0 ||
		len(retarget.Targets) != 0 ||
		!referencesBindTo(retarget.References, compiler.ReferenceBindingTarget, 0) {
		return game.AbilityContent{}, false
	}
	targetSpec, ok := stackSpellTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{Primitive: game.ChangeStackObjectController{
				Object:     game.TargetStackObjectReference(0),
				Controller: game.ControllerReference(),
			}},
			{
				Primitive: game.ChooseNewTargets{Object: game.TargetStackObjectReference(0)},
				Optional:  true,
			},
		},
	}.Ability(), true
}
