package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// exileUntilMonarchKey links a permanent exiled "until an opponent becomes the
// monarch" to its source so the return trigger can put it back.
const exileUntilMonarchKey = game.LinkedKey("exile-until-opponent-monarch")

// lowerExileUntilOpponentBecomesMonarchContent lowers the monarch exile clause
// "exile <target> until an opponent becomes the monarch." (Palace Jailer) into a
// linked Exile paired with a persistent event delayed trigger that returns the
// exiled card the next time an opponent becomes the monarch. The return is a
// delayed trigger — not a face ability — so it fires even after the source has
// left the battlefield (Palace Jailer leaving does not return the card; the game
// keeps watching for the next monarch change), and its controller is locked to
// the exiling player at schedule time. It mirrors the O-Ring exile-until-leaves
// lowering but anchors the return to a monarch change rather than the source
// leaving. The target is a single permanent (MaxTargets 1), matching the one-card
// link key; every other exile shape leaves the clause unrecognized so lowering
// fails closed.
func lowerExileUntilOpponentBecomesMonarchContent(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.enclosingKind != compiler.AbilityTriggered ||
		ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectExile ||
		!effect.ExileUntilOpponentBecomesMonarch ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	targetSpec, ok := permanentTargetSpecWithCardinality(ctx.content.Targets[0])
	if !ok || targetSpec.MaxTargets != 1 {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{
				Primitive: game.Exile{
					Object:         game.TargetPermanentReference(0),
					ExileLinkedKey: exileUntilMonarchKey,
				},
			},
			{
				Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
					EventPattern: opt.Val(game.TriggerPattern{
						Event:  game.EventBecameMonarch,
						Player: game.TriggerPlayerOpponent,
					}),
					OneShot: true,
					Window:  game.DelayedWindowUntilFires,
					Content: game.Mode{Sequence: []game.Instruction{{
						Primitive: game.PutOnBattlefield{
							Source: game.LinkedBattlefieldSource(exileUntilMonarchKey),
						},
					}}}.Ability(),
				}},
			},
		},
	}.Ability(), true
}
