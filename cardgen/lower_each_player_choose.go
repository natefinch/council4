package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// lowerEachPlayerChooseDestroyContent lowers the recognized "Starting with you,
// each player may choose <permanent>. Destroy each permanent chosen this way."
// construct (Druid of Purification) into a single EachPlayerChooseDestroy
// primitive over the shared candidate pool. Each player, in turn order beginning
// with the controller, may choose up to one permanent matching the pool selector
// — evaluated relative to the controller, so a "you don't control" filter offers
// every chooser the same permanents — and every permanent chosen this way is
// destroyed simultaneously.
//
// It fails closed for any shape it does not fully consume: a non-controller or
// non-exact destroy context, a negated effect, a resolving target, condition,
// mode, keyword, or reference, or a pool selector the backend cannot project.
func lowerEachPlayerChooseDestroyContent(ctx contentCtx) (game.AbilityContent, bool) {
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
	if !effect.EachPlayerChooseDestroy ||
		effect.Kind != compiler.EffectDestroy ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	selection, ok := SelectionForSelector(effect.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.EachPlayerChooseDestroy{
				Selection:           selection,
				Optional:            effect.EachPlayerChooseDestroyOptional,
				PreventRegeneration: effect.PreventRegeneration,
			},
		}},
	}.Ability(), true
}
