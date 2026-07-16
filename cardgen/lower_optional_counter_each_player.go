package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
)

const optionalCounterForEachPlayerLinkKey game.LinkedKey = "optional-counter-for-each-player"

// lowerOptionalCounterForEachPlayerContent emits an atomic APNAP per-player
// optional counter choice followed by a consuming goad of exactly the permanents
// that actually received counters.
func lowerOptionalCounterForEachPlayerContent(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if !effect.OptionalCounterForEachPlayer ||
		effect.Kind != compiler.EffectPut ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		!effect.CounterKindKnown ||
		!compiler.CounterKindPlacementSupported(effect.CounterKind) ||
		effect.CounterKind.PlayerOnly() ||
		!effect.Amount.Known ||
		effect.Amount.Value <= 0 {
		return game.AbilityContent{}, false
	}
	players, ok := groupMayHaveScope(effect.Context)
	if !ok {
		return game.AbilityContent{}, false
	}
	selection, ok := SelectionForSelector(effect.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{
		{Primitive: game.OptionalCounterForEachPlayer{
			Players:       players,
			Selection:     selection,
			Amount:        game.Fixed(effect.Amount.Value),
			CounterKind:   effect.CounterKind,
			PublishLinked: optionalCounterForEachPlayerLinkKey,
		}},
		{Primitive: game.Goad{
			Group:         game.LinkedObjectsGroup(optionalCounterForEachPlayerLinkKey),
			ConsumeLinked: true,
		}},
	}}.Ability(), true
}
