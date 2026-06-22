package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// lowerCastForFreeSpell lowers a single EffectCast carrying the free-cast rider
// "... without paying its mana cost" — the resolving "(You may) cast a spell
// [with mana value N or less] from your hand without paying its mana cost."
// family (Rishkar's Expertise and similar). It produces one CastForFree
// primitive that has the controller choose an eligible card from hand and cast
// it for free; the enclosing "you may" optionality is applied by the caller.
//
// It fails closed for every cast effect outside that exact envelope: a paid or
// alternative-cost cast, a cast from a zone other than hand, a non-controller
// caster, a targeted/negated/delayed cast, an adventure cast, residual ability
// content, or a selector restriction this backend cannot express.
func lowerCastForFreeSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func(detail string) (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(ctx, "unsupported cast effect", detail)
	}
	effect := ctx.content.Effects[0]
	if !effect.CastWithoutPayingManaCost {
		return unsupported("only cast-without-paying-mana-cost spells are supported")
	}
	if effect.Negated || effect.DelayedTiming != 0 || ctx.optional || effect.Optional {
		return unsupported("unsupported cast effect modifiers")
	}
	if effect.Context != parser.EffectContextController {
		return unsupported("only the controller casting is supported")
	}
	if effect.FromZone != zone.Hand {
		return unsupported("only casting from the hand is supported")
	}
	if effect.CastAsAdventure || effect.Duration != 0 {
		return unsupported("unsupported cast rider")
	}
	if len(ctx.content.Targets) != 0 {
		return unsupported("cast-for-free spells take no targets")
	}
	// The lone reference is the implicit cast card ("its mana cost"); the
	// primitive casts whichever card the controller chooses, so the reference
	// needs no separate wiring.
	consumed := ctx
	consumed.content.References = nil
	if consumed.content.Unconsumed() {
		return unsupported("unexpected residual ability content")
	}
	selection, ok := castForFreeSelection(effect.Selector)
	if !ok {
		return unsupported("cast restriction is not expressible")
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.CastForFree{
			Player:    game.ControllerReference(),
			Selection: selection,
			Zone:      zone.Hand,
		},
	}}}.Ability(), nil
}

// DO-NOT-COPY(filter): projects the "a spell" (SelectorSpell nonland card)
// free-cast selector and delegates typed cards to cardSelectionForSelector,
// neither of which the battlefield-only canonical projector represents; prefer
// SelectionForSelectorMasked for new code. (retire: #1393)
//
// castForFreeSelection builds the runtime card filter for a free cast. A typed
// selector ("a creature card", "an instant or sorcery card") reuses the shared
// card-selection conversion. The bare "a spell" selector matches any nonland
// card (lands can't be cast), optionally narrowed by mana value and color; it
// fails closed for any other qualifier this backend cannot model on a spell.
func castForFreeSelection(selector compiler.CompiledSelector) (game.Selection, bool) {
	if selector.Kind != compiler.SelectorSpell {
		return cardSelectionForSelector(selector)
	}
	if selector.Keyword != parser.KeywordUnknown ||
		selector.ExcludedKeyword != parser.KeywordUnknown ||
		selector.MatchPower || selector.MatchToughness ||
		selector.BasicLandType ||
		len(selector.RequiredTypesAny()) != 0 ||
		len(selector.ExcludedTypes()) != 0 ||
		len(selector.Supertypes()) != 0 ||
		len(selector.SubtypesAny()) != 0 ||
		len(selector.ColorsAny()) != 0 ||
		len(selector.ExcludedColors()) != 0 ||
		len(selector.Alternatives) != 0 {
		return game.Selection{}, false
	}
	if selector.Controller != compiler.ControllerYou && selector.Controller != compiler.ControllerAny {
		return game.Selection{}, false
	}
	selection := game.Selection{ExcludedTypes: []types.Card{types.Land}}
	if selector.MatchManaValue {
		// The runtime spell filter uses a fixed mana-value comparison and cannot
		// express the spell's chosen {X} ("a spell with mana value X or less"), so
		// fail closed rather than lowering to a wrong fixed bound.
		if selector.ManaValueX {
			return game.Selection{}, false
		}
		selection.ManaValue = opt.Val(selector.ManaValue)
	}
	selection.Colorless = selector.Colorless
	selection.Multicolored = selector.Multicolored
	return selection, true
}
