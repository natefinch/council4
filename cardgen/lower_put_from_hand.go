package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// lowerPutEffectSpell dispatches a single EffectPut clause to its supported
// shapes: a targeted graveyard return, a put-from-hand ramp effect, the source
// permanent onto its owner's library, or counter placement. A put with any
// other library destination is rejected as an unsupported library placement.
func lowerPutEffectSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	if content, ok := lowerCommanderFromCommandZone(ctx); ok {
		return content, nil
	}
	if content, ok := lowerTargetedGraveyardReturn(ctx); ok {
		return content, nil
	}
	if content, ok := lowerMassGraveyardReturn(ctx); ok {
		return content, nil
	}
	if content, ok := lowerPutFromHandSpell(ctx); ok {
		return content, nil
	}
	if content, ok := lowerPutSourceOnLibrary(ctx); ok {
		return content, nil
	}
	if content, ok := lowerPutThoseCountersSpell(ctx); ok {
		return content, nil
	}
	if ctx.content.Effects[0].ToZone == zone.Library {
		return game.AbilityContent{}, unsupportedLibraryPlacementDiagnostic(ctx)
	}
	return lowerCounterPlacementSpell(ctx)
}

// lowerPutSourceOnLibrary lowers "put this [permanent] on top of its owner's
// library" — Sensei's Divining Top's "put this artifact on top of its owner's
// library" — and the corresponding bottom wording, into a single
// PutPermanentOnLibrary instruction moving the source permanent to the top (or
// bottom) of its owner's library without shuffling.
//
// It is card-name-blind and fails closed on any shape it does not fully model: a
// destination other than the recognized top/bottom, a non-self subject (every
// reference must bind to the source, and a "this <type>" reference must be
// present), targets, an "enters tapped" or under-your-control rider, negation,
// division, a delayed timing, or a non-instant duration.
func lowerPutSourceOnLibrary(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Targets) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectPut ||
		effect.Negated ||
		effect.Divided ||
		effect.Optional ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		effect.ToZone != zone.Library ||
		effect.EntersTapped ||
		effect.UnderYourControl {
		return game.AbilityContent{}, false
	}
	var bottom bool
	switch effect.Destination {
	case parser.EffectDestinationTop:
		bottom = false
	case parser.EffectDestinationBottom:
		bottom = true
	default:
		return game.AbilityContent{}, false
	}
	if len(effect.References) == 0 {
		return game.AbilityContent{}, false
	}
	sawThis := false
	for i := range effect.References {
		reference := effect.References[i]
		if reference.Binding != compiler.ReferenceBindingSource {
			return game.AbilityContent{}, false
		}
		if reference.Kind == compiler.ReferenceThisObject {
			sawThis = true
		}
	}
	if !sawThis {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.PutPermanentOnLibrary{
			Object: game.SourcePermanentReference(),
			Bottom: bottom,
		},
	}}}.Ability(), true
}

// lowerPutFromHandSpell lowers "put a <filter> card from your hand onto the
// battlefield" — a ramp / cheat-into-play effect (Growth Spiral's "you may put a
// land card from your hand onto the battlefield", Dramatic Entrance, Elvish
// Pioneer, ...). It produces one game.ChooseFromZone instruction that has the
// controller choose one matching card from their own hand and put it onto the
// battlefield. A "you may" wrapper is carried by the enclosing instruction's
// Optional flag, applied by the optional-flow machinery after this lowers, so
// this path lowers only the mandatory core.
//
// It is card-name-blind and fails closed (ok=false) on any shape it does not
// fully model — references or targets, a non-hand source or non-battlefield
// destination, a selector qualifier it cannot express, or an amount other than
// exactly one card — so an unmodeled wording falls through to the generic put
// path's diagnostic rather than lowering to a silently-wrong instruction. An
// "enters tapped" rider ("onto the battlefield tapped") is honored and carried
// through to the produced instruction.
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
		effect.UnderYourControl {
		return game.AbilityContent{}, false
	}
	selector := effect.Selector
	// The parser reflects the trailing "tapped" entry rider of "onto the
	// battlefield tapped" into the selector's Tapped flag as well as setting
	// EntersTapped. A card chosen from hand is never literally tapped, so when
	// EntersTapped is set the selector's Tapped is that same entry rider rather
	// than a selection qualifier and is not a blocker; cardSelectionForSelector
	// ignores Tapped, so the produced selection stays correct either way.
	tappedSelection := selector.Tapped && !effect.EntersTapped
	if selector.Zone != zone.Hand ||
		selector.Controller != compiler.ControllerAny ||
		selector.All ||
		selector.Another ||
		selector.Other ||
		selector.Attacking ||
		selector.Blocking ||
		tappedSelection ||
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
		Primitive: game.PutFromHandChoice(
			game.ControllerReference(),
			selection,
			game.Fixed(1),
			effect.EntersTapped,
		),
	}}}.Ability(), true
}
