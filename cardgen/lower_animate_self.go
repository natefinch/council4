package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// lowerAnimateSelfContent lowers the one-shot continuous self-animation "This
// <land|artifact|creature|permanent> becomes a N/N [<color>...] [artifact]
// <subtype>... creature [with <keyword>...|all creature types] until end of
// turn." (Faerie Conclave, Mishra's Factory, the Keyrune and Monument mana
// rocks, Mutavault) into a single ApplyContinuous over the source for the turn.
//
// The continuous effects span the layers the animation touches: LayerColor sets
// the stated colors, LayerType adds the creature card type (plus the artifact
// type when stated) and the named subtypes or every creature type, LayerAbility
// grants the keywords, and LayerPowerToughnessSet sets the literal base
// power/toughness. The source keeps its existing land or artifact types because
// the type layer adds rather than sets. Any richer shape — a target, condition,
// keyword, mode, reference, or an unsupported animated keyword — fails closed.
func lowerAnimateSelfContent(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	payload := effect.AnimateSelf
	unsupported := func(reason string) (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(ctx, "unsupported self-animation effect", reason)
	}
	if payload == nil {
		return unsupported("the effect carries no typed self-animation payload")
	}
	if effect.Negated ||
		effect.Optional ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Targets) != 0 {
		return unsupported("the self-animation accepts no targets, references, conditions, keywords, or modes")
	}

	continuousEffects, ok := animationContinuousEffects(payload)
	if !ok {
		return unsupported("unsupported animated color or keyword")
	}
	return continuousSourceMode(continuousEffects, game.DurationUntilEndOfTurn), nil
}

// animationContinuousEffects builds the layered continuous effects shared by the
// self- and target-animation lowerers from an animation payload: LayerColor sets
// the stated colors, LayerType adds the creature card type (plus the artifact
// type when stated) and the named subtypes or every creature type, LayerAbility
// grants the keywords, and LayerPowerToughnessSet sets the literal base
// power/toughness. The animated permanent keeps its existing land or artifact
// types because the type layer adds rather than sets. It fails closed for an
// unsupported animated color or keyword. The target-animation parser never sets
// AddArtifact or EveryCreatureType, so those riders apply only to self-animation.
func animationContinuousEffects(payload *parser.AnimateSelfSyntax) ([]game.ContinuousEffect, bool) {
	continuousEffects := make([]game.ContinuousEffect, 0, 4)
	if len(payload.Colors) != 0 {
		colors := make([]color.Color, 0, len(payload.Colors))
		for _, parserColor := range payload.Colors {
			runtimeColor, ok := animateSelfColor(parserColor)
			if !ok {
				return nil, false
			}
			colors = append(colors, runtimeColor)
		}
		continuousEffects = append(continuousEffects, game.ContinuousEffect{
			Layer:     game.LayerColor,
			SetColors: colors,
		})
	}

	addTypes := []types.Card{types.Creature}
	if payload.AddArtifact {
		addTypes = append(addTypes, types.Artifact)
	}
	continuousEffects = append(continuousEffects, game.ContinuousEffect{
		Layer:                game.LayerType,
		AddTypes:             addTypes,
		AddSubtypes:          slices.Clone(payload.Subtypes),
		AddEveryCreatureType: payload.EveryCreatureType,
	})

	if len(payload.Keywords) != 0 {
		keywords := make([]game.Keyword, 0, len(payload.Keywords))
		for _, kind := range payload.Keywords {
			keyword, ok := runtimeKeyword(kind)
			if !ok {
				return nil, false
			}
			keywords = append(keywords, keyword)
		}
		continuousEffects = append(continuousEffects, game.ContinuousEffect{
			Layer:       game.LayerAbility,
			AddKeywords: keywords,
		})
	}

	continuousEffects = append(continuousEffects, game.ContinuousEffect{
		Layer:        game.LayerPowerToughnessSet,
		SetPower:     opt.Val(game.PT{Value: payload.Power}),
		SetToughness: opt.Val(game.PT{Value: payload.Toughness}),
	})
	return continuousEffects, true
}

// animateSelfColor maps a parser color to its runtime color, failing closed for
// the unknown color. The parser only yields colors recognized from atoms, so a
// well-formed self-animation never drops a color here.
func animateSelfColor(parserColor parser.Color) (color.Color, bool) {
	switch parserColor {
	case parser.ColorWhite:
		return color.White, true
	case parser.ColorBlue:
		return color.Blue, true
	case parser.ColorBlack:
		return color.Black, true
	case parser.ColorRed:
		return color.Red, true
	case parser.ColorGreen:
		return color.Green, true
	default:
		return "", false
	}
}
