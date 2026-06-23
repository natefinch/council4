package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// exileHandReturnKey is the constant linked key binding an entire-hand exile to
// the source permanent that exiled it. The runtime keys linked objects by source
// card-instance id plus this string, so a fixed key still keeps each source's
// exiled hand distinct. The enters-the-battlefield exile publishes it and the
// paired leaves-the-battlefield return-to-hand trigger consumes it to return the
// exact exiled set (Wormfang Behemoth).
const exileHandReturnKey = game.LinkedKey("exile-hand-return")

// lowerExileEntireHandContent lowers the involuntary whole-hand exile clause
// "Exile all cards from your hand." (Wormfang Behemoth) into a linked
// ExileEntireHand of the controller's hand. The exiled set is published under
// exileHandReturnKey, keyed by the source permanent, so the paired
// leaves-the-battlefield return-to-hand trigger returns exactly that set. The
// clause names no target.
//
// It returns ok=false for any shape it does not fully consume: a target,
// reference, condition, mode, or keyword rider, an optional or negated effect,
// or a non-controller context, so an unmodeled wording fails closed.
func lowerExileEntireHandContent(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectExile ||
		!effect.ExileEntireHand ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.ExileEntireHand{
			Player:    game.ControllerReference(),
			LinkedKey: exileHandReturnKey,
		},
	}}}.Ability(), true
}

// lowerReturnExiledCardsToHandContent lowers the leaves-the-battlefield clause
// "Return the exiled cards to their owner's hand." (Wormfang Behemoth) into a
// linked ReturnExiledCardsToHand reading exileHandReturnKey. The returned cards
// are the set the sibling entire-hand exile published, identified by the source
// link rather than a target, so the clause carries no target and its pronoun
// reference ("their") is consumed in place of a target binding.
//
// It returns ok=false for any shape it does not fully consume: a target,
// condition, mode, or keyword rider, an optional or negated effect, or a
// non-controller context, so an unmodeled wording fails closed.
func lowerReturnExiledCardsToHandContent(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectReturn ||
		!effect.ReturnExiledCardsToHand ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.ReturnExiledCardsToHand{
			LinkedKey: exileHandReturnKey,
		},
	}}}.Ability(), true
}
