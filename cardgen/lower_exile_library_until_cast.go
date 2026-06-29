package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// lowerExileLibraryUntilNonlandCast lowers the exact body "Exile cards from the
// top of your library until you exile a nonland card. You may cast that card
// without paying its mana cost." into one ExileLibraryUntilNonlandCast
// primitive. The compiler splits the "exile cards ... until you exile a nonland
// card" clause into two identical controller exile effects (the dug pile and
// the stopping nonland), followed by the optional free cast of "that card", so
// the recognizer accepts exactly that three-effect shape: two excluded-Land
// controller card exiles plus a controller free-cast of a card. The dig depth
// and the free-cast target are the same nonland card, so the whole sequence is
// one primitive.
//
// It fails closed (ok=false) for every variant outside the envelope: a casting
// window other than "without paying its mana cost" (Territorial Bruntar's "this
// turn"), any mana-value cap or other condition (Solstice Revelations), an
// each-opponent dig (Fevered Suspicion), or any modes, targets, keywords, or
// stray effects.
func lowerExileLibraryUntilNonlandCast(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Modes) != 0 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Effects) != 3 {
		return game.AbilityContent{}, false
	}
	if !nonlandTopExile(ctx.content.Effects[0]) || !nonlandTopExile(ctx.content.Effects[1]) {
		return game.AbilityContent{}, false
	}
	cast := ctx.content.Effects[2]
	if cast.Kind != compiler.EffectCast ||
		!cast.Optional ||
		!cast.CastWithoutPayingManaCost ||
		cast.CastAsAdventure ||
		cast.Negated ||
		cast.DelayedTiming != 0 ||
		cast.Context != parser.EffectContextController ||
		cast.Selector.Kind != compiler.SelectorCard {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ExileLibraryUntilNonlandCast{Player: game.ControllerReference()},
		}},
	}.Ability(), true
}

// nonlandTopExile reports whether effect is the mandatory controller exile of a
// nonland card from the top of their library, the dig clause the parser repeats
// for the "exile cards from the top of your library until you exile a nonland
// card" family.
func nonlandTopExile(effect compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectExile &&
		!effect.Optional &&
		!effect.Negated &&
		effect.Context == parser.EffectContextController &&
		effect.Selector.Kind == compiler.SelectorCard &&
		slices.Contains(effect.Selector.ExcludedTypes(), types.Land)
}
