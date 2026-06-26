package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// lowerTargetedGraveyardFreeCast lowers the optional free cast of a targeted
// graveyard card "you may cast target <card> from [your/an opponent's] graveyard
// without paying its mana cost" (Memory Plunder, Torrential Gearhulk) into a
// CastForFree that casts exactly the targeted card without paying its cost. The
// controller targets the card when the spell or ability goes on the stack and
// the enclosing instruction's Optional flag carries the "you may", so at
// resolution the controller decides whether to cast the locked-in target.
//
// Unlike lowerGraveyardTargetCastThatCard, whose cast effect carries the card
// filter on a separate target selector, this family records the whole
// "target instant or sorcery card ... from a graveyard" filter on the cast
// effect's own selector (including its graveyard source zone and any mana-value
// bound), leaving the parsed target an empty placeholder. The target spec is
// therefore built from the cast effect's selector.
//
// It also recognizes the recurring trailing rider "If that spell would be put
// into your graveyard, exile it instead.", which the compiler renders as a bare
// condition plus an EffectPut-to-graveyard and EffectExile referring to the cast
// spell. The rider lowers to CastForFree.ExileOnResolution, redirecting the
// resolved spell to exile in place of its owner's graveyard.
//
// It returns ok=false (so the caller falls through to the generic optional path,
// which fails closed) for every shape outside that envelope: any modal or
// keyword content, a missing or extra target, a non-optional or non-free cast, a
// negated, delayed, adventure, or non-controller cast, a cast that does not draw
// from the graveyard, a graveyard-card selector this backend cannot express
// (a dynamic or {X} mana-value bound), or a trailing-effect shape that is
// neither empty nor exactly the exile-instead rider.
func lowerTargetedGraveyardFreeCast(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Modes) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Effects) == 0 {
		return game.AbilityContent{}, false
	}
	cast := ctx.content.Effects[0]
	if cast.Kind != compiler.EffectCast ||
		!cast.Optional ||
		!cast.CastWithoutPayingManaCost ||
		cast.CastAsAdventure ||
		cast.Negated ||
		cast.DelayedTiming != 0 ||
		cast.Context != parser.EffectContextController ||
		cast.FromZone != zone.Graveyard ||
		cast.Selector.Zone != zone.Graveyard {
		return game.AbilityContent{}, false
	}
	exileOnResolution, ok := graveyardFreeCastRider(ctx.content)
	if !ok {
		return game.AbilityContent{}, false
	}
	target := ctx.content.Targets[0]
	target.Selector = cast.Selector
	spec, ok := cardInZoneTargetSpec(target, zone.Graveyard)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{spec},
		Sequence: []game.Instruction{{
			Optional: true,
			Primitive: game.CastForFree{
				Player:            game.ControllerReference(),
				Card:              game.CardReference{Kind: game.CardReferenceTarget},
				Zone:              zone.Graveyard,
				ExileOnResolution: exileOnResolution,
			},
		}},
	}.Ability(), true
}

// graveyardFreeCastRider reports whether the free-cast content carries no
// trailing effects (a clean cast) or exactly the "If that spell would be put
// into your graveyard, exile it instead." rider, returning whether that rider is
// present. The rider compiles to one bare ConditionIf condition plus an
// EffectPut moving the referenced spell to the graveyard and an EffectExile of
// that spell, so it lowers to ExileOnResolution. Any other trailing shape (a
// stray condition, an unexpected effect, or an effect that is not the
// graveyard-put/exile pair) leaves the content outside the envelope and fails
// closed (ok=false).
func graveyardFreeCastRider(content compiler.AbilityContent) (exileOnResolution bool, ok bool) {
	switch {
	case len(content.Effects) == 1 && len(content.Conditions) == 0:
		return false, true
	case len(content.Effects) == 3 && len(content.Conditions) == 1:
		condition := content.Conditions[0]
		put := content.Effects[1]
		exile := content.Effects[2]
		if condition.Kind != compiler.ConditionIf ||
			condition.Negated ||
			put.Kind != compiler.EffectPut ||
			put.Negated ||
			put.ToZone != zone.Graveyard ||
			put.Context != parser.EffectContextReferencedObject ||
			exile.Kind != compiler.EffectExile ||
			exile.Negated ||
			exile.Context != parser.EffectContextPriorSubject {
			return false, false
		}
		return true, true
	default:
		return false, false
	}
}
