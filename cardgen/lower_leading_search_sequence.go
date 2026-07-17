package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// lowerLeadingSequenceThenSearch lowers a body whose leading effect sequence is
// followed by a single trailing controller library-search group — "<effect
// sequence>. Search your library for <filter>, put it ..., then shuffle."
// (Proctor's Gaze's "Return up to one target nonland permanent to its owner's
// hand. Search your library for a basic land card, put it onto the battlefield
// tapped, then shuffle."). The trailing search clause resolves exactly as a
// standalone tutor would; the leading sentence(s) are lowered through the shared
// ordered-sequence path so any already-supported effect sequence composes before
// the search.
//
// It is the mirror of lowerSearchThenTrailingSequence (which lowers a leading
// search followed by trailing effects), and the generalization of the dedicated
// lowerSacrificeThenSearch and lowerDestroyThenSearch lowerers, which keep their
// tighter, target-threading contracts and run first. This generic fallback only
// runs once those have already failed closed, so it can never change a card they
// already lower.
//
// The search group is collapsed to a single game.Search instruction via the
// shared searchGroupSpec/searchGroupInstructions helpers; the leading effects are
// re-lowered as their own content and the search instruction is appended. It
// fails closed unless the trailing clause is a plain mandatory controller search
// (no "under target player's control" rider, no in-clause life/discard rider)
// moving the found card to hand or the battlefield, whose own references are all
// bound to the search result, and the leading content lowers to a single
// non-modal mode with no shared targets.
func lowerLeadingSequenceThenSearch(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, bool) {
	effects := ctx.content.Effects
	if ctx.optional ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(effects) < 3 {
		return game.AbilityContent{}, false
	}
	// The search group is the trailing run; locate its single EffectSearch and
	// require it to be the only search, with at least one leading effect ahead of
	// it. The group must consume every effect from that point to the end.
	searchIndex := -1
	for i := range effects {
		if effects[i].Kind == compiler.EffectSearch {
			if searchIndex >= 0 {
				return game.AbilityContent{}, false
			}
			searchIndex = i
		}
	}
	if searchIndex < 1 {
		return game.AbilityContent{}, false
	}
	search := effects[searchIndex]
	if search.Context != parser.EffectContextController ||
		search.Optional ||
		search.Negated ||
		search.SearchControl != parser.SearchControlRiderNone {
		return game.AbilityContent{}, false
	}
	group, ok := searchGroupSpec(effects[searchIndex:])
	if !ok || group.RiderIndex != 0 || group.Length != len(effects)-searchIndex {
		return game.AbilityContent{}, false
	}
	// Library-top/bottom tutors are a distinct, sensitive family with their own
	// dedicated lowering and fail-closed coverage; this fallback stays away from
	// them and only composes searches that move the found card to hand or the
	// battlefield.
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
	searchReferences := referencesForEffects(effects[searchIndex:])
	leadingReferences := referencesForEffects(effects[:searchIndex])
	if len(searchReferences)+len(leadingReferences) != len(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	for _, ref := range searchReferences {
		if ref.Binding != compiler.ReferenceBindingPriorInstructionResult {
			return game.AbilityContent{}, false
		}
	}

	leadingCtx := ctx
	leadingCtx.content.Effects = effects[:searchIndex]
	leadingCtx.content.References = leadingReferences
	leadingContent, diagnostic := lowerOrderedEffectSequence(cardName, leadingCtx, syntax)
	if diagnostic != nil ||
		leadingContent.IsModal() ||
		len(leadingContent.Modes) != 1 ||
		len(leadingContent.SharedTargets) != 0 {
		return game.AbilityContent{}, false
	}
	mode := leadingContent.Modes[0]
	mode.Sequence = append(mode.Sequence, searchSeq...)
	leadingContent.Modes[0] = mode
	return leadingContent, true
}

func referencesForEffects(effects []compiler.CompiledEffect) []compiler.CompiledReference {
	var references []compiler.CompiledReference
	for i := range effects {
		references = append(references, effects[i].References...)
	}
	return references
}
