package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// lowerSearchThenTrailingSequence lowers a controller library-search group that
// is followed by one or more additional independent effects — "Search your
// library for <filter>, put it ..., then shuffle. <trailing effects>" (Growth
// Spasm, They Went This Way, Scout the Wilderness). The leading search clause
// resolves exactly as a standalone tutor would; the trailing sentence(s) are
// lowered through the shared content path so any already-supported effect
// sequence (a created token, Investigate, a kicker-gated payoff) composes after
// the search.
//
// It applies only to spell abilities whose search moves the found card to hand
// or battlefield; activated/triggered land searches and library-top tutors keep
// their own dedicated lowering and fail-closed coverage.
//
// It exists as an additive fallback after lowerSearchSpell, which fails closed
// whenever the search group is not the whole ability (it requires every effect
// and reference to be consumed by the search). Because this lowerer only runs
// when lowerSearchSpell has already returned a diagnostic, it can never change a
// card that lowerSearchSpell already lowers, keeping the standalone-tutor output
// byte-identical.
//
// The search group is collapsed to a single game.Search instruction via the
// shared searchGroupSpec/searchGroupInstructions helpers; the trailing effects
// are re-lowered as their own content and the search instruction is prepended.
// It fails closed unless the leading clause is a plain mandatory controller
// search (no "under target player's control" rider, no in-clause life/discard
// rider) whose own references are all bound to the search result, and the
// trailing content carries no reference back to the found card.
func lowerSearchThenTrailingSequence(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, bool) {
	if ctx.optional ||
		ctx.enclosingKind != compiler.AbilitySpell ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Effects) < 4 {
		return game.AbilityContent{}, false
	}
	search := ctx.content.Effects[0]
	if search.Kind != compiler.EffectSearch ||
		search.Context != parser.EffectContextController ||
		search.Optional ||
		search.Negated ||
		search.SearchControl != parser.SearchControlRiderNone {
		return game.AbilityContent{}, false
	}
	group, ok := searchGroupSpec(ctx.content.Effects)
	if !ok || group.RiderIndex != 0 || group.Length >= len(ctx.content.Effects) {
		return game.AbilityContent{}, false
	}
	// Restrict to searches that move the found card out of the library to hand or
	// battlefield. Library-top/bottom tutors are a distinct, sensitive family with
	// their own dedicated lowering and fail-closed coverage; this trailing-sequence
	// fallback stays away from them.
	if group.Spec.Destination != zone.Hand && group.Spec.Destination != zone.Battlefield {
		return game.AbilityContent{}, false
	}
	searchSeq, ok := searchGroupInstructions(group)
	if !ok {
		return game.AbilityContent{}, false
	}
	// Every reference inside the search clause must be a search-result reference
	// (the put/reveal "it"/"that card"), already consumed by collapsing the group
	// into one Search instruction. Any other binding there would be silently
	// dropped, so fail closed.
	searchSpan := search.Span
	for _, ref := range referencesWithinSpan(ctx.content.References, searchSpan) {
		if ref.Binding != compiler.ReferenceBindingPriorInstructionResult {
			return game.AbilityContent{}, false
		}
	}
	// The trailing content must not reference the found card. A search-result
	// reference would point at the collapsed search instruction, which is no
	// longer part of the re-lowered trailing sequence; rather than risk
	// misbinding it to a trailing instruction, reject the whole card.
	trailingRefs := referencesOutsideSpan(ctx.content.References, searchSpan)
	for _, ref := range trailingRefs {
		if ref.Binding == compiler.ReferenceBindingPriorInstructionResult {
			return game.AbilityContent{}, false
		}
	}

	payoffCtx := ctx
	payoffCtx.content.Effects = ctx.content.Effects[group.Length:]
	payoffCtx.content.References = trailingRefs
	payoffContent, diagnostic := lowerOrderedEffectSequence(cardName, payoffCtx, syntax)
	if diagnostic != nil ||
		payoffContent.IsModal() ||
		len(payoffContent.Modes) != 1 ||
		len(payoffContent.SharedTargets) != 0 {
		return game.AbilityContent{}, false
	}
	mode := payoffContent.Modes[0]
	mode.Sequence = append(slices.Clone(searchSeq), mode.Sequence...)
	payoffContent.Modes[0] = mode
	return payoffContent, true
}
