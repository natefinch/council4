package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

const repeatProcessContinueResultKey = game.ResultKey("repeat-process-continue")

// lowerRepeatProcessSpell lowers a "Repeat the following process <count> times.
// <body>" loop (EffectRepeatProcess) to a single RepeatProcess instruction whose
// Times is the repeat count (a fixed cardinal or the spell's variable X) and
// whose Body is the recursively lowered ordered process executed each iteration. It
// fails closed for any targets, conditions, keywords, or modes on the loop
// itself, for an unsupported count, an empty body, or when the body itself does
// not lower.
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
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(effect.RepeatBody) == 0 {
		return unsupported()
	}
	if effect.RepeatUntilFailure {
		return lowerRepeatUntilFailure(cardName, ctx, syntax, effect, unsupported)
	}
	if len(ctx.content.Conditions) != 0 {
		return unsupported()
	}
	times, ok := createTokenAmount(ctx, &effect, game.ObjectReference{})
	if !ok {
		return unsupported()
	}
	bodyContent := ctx.content
	bodyContent.Effects = effect.RepeatBody
	bodyCtx := ctx
	bodyCtx.content = bodyContent
	body, diagnostic := lowerContent(cardName, bodyCtx, syntax)
	if diagnostic != nil {
		return game.AbilityContent{}, diagnostic
	}
	return game.Mode{Sequence: []game.Instruction{{Primitive: game.RepeatProcess{Times: times, Body: body}}}}.Ability(), nil
}

func lowerRepeatUntilFailure(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
	effect compiler.CompiledEffect,
	unsupported func() (game.AbilityContent, *shared.Diagnostic),
) (game.AbilityContent, *shared.Diagnostic) {
	if effect.Amount.Known ||
		effect.Amount.RangeKnown ||
		effect.Amount.VariableX ||
		effect.Amount.DynamicKind != 0 ||
		len(ctx.content.Conditions) != 1 ||
		ctx.content.Conditions[0].Predicate != compiler.ConditionPredicatePriorInstructionAccepted {
		return unsupported()
	}
	bodyContent := ctx.content
	bodyContent.Effects = effect.RepeatBody
	bodyCtx := ctx
	bodyCtx.content = bodyContent
	body, diagnostic := lowerContent(cardName, bodyCtx, syntax)
	if diagnostic != nil {
		return game.AbilityContent{}, diagnostic
	}
	if body.IsModal() ||
		len(body.SharedTargets) != 0 ||
		len(body.Modes) != 1 ||
		len(body.Modes[0].Targets) != 0 ||
		len(body.Modes[0].Sequence) < 2 {
		return unsupported()
	}
	last := &body.Modes[0].Sequence[len(body.Modes[0].Sequence)-1]
	if last.Optional || last.PublishResult != "" || !last.ResultGate.Exists {
		return unsupported()
	}
	last.PublishResult = repeatProcessContinueResultKey
	return game.Mode{Sequence: []game.Instruction{{Primitive: game.RepeatProcess{
		Body:           body,
		ContinueResult: repeatProcessContinueResultKey,
	}}}}.Ability(), nil
}
