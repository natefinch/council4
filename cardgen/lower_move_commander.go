package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// lowerCommanderFromCommandZone lowers "Put your commander into your hand from
// the command zone." (Command Beacon, Road of Return, Netherborn Altar) to a
// MoveCommander instruction that relocates the controller's commander(s) from
// the command zone to their hand. It is card-name-blind and fails closed on any
// shape it does not fully model — targets, references, conditions, a non-command
// source, a destination other than hand, an opponent-controlled selector, or any
// extra selector qualifier — so unmodeled wordings fall through to the generic
// put path's diagnostic.
func lowerCommanderFromCommandZone(ctx contentCtx) (game.AbilityContent, bool) {
	// This recognizer's sole caller lowerPutEffectSpell is reached only through
	// the EffectPut arm of lowerImmediateSingleEffectSpell, whose content is
	// always single-effect and whose sole effect is an EffectPut. So neither a
	// count other than one nor a kind other than EffectPut can reach here —
	// either is a dispatch bug rather than an unsupported card.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf(
			"lowerCommanderFromCommandZone: reached with %d effects; lowerPutEffectSpell dispatches only single-effect content",
			len(ctx.content.Effects)))
	}
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectPut {
		panic(fmt.Sprintf(
			"lowerCommanderFromCommandZone: reached with effect kind %v; the EffectPut dispatch guarantees EffectPut",
			effect.Kind))
	}
	if effect.Negated ||
		effect.Divided ||
		effect.Optional ||
		ctx.optional ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		effect.UnderYourControl ||
		effect.EntersTapped ||
		effect.FromZone != zone.Command ||
		effect.ToZone != zone.Hand {
		return game.AbilityContent{}, false
	}
	selector := effect.Selector
	if selector.Kind != compiler.SelectorCommander ||
		(selector.Controller != compiler.ControllerAny && selector.Controller != compiler.ControllerYou) ||
		selector.Another ||
		selector.Other ||
		selector.Attacking ||
		selector.Blocking ||
		selector.Tapped ||
		selector.Untapped {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.MoveCommander{
			Player:      game.ControllerReference(),
			Destination: zone.Hand,
		},
	}}}.Ability(), true
}
