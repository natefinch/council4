package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerCantBeBlockedSpell lowers the temporary combat-evasion effect "Target
// creature can't be blocked this turn." into an ApplyRule instruction that
// places a RuleEffectCantBeBlocked restriction on the single targeted creature
// for the turn (game.DurationThisTurn, removed during cleanup). It accepts only
// the exact single-creature-target shape recognized by the parser; every other
// recipient, duration, condition, mode, or reference fails closed so the broader
// "can't be blocked this turn" family stays faithful and bounded.
func lowerCantBeBlockedSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported can't-be-blocked effect",
			"the executable source backend supports only exact \"Target creature can't be blocked this turn.\"",
		)
	}
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.Context != parser.EffectContextTarget ||
		effect.Duration != compiler.DurationThisTurn ||
		ctx.optional ||
		len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Selector.Kind != compiler.SelectorCreature ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return unsupported()
	}
	targetSpec, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return unsupported()
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{
				Primitive: game.ApplyRule{
					Object: opt.Val(game.TargetPermanentReference(0)),
					RuleEffects: []game.RuleEffect{
						{Kind: game.RuleEffectCantBeBlocked},
					},
					Duration: game.DurationThisTurn,
				},
			},
		},
	}.Ability(), nil
}
