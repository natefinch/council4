package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// lowerPutEffectSpell dispatches a single EffectPut clause to its supported
// shapes: a targeted graveyard return, a put-from-hand ramp effect, or counter
// placement. A put with a library destination is rejected as an unsupported
// library placement.
func lowerPutEffectSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	if content, ok := lowerTargetedGraveyardReturn(ctx); ok {
		return content, nil
	}
	if content, ok := lowerPutFromHandSpell(ctx); ok {
		return content, nil
	}
	if ctx.content.Effects[0].ToZone == zone.Library {
		return game.AbilityContent{}, unsupportedLibraryPlacementDiagnostic(ctx)
	}
	return lowerCounterPlacementSpell(ctx)
}

// lowerPutFromHandSpell lowers "put a <filter> card from your hand onto the
// battlefield" — a ramp / cheat-into-play effect (Growth Spiral's "you may put a
// land card from your hand onto the battlefield", Dramatic Entrance, Elvish
// Pioneer, ...). It produces one game.PutFromHand instruction that has the
// controller choose one matching card from their own hand and put it onto the
// battlefield. A "you may" wrapper is carried by the enclosing instruction's
// Optional flag, applied by the optional-flow machinery after this lowers, so
// this path lowers only the mandatory core.
//
// It is card-name-blind and fails closed (ok=false) on any shape it does not
// fully model — references or targets, a non-hand source or non-battlefield
// destination, a selector qualifier it cannot express, an "enters tapped" rider,
// or an amount other than exactly one card — so an unmodeled wording falls
// through to the generic put path's diagnostic rather than lowering to a
// silently-wrong instruction.
func lowerPutFromHandSpell(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectPut ||
		effect.Negated ||
		effect.Divided ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		effect.FromZone != zone.Hand ||
		effect.ToZone != zone.Battlefield ||
		effect.EntersTapped ||
		effect.UnderYourControl {
		return game.AbilityContent{}, false
	}
	selector := effect.Selector
	if selector.Zone != zone.Hand ||
		selector.Controller != compiler.ControllerAny ||
		selector.All ||
		selector.Another ||
		selector.Other ||
		selector.Attacking ||
		selector.Blocking ||
		selector.Tapped ||
		selector.Untapped {
		return game.AbilityContent{}, false
	}
	if !effect.Amount.Known ||
		effect.Amount.RangeKnown ||
		effect.Amount.VariableX ||
		effect.Amount.DynamicKind != 0 ||
		effect.Amount.Value != 1 {
		return game.AbilityContent{}, false
	}
	selection, ok := cardSelectionForSelector(selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.PutFromHand{
			Player:    game.ControllerReference(),
			Selection: selection,
			Amount:    game.Fixed(1),
		},
	}}}.Ability(), true
}
