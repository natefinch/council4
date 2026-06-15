package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// lowerCreateTokenSpell lowers the simplest vanilla creature-token creation:
// the controller creates a single fixed-power/toughness creature token with one
// subtype and at most one color, no abilities. Richer token shapes (count > 1,
// keywords, multiple colors/subtypes, named artifact tokens, modifiers) fail
// closed pending follow-up work under the token-creation epic.
func lowerCreateTokenSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if len(ctx.content.Effects) != 1 ||
		effect.Kind != compiler.EffectCreate ||
		!effect.Exact ||
		effect.Context != parser.EffectContextController ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.References) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	def, ok := synthesizeCreatureTokenDef(&effect)
	if !ok || effect.Amount.Value < 1 {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.CreateToken{
				Amount: game.Fixed(effect.Amount.Value),
				Source: game.TokenDef(def),
			},
		}},
	}.Ability(), nil
}

// synthesizeCreatureTokenDef builds a token CardDef from a recognized create
// effect: a creature with exactly one subtype, at most one color, and a fixed
// power/toughness. The token's name is its subtype, matching paper tokens.
func synthesizeCreatureTokenDef(effect *compiler.CompiledEffect) (*game.CardDef, bool) {
	if !effect.TokenPTKnown {
		return nil, false
	}
	subtypes := effect.Selector.SubtypesAny()
	if len(subtypes) != 1 {
		return nil, false
	}
	colors := effect.Selector.ColorsAny()
	if len(colors) > 1 {
		return nil, false
	}
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      string(subtypes[0]),
			Colors:    slices.Clone(colors),
			Types:     []types.Card{types.Creature},
			Subtypes:  slices.Clone(subtypes),
			Power:     opt.Val(game.PT{Value: effect.TokenPower}),
			Toughness: opt.Val(game.PT{Value: effect.TokenToughness}),
		},
	}, true
}

func unsupportedTokenCreationDiagnostic(ctx contentCtx) *shared.Diagnostic {
	return contentDiagnostic(
		ctx,
		"unsupported token creation",
		"the executable source backend supports only a single fixed-power/toughness creature token with one subtype and at most one color",
	)
}
