package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

func lifeEventPlayerRecipient(ctx contentCtx, references []compiler.CompiledReference) (playerRecipient, bool) {
	if len(ctx.content.Effects) != 1 ||
		ctx.content.Effects[0].Context != parser.EffectContextReferencedPlayer ||
		len(ctx.content.Targets) != 0 || len(references) != 1 ||
		references[0].Kind != compiler.ReferenceThatPlayer ||
		ctx.enclosingKind != compiler.AbilityTriggered {
		return playerRecipient{}, false
	}
	ref := references[0]
	switch ctx.triggerEvent {
	case game.EventZoneChanged:
		if ref.Binding == compiler.ReferenceBindingEventPermanent || ref.Binding == compiler.ReferenceBindingEventCard {
			return playerRecipient{player: game.EventPlayerReference()}, true
		}
	case game.EventPermanentEnteredBattlefield, game.EventPermanentDied:
		if ref.Binding == compiler.ReferenceBindingEventPermanent &&
			ctx.triggerPattern != nil && !ctx.triggerPattern.OneOrMore &&
			ctx.triggerPattern.Controller != game.TriggerControllerAny {
			return playerRecipient{player: game.ObjectControllerReference(game.EventPermanentReference())}, true
		}
	default:
	}
	return playerRecipient{}, false
}
