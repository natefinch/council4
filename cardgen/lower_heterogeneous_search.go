package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// lowerHeterogeneousSearch lowers a heterogeneous multi-slot controller library
// search whose noun phrase names two distinct single-subtype card slots joined
// by a plain "and" and puts the found cards onto the battlefield tapped — "Search
// your library for a Forest card and a Plains card, put them onto the battlefield
// tapped, then shuffle." (Krosan Verge). The parser marks the search clause with
// the per-slot subtypes (CompiledEffect.SearchSlots); this lowerer builds a
// single game.Search whose SearchSpec.SlotFilters carries one subtype filter per
// slot, so the runtime finds one card matching each slot and places both at the
// shared battlefield destination.
//
// It exists as an additive fallback after lowerSearchSpell, which fails closed on
// the heterogeneous noun phrase (its byte-exact reconstruction expects a single
// "or" union, not "a X card and a Y card"). Because this lowerer requires the
// parser-set SearchSlots marker — which lowerSearchSpell's single-filter path
// never produces — it can never change a card lowerSearchSpell already lowers.
//
// It fails closed unless the ability is exactly the three-effect sequence
// search → put-onto-battlefield-tapped → shuffle, performed by the controller
// with no targets, no optionality, no control rider, and only search-result
// references, so an unsupported wording is never silently dropped.
func lowerHeterogeneousSearch(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Effects) != 3 {
		return game.AbilityContent{}, false
	}
	search := ctx.content.Effects[0]
	// lowerContent dispatches here only inside its
	// Effects[0].Kind == EffectSearch block, so a different lead kind is a
	// dispatch bug rather than an unsupported card.
	if search.Kind != compiler.EffectSearch {
		panic(fmt.Sprintf("lowerHeterogeneousSearch: reached with lead effect kind %v; lowerContent dispatches here only for an EffectSearch lead", search.Kind))
	}
	if len(search.SearchSlots) != 2 ||
		search.Context != parser.EffectContextController ||
		search.Optional ||
		search.Negated ||
		search.SearchControl != parser.SearchControlRiderNone ||
		search.SearchSharedSubtype {
		return game.AbilityContent{}, false
	}
	put := ctx.content.Effects[1]
	if put.Kind != compiler.EffectPut ||
		put.ToZone != zone.Battlefield ||
		!put.EntersTapped ||
		put.SearchSplit.Present ||
		put.Optional ||
		put.Negated {
		return game.AbilityContent{}, false
	}
	shuffle := ctx.content.Effects[2]
	if shuffle.Kind != compiler.EffectShuffle ||
		shuffle.Connection != parser.EffectConnectionThen {
		return game.AbilityContent{}, false
	}
	for i := range ctx.content.Effects {
		effect := &ctx.content.Effects[i]
		if effect.Span != search.Span ||
			effect.DelayedTiming != 0 ||
			effect.Duration != compiler.DurationNone {
			return game.AbilityContent{}, false
		}
	}
	// The only reference in the clause is the put's "them", which binds to the
	// search result the collapsed Search instruction already produces. Any other
	// binding would be silently dropped, so fail closed.
	for _, ref := range ctx.content.References {
		if ref.Binding != compiler.ReferenceBindingPriorInstructionResult {
			return game.AbilityContent{}, false
		}
	}
	slots := make([]game.Selection, 0, len(search.SearchSlots))
	for _, sub := range search.SearchSlots {
		slots = append(slots, game.Selection{SubtypesAny: []types.Sub{sub}})
	}
	sequence := []game.Instruction{{Primitive: game.Search{
		Player: game.ControllerReference(),
		Spec: game.SearchSpec{
			SourceZone:   zone.Library,
			Destination:  zone.Battlefield,
			EntersTapped: true,
			SlotFilters:  slots,
		},
		Amount: game.Fixed(len(slots)),
	}}}
	return game.Mode{Sequence: sequence}.Ability(), true
}
