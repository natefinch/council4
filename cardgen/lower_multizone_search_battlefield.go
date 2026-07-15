package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// searchedLibraryResultKey links the multi-zone search's published
// "searched-library" result to the conditional shuffle that consumes it.
const searchedLibraryResultKey = game.ResultKey("multizone-search-library")

// lowerMultiZoneSearchToBattlefieldSequence lowers the composite spell
//
//	Search your library and/or graveyard for a creature card with mana value X
//	or less and put it onto the battlefield. If you search your library this way,
//	shuffle. If X is 10 or more, creatures you control get +X/+X and gain haste
//	until end of turn.
//
// (Finale of Devastation) into three instructions: one multi-zone battlefield
// search whose published result records whether the library was searched, a
// ShuffleLibrary gated on that result (the "If you search your library this way,
// shuffle." step), and an X-gated group pump-and-haste rider. It recognizes the
// shape structurally — a graveyard search bearing the creature/mana-value filter,
// a battlefield put, a library-search marker, a shuffle, and a modify/gain pump
// pair — attaching each condition to its sentence group by span containment, so
// no card text or name is inspected. It fails closed for any other shape, keeping
// unsupported multi-zone searches out of the generated corpus.
func lowerMultiZoneSearchToBattlefieldSequence(ctx contentCtx) (game.AbilityContent, bool) {
	content := ctx.content
	if len(content.Effects) != 6 ||
		len(content.Conditions) != 2 ||
		len(content.Keywords) != 1 ||
		len(content.Targets) != 0 ||
		len(content.Modes) != 0 ||
		ctx.optional {
		return game.AbilityContent{}, false
	}
	// The "it" in "put it onto the battlefield" binds to the search's found card;
	// folding the put into the search's battlefield destination consumes it.
	if len(content.References) != 1 ||
		content.References[0].Binding != compiler.ReferenceBindingPriorInstructionResult {
		return game.AbilityContent{}, false
	}
	spec, amount, ok := multiZoneSearchBattlefieldSpec(content.Effects)
	if !ok {
		return game.AbilityContent{}, false
	}
	// "creatures you control get +X/+X and gain haste until end of turn." over
	// effects[4..5] and the Haste keyword, reusing the shared group-pump builder.
	pump, ok := groupTemporaryPTKeywordContinuousEffects(content.Effects[4:6], content.Keywords)
	if !ok {
		return game.AbilityContent{}, false
	}
	// Conditions attach to their sentence group by span containment (text-blind):
	// the shuffle gate ("If you search your library this way") lands in the
	// library-search/shuffle group and becomes a SearchedLibrary result gate; the
	// rider gate ("If X is 10 or more") lands in the pump group and lowers to an
	// effect-gate condition.
	shuffleCondIdx, ok := conditionContainedInEffect(content.Conditions, content.Effects[2].Span)
	if !ok {
		return game.AbilityContent{}, false
	}
	riderCondIdx, ok := conditionContainedInEffect(content.Conditions, content.Effects[4].Span)
	if !ok || riderCondIdx == shuffleCondIdx {
		return game.AbilityContent{}, false
	}
	shuffleCond := content.Conditions[shuffleCondIdx]
	if shuffleCond.Kind != compiler.ConditionIf ||
		shuffleCond.Predicate != compiler.ConditionPredicateUnsupported ||
		shuffleCond.Negated {
		// The shuffle gate carries no typed predicate; it is recognized only as
		// the untyped "If you search your library this way" clause gating this
		// group. A typed predicate here is a different card, so fail closed.
		return game.AbilityContent{}, false
	}
	riderGate, ok := lowerCondition(content.Conditions[riderCondIdx], conditionContextEffectGate)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{
				Primitive: game.Search{
					Player: game.ControllerReference(),
					Spec:   spec,
					Amount: amount,
				},
				PublishResult: searchedLibraryResultKey,
			},
			{
				Primitive: game.ShuffleLibrary{Player: game.ControllerReference()},
				ResultGate: opt.Val(game.InstructionResultGate{
					Key:             searchedLibraryResultKey,
					SearchedLibrary: game.TriTrue,
				}),
			},
			{
				Primitive: game.ApplyContinuous{
					ContinuousEffects: pump,
					Duration:          game.DurationUntilEndOfTurn,
				},
				Condition: opt.Val(game.EffectCondition{Condition: opt.Val(riderGate)}),
			},
		},
	}.Ability(), true
}

// multiZoneSearchBattlefieldSpec validates the four search/shuffle effects of the
// multi-zone battlefield search and builds the SearchSpec and amount for the
// folded Search instruction. effects[0] is the graveyard search carrying the
// "creature card with mana value X or less" filter, effects[1] the battlefield
// put, effects[2] the library-search marker, and effects[3] the shuffle. The
// filter is taken from the graveyard search and applied to both zones by the
// runtime; the dynamic X mana-value bound is required. It fails closed for any
// other search shape, destination, or amount.
func multiZoneSearchBattlefieldSpec(effects []compiler.CompiledEffect) (game.SearchSpec, game.Quantity, bool) {
	graveyardSearch := effects[0]
	put := effects[1]
	librarySearch := effects[2]
	shuffle := effects[3]
	if graveyardSearch.Kind != compiler.EffectSearch ||
		graveyardSearch.Context != parser.EffectContextController ||
		graveyardSearch.ToZone != zone.Graveyard ||
		graveyardSearch.Negated ||
		graveyardSearch.Duration != compiler.DurationNone ||
		!graveyardSearch.Amount.Known ||
		graveyardSearch.Amount.Value != 1 {
		return game.SearchSpec{}, game.Quantity{}, false
	}
	if put.Kind != compiler.EffectPut ||
		put.Context != parser.EffectContextController ||
		put.ToZone != zone.Battlefield ||
		put.Connection != parser.EffectConnectionAnd ||
		put.Negated {
		return game.SearchSpec{}, game.Quantity{}, false
	}
	if librarySearch.Kind != compiler.EffectSearch ||
		librarySearch.Context != parser.EffectContextController ||
		librarySearch.ToZone != zone.None ||
		librarySearch.Negated {
		return game.SearchSpec{}, game.Quantity{}, false
	}
	if shuffle.Kind != compiler.EffectShuffle ||
		shuffle.Context != parser.EffectContextController ||
		shuffle.Negated {
		return game.SearchSpec{}, game.Quantity{}, false
	}
	spec, ok := searchSpecForSelector(graveyardSearch.Selector)
	if !ok || !spec.MaxManaValueFromX {
		return game.SearchSpec{}, game.Quantity{}, false
	}
	spec.SourceZone = zone.Library
	spec.AlsoGraveyard = true
	spec.ConditionalShuffle = true
	spec.Destination = zone.Battlefield
	return spec, game.Fixed(1), true
}

// conditionContainedInEffect returns the index of the single condition whose span
// is contained in the given effect span, associating a condition with the
// sentence group that owns it. It fails closed when no condition or more than one
// condition falls inside the span, keeping the group-to-condition mapping
// unambiguous.
func conditionContainedInEffect(conditions []compiler.CompiledCondition, span shared.Span) (int, bool) {
	found := -1
	for i := range conditions {
		if spanContains(span, conditions[i].Span) {
			if found >= 0 {
				return 0, false
			}
			found = i
		}
	}
	if found < 0 {
		return 0, false
	}
	return found, true
}
