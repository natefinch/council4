package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// lowerGraveyardTargetCastThatCard lowers the two-sentence graveyard
// cast-permission body "Choose target <card> in your graveyard. You may cast
// that card this turn." (Emry, Lurker of the Loch) into a GrantCastPermission
// that lets the controller cast the chosen graveyard card until end of turn.
//
// It is the split-sentence sibling of lowerCastFromGraveyardPermission, which
// handles the single-sentence form "you may cast target <card> from your
// graveyard this turn." (Norika Yamazaki). Here the target is chosen by a
// leading sentence and the optional cast effect refers back to it as "that
// card", so the cast effect itself carries no FromZone and binds its referent
// through a ReferenceThatObject reference to the lone target. The "you may" is
// the permission itself — the grant is unconditional and the controller later
// chooses whether to cast — so the effect's optional flag is intentionally not
// gated here.
//
// It returns ok=false (so the caller falls through to the generic optional
// path, which fails closed) for every shape outside that envelope: a body-level
// optional, any modal, conditional, or keyword content, a missing or extra
// target, a free or adventure cast, a negated or delayed cast, a cast by anyone
// but the controller, a cast effect that already carries its own source zone, a
// duration other than this-turn, a graveyard-card target this backend cannot
// express, or a reference that does not bind exactly the lone target as "that
// card".
func lowerGraveyardTargetCastThatCard(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Effects) != 1 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectCast ||
		!effect.Optional ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.CastWithoutPayingManaCost ||
		effect.CastAsAdventure ||
		effect.Context != parser.EffectContextController ||
		effect.FromZone != zone.None {
		return game.AbilityContent{}, false
	}
	duration, ok := graveyardCastPermissionDuration(effect.Duration)
	if !ok {
		return game.AbilityContent{}, false
	}
	spec, ok := cardInZoneTargetSpec(ctx.content.Targets[0], zone.Graveyard)
	if !ok {
		return game.AbilityContent{}, false
	}
	if !referencesBindOnlyTarget(ctx.content.References) ||
		!referencesBindOnlyTarget(effect.References) {
		return game.AbilityContent{}, false
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
	}.Ability(), true
}

// referencesBindOnlyTarget reports whether every reference is a "that
// card"/"that object" reference bound to the ability's lone target. The
// graveyard cast-permission body refers to its chosen target as "that card";
// any other reference (a source pronoun, an ambiguous or unsupported antecedent,
// or a second target occurrence) leaves the body outside the recognized
// envelope.
func referencesBindOnlyTarget(references []compiler.CompiledReference) bool {
	for i := range references {
		if references[i].Kind != compiler.ReferenceThatObject ||
			references[i].Binding != compiler.ReferenceBindingTarget ||
			references[i].Occurrence != 0 {
			return false
		}
	}
	return true
}
