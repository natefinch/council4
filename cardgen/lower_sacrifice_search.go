package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerOptionalSacrificeThenSearchSequence lowers the resolving-optional
// "You may sacrifice <permanent>. If you do, <library-search group>." family
// (The Huntsman's Redemption chapter II, Blood Speaker, Sanctum of Ugin). Unlike
// the mandatory cost-prelude shape handled by lowerSacrificeThenSearchSequence,
// the leading sacrifice is an optional resolving effect: the controller chooses
// whether to perform it, and the search runs only when they did. Optionality is
// realized through the shared optional-flow envelope — the sacrifice instruction
// is marked Optional and publishes its result, and the single search instruction
// is gated on that result having succeeded.
//
// The search group spans several effects (search, reveal, put, then shuffle)
// that the per-effect ordered-sequence loop cannot split, so this dedicated
// lowerer groups them via splitSequenceSearchGroups and emits one game.Search.
// It fails closed unless the body is exactly one optional sacrifice followed by
// one "if you do"-gated search group with no other conditions, modes, targets,
// or keywords.
func lowerOptionalSacrificeThenSearchSequence(ctx contentCtx) (game.AbilityContent, bool) {
	content := ctx.content
	if ctx.optional ||
		len(content.Modes) != 0 ||
		len(content.Targets) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.Effects) < 4 ||
		content.Effects[0].Kind != compiler.EffectSacrifice ||
		!content.Effects[0].Optional {
		return game.AbilityContent{}, false
	}
	plan, ok := planOptionalFlow(content)
	if !ok ||
		!plan.enabled ||
		plan.publishWithoutOptional ||
		plan.optionalIndex != 0 ||
		plan.gateIndex != 1 ||
		plan.elseIndex >= 0 ||
		plan.bareIndex >= 0 {
		return game.AbilityContent{}, false
	}
	// The gate condition is consumed by the optional flow; no other per-effect
	// condition may survive, or one would be silently dropped.
	if len(optionalFlowGateConditions(content.Conditions, plan)) != 0 {
		return game.AbilityContent{}, false
	}
	// Every reference must belong either to the leading sacrifice clause (the
	// "this creature" self-reference the sacrifice lowerer resolves) or to the
	// gated search clause (the tutor's "reveal it"/"put it" pronouns the
	// search-group shape consumes). A reference outside both spans would be
	// silently dropped, so fail closed.
	clauseSpans := []shared.Span{content.Effects[0].Span, content.Effects[1].Span}
	for i := range content.References {
		if !spanCovered(content.References[i].Span, clauseSpans) {
			return game.AbilityContent{}, false
		}
	}
	searchEffects := searchGroupEffectsWithoutGatePrefix(content.Effects[1:], content.Conditions[plan.gateCondition].Span)
	groups, _, ok := splitSequenceSearchGroups(searchEffects)
	if !ok || len(groups) != 1 {
		return game.AbilityContent{}, false
	}
	sacrifice, ok := lowerSacrificeLeadInstruction(ctx)
	if !ok {
		return game.AbilityContent{}, false
	}
	sacrificeSeq := []game.Instruction{sacrifice}
	if !applyOptionalFlowPublish(sacrificeSeq) {
		return game.AbilityContent{}, false
	}
	searchSeq, ok := searchGroupInstructions(groups[0])
	if !ok || !applyOptionalFlowGate(searchSeq, game.TriTrue) {
		return game.AbilityContent{}, false
	}
	sequence := make([]game.Instruction, 0, len(sacrificeSeq)+len(searchSeq))
	sequence = append(sequence, sacrificeSeq...)
	sequence = append(sequence, searchSeq...)
	return game.Mode{Sequence: sequence}.Ability(), true
}

// searchGroupEffectsWithoutGatePrefix returns a copy of a gated search group's
// effects with the leading search effect's span trimmed past the "if you do"
// gate condition. In trigger-body context the gate condition prefixes the
// leading search effect's span, widening it past the reveal/put/shuffle effects
// that follow; searchGroupSpec's shared-span membership guard would then reject
// the group. When the leading effect's span begins at the gate condition, it is
// realigned to the following effect's span start (the rest of the search
// sentence), which restores the uniform span searchGroupSpec expects without
// altering the recognized search shape (which keys on effect kinds and
// connections, not spans).
func searchGroupEffectsWithoutGatePrefix(effects []compiler.CompiledEffect, gateSpan shared.Span) []compiler.CompiledEffect {
	trimmed := slices.Clone(effects)
	if len(trimmed) >= 2 &&
		trimmed[0].Span.Start.Offset == gateSpan.Start.Offset &&
		trimmed[0].Span.End == trimmed[1].Span.End &&
		trimmed[0].Span.Start.Offset < trimmed[1].Span.Start.Offset {
		trimmed[0].Span.Start = trimmed[1].Span.Start
	}
	return trimmed
}

// lowerSacrificeLeadInstruction lowers the leading sacrifice effect of a
// sacrifice-then-search sequence to its single runtime instruction. It passes the
// effect's own references so a self-referential sacrifice ("sacrifice this
// creature", the Cabaretti Courtyard "sacrifice it" reflexive backreference)
// resolves, and it accepts either the source-bound Sacrifice primitive or the
// chosen-permanent SacrificePermanents primitive. It fails closed unless the
// effect lowers to exactly one such instruction with no target.
func lowerSacrificeLeadInstruction(ctx contentCtx) (game.Instruction, bool) {
	effect := ctx.content.Effects[0]
	sacrificeCtx := ctx
	sacrificeCtx.content = compiler.AbilityContent{
		Effects:    []compiler.CompiledEffect{effect},
		References: effect.References,
	}
	content, diagnostic := lowerSacrificeSpell(sacrificeCtx)
	if diagnostic != nil ||
		len(content.Modes) != 1 ||
		len(content.Modes[0].Targets) != 0 ||
		len(content.Modes[0].Sequence) != 1 {
		return game.Instruction{}, false
	}
	instruction := content.Modes[0].Sequence[0]
	if instruction.Primitive == nil {
		return game.Instruction{}, false
	}
	kind := instruction.Primitive.Kind()
	if kind != game.PrimitiveSacrifice && kind != game.PrimitiveSacrificePermanents {
		return game.Instruction{}, false
	}
	return instruction, true
}

// lowerSacrificeThenSearchSequence lowers the "Sacrifice <permanent>.
// <library-search group>" cost-prelude family, optionally followed by a single
// conditional "instead" alternative that performs a larger search when a
// resolving condition holds (Entish Restoration, Harrow-style ramp). The leading
// sacrifice is a non-optional cost that always resolves; the search group(s)
// follow. When two groups are present the second carries an "instead"
// replacement gated on one resolving condition, so the first group runs on the
// negation and the second on the condition — exactly one performs the search.
//
// A single trailing rider may close the sentence after the last group's shuffle
// — the Cabaretti Courtyard tapped-fetch land cycle ends "..., then shuffle and
// you gain 1 life." (the "When this land enters, sacrifice it. When you do,
// search ..." form, whose mandatory reflexive trigger flattens into this
// sacrifice-then-search shape). The rider is lowered as its own instruction
// after the search, mirroring the post-search rider handling in lowerSearchSpell.
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
	// Every reference must belong either to the leading sacrifice clause (the
	// "it"/"this land" self-reference the sacrifice lowerer resolves) or to the
	// trailing search/rider clauses (the tutor's "put it" pronoun the
	// search-group shape consumes, bound as a prior-instruction result). A
	// reference outside the sacrifice clause that is not a prior-instruction
	// result would be silently dropped, so fail closed.
	for i := range content.References {
		if content.References[i].Binding == compiler.ReferenceBindingPriorInstructionResult {
			continue
		}
		if spanCovered(content.References[i].Span, []shared.Span{content.Effects[0].Span}) {
			continue
		}
		return game.AbilityContent{}, false
	}
	searchEffects := content.Effects[1:]
	trailingRider, searchEffects, hasRider := peelTrailingSearchRider(searchEffects)
	groups, starts, ok := splitSequenceSearchGroups(searchEffects)
	if !ok || len(groups) < 1 || len(groups) > 2 {
		return game.AbilityContent{}, false
	}

	sacrifice, ok := lowerSacrificeLeadInstruction(ctx)
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
		if hasRider {
			sequence = append(sequence, trailingRider)
		}
		return game.Mode{Sequence: sequence}.Ability(), true
	}
	// The conditional "instead" two-group form is not modeled with a trailing
	// rider, so reject that pairing rather than silently dropping the rider.
	if hasRider {
		return game.AbilityContent{}, false
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

// peelTrailingSearchRider removes a single rider effect that closes a
// sacrifice-then-search sentence after the final group's shuffle — the "you gain
// N life" reward on the Cabaretti Courtyard tapped-fetch land cycle. The rider
// must immediately follow a shuffle effect and lower through lowerSearchRider
// (the same fixed controller life-change or random-discard riders the spell
// search path lowers after its group). It returns the rider instruction, the
// effects with the rider removed, and true when a rider is present; otherwise it
// returns the unchanged effects and false so the no-rider shape is unaffected.
func peelTrailingSearchRider(effects []compiler.CompiledEffect) (game.Instruction, []compiler.CompiledEffect, bool) {
	n := len(effects)
	if n < 2 || effects[n-2].Kind != compiler.EffectShuffle {
		return game.Instruction{}, effects, false
	}
	inst, ok := lowerSearchRider(&effects[n-1])
	if !ok {
		return game.Instruction{}, effects, false
	}
	return inst, effects[:n-1], true
}

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
