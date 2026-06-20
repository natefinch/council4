package cardgen

import (
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
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectPut ||
		effect.Negated ||
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
