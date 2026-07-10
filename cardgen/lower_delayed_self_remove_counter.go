package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerDelayedSelfRemoveCounter lowers the Clockwork family "remove a <kind>
// counter from it at end of combat." The combat event permanent is captured at
// scheduling time so blinking the creature cannot redirect the delayed removal
// to the new object.
func lowerDelayedSelfRemoveCounter(
	ctx contentCtx,
	timing game.DelayedTriggerTiming,
) (game.AbilityContent, bool) {
	if !ctx.selfTrigger ||
		timing != game.DelayedAtEndOfCombat ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectRemoveCounter ||
		effect.Negated ||
		!effect.Amount.Known ||
		effect.Amount.Value < 1 ||
		!effect.CounterKindKnown ||
		!compiler.CounterKindPlacementSupported(effect.CounterKind) ||
		effect.CounterKind.PlayerOnly() ||
		!referencesBindTo(ctx.content.References, compiler.ReferenceBindingEventPermanent, 0) ||
		!referencesDenoteSelf(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	content := game.Mode{Sequence: []game.Instruction{{
		Primitive: game.RemoveCounter{
			Object:      game.CapturedObjectReference(),
			CounterKind: effect.CounterKind,
			Amount:      game.Fixed(effect.Amount.Value),
		},
	}}}.Ability()
	return game.Mode{Sequence: []game.Instruction{{Primitive: game.CreateDelayedTrigger{
		Trigger: game.DelayedTriggerDef{
			Timing:         timing,
			CapturedObject: opt.Val(game.EventPermanentReference()),
			Content:        content,
		},
	}}}}.Ability(), true
}
