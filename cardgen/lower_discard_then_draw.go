package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// lowerDiscardThenDrawSpell lowers a discard clause that the parser fused with a
// following "then draw that many cards[ plus K]" clause into a single variable
// looter primitive. The compiler carries the upper bound (DiscardThenDrawMax,
// zero meaning "any number") and the draw offset as typed parameters; this
// lowering reads only those typed values, so it never inspects Oracle words. The
// fused clause is the controller's own looter, so it accepts only a non-negated,
// non-optional, controller-context effect with no leftover targets, references,
// conditions, keywords, or modes.
func lowerDiscardThenDrawSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if effect.Negated ||
		effect.Optional ||
		ctx.optional ||
		effect.Context != parser.EffectContextController ||
		effect.DiscardThenDrawMax < 0 ||
		effect.DiscardThenDrawOffset < 0 ||
		ctx.content.Unconsumed() {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported discard-then-draw spell",
			"the executable source backend supports only a controller variable looter "+
				"(discard up to N or any number of cards, then draw that many)",
		)
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.DiscardThenDraw{
			Player:     game.ControllerReference(),
			Max:        effect.DiscardThenDrawMax,
			DrawOffset: effect.DiscardThenDrawOffset,
		},
	}}}.Ability(), nil
}
