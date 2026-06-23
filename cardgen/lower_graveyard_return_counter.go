package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// lowerGraveyardReturnThenCounterPlacement lowers the ordered pair "Return target
// <permanent> card from your graveyard to the battlefield. Put a <kind> counter
// [or a <kind> counter] on it." (Elspeth Conquers Death chapter III), where the
// second clause places counters on the just-returned permanent named by "it".
// The returned permanent is a new object the target reference cannot name (the
// target denotes a graveyard card), so the return's put-onto-battlefield
// instruction publishes the entered permanent under a link key and the counter
// placement reads that linked object. Both the single recognized kind and the
// binary "a <X> counter or a <Y> counter" controller choice are supported. It
// fails closed for any other shape: a non-single-target return, a non-controller
// or negated clause, a counter rider already on the return, a dynamic or
// non-positive count, or a counter recipient that is not the returned permanent.
func lowerGraveyardReturnThenCounterPlacement(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		ctx.optional {
		return game.AbilityContent{}, false
	}
	returnEffect := ctx.content.Effects[0]
	counterEffect := ctx.content.Effects[1]
	if returnEffect.Kind != compiler.EffectReturn ||
		!returnEffect.Exact ||
		returnEffect.Negated ||
		returnEffect.Optional ||
		returnEffect.Context != parser.EffectContextController ||
		returnEffect.FromZone != zone.Graveyard ||
		returnEffect.ToZone != zone.Battlefield ||
		returnEffect.DelayedTiming != 0 ||
		returnEffect.CounterKindKnown ||
		len(returnEffect.Targets) != 1 {
		return game.AbilityContent{}, false
	}
	if counterEffect.Kind != compiler.EffectPut ||
		!counterEffect.Exact ||
		counterEffect.Negated ||
		counterEffect.Optional ||
		counterEffect.Context != parser.EffectContextController ||
		counterEffect.Duration != compiler.DurationNone ||
		len(counterEffect.Targets) != 0 ||
		!counterEffect.Amount.Known ||
		counterEffect.Amount.Value < 1 ||
		!referencesBindTo(counterEffect.References, compiler.ReferenceBindingTarget, 0) {
		return game.AbilityContent{}, false
	}
	singleKind := counterEffect.CounterKindKnown &&
		compiler.CounterKindPlacementSupported(counterEffect.CounterKind) &&
		!counterEffect.CounterKind.PlayerOnly()
	choiceKinds := counterEffect.CounterKindChoices
	if !singleKind {
		if len(choiceKinds) < 2 {
			return game.AbilityContent{}, false
		}
		for _, kind := range choiceKinds {
			if !compiler.CounterKindPlacementSupported(kind) || kind.PlayerOnly() {
				return game.AbilityContent{}, false
			}
		}
	}
	returnContent, ok := lowerTargetedGraveyardReturn(contextForEffect(ctx, &returnEffect))
	if !ok ||
		len(returnContent.Modes) != 1 ||
		len(returnContent.Modes[0].Targets) != 1 ||
		len(returnContent.Modes[0].Sequence) != 1 {
		return game.AbilityContent{}, false
	}
	put, ok := returnContent.Modes[0].Sequence[0].Primitive.(game.PutOnBattlefield)
	if !ok || put.PublishLinked != "" {
		return game.AbilityContent{}, false
	}
	key := game.LinkedKey("reanimate-counter")
	put.PublishLinked = key
	addCounter := game.AddCounter{
		Amount: game.Fixed(counterEffect.Amount.Value),
		Object: game.LinkedObjectReference(string(key)),
	}
	if singleKind {
		addCounter.CounterKind = counterEffect.CounterKind
	} else {
		addCounter.KindChoices = slices.Clone(choiceKinds)
	}
	return game.Mode{
		Targets: returnContent.Modes[0].Targets,
		Sequence: []game.Instruction{
			{Primitive: put},
			{Primitive: addCounter},
		},
	}.Ability(), true
}
