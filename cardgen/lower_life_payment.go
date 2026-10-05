package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/opt"
)

func lowerLifePaymentClause(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if !effect.Exact || !effect.LifeObject ||
		effect.LifePayment != parser.EffectLifePaymentOptional ||
		effect.Context != parser.EffectContextController || effect.Player != parser.EffectPlayerNone ||
		effect.Negated || effect.DelayedTiming != 0 || effect.Duration != 0 ||
		!effect.Amount.Known || effect.Amount.Value <= 0 ||
		len(ctx.content.Effects) != 1 || len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Targets) != 0 || len(ctx.content.References) != 0 ||
		len(ctx.content.Keywords) != 0 || len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, contentDiagnostic(ctx, "unsupported life payment",
			"resolution life payment requires one optional fixed controller cost")
	}
	return game.Mode{Sequence: []game.Instruction{{Primitive: game.Pay{
		Payment: game.ResolutionPayment{
			Payer: opt.Val(game.ControllerReference()),
			AdditionalCosts: []cost.Additional{{
				Kind: cost.AdditionalPayLife, Amount: effect.Amount.Value,
			}},
		},
	}}}}.Ability(), nil
}
