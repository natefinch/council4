package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerPolymorphContent lowers a targeted or target-referenced resolving
// characteristic set into a single ApplyContinuous over the target. Its typed
// payload independently controls ability removal, color, types, subtypes, name,
// and base power/toughness across the corresponding layers.
// The target's filter flows through the canonical target machinery, so
// single-target and "up to one target" forms both lower here. Any richer shape
// (riders, conditions, keyword grants, or a missing target) fails closed.
func lowerPolymorphContent(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported polymorph effect",
			"the executable source backend requires one target and an exact fixed characteristic set",
		)
	}
	referencesOK := effect.Context == parser.EffectContextController && len(ctx.content.References) == 0 ||
		effect.Context == parser.EffectContextReferencedObject &&
			len(ctx.content.References) == 1 &&
			referencesBindTo(ctx.content.References, compiler.ReferenceBindingTarget, 0)
	if effect.Negated ||
		effect.Optional ||
		len(effect.PolymorphTypes) == 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!referencesOK ||
		len(ctx.content.Targets) != 1 {
		return unsupported()
	}
	var continuousEffects []game.ContinuousEffect
	if effect.PolymorphLosesAllAbilities {
		continuousEffects = append(continuousEffects, game.ContinuousEffect{
			Layer:              game.LayerAbility,
			RemoveAllAbilities: true,
		})
	}
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
		SetTypes:      slices.Clone(effect.PolymorphTypes),
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

// lowerTurnFaceDownContent lowers a generic single-target turn-face-down action.
func lowerTurnFaceDownContent(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported turn-face-down effect",
			"turn face down requires exactly one permanent target",
		)
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
	target, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return unsupported()
	}
	return game.Mode{
		Targets: []game.TargetSpec{target},
		Sequence: []game.Instruction{{
			Primitive: game.TurnFaceDown{Object: game.TargetPermanentReference(0)},
		}},
	}.Ability(), nil
}

// lowerTurnFaceDownCharacteristicsContent composes a turn-face-down action with
// the fixed characteristics defined for permanents turned face down by that
// action. The values are carried by the atomic primitive because CR 708.2 makes
// them copiable face-down characteristics that end when the permanent turns
// face up, not an independent continuous effect.
func lowerTurnFaceDownCharacteristicsContent(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	turn, characteristics := ctx.content.Effects[0], ctx.content.Effects[1]
	if turn.Kind != compiler.EffectTurnFaceDown ||
		turn.Negated || turn.Optional || len(turn.References) != 0 ||
		characteristics.Kind != compiler.EffectPolymorph ||
		characteristics.Negated || characteristics.Optional ||
		!characteristics.PolymorphPermanent ||
		len(characteristics.PolymorphTypes) == 0 ||
		len(characteristics.References) != 1 ||
		!referencesBindTo(characteristics.References, compiler.ReferenceBindingTarget, 0) {
		return game.AbilityContent{}, false
	}
	target, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{target},
		Sequence: []game.Instruction{{
			Primitive: game.TurnFaceDown{
				Object: game.TargetPermanentReference(0),
				Characteristics: opt.Val(game.FaceDownCharacteristics{
					Name:       characteristics.PolymorphName,
					Colors:     slices.Clone(characteristics.PolymorphColors),
					Supertypes: slices.Clone(characteristics.PolymorphSupertypes),
					Types:      slices.Clone(characteristics.PolymorphTypes),
					Subtypes:   slices.Clone(characteristics.PolymorphSubtypes),
					Power:      game.PT{Value: characteristics.PolymorphBasePower},
					Toughness:  game.PT{Value: characteristics.PolymorphBaseToughness},
				}),
			},
		}},
	}.Ability(), true
}
