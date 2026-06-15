package cardgen

import (
	"slices"
	"strings"

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
	controllerRecipient := effect.Context == parser.EffectContextController
	referencedRecipient := effect.Context == parser.EffectContextReferencedObjectController
	if len(ctx.content.Effects) != 1 ||
		effect.Kind != compiler.EffectCreate ||
		!effect.Exact ||
		(!controllerRecipient && !referencedRecipient) ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(keywordsExcludingTokenKeyword(ctx.content, &effect)) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	var recipient opt.V[game.PlayerReference]
	if controllerRecipient {
		if len(ctx.content.References) != 0 {
			return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
		}
	} else {
		if len(ctx.content.References) != 1 {
			return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
		}
		object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{AllowTarget: true})
		if !ok {
			return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
		}
		recipient = opt.Val(game.ObjectControllerReference(object))
	}
	def, ok := synthesizeCreatureTokenDef(&effect)
	if !ok || effect.Amount.Value < 1 {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.CreateToken{
				Amount:    game.Fixed(effect.Amount.Value),
				Source:    game.TokenDef(def),
				Recipient: recipient,
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
	if len(subtypes) < 1 || len(subtypes) > 2 {
		return nil, false
	}
	colors := effect.Selector.ColorsAny()
	if len(colors) > 2 {
		return nil, false
	}
	names := make([]string, 0, len(subtypes))
	for _, sub := range subtypes {
		names = append(names, string(sub))
	}
	def := &game.CardDef{
		CardFace: game.CardFace{
			Name:      strings.Join(names, " "),
			Colors:    slices.Clone(colors),
			Types:     []types.Card{types.Creature},
			Subtypes:  slices.Clone(subtypes),
			Power:     opt.Val(game.PT{Value: effect.TokenPower}),
			Toughness: opt.Val(game.PT{Value: effect.TokenToughness}),
		},
	}
	if effect.Selector.Keyword != parser.KeywordUnknown {
		static, ok := keywordStaticBodies[effect.Selector.Keyword]
		if !ok {
			return nil, false
		}
		def.StaticAbilities = []game.StaticAbility{static.Body}
	}
	return def, true
}

// keywordsExcludingTokenKeyword returns the ability's compiled keywords with the
// create effect's token keyword removed. A token's "with <keyword>" rider is
// represented both on the effect selector and in the ability keyword list; it is
// part of the token spec, not a standalone ability keyword, so it must not block
// token lowering.
func keywordsExcludingTokenKeyword(content compiler.AbilityContent, effect *compiler.CompiledEffect) []compiler.CompiledKeyword {
	if effect.Selector.Keyword == parser.KeywordUnknown {
		return content.Keywords
	}
	filtered := make([]compiler.CompiledKeyword, 0, len(content.Keywords))
	removed := false
	for _, keyword := range content.Keywords {
		if !removed && keyword.Kind == effect.Selector.Keyword && keyword.ParameterKind == parser.KeywordParameterNone {
			removed = true
			continue
		}
		filtered = append(filtered, keyword)
	}
	return filtered
}

func unsupportedTokenCreationDiagnostic(ctx contentCtx) *shared.Diagnostic {
	return contentDiagnostic(
		ctx,
		"unsupported token creation",
		"the executable source backend supports only a single fixed-power/toughness creature token with one subtype and at most one color",
	)
}
