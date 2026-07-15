package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// lowerPaidTargetedGraveyardCast lowers the resolving optional paid cast of a
// targeted graveyard card "you may cast that card" (Conduit of Worlds), where a
// leading sentence chose the target ("Choose target nonland permanent card in
// your graveyard.") and this effect refers back to it as "that card". It emits a
// CastForFree that casts exactly the targeted card while paying its normal and
// additional costs (PayManaCost), so at resolution the controller may cast the
// locked-in target for its cost, ignoring timing but obeying cast prohibitions
// and per-turn cast limits. The enclosing sequence marks the instruction Optional
// (the "you may") and, when the cast is the optional-flow source, records its
// PublishResult so a following "if you do" effect gates on the actual cast.
//
// It is the paid, resolving sibling of lowerGraveyardTargetCastThatCard, whose
// "you may cast that card this turn." grants a lasting cast permission rather
// than casting during resolution, and of lowerTargetedGraveyardFreeCast, whose
// cast is free. The paid rider (CastWithoutPayingManaCost == false) is the
// distinguishing signal: a free or permission cast is handled by those paths,
// not here.
//
// It returns ok=false (so the caller falls through to the free-cast path, which
// fails closed) for every shape outside that envelope: any modal, conditional,
// or keyword content, a missing or extra target, a free or adventure cast, a
// negated or delayed cast, a cast by anyone but the controller, a cast effect
// that already carries its own source zone (rather than binding the target as
// "that card"), a graveyard-card target this backend cannot express, or a
// reference that does not bind exactly the lone target as "that card".
func lowerPaidTargetedGraveyardCast(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic, bool) {
	if len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Effects) != 1 {
		return game.AbilityContent{}, nil, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectCast ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.CastWithoutPayingManaCost ||
		effect.CastAsAdventure ||
		effect.Context != parser.EffectContextController ||
		effect.FromZone != zone.None {
		return game.AbilityContent{}, nil, false
	}
	spec, ok := cardInZoneTargetSpec(ctx.content.Targets[0], zone.Graveyard)
	if !ok {
		return game.AbilityContent{}, nil, false
	}
	if !referencesBindOnlyTarget(ctx.content.References) ||
		!referencesBindOnlyTarget(effect.References) {
		return game.AbilityContent{}, nil, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{spec},
		Sequence: []game.Instruction{{
			Primitive: game.CastForFree{
				Player:      game.ControllerReference(),
				Card:        game.CardReference{Kind: game.CardReferenceTarget},
				Zone:        zone.Graveyard,
				PayManaCost: true,
			},
		}},
	}.Ability(), nil, true
}
