package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerSacrificeThenSearchSequence lowers the "Sacrifice <permanent>.
// <library-search group>" cost-prelude family, optionally followed by a single
// conditional "instead" alternative that performs a larger search when a
// resolving condition holds (Entish Restoration, Harrow-style ramp). The leading
// sacrifice is a non-optional cost that always resolves; the search group(s)
// follow. When two groups are present the second carries an "instead"
// replacement gated on one resolving condition, so the first group runs on the
// negation and the second on the condition — exactly one performs the search.
//
// Each search group is a multi-effect run (search, put, then shuffle) that the
// per-effect ordered-sequence loop cannot split, so this dedicated lowerer
// groups them via searchGroupSpec and emits one game.Search per group.
func lowerSacrificeThenSearchSequence(ctx contentCtx) (game.AbilityContent, bool) {
	content := ctx.content
	if ctx.optional ||
		len(content.Modes) != 0 ||
		len(content.Targets) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.Effects) < 4 ||
		content.Effects[0].Kind != compiler.EffectSacrifice {
		return game.AbilityContent{}, false
	}
	for i := range content.References {
		if content.References[i].Binding != compiler.ReferenceBindingPriorInstructionResult {
			return game.AbilityContent{}, false
		}
	}
	groups, starts, ok := splitSequenceSearchGroups(content.Effects[1:])
	if !ok || len(groups) < 1 || len(groups) > 2 {
		return game.AbilityContent{}, false
	}

	sacrifice, ok := lowerSequenceSacrificeInstruction(ctx)
	if !ok {
		return game.AbilityContent{}, false
	}
	sequence := []game.Instruction{sacrifice}

	if len(groups) == 1 {
		if len(content.Conditions) != 0 {
			return game.AbilityContent{}, false
		}
		searchSeq, ok := searchGroupInstructions(groups[0])
		if !ok {
			return game.AbilityContent{}, false
		}
		sequence = append(sequence, searchSeq...)
		return game.Mode{Sequence: sequence}.Ability(), true
	}

	// Two groups: the second is a conditional "instead" replacement of the
	// first, gated on exactly one resolving condition contained in its clause.
	insteadStart := starts[1] + 1
	insteadSearch := &content.Effects[insteadStart]
	if insteadSearch.Replacement.Kind != parser.EffectReplacementInstead ||
		len(content.Conditions) != 1 {
		return game.AbilityContent{}, false
	}
	condition := content.Conditions[0]
	if !spanCovered(condition.Span, []shared.Span{insteadSearch.Span}) {
		return game.AbilityContent{}, false
	}
	lowered, ok := lowerCondition(condition, conditionContextEffectGate)
	if !ok {
		return game.AbilityContent{}, false
	}
	gate := game.EffectCondition{Condition: opt.Val(lowered)}
	negated, ok := negatedEffectCondition(&gate)
	if !ok {
		return game.AbilityContent{}, false
	}
	baseSeq, ok := searchGroupInstructions(groups[0])
	if !ok || !applyEffectConditionGate(baseSeq, &negated) {
		return game.AbilityContent{}, false
	}
	insteadSeq, ok := searchGroupInstructions(groups[1])
	if !ok || !applyEffectConditionGate(insteadSeq, &gate) {
		return game.AbilityContent{}, false
	}
	sequence = append(sequence, baseSeq...)
	sequence = append(sequence, insteadSeq...)
	return game.Mode{Sequence: sequence}.Ability(), true
}

// splitSequenceSearchGroups partitions a run of effects into consecutive
// library-search groups (each a search/put/then-shuffle run terminated by a
// shuffle effect), returning each group's spec and its start index within the
// supplied slice. It fails closed if any effect is not part of an exact search
// group.
func splitSequenceSearchGroups(effects []compiler.CompiledEffect) ([]searchGroup, []int, bool) {
	var groups []searchGroup
	var starts []int
	start := 0
	for start < len(effects) {
		if effects[start].Kind != compiler.EffectSearch {
			return nil, nil, false
		}
		end := -1
		for j := start; j < len(effects); j++ {
			if effects[j].Kind == compiler.EffectShuffle {
				end = j
				break
			}
		}
		if end < 0 {
			return nil, nil, false
		}
		group, ok := searchGroupSpec(effects[start : end+1])
		if !ok || group.Length != end+1-start {
			return nil, nil, false
		}
		groups = append(groups, group)
		starts = append(starts, start)
		start = end + 1
	}
	return groups, starts, true
}

// searchGroupInstructions builds the runtime instructions for one library-search
// group: a single Search primitive carrying the group spec and fixed count. A
// group carrying an in-clause rider is not modeled here and fails closed.
func searchGroupInstructions(group searchGroup) ([]game.Instruction, bool) {
	if group.RiderIndex != 0 {
		return nil, false
	}
	return []game.Instruction{{Primitive: game.Search{
		Player: game.ControllerReference(),
		Spec:   group.Spec,
		Amount: game.Fixed(group.Amount),
	}}}, true
}
