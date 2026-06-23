package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// lowerCastFromGraveyardPermission lowers the targeted graveyard cast-permission
// effect "you may cast target <card> from your graveyard this turn" (Norika
// Yamazaki, the Poet) into a GrantCastPermission that lets the controller cast
// the chosen graveyard card normally until end of turn. The "you may" is the
// permission itself — the grant is unconditional and the controller later
// chooses whether to cast — so the effect's optional flag is intentionally not
// gated here.
//
// It returns ok=false (so the caller falls through to the free-cast path) for
// every cast effect outside that envelope: a free or adventure cast, a negated
// or delayed cast, a cast by anyone but the controller, a source zone other than
// the graveyard, an unbounded or unsupported duration, or a target whose
// graveyard-card selector this backend cannot express.
func lowerCastFromGraveyardPermission(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic, bool) {
	effect := ctx.content.Effects[0]
	if effect.CastWithoutPayingManaCost || effect.CastAsAdventure ||
		effect.Negated || effect.DelayedTiming != 0 ||
		effect.Context != parser.EffectContextController ||
		effect.FromZone != zone.Graveyard {
		return game.AbilityContent{}, nil, false
	}
	duration, ok := graveyardCastPermissionDuration(effect.Duration)
	if !ok {
		return game.AbilityContent{}, nil, false
	}
	if len(ctx.content.Targets) != 1 {
		return game.AbilityContent{}, nil, false
	}
	spec, ok := cardInZoneTargetSpec(ctx.content.Targets[0], zone.Graveyard)
	if !ok {
		return game.AbilityContent{}, nil, false
	}
	consumed := ctx
	consumed.content.Targets = nil
	if consumed.content.Unconsumed() {
		return game.AbilityContent{}, nil, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{spec},
		Sequence: []game.Instruction{{
			Primitive: game.GrantCastPermission{
				Card:     game.CardReference{Kind: game.CardReferenceTarget},
				FromZone: zone.Graveyard,
				Face:     game.FaceFront,
				Duration: duration,
			},
		}},
	}.Ability(), nil, true
}

// graveyardCastPermissionDuration maps a compiled "this turn" / "until end of
// turn" cast window to the runtime end-of-turn duration. Both wordings grant the
// card a single-turn cast window that expires in the same turn's cleanup. Any
// other duration is unsupported.
func graveyardCastPermissionDuration(duration compiler.DurationKind) (game.EffectDuration, bool) {
	switch duration {
	case compiler.DurationThisTurn, compiler.DurationUntilEndOfTurn:
		return game.DurationUntilEndOfTurn, true
	default:
		return 0, false
	}
}
