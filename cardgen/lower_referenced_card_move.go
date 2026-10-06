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
	if !ctx.content.OwnsSubject(reference) {
		return game.CardReference{}, false
	}
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
	capturedDeparture := from == zone.None && destination == zone.Battlefield &&
		ctx.capturedSubject != nil && ctx.capturedSubject.card
	if from == zone.Battlefield {
		return lowerReachedBattlefieldCardMove(ctx, effect, destination)
	}
	if from != zone.Exile && from != zone.Graveyard && !capturedDeparture {
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

func lowerReachedBattlefieldCardMove(ctx contentCtx, effect compiler.CompiledEffect, destination zone.Type) (game.AbilityContent, bool) {
	if destination != zone.Hand && destination != zone.Exile ||
		effect.EntersTapped || effect.EntersTransformed || effect.UnderYourControl ||
		effect.CounterKindKnown || effect.ReturnAsEnchantment {
		return game.AbilityContent{}, false
	}
	var object game.ObjectReference
	for i, reference := range ctx.content.References {
		if reference.SubjectDomain() != compiler.ReferenceSubjectCard ||
			!reference.EnteredSubjectSupported() ||
			reference.Binding != compiler.ReferenceBindingPriorInstructionResult {
			return game.AbilityContent{}, false
		}
		resolved, ok := lowerObjectReference(reference, referenceLoweringContext{
			PriorInstruction: ctx.priorInstruction, PriorLinkedKey: ctx.priorLinkedKey,
		})
		if !ok || i > 0 && object != resolved {
			return game.AbilityContent{}, false
		}
		object = resolved
	}
	return game.Mode{Sequence: []game.Instruction{{Primitive: game.MovePermanent{
		Object: object, Destination: destination,
	}}}}.Ability(), true
}
