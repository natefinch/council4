package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const eventCardExileCopyLink = game.LinkedKey("event-card-exile-copy")

// lowerEventCardExileCopySequence lowers the reusable triggered sequence
// "you may exile [the triggering card]. If you do, create a token copy of that
// object [with copy exceptions]." The exile publishes the exact event card and
// its battlefield object identity only when the card actually reaches exile; the
// copy then reads that object's last-known copiable values through the link.
func lowerEventCardExileCopySequence(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		ctx.triggerEvent != game.EventPermanentDied ||
		len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 1 {
		return game.AbilityContent{}, false
	}
	exile := ctx.content.Effects[0]
	create := ctx.content.Effects[1]
	condition := ctx.content.Conditions[0]
	if exile.Kind != compiler.EffectExile ||
		exile.Context != parser.EffectContextController ||
		!exile.Exact ||
		!exile.Optional ||
		exile.Negated ||
		exile.DelayedTiming != 0 ||
		len(exile.References) != 1 ||
		exile.References[0].Binding != compiler.ReferenceBindingEventCard ||
		create.Kind != compiler.EffectCreate ||
		create.Context != parser.EffectContextController ||
		!create.Exact ||
		create.Optional ||
		create.Negated ||
		create.DelayedTiming != 0 ||
		!create.TokenCopyOfReference ||
		len(create.References) < 1 ||
		create.References[0].Binding != compiler.ReferenceBindingPriorInstructionResult ||
		create.References[0].PriorInstruction != 0 ||
		!tokenCopyAuxiliaryReferencesOK(create.References[1:]) ||
		condition.Predicate != compiler.ConditionPredicatePriorInstructionAccepted ||
		condition.Kind != compiler.ConditionIf ||
		condition.Negated ||
		condition.Intervening ||
		!create.Order.Contains(condition.Order) ||
		len(ctx.content.References) != len(exile.References)+len(create.References) {
		return game.AbilityContent{}, false
	}
	spec, ok := tokenCopyModifiers(&create, game.LinkedObjectReference(string(eventCardExileCopyLink)))
	if !ok {
		return game.AbilityContent{}, false
	}
	amount, ok := createTokenAmount(ctx, &create, game.ObjectReference{})
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{
		{
			Optional: true,
			Primitive: game.MoveCard{
				Card:                            game.CardReference{Kind: game.CardReferenceEvent},
				FromZone:                        zone.Graveyard,
				Destination:                     zone.Exile,
				PublishLinked:                   eventCardExileCopyLink,
				ReplacePublishedLinked:          true,
				IncludeEventPermanentComponents: true,
			},
			PublishResult: optionalIfYouDoResultKey,
		},
		{
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:       optionalIfYouDoResultKey,
				Succeeded: game.TriTrue,
			}),
			Primitive: game.CreateToken{
				Amount: amount,
				Source: game.TokenCopyOf(spec),
			},
		},
	}}.Ability(), true
}
