package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// exiledCardsToHandKey is the constant linked key binding the cards a source
// exiled from its controller's hand to that source. The runtime keys linked
// objects by source card-instance id plus this string, so a fixed key still
// keeps each permanent's exiled set distinct. An enters-the-battlefield "exile
// all cards from your hand" clause publishes it and a leaves-the-battlefield
// "return the exiled cards to their owner's hand" clause consumes it to return
// the set (Wormfang Behemoth). It mirrors exileUntilLeavesKey, but the return
// goes to hand rather than the battlefield.
const exiledCardsToHandKey = game.LinkedKey("exiled-cards-to-hand")

// lowerExileHandToLinkedSetContent lowers the clause "exile all cards from your
// hand." (Wormfang Behemoth's enters-the-battlefield trigger) into a single
// player-zone-group move of the controller's whole hand to exile that publishes
// the exiled set under exiledCardsToHandKey. A sibling leaves-the-battlefield
// trigger consumes the same key to return the set to hand
// (lowerReturnExiledCardsToHandContent); the cross-ability link is validated at
// the face level (collectFacePublishedLinkedKeys).
//
// It returns ok=false for any shape it does not fully consume: a target,
// reference, condition, mode, or keyword rider, an optional or negated effect, a
// non-"all" selector, a non-hand source zone, an opponent-scoped controller, or
// any selector qualifier beyond the plain whole-hand match — so an unmodeled
// wording falls through to the generic exile path's diagnostic rather than
// lowering to a silently-wrong instruction.
func lowerExileHandToLinkedSetContent(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		ctx.content.Unconsumed() {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectExile ||
		!effect.ExileEntireHand ||
		!effect.Exact ||
		effect.Optional ||
		effect.Negated ||
		effect.Divided ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		effect.Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.MoveCard{
			Player:        game.ControllerReference(),
			FromZone:      zone.Hand,
			Destination:   zone.Exile,
			PublishLinked: exiledCardsToHandKey,
		},
	}}}.Ability(), true
}

// lowerReturnExiledCardsToHandContent lowers the explicit leaves-the-battlefield
// clause "return the exiled card(s) to its/their owner's hand." (Wormfang
// Behemoth) into a linked-set move reading exiledCardsToHandKey: every card the
// sibling exile published returns to its owner's hand. The exiled cards are
// identified by the source link rather than a target, so the clause carries no
// target.
//
// It returns ok=false for any shape it does not fully consume: a target, a
// condition, mode, or keyword rider, an optional or negated effect, or a
// non-controller context.
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
		!effect.ReturnExiledCardToHand ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.MoveCard{
			FromLinked:  exiledCardsToHandKey,
			FromZone:    zone.Exile,
			Destination: zone.Hand,
		},
	}}}.Ability(), true
}
