package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// exiledForEachOpponentKey links the permanents a distributive per-opponent
// exile clause exiles to the per-controller draw payoff that draws one card for
// each. It is distinct from the destroy-for-each-player and exile-until-leaves
// links so the distributive mechanisms never share a record set.
const exiledForEachOpponentKey = game.LinkedKey("exiled-for-each-opponent")

// lowerExileForEachOpponentDrawChainContent lowers the distributive enters
// trigger body "for each opponent, exile up to one target permanent that player
// controls with mana value 3 or greater. For each permanent exiled this way, its
// controller draws a card." (King Solomon's Frogs) into an ExileForEachOpponent
// primitive paired with a DrawForEachExiled payoff. The controller chooses up to
// one matching permanent each opponent controls at resolution; the runtime links
// every exiled permanent under exiledForEachOpponentKey, and the payoff draws one
// card under each exiled permanent's last-known controller. The exile clause's
// per-opponent target is consumed by the distributive walk rather than emitted as
// a resolving target. The "if you cast it" intervening condition gates the
// trigger itself and is stripped from the body before this lowering runs, so the
// body carries no conditions.
//
// It returns ok=false for any shape it does not fully consume: a non-triggered
// host, an optional, condition, mode, or keyword rider, a non-controller exile
// context, a non-referenced-controller draw context, a selector it cannot
// project, or references beyond the distributive "that player" anchor and the
// per-controller "its" pronoun.
func lowerExileForEachOpponentDrawChainContent(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.enclosingKind != compiler.AbilityTriggered ||
		ctx.optional ||
		len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	exileEffect := ctx.content.Effects[0]
	if exileEffect.Kind != compiler.EffectExile ||
		!exileEffect.ExileForEachOpponent ||
		!exileEffect.Exact ||
		exileEffect.Negated ||
		exileEffect.Optional ||
		exileEffect.Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	drawEffect := ctx.content.Effects[1]
	if drawEffect.Kind != compiler.EffectDraw ||
		!drawEffect.DrawForEachExiledThisWay ||
		!drawEffect.Exact ||
		drawEffect.Negated ||
		drawEffect.Optional ||
		drawEffect.Context != parser.EffectContextReferencedObjectController {
		return game.AbilityContent{}, false
	}
	if !referencesAreThatPlayerOrItsController(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	selection, ok := SelectionForSelector(exileEffect.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{
				Primitive: game.ExileForEachOpponent{
					Chooser:   game.ControllerReference(),
					Selection: selection,
					LinkedKey: exiledForEachOpponentKey,
				},
			},
			{
				Primitive: game.DrawForEachExiled{
					LinkedKey: exiledForEachOpponentKey,
				},
			},
		},
	}.Ability(), true
}
