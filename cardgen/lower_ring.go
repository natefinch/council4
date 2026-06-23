package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// lowerRingTemptsSpell lowers the fixed designation effect "The Ring tempts
// you." (CR 701.51) to a RingTempts primitive scoped to the resolving
// controller. The wording carries no targets, conditions, keywords, or modes;
// any other shape fails closed.
func lowerRingTemptsSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	unsupported := contentDiagnostic(
		ctx,
		"unsupported ring tempts effect",
		"the executable source backend supports only the exact controller-scoped 'the Ring tempts you'",
	)
	if effect.Negated || ctx.optional || !effect.Exact ||
		effect.Context != parser.EffectContextController ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, unsupported
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.RingTempts{Player: game.ControllerReference()},
		}},
	}.Ability(), nil
}
