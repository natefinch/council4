package cardgen

import (
	"slices"

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
// Targets belong to the body but are announced once by the enclosing mode.
// Conditions, keywords, modes, unsupported counts, and bodies fail closed.
func lowerRepeatProcessSpell(cardName string, ctx contentCtx, syntax *parser.Ability) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported repeat effect",
			"the executable source backend does not yet lower this repeated effect",
		)
	}
	effect := ctx.content.Effects[0]
	if !effect.Exact || effect.Negated ||
		effect.Context != parser.EffectContextController ||
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
	times, ok := repeatProcessQuantity(effect.Amount)
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
	content, ok := repeatProcessContent(times, body)
	if !ok {
		return unsupported()
	}
	return content, nil
}

func repeatProcessQuantity(amount compiler.CompiledAmount) (game.Quantity, bool) {
	if amount.RangeKnown || amount.AnyNumber ||
		amount.Known && (amount.VariableX || amount.DynamicKind != compiler.DynamicAmountNone) ||
		amount.VariableX && amount.DynamicKind != compiler.DynamicAmountNone {
		return game.Quantity{}, false
	}
	if amount.Known || amount.VariableX {
		if amount.DynamicForm != compiler.DynamicAmountFormNone ||
			amount.Multiplier != 0 || amount.Addend != 0 {
			return game.Quantity{}, false
		}
		if amount.Known && amount.Value == 0 {
			return game.Fixed(0), true
		}
	}
	return positiveQuantity(amount, true)
}

func repeatProcessContent(times game.Quantity, body game.AbilityContent) (game.AbilityContent, bool) {
	if body.IsModal() || len(body.SharedTargets) != 0 || len(body.Modes) != 1 {
		return game.AbilityContent{}, false
	}
	targets := body.Modes[0].Targets
	body.Modes = slices.Clone(body.Modes)
	body.Modes[0].Targets = nil
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{{
			Primitive: game.RepeatProcess{Times: times, Body: body},
		}},
	}.Ability(), true
}

func transformRepeatProcessTargets(value game.RepeatProcess, transform targetIndexTransform) (game.Primitive, bool) {
	if value.Body.IsModal() || len(value.Body.SharedTargets) != 0 || len(value.Body.Modes) != 1 ||
		len(value.Body.Modes[0].Targets) != 0 {
		return nil, false
	}
	times, ok := transformQuantity(value.Times, transform)
	if !ok {
		return nil, false
	}
	value.Times = times
	value.Body.Modes = slices.Clone(value.Body.Modes)
	value.Body.Modes[0].Sequence = slices.Clone(value.Body.Modes[0].Sequence)
	if !transformTargetedSequence(value.Body.Modes[0].Sequence, transform) {
		return nil, false
	}
	if !transformRepeatEnvelopes(value.Body.Modes[0].Sequence, transform) {
		return nil, false
	}
	return value, true
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
