package cardgen

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// lowerCreateTokenSpell lowers vanilla creature-token creation: the controller
// (or a referenced object's controller) creates a fixed-power/toughness creature
// token with one or two subtypes, up to two colors (or colorless), an optional
// leading artifact/enchantment permanent type, and an optional single creature
// keyword. Richer token shapes (tapped/attacking entry, quoted abilities,
// modifiers) fail closed pending follow-up work under the token-creation epic.
func lowerCreateTokenSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if effect.TokenCopyOfTarget {
		return lowerCreateCopyTokenSpell(ctx)
	}
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
	if !ok {
		def, ok = synthesizeNamedArtifactTokenDef(&effect)
	}
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	amount, ok := createTokenAmount(&effect)
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.CreateToken{
				Amount:    amount,
				Source:    game.TokenDef(def),
				Recipient: recipient,
			},
		}},
	}.Ability(), nil
}

// createTokenAmount resolves a recognized create-token effect's token count. A
// "for each <X>" iterator lowers to a dynamic count of the iterated objects
// (one token per object); every other recognized shape is a fixed literal.
func createTokenAmount(effect *compiler.CompiledEffect) (game.Quantity, bool) {
	if effect.Amount.DynamicForm == compiler.DynamicAmountForEach {
		if dynamic, ok := lowerDynamicAmount(effect.Amount, game.ObjectReference{}); ok {
			return game.Dynamic(dynamic), true
		}
		return game.Fixed(max(effect.Amount.Multiplier, 1)), true
	}
	if effect.Amount.Value < 1 {
		return game.Quantity{}, false
	}
	return game.Fixed(effect.Amount.Value), true
}

// lowerCreateCopyTokenSpell lowers "Create a token that's a copy of <target>."
// to a CreateToken whose source copies the lone target object. The runtime
// already supports object-copy tokens (TokenCopySourceObject); only a single
// fixed target, controller recipient, and no extra clauses are accepted here.
func lowerCreateCopyTokenSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if len(ctx.content.Effects) != 1 ||
		effect.Context != parser.EffectContextController ||
		!effect.Exact ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	targetSpec, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{{
			Primitive: game.CreateToken{
				Amount: game.Fixed(1),
				Source: game.TokenCopyOf(game.TokenCopySpec{
					Source: game.TokenCopySourceObject,
					Object: game.TargetPermanentReference(0),
				}),
			},
		}},
	}.Ability(), nil
}

// synthesizeCreatureTokenDef builds a token CardDef from a recognized create
// effect: a creature with one or two subtypes, up to two colors (or colorless),
// an optional leading artifact/enchantment permanent type, a fixed
// power/toughness, and an optional single creature keyword. The token's name is
// its joined subtypes, matching paper tokens.
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
	cardTypes, ok := creatureTokenCardTypes(effect.Selector)
	if !ok {
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
			Types:     cardTypes,
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

// creatureTokenCardTypes returns the card types for a synthesized creature
// token. A bare creature token compiles to [Creature]; an artifact- or
// enchantment-creature token prepends the additional permanent type, matching
// the Oracle "<type> creature" ordering. Any other required-type set fails
// closed.
func creatureTokenCardTypes(selector compiler.CompiledSelector) ([]types.Card, bool) {
	required := selector.RequiredTypesAny()
	if len(required) == 0 {
		return []types.Card{types.Creature}, true
	}
	hasCreature := false
	var extra types.Card
	for _, cardType := range required {
		switch cardType {
		case types.Creature:
			hasCreature = true
		case types.Artifact, types.Enchantment:
			if extra != "" {
				return nil, false
			}
			extra = cardType
		default:
			return nil, false
		}
	}
	if !hasCreature {
		return nil, false
	}
	if extra == "" {
		return []types.Card{types.Creature}, true
	}
	return []types.Card{extra, types.Creature}, true
}

// synthesizeNamedArtifactTokenDef builds a token CardDef for a predefined
// artifact token with no printed power/toughness (Treasure, Food, Clue, Blood).
// These tokens carry a fixed Oracle ability the runtime CreateToken/TokenDef
// model already represents. Any other named token fails closed.
func synthesizeNamedArtifactTokenDef(effect *compiler.CompiledEffect) (*game.CardDef, bool) {
	if effect.TokenPTKnown {
		return nil, false
	}
	subtypes := effect.Selector.SubtypesAny()
	if len(subtypes) != 1 ||
		len(effect.Selector.ColorsAny()) != 0 ||
		effect.Selector.Keyword != parser.KeywordUnknown {
		return nil, false
	}
	return namedArtifactTokenDef(subtypes[0])
}

// namedArtifactTokenDef returns the synthesized CardDef for a recognized
// predefined artifact token, or false for an unrepresented one.
func namedArtifactTokenDef(sub types.Sub) (*game.CardDef, bool) {
	switch sub {
	case types.Treasure:
		return treasureTokenDef(), true
	case types.Food:
		return foodTokenDef(), true
	case types.Clue:
		return clueTokenDef(), true
	case types.Blood:
		return bloodTokenDef(), true
	default:
		return nil, false
	}
}

// sacrificeArtifactCost is the "Sacrifice this artifact" additional cost shared
// by predefined artifact tokens.
func sacrificeArtifactCost() cost.Additional {
	return cost.Additional{
		Kind:               cost.AdditionalSacrificeSource,
		Text:               "Sacrifice this artifact",
		Amount:             1,
		MatchPermanentType: true,
		PermanentType:      types.Artifact,
	}
}

// artifactTokenDef builds a single-ability colorless artifact token CardDef
// named for its subtype, matching paper predefined tokens.
func artifactTokenDef(sub types.Sub, ability *game.ActivatedAbility) *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:               string(sub),
			Types:              []types.Card{types.Artifact},
			Subtypes:           []types.Sub{sub},
			ActivatedAbilities: []game.ActivatedAbility{*ability},
		},
	}
}

// treasureTokenDef builds the Treasure token: tap and sacrifice it to add one
// mana of any color.
func treasureTokenDef() *game.CardDef {
	ability := game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G)
	ability.Text = "{T}, Sacrifice this artifact: Add one mana of any color."
	ability.AdditionalCosts = append(slices.Clone(ability.AdditionalCosts), sacrificeArtifactCost())
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:          string(types.Treasure),
			Types:         []types.Card{types.Artifact},
			Subtypes:      []types.Sub{types.Treasure},
			ManaAbilities: []game.ManaAbility{ability},
		},
	}
}

// foodTokenDef builds the Food token: pay two generic mana, tap, and sacrifice
// it to gain 3 life.
func foodTokenDef() *game.CardDef {
	return artifactTokenDef(types.Food, &game.ActivatedAbility{
		Text:            "{2}, {T}, Sacrifice this artifact: You gain 3 life.",
		ManaCost:        opt.Val(cost.Mana{cost.O(2)}),
		AdditionalCosts: []cost.Additional{cost.T, sacrificeArtifactCost()},
		Content: game.Mode{Sequence: []game.Instruction{
			{Primitive: game.GainLife{Amount: game.Fixed(3), Player: game.ControllerReference()}},
		}}.Ability(),
	})
}

// clueTokenDef builds the Clue token: pay two generic mana and sacrifice it to
// draw a card.
func clueTokenDef() *game.CardDef {
	return artifactTokenDef(types.Clue, &game.ActivatedAbility{
		Text:            "{2}, Sacrifice this artifact: Draw a card.",
		ManaCost:        opt.Val(cost.Mana{cost.O(2)}),
		AdditionalCosts: []cost.Additional{sacrificeArtifactCost()},
		Content: game.Mode{Sequence: []game.Instruction{
			{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}},
		}}.Ability(),
	})
}

// bloodTokenDef builds the Blood token: pay one generic mana, tap, discard a
// card, and sacrifice it to draw a card.
func bloodTokenDef() *game.CardDef {
	return artifactTokenDef(types.Blood, &game.ActivatedAbility{
		Text:     "{1}, {T}, Discard a card, Sacrifice this artifact: Draw a card.",
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
		AdditionalCosts: []cost.Additional{
			cost.T,
			{Kind: cost.AdditionalDiscard, Text: "Discard a card", Amount: 1, Source: zone.Hand},
			sacrificeArtifactCost(),
		},
		Content: game.Mode{Sequence: []game.Instruction{
			{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}},
		}}.Ability(),
	})
}

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
