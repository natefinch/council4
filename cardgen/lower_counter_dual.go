package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// lowerDualReferencedCounterPlacement lowers a single counter-placement effect
// that names two distinct referenced permanents — the triggering event
// permanent and the ability's own source — placing one fixed batch of the same
// counter kind on each ("Whenever another creature you control enters, put a
// +1/+1 counter on that creature and a +1/+1 counter on this creature." —
// Juniper Order Ranger; "Whenever another Mutant you control enters, put a
// +1/+1 counter on that creature and a +1/+1 counter on X-23." — X-23, Deadly
// Weapon). The parser models both recipients as one effect carrying one counter
// kind and amount with two object references, so the same kind and count apply
// to each. The runtime models this as two AddCounter instructions, one per
// referenced permanent, emitted in source order.
//
// It requires exactly one source reference and one event-permanent reference,
// no targets, conditions, or modes, a fixed positive amount, and a single
// recognized placeable permanent counter kind. The event reference observes the
// same sequence-clause gate as the single-reference branch, so an "it"/"that
// creature" pronoun that could denote a prior instruction's product fails
// closed. Every other shape — a player counter, a kind choice, a dynamic
// amount, a non-controller or negated effect, two source or two event
// references — fails closed.
func lowerDualReferencedCounterPlacement(ctx contentCtx) (game.AbilityContent, bool) {
	effect := ctx.content.Effects[0]
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 2 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(effect.CounterKindChoices) != 0 ||
		!effect.Exact ||
		effect.Negated ||
		effect.Context != parser.EffectContextController ||
		!effect.CounterKindKnown ||
		!compiler.CounterKindPlacementSupported(effect.CounterKind) ||
		effect.CounterKind.PlayerOnly() ||
		!effect.Amount.Known ||
		effect.Amount.Value <= 0 {
		return game.AbilityContent{}, false
	}
	if !dualReferenceSourceAndEvent(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	refCtx := referenceLoweringContext{
		AllowSource: true,
		AllowEvent:  !ctx.sequenceClause || ctx.allowEventPronoun,
	}
	amount := game.Fixed(effect.Amount.Value)
	ordered := []compiler.CompiledReference{ctx.content.References[0], ctx.content.References[1]}
	if ordered[1].Span.Start.Offset < ordered[0].Span.Start.Offset {
		ordered[0], ordered[1] = ordered[1], ordered[0]
	}
	sequence := make([]game.Instruction, 0, len(ordered))
	for _, reference := range ordered {
		object, ok := lowerObjectReference(reference, refCtx)
		if !ok {
			return game.AbilityContent{}, false
		}
		sequence = append(sequence, game.Instruction{Primitive: game.AddCounter{
			Amount:      amount,
			Object:      object,
			CounterKind: effect.CounterKind,
		}})
	}
	return game.Mode{Sequence: sequence}.Ability(), true
}

// dualReferenceSourceAndEvent reports whether the two references denote exactly
// one source permanent and one triggering event permanent, in either order.
// Requiring distinct source and event bindings keeps the dual placement closed
// to any pair that does not name the two different permanents the wording does.
func dualReferenceSourceAndEvent(references []compiler.CompiledReference) bool {
	if len(references) != 2 {
		return false
	}
	var sources, events int
	for _, reference := range references {
		switch reference.Binding {
		case compiler.ReferenceBindingSource:
			sources++
		case compiler.ReferenceBindingEventPermanent:
			events++
		default:
			return false
		}
	}
	return sources == 1 && events == 1
}
