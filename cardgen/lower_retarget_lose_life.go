package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// lowerRetargetThenLoseLifeContent lowers the exact two-effect redirect body
// "Change the target of target spell with a single target. You lose life equal
// to that spell's mana value." (Imp's Mischief) into a ChooseNewTargets over the
// targeted spell followed by a controller life loss whose amount reads that same
// spell's live mana value. The retarget leaves the spell on the stack, so the
// life-loss amount is the live target mana value (DynamicAmountObjectManaValue
// over the target stack object) rather than the counter-captured value used by
// the Mana Drain family. Any other shape returns ok=false so the body falls
// through to the generic ordered-sequence path and fails closed there.
func lowerRetargetThenLoseLifeContent(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 1 {
		return game.AbilityContent{}, false
	}
	retarget := &ctx.content.Effects[0]
	loseLife := &ctx.content.Effects[1]
	if retarget.Kind != compiler.EffectChooseNewTargets ||
		!isExactMandatoryEffect(retarget) ||
		retarget.Context != parser.EffectContextController ||
		retarget.Amount.Known ||
		retarget.Connection != parser.EffectConnectionNone ||
		len(retarget.Targets) != 1 ||
		len(retarget.References) != 0 {
		return game.AbilityContent{}, false
	}
	if loseLife.Kind != compiler.EffectLose ||
		!loseLife.LifeObject ||
		!isExactMandatoryEffect(loseLife) ||
		loseLife.Context != parser.EffectContextController ||
		loseLife.Amount.Known ||
		loseLife.Amount.DynamicKind != compiler.DynamicAmountSourceManaValue ||
		loseLife.Amount.DynamicForm != compiler.DynamicAmountEqual ||
		loseLife.Amount.Multiplier != 1 ||
		len(loseLife.Targets) != 0 ||
		!referencesBindTo(loseLife.References, compiler.ReferenceBindingTarget, 0) {
		return game.AbilityContent{}, false
	}
	targetSpec, ok := counterTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{Primitive: game.ChooseNewTargets{Object: game.TargetStackObjectReference(0)}},
			{Primitive: game.LoseLife{
				Player: game.ControllerReference(),
				Amount: game.Dynamic(game.DynamicAmount{
					Kind:   game.DynamicAmountObjectManaValue,
					Object: game.TargetStackObjectReference(0),
				}),
			}},
		},
	}.Ability(), true
}
