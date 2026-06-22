package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// lowerControllerDrawLoseShareXSpell handles the two-clause controller pattern
// "you draw X cards and lose X life, where X is <dynamic>", where both clauses
// move the same variable amount X that a single "where X is" definition fixes
// for the whole sentence. The Speed Demon's end-step ability is the anchor:
// "you draw X cards and lose X life, where X is your speed." emits a Draw and a
// LoseLife for the controller that both read the controller's speed.
//
// The compiler attaches the "where X is" definition to only the trailing clause
// (the lose), leaving the leading draw with the bare variable X. This lowerer
// resolves the shared dynamic amount from the definition clause and applies it
// to both, so the draw is not left reading an unbound X. It is restricted to the
// dynamic "where X is" form (not equal fixed amounts) so it never intercepts
// plain fixed draw/lose sequences that already lower as independent effects.
func lowerControllerDrawLoseShareXSpell(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Targets) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		ctx.optional {
		return game.AbilityContent{}, false
	}
	draw := &ctx.content.Effects[0]
	lose := &ctx.content.Effects[1]
	if draw.Kind != compiler.EffectDraw ||
		lose.Kind != compiler.EffectLose ||
		draw.Context != parser.EffectContextController ||
		lose.Context != parser.EffectContextController ||
		lose.Connection != parser.EffectConnectionAnd ||
		draw.Negated || lose.Negated ||
		!draw.Exact || !lose.Exact ||
		len(draw.Targets) != 0 || len(lose.Targets) != 0 {
		return game.AbilityContent{}, false
	}
	amount, ok := controllerSharedDynamicAmount(draw, lose, ctx.content.References)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{Primitive: game.Draw{Player: game.ControllerReference(), Amount: amount}},
			{Primitive: game.LoseLife{Player: game.ControllerReference(), Amount: amount}},
		},
	}.Ability(), true
}

// controllerSharedDynamicAmount resolves the single dynamic Quantity X that both
// sentence clauses share. Exactly one clause must carry the "where X is
// <dynamic>" definition and the other the bare variable X; equal fixed amounts
// are intentionally rejected so this path only owns the shared-dynamic form.
func controllerSharedDynamicAmount(
	draw, lose *compiler.CompiledEffect,
	references []compiler.CompiledReference,
) (game.Quantity, bool) {
	definition, bare := drainDefinitionAndBare(draw, lose)
	if definition == nil ||
		definition.Amount.DynamicForm != compiler.DynamicAmountWhereX ||
		!bare.Amount.VariableX ||
		bare.Amount.DynamicKind != compiler.DynamicAmountNone {
		return game.Quantity{}, false
	}
	if !drainDynamicReferencesSafe(references, definition.Amount.DynamicKind) {
		return game.Quantity{}, false
	}
	dynamic, ok := lowerDynamicAmount(definition.Amount, game.SourcePermanentReference())
	if !ok {
		return game.Quantity{}, false
	}
	return game.Dynamic(dynamic), true
}
