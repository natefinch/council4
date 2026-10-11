package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func lowerHandLibraryPutClause(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic, bool) {
	effect := ctx.content.Effects[0]
	put := effect.HandLibraryPut
	if put.UnsupportedParameter != "" {
		return game.AbilityContent{}, &shared.Diagnostic{
			Severity: shared.SeverityWarning,
			Summary:  "unsupported hand-library " + put.UnsupportedParameter,
			Detail:   "hand-to-library placement requires a positive literal or bare X card count, your own hand and library, top or bottom position, and player-chosen order",
			Span:     put.ParameterSpan,
		}, true
	}
	if !put.Present {
		return game.AbilityContent{}, nil, false
	}
	if !effect.Exact || effect.Negated || effect.Divided ||
		effect.Duration != 0 || effect.DelayedTiming != 0 ||
		effect.Context != parser.EffectContextController ||
		effect.FromZone != zone.Hand || effect.ToZone != zone.Library ||
		len(ctx.content.Targets) != 0 || len(ctx.content.References) != 0 ||
		len(effect.References) != 0 || len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 || len(ctx.content.Conditions) != 0 {
		return game.AbilityContent{}, contentDiagnostic(ctx, "unsupported hand-library clause",
			"hand-to-library placement requires an exact independent controller clause with no unconsumed references or riders"), true
	}
	count := effect.Amount
	if count.RangeKnown || count.AnyNumber ||
		count.DynamicKind != compiler.DynamicAmountNone || count.DynamicForm != compiler.DynamicAmountFormNone ||
		count.Multiplier != 0 || count.Addend != 0 ||
		count.Known == count.VariableX || count.Known && count.Value <= 0 {
		return game.AbilityContent{}, contentDiagnostic(ctx, "unsupported hand-library amount",
			"hand-to-library placement requires a positive literal or bare X card count"), true
	}
	amount, ok := positiveQuantity(count, true)
	if !ok {
		return game.AbilityContent{}, contentDiagnostic(ctx, "unsupported hand-library amount",
			"hand-to-library placement requires a positive literal or bare X card count"), true
	}
	return game.Mode{Sequence: []game.Instruction{{Primitive: game.MoveCard{
		Player:            game.ControllerReference(),
		Amount:            amount,
		FromZone:          zone.Hand,
		Destination:       zone.Library,
		DestinationBottom: put.Bottom,
	}}}}.Ability(), nil, true
}

func handLibraryResolvingCostDiagnostic(content compiler.AbilityContent) *shared.Diagnostic {
	for _, condition := range content.Conditions {
		if condition.Predicate != compiler.ConditionPredicatePriorInstructionAccepted &&
			condition.Predicate != compiler.ConditionPredicatePriorInstructionNotAccepted {
			continue
		}
		for _, effect := range content.Effects {
			if !effect.HandLibraryPut.Present || effect.ClauseID <= 0 ||
				effect.ClauseID != condition.Ownership.ResultProducerClauseID {
				continue
			}
			return &shared.Diagnostic{
				Severity: shared.SeverityWarning,
				Summary:  "unsupported hand-library resolving cost",
				Detail:   "hand placement is best-effort, not an all-or-nothing resolving cost; partial payment and zero-card payment cannot gate an If/When you do consequence",
				Span:     effect.ClauseSpan,
			}
		}
	}
	return nil
}
