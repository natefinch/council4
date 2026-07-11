package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// lowerExileEachTopCastAnyForFree lowers the exact body "exile the top card of
// each player's library, then you may cast any number of spells from among those
// cards without paying their mana costs." (Etali, Primal Storm) into one
// ExileTopEachLibraryCastFree primitive. The compiler splits it into an
// each-player top-of-library exile followed by an optional controller free-cast
// of any number of "those" cards, so the recognizer accepts exactly that
// two-effect shape and reads the per-library exile count off the exile effect.
//
// It fails closed (ok=false) for every variant outside the envelope: a per-
// controller or single-player exile (the exile is not each-player), a bounded
// or restricted cast selector ("cast any number of spells with mana value 3 or
// less", Kotis; a typed or colored spell), a singular "cast a spell" without the
// any-number count, a paid or timed cast window, or any modes, targets,
// keywords, conditions, or stray effects.
func lowerExileEachTopCastAnyForFree(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Modes) != 0 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Effects) != 2 {
		return game.AbilityContent{}, false
	}
	exile := ctx.content.Effects[0]
	if exile.Kind != compiler.EffectExile ||
		exile.Optional ||
		exile.Negated ||
		exile.Context != parser.EffectContextEachPlayer ||
		exile.CardSource != parser.EffectCardSourceTopOfPlayerLibrary ||
		exile.Selector.Kind != compiler.SelectorCard ||
		!exile.Amount.Known ||
		exile.Amount.Value < 1 ||
		exile.Amount.AnyNumber ||
		exile.DelayedTiming != 0 ||
		len(exile.Targets) != 0 ||
		len(exile.References) != 0 {
		return game.AbilityContent{}, false
	}
	cast := ctx.content.Effects[1]
	if cast.Kind != compiler.EffectCast ||
		!cast.Optional ||
		!cast.CastWithoutPayingManaCost ||
		!cast.Amount.AnyNumber ||
		cast.CastAsAdventure ||
		cast.Negated ||
		cast.DelayedTiming != 0 ||
		cast.Duration != 0 ||
		cast.Connection != parser.EffectConnectionThen ||
		cast.Context != parser.EffectContextController ||
		len(cast.Targets) != 0 ||
		!castsPriorInstructionResult(cast.References) ||
		!bareSpellSelector(cast.Selector) {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ExileTopEachLibraryCastFree{Amount: game.Fixed(exile.Amount.Value)},
		}},
	}.Ability(), true
}

// castsPriorInstructionResult reports whether references are exactly the two
// pronouns ("those cards", "their mana costs") that bind the free cast to the
// just-exiled pile, the compiler's signal that "those cards" are the prior
// exile's result rather than an unrelated set.
func castsPriorInstructionResult(references []compiler.CompiledReference) bool {
	if len(references) != 2 {
		return false
	}
	for _, reference := range references {
		if reference.Binding != compiler.ReferenceBindingPriorInstructionResult {
			return false
		}
	}
	return true
}

// bareSpellSelector reports whether selector is the unrestricted "spells" noun
// with no qualifier this primitive cannot express. Any mana-value bound, keyword,
// type, supertype, subtype, color, or alternative narrows the castable pool, so
// the recognizer fails closed on those variants rather than casting cards the
// text forbids.
func bareSpellSelector(selector compiler.CompiledSelector) bool {
	return selector.Kind == compiler.SelectorSpell &&
		(selector.Controller == compiler.ControllerYou || selector.Controller == compiler.ControllerAny) &&
		selector.Keyword == parser.KeywordUnknown &&
		selector.ExcludedKeyword == parser.KeywordUnknown &&
		!selector.MatchManaValue &&
		!selector.MatchPower &&
		!selector.MatchToughness &&
		!selector.BasicLandType &&
		!selector.Colorless &&
		!selector.Multicolored &&
		len(selector.RequiredTypesAny()) == 0 &&
		len(selector.ExcludedTypes()) == 0 &&
		len(selector.Supertypes()) == 0 &&
		len(selector.SubtypesAny()) == 0 &&
		len(selector.ColorsAny()) == 0 &&
		len(selector.ExcludedColors()) == 0 &&
		len(selector.Alternatives) == 0
}
