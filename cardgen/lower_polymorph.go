package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// lowerPolymorphContent lowers the targeted resolving polymorph effect "Until
// end of turn, target creature loses all abilities and becomes a <color>
// <subtype> with base power and toughness N/N." (Turn to Frog, Snakeform, Gift
// of Tusks) into a single ApplyContinuous over the target. The continuous
// effects span four layers, mirroring the static-aura polymorph: LayerAbility
// removes all abilities, LayerColor sets the new color, LayerType sets the
// creature type, and LayerPowerToughnessSet sets the base power and toughness.
// The target's filter flows through the canonical target machinery, so
// single-target and "up to one target" forms both lower here. Any richer shape
// (riders, conditions, keyword grants, or a missing target) fails closed.
func lowerPolymorphContent(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported polymorph effect",
			"the executable source backend supports only a target creature that loses all abilities and becomes a fixed-power/toughness creature with one subtype until end of turn",
		)
	}
	if effect.Negated ||
		effect.Optional ||
		len(effect.PolymorphSubtypes) == 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Targets) != 1 {
		return unsupported()
	}
	continuousEffects := []game.ContinuousEffect{{
		Layer:              game.LayerAbility,
		RemoveAllAbilities: true,
	}}
	switch {
	case effect.PolymorphColorless:
		continuousEffects = append(continuousEffects, game.ContinuousEffect{
			Layer:        game.LayerColor,
			SetColorless: true,
		})
	case len(effect.PolymorphColors) != 0:
		continuousEffects = append(continuousEffects, game.ContinuousEffect{
			Layer:     game.LayerColor,
			SetColors: slices.Clone(effect.PolymorphColors),
		})
	default:
	}
	continuousEffects = append(continuousEffects, game.ContinuousEffect{
		Layer:         game.LayerType,
		SetTypes:      []types.Card{types.Creature},
		SetSubtypes:   slices.Clone(effect.PolymorphSubtypes),
		AddSupertypes: slices.Clone(effect.PolymorphSupertypes),
	})
	if effect.PolymorphName != "" {
		continuousEffects = append(continuousEffects, game.ContinuousEffect{
			Layer:   game.LayerText,
			SetName: effect.PolymorphName,
		})
	}
	continuousEffects = append(continuousEffects, game.ContinuousEffect{
		Layer:        game.LayerPowerToughnessSet,
		SetPower:     opt.Val(game.PT{Value: effect.PolymorphBasePower}),
		SetToughness: opt.Val(game.PT{Value: effect.PolymorphBaseToughness}),
	})
	if effect.PolymorphPermanent {
		return continuousTargetMode(ctx.content.Targets[0], continuousEffects, game.DurationPermanent, unsupported)
	}
	return temporaryKeywordTargetMode(ctx.content.Targets[0], continuousEffects, unsupported)
}
