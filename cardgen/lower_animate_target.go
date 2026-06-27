package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// lowerAnimateTargetContent lowers the one-shot continuous target-animation
// "[Until end of turn,] target land becomes a N/N [<color>...] <subtype>...
// creature [with <keyword>...] until end of turn [that's still a land]." (Animate
// Land, Vivify, Hydroform, Kamahl, Soilshaper, Lifespark Spellbomb) into an
// ApplyContinuous over the single targeted land for the turn.
//
// The continuous effects span the layers the animation touches: LayerColor sets
// the stated colors, LayerType adds the creature card type and the named
// subtypes, LayerAbility grants the keywords, and LayerPowerToughnessSet sets the
// literal base power/toughness. The targeted land keeps its land type because the
// type layer adds rather than sets. It mirrors lowerAnimateSelfContent but binds
// the effects to a target slot rather than the source. Any richer shape — no
// target, an extra condition, keyword, mode, or reference, or an unsupported
// animated color or keyword — fails closed.
func lowerAnimateTargetContent(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	payload := effect.AnimateTarget
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported target-animation effect",
			"the executable source backend supports only a single target land becoming a fixed N/N creature until end of turn",
		)
	}
	if payload == nil {
		return unsupported()
	}
	if effect.Negated ||
		effect.Optional ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Targets) != 1 {
		return unsupported()
	}

	continuousEffects := make([]game.ContinuousEffect, 0, 4)
	if len(payload.Colors) != 0 {
		colors := make([]color.Color, 0, len(payload.Colors))
		for _, parserColor := range payload.Colors {
			runtimeColor, ok := animateSelfColor(parserColor)
			if !ok {
				return unsupported()
			}
			colors = append(colors, runtimeColor)
		}
		continuousEffects = append(continuousEffects, game.ContinuousEffect{
			Layer:     game.LayerColor,
			SetColors: colors,
		})
	}

	continuousEffects = append(continuousEffects, game.ContinuousEffect{
		Layer:       game.LayerType,
		AddTypes:    []types.Card{types.Creature},
		AddSubtypes: slices.Clone(payload.Subtypes),
	})

	if len(payload.Keywords) != 0 {
		keywords := make([]game.Keyword, 0, len(payload.Keywords))
		for _, kind := range payload.Keywords {
			keyword, ok := runtimeKeyword(kind)
			if !ok {
				return unsupported()
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

	return continuousTargetMode(ctx.content.Targets[0], continuousEffects, game.DurationUntilEndOfTurn, unsupported)
}
