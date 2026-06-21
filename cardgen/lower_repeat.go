package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// lowerRepeatProcessSpell lowers a "Repeat the following process <count> times.
// <body>" loop (EffectRepeatProcess) to a single RepeatProcess instruction whose
// Times is the repeat count (a fixed cardinal or the spell's variable X) and
// whose Body is the recursively lowered sub-effect executed each iteration. It
// fails closed for any targets, conditions, keywords, or modes on the loop
// itself, for an unsupported count, for anything but a single body effect, or
// when the body itself does not lower.
func lowerRepeatProcessSpell(cardName string, ctx contentCtx, syntax *parser.Ability) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported repeat effect",
			"the executable source backend does not yet lower this repeated effect",
		)
	}
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(effect.RepeatBody) != 1 {
		return unsupported()
	}
	times, ok := createTokenAmount(&effect)
	if !ok {
		return unsupported()
	}
	bodyContent := ctx.content
	bodyContent.Effects = effect.RepeatBody
	bodyCtx := ctx
	bodyCtx.content = bodyContent
	body, diagnostic := lowerImmediateSingleEffectSpell(cardName, bodyCtx, syntax)
	if diagnostic != nil {
		return game.AbilityContent{}, diagnostic
	}
	return game.Mode{Sequence: []game.Instruction{{Primitive: game.RepeatProcess{Times: times, Body: body}}}}.Ability(), nil
}
