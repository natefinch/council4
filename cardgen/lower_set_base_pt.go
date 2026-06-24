package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerSetBasePTContent lowers the one-shot continuous base power/toughness SET
// effect "[Until end of turn,] <subject> ha(s|ve) base power and toughness
// <N/N|X/X>[ and become every creature type] until end of turn." (Mirror Entity,
// Square Up, Biomass Mutation, Marsh Flitter) into a single ApplyContinuous over
// the affected permanents for the turn.
//
// The base power and toughness are set on LayerPowerToughnessSet: a literal N/N
// uses the fixed SetPower/SetToughness fields, while the variable X/X uses the
// dynamic SetPowerDynamic/SetToughnessDynamic fields holding the activation's X,
// which the resolution-time snapshot freezes into the fixed value. When the
// "become every creature type" rider is present, a LayerType effect adds every
// creature subtype. The subject selects the application scope: a controlled or
// battlefield creature group (StaticSubject), a single targeted creature, or the
// source permanent. Any richer shape fails closed.
func lowerSetBasePTContent(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported set base power/toughness effect",
			"the executable source backend supports only setting a creature group, a single target creature, or the source to a fixed or X base power and toughness, optionally gaining every creature type, until end of turn",
		)
	}
	if effect.Negated ||
		effect.Optional ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return unsupported()
	}

	setEffect := game.ContinuousEffect{Layer: game.LayerPowerToughnessSet}
	if effect.SetBasePTVariableX {
		setEffect.SetPowerDynamic = opt.Val(game.DynamicAmount{Kind: game.DynamicAmountX})
		setEffect.SetToughnessDynamic = opt.Val(game.DynamicAmount{Kind: game.DynamicAmountX})
	} else {
		setEffect.SetPower = opt.Val(game.PT{Value: effect.SetBasePower})
		setEffect.SetToughness = opt.Val(game.PT{Value: effect.SetBaseToughness})
	}
	continuousEffects := []game.ContinuousEffect{setEffect}
	if effect.SetBasePTEveryCreatureType {
		continuousEffects = append(continuousEffects, game.ContinuousEffect{
			Layer:                game.LayerType,
			AddEveryCreatureType: true,
		})
	}

	switch {
	case effect.SetBasePTSource:
		if len(ctx.content.Targets) != 0 {
			return unsupported()
		}
		return sourceContinuousMode(continuousEffects), nil
	case effect.StaticSubject != compiler.StaticSubjectNone:
		if len(ctx.content.Targets) != 0 {
			return unsupported()
		}
		group, ok := resolvingStaticSubjectGroup(&effect)
		if !ok {
			return unsupported()
		}
		for i := range continuousEffects {
			continuousEffects[i].Group = group
		}
		return game.Mode{
			Sequence: []game.Instruction{{
				Primitive: game.ApplyContinuous{
					ContinuousEffects: continuousEffects,
					Duration:          game.DurationUntilEndOfTurn,
				},
			}},
		}.Ability(), nil
	default:
		if len(ctx.content.Targets) != 1 {
			return unsupported()
		}
		return temporaryKeywordTargetMode(ctx.content.Targets[0], continuousEffects, unsupported)
	}
}

// sourceContinuousMode builds an ApplyContinuous mode applying the given
// continuous effects to the source permanent until end of turn.
func sourceContinuousMode(continuousEffects []game.ContinuousEffect) game.AbilityContent {
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ApplyContinuous{
				Object:            opt.Val(game.SourcePermanentReference()),
				ContinuousEffects: continuousEffects,
				Duration:          game.DurationUntilEndOfTurn,
			},
		}},
	}.Ability()
}
