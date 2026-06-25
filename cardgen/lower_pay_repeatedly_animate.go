package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// lowerPayRepeatedlyAnimateContent lowers Primal Adversary's enters trigger into
// the three-instruction resolution sequence proved by the runtime tests: a
// PayRepeatedly that offers the repeatable mana cost any number of times and
// publishes the payment count, an AddCounter that puts that many +1/+1 counters
// on the source, and an ApplyContinuous that lets the controller choose up to
// that many lands they control and animates each chosen land into a creature
// with the recorded base power/toughness, added subtype(s), and keyword(s) while
// it remains a land. The number paid drives both the counter amount and the
// land-selection maximum. Any shape the runtime sequence cannot represent — a
// non-uniform counter, an unsupported keyword, or stray targets, references,
// conditions, keywords, or modes — fails closed.
func lowerPayRepeatedlyAnimateContent(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	payload := effect.PayRepeatedlyAnimate
	unsupported := func(reason string) (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(ctx, "unsupported pay-repeatedly land animation", reason)
	}
	if payload == nil {
		return unsupported("the effect carries no typed pay-repeatedly land-animation payload")
	}
	if effect.Negated ||
		effect.Optional ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Targets) != 0 {
		return unsupported("the pay-repeatedly land animation accepts no targets, references, conditions, keywords, or modes")
	}
	counterKind, ok := powerToughnessCounterKind(payload.CounterPower, payload.CounterToughness)
	if !ok {
		return unsupported("only +1/+1 counters are supported")
	}
	if len(payload.LandSubtypes) == 0 {
		return unsupported("the animated lands gain no creature subtype")
	}
	keywords := make([]game.Keyword, 0, len(payload.LandKeywords))
	for _, kind := range payload.LandKeywords {
		keyword, ok := runtimeKeyword(kind)
		if !ok {
			return unsupported("unsupported animated-land keyword")
		}
		keywords = append(keywords, keyword)
	}

	const countKey = game.ResultKey("pay-repeatedly-count")
	paidCount := game.Dynamic(game.DynamicAmount{
		Kind:      game.DynamicAmountChosenNumber,
		ResultKey: countKey,
	})
	continuousEffects := []game.ContinuousEffect{
		{
			Layer:       game.LayerType,
			AddTypes:    []types.Card{types.Creature},
			AddSubtypes: slices.Clone(payload.LandSubtypes),
		},
	}
	if len(keywords) != 0 {
		continuousEffects = append(continuousEffects, game.ContinuousEffect{
			Layer:       game.LayerAbility,
			AddKeywords: keywords,
		})
	}
	continuousEffects = append(continuousEffects, game.ContinuousEffect{
		Layer:        game.LayerPowerToughnessSet,
		SetPower:     opt.Val(game.PT{Value: payload.LandPower}),
		SetToughness: opt.Val(game.PT{Value: payload.LandToughness}),
	})

	return game.Mode{
		Sequence: []game.Instruction{
			{
				Primitive: game.PayRepeatedly{
					Payment: game.ResolutionPayment{
						ManaCost: opt.Val(slices.Clone(payload.Cost)),
					},
					PublishCount: countKey,
				},
			},
			{
				Primitive: game.AddCounter{
					Amount:      paidCount,
					Object:      game.SourcePermanentReference(),
					CounterKind: counterKind,
				},
			},
			{
				Primitive: game.ApplyContinuous{
					ChooseFrom: game.ObjectControlledGroup(
						game.SourcePermanentReference(),
						game.Selection{RequiredTypes: []types.Card{types.Land}},
					),
					ChooseUpTo:        paidCount,
					ContinuousEffects: continuousEffects,
					Duration:          game.DurationPermanent,
				},
			},
		},
	}.Ability(), nil
}

// powerToughnessCounterKind maps the +N/+N counter dimensions of a
// pay-repeatedly land animation to a named counter kind. Only the +1/+1 counter
// (CR 122.1) has a runtime kind, so any other dimension fails closed.
func powerToughnessCounterKind(power, toughness int) (counter.Kind, bool) {
	if power == 1 && toughness == 1 {
		return counter.PlusOnePlusOne, true
	}
	return counter.PlusOnePlusOne, false
}
