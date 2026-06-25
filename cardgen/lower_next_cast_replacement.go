package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// lowerNextCastEntersWithCountersReplacement lowers the body of a "When you next
// cast a creature spell this turn, that creature enters with an additional +1/+1
// counter on it." delayed trigger (Summon: Fenrir chapter II). After the parser
// rewrites the one-shot "When you next cast ..." wording into a "Whenever you
// cast a creature spell, ..." inner triggered ability, that ability's body is a
// single enters-with-counters effect whose subject ("that creature") binds to
// the triggering spell-cast event's stack object. The body lowers to a
// CreateReplacement bound to that spell's card instance: a one-shot
// enters-the-battlefield replacement, lasting until end of turn, that adds the
// extra counters when that specific card resolves onto the battlefield. The
// replacement is bound by card instance ID (carried on the entered event)
// because a permanent spell gains a fresh object ID as it resolves, so an
// object-ID binding could not match the entering permanent. It fails closed on
// any other shape.
func lowerNextCastEntersWithCountersReplacement(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Conditions) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectEnterTapped ||
		effect.Context != parser.EffectContextReferencedObject ||
		!effect.EntersWithCounters ||
		!effect.CounterKindKnown ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.EntersWithCountersGroup() ||
		effect.EntersTypeChoice ||
		effect.EntersColorChoice ||
		effect.Selector.Tapped {
		return game.AbilityContent{}, false
	}
	if !effect.Amount.Known || effect.Amount.VariableX || effect.Amount.Value < 1 {
		return game.AbilityContent{}, false
	}
	if !referencesBindTo(ctx.content.References, compiler.ReferenceBindingEventStackObject, 0) {
		return game.AbilityContent{}, false
	}
	create := game.CreateReplacement{
		Object:   game.EventStackObjectReference(),
		Duration: game.DurationUntilEndOfTurn,
		Replacement: &game.ReplacementEffect{
			Description: ctx.text,
			MatchEvent:  game.EventPermanentEnteredBattlefield,
			EntersWithCounters: []game.CounterPlacement{{
				Kind:   effect.CounterKind,
				Amount: effect.Amount.Value,
			}},
		},
	}
	return game.Mode{Sequence: []game.Instruction{{Primitive: create}}}.Ability(), true
}
