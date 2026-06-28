package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// lowerOptionalLibraryGraveyardTutor lowers the planeswalker-companion tutor:
//
//	You may search your library and/or graveyard for a card named X, reveal it,
//	and put it into your hand. If you search your library this way, shuffle.
//
// The compiler models the body as five effects guarded by one resolving "if"
// condition: an optional EffectSearch whose graveyard destination marker and
// RequiredName selector carry the "search your graveyard for a card named X"
// half, a mandatory EffectReveal, a mandatory EffectPut into the controller's
// hand, a second mandatory EffectSearch for the "and/or your library" half, and
// a mandatory EffectShuffle. The lone condition is the "If you search your
// library this way" shuffle gate, which always holds because the library is
// among the searched zones. This lowers the whole body to one optional
// game.Search whose AlsoGraveyard spec searches both the library and graveyard,
// reveals the found card, and puts it into the controller's hand, shuffling the
// library afterward.
//
// It fails closed unless the body is exactly this four- or five-effect,
// one-condition shape with a name-only card selector and only the internal
// "it"/"this way" anaphors as references: a body-level optional, a modal,
// targeted, or keyword-bearing body, a non-controller subject, a negated or
// delayed effect, a selector that is not a plain "card named X" filter, an
// unexpected condition, or a reference needing its own instruction all leave the
// body unsupported rather than lowering a silently-wrong sequence.
func lowerOptionalLibraryGraveyardTutor(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Conditions) != 1 {
		return game.AbilityContent{}, false
	}
	// The "and/or ... and put it ... If you search" wording compiles the library
	// half to its own EffectSearch before the EffectShuffle (five effects); the
	// "and/or ... then put it ... If you searched" wording folds the library half
	// into the leading search and compiles to four effects. Both end with the
	// shuffle and otherwise share the search/reveal/put core.
	var searchGraveyard, reveal, putHand, shuffle compiler.CompiledEffect
	switch len(ctx.content.Effects) {
	case 4:
		searchGraveyard = ctx.content.Effects[0]
		reveal = ctx.content.Effects[1]
		putHand = ctx.content.Effects[2]
		shuffle = ctx.content.Effects[3]
	case 5:
		searchGraveyard = ctx.content.Effects[0]
		reveal = ctx.content.Effects[1]
		putHand = ctx.content.Effects[2]
		searchLibrary := ctx.content.Effects[3]
		shuffle = ctx.content.Effects[4]
		if !mandatoryControllerTutorEffect(searchLibrary, compiler.EffectSearch) {
			return game.AbilityContent{}, false
		}
	default:
		return game.AbilityContent{}, false
	}

	if searchGraveyard.Kind != compiler.EffectSearch ||
		!searchGraveyard.Optional ||
		searchGraveyard.Negated ||
		searchGraveyard.DelayedTiming != 0 ||
		searchGraveyard.Duration != compiler.DurationNone ||
		searchGraveyard.Context != parser.EffectContextController ||
		searchGraveyard.ToZone != zone.Graveyard ||
		!searchGraveyard.Amount.Known ||
		searchGraveyard.Amount.Value != 1 {
		return game.AbilityContent{}, false
	}
	if !mandatoryControllerTutorEffect(reveal, compiler.EffectReveal) ||
		!mandatoryControllerTutorEffect(putHand, compiler.EffectPut) ||
		putHand.ToZone != zone.Hand ||
		!mandatoryControllerTutorEffect(shuffle, compiler.EffectShuffle) {
		return game.AbilityContent{}, false
	}

	// The lone condition is the "If you search your library this way, shuffle"
	// gate. It always holds because the library is among the searched zones, so
	// the runtime always shuffles after the search; an unexpected condition shape
	// fails closed.
	condition := ctx.content.Conditions[0]
	if condition.Kind != compiler.ConditionIf ||
		condition.Negated ||
		condition.Intervening {
		return game.AbilityContent{}, false
	}

	// The body's only references are the internal "it"/"this way" anaphors back
	// to the found card and the resolving source, both of which the Search
	// primitive models directly. A reference binding to anything else would need
	// its own instruction.
	for ri := range ctx.content.References {
		switch ctx.content.References[ri].Binding {
		case compiler.ReferenceBindingPriorInstructionResult, compiler.ReferenceBindingSource:
		default:
			return game.AbilityContent{}, false
		}
	}

	spec, ok := searchSpecForSelector(searchGraveyard.Selector)
	if !ok || spec.Name == "" || !spec.Filter.Empty() {
		return game.AbilityContent{}, false
	}
	spec.SourceZone = zone.Library
	spec.Destination = zone.Hand
	spec.Reveal = true
	spec.AlsoGraveyard = true

	searcher := game.ControllerReference()
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.Search{
			Player: searcher,
			Spec:   spec,
			Amount: game.Fixed(1),
		},
		Optional:      true,
		OptionalActor: opt.Val(searcher),
	}}}.Ability(), true
}

// mandatoryControllerTutorEffect reports whether effect is a non-optional,
// non-negated, non-delayed effect of the given kind resolved by the controller,
// the shared shape every trailing effect of the library-and-graveyard tutor must
// have.
func mandatoryControllerTutorEffect(effect compiler.CompiledEffect, kind compiler.EffectKind) bool {
	return effect.Kind == kind &&
		!effect.Optional &&
		!effect.Negated &&
		effect.DelayedTiming == 0 &&
		effect.Duration == compiler.DurationNone &&
		effect.Context == parser.EffectContextController &&
		len(effect.Targets) == 0
}
