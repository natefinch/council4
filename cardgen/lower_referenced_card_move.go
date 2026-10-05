package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func lowerContentCardReference(ctx contentCtx, reference compiler.CompiledReference, bindings referenceLoweringContext) (game.CardReference, bool) {
	if ctx.capturedSubject != nil {
		if !ctx.capturedSubject.card || !slices.Contains(ctx.capturedSubject.references, reference.NodeID) {
			return game.CardReference{}, false
		}
		return game.CapturedCardReference(), true
	}
	return lowerCardReference(reference, bindings)
}

func returnedCardEnchantmentEffects(asEnchantment bool) []game.ContinuousEffect {
	if !asEnchantment {
		return nil
	}
	return []game.ContinuousEffect{{Layer: game.LayerType, SetTypes: []types.Card{types.Enchantment}}}
}

func lowerReferencedCardMove(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 1 || len(ctx.content.References) == 0 ||
		len(ctx.content.Targets) != 0 || len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 || len(ctx.content.Keywords) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	destination := effect.ToZone
	if effect.Kind == compiler.EffectExile {
		destination = zone.Exile
	}
	if !effect.Exact || effect.Negated || effect.Optional ||
		effect.Context != parser.EffectContextController ||
		effect.Amount.DynamicKind != compiler.DynamicAmountNone ||
		(effect.Kind != compiler.EffectReturn && effect.Kind != compiler.EffectPut && effect.Kind != compiler.EffectExile) ||
		(destination != zone.Hand && destination != zone.Battlefield && destination != zone.Exile) {
		return game.AbilityContent{}, false
	}
	from := effect.FromZone
	if ctx.capturedSubject != nil {
		if !ctx.capturedSubject.card {
			return game.AbilityContent{}, false
		}
		from = ctx.capturedSubject.fromZone
	}
	if from != zone.Exile && from != zone.Graveyard &&
		!(from == zone.None && destination == zone.Battlefield && ctx.capturedSubject != nil && ctx.capturedSubject.card) {
		return game.AbilityContent{}, false
	}
	if effect.Amount.Known && (!effect.CounterKindKnown || effect.ToZone != zone.Battlefield) {
		return game.AbilityContent{}, false
	}
	card, ok := lowerContentCardReference(ctx, ctx.content.References[0], referenceLoweringContext{
		AllowSource: true, AllowEvent: true,
		PriorInstruction: ctx.priorInstruction, PriorLinkedKey: ctx.priorLinkedKey,
	})
	if !ok {
		return game.AbilityContent{}, false
	}
	for _, reference := range ctx.content.References[1:] {
		other, ok := lowerContentCardReference(ctx, reference, referenceLoweringContext{
			AllowSource: true, AllowEvent: true,
			PriorInstruction: ctx.priorInstruction, PriorLinkedKey: ctx.priorLinkedKey,
		})
		if !ok || other != card {
			return game.AbilityContent{}, false
		}
	}
	if destination == zone.Battlefield {
		put := game.PutOnBattlefield{
			Source: game.CardBattlefieldSource(card), EntryTapped: effect.EntersTapped, EntryTransformed: effect.EntersTransformed,
			ContinuousEffects: returnedCardEnchantmentEffects(effect.ReturnAsEnchantment),
		}
		if effect.UnderYourControl {
			put.Recipient = opt.Val(game.ControllerReference())
		}
		var countersOK bool
		put.EntryCounters, countersOK = blinkEntryCounters(effect)
		if !countersOK {
			return game.AbilityContent{}, false
		}
		return game.Mode{Sequence: []game.Instruction{{Primitive: put}}}.Ability(), true
	}
	if effect.EntersTapped || effect.EntersTransformed || effect.UnderYourControl || effect.CounterKindKnown || effect.ReturnAsEnchantment {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{{Primitive: game.MoveCard{
		Card: card, FromZone: from, Destination: destination,
	}}}}.Ability(), true
}
