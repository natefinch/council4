package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerConditionalInsteadSearchSequence lowers two complete library-search
// groups where the second is a conditional "instead" replacement. The groups
// retain their independently typed selectors, quantities, destinations, entry
// state, and shuffle behavior; only their mutually exclusive gates are added.
func lowerConditionalInsteadSearchSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 {
		return game.AbilityContent{}, false
	}
	sequence, condition, ok := conditionalInsteadSearchInstructions(
		ctx.content.Effects,
		ctx.content.Conditions,
	)
	if !ok {
		return game.AbilityContent{}, false
	}
	for i := range ctx.content.References {
		reference := ctx.content.References[i]
		if reference.Binding == compiler.ReferenceBindingPriorInstructionResult ||
			spanCovered(reference.Span, []shared.Span{condition.Span}) {
			continue
		}
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: sequence}.Ability(), true
}

// conditionalInsteadSearchInstructions is shared by search-first spells and
// sacrifice-then-search sequences. It consumes exactly two complete search
// groups and one resolving condition contained by the replacement search.
func conditionalInsteadSearchInstructions(
	effects []compiler.CompiledEffect,
	conditions []compiler.CompiledCondition,
) ([]game.Instruction, compiler.CompiledCondition, bool) {
	groups, starts, ok := splitSequenceSearchGroups(effects)
	if !ok || len(groups) != 2 || len(conditions) != 1 {
		return nil, compiler.CompiledCondition{}, false
	}
	insteadSearch := &effects[starts[1]]
	if insteadSearch.Replacement.Kind != parser.EffectReplacementInstead {
		return nil, compiler.CompiledCondition{}, false
	}
	condition := conditions[0]
	if !spanCovered(condition.Span, []shared.Span{insteadSearch.Span}) {
		return nil, compiler.CompiledCondition{}, false
	}
	lowered, ok := lowerCondition(condition, conditionContextEffectGate)
	if !ok {
		return nil, compiler.CompiledCondition{}, false
	}
	gate := game.EffectCondition{Condition: opt.Val(lowered)}
	negated, ok := negatedEffectCondition(&gate)
	if !ok {
		return nil, compiler.CompiledCondition{}, false
	}
	base, ok := searchGroupInstructions(groups[0])
	if !ok || !applyEffectConditionGate(base, &negated) {
		return nil, compiler.CompiledCondition{}, false
	}
	replacement, ok := searchGroupInstructions(groups[1])
	if !ok || !applyEffectConditionGate(replacement, &gate) {
		return nil, compiler.CompiledCondition{}, false
	}
	return append(base, replacement...), condition, true
}
