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

const sacrificeSucceededResultKey = game.ResultKey("sacrifice-succeeded")

func lowerSacrificeConditionedReanimationSequence(
	ctx contentCtx,
) (game.AbilityContent, bool) {
	if !isSacrificeConditionedChosenCardsCategory(ctx.content) ||
		!matchesSacrificeConditionedReanimation(ctx) {
		return game.AbilityContent{}, false
	}

	target, ok := cardInZoneTargetSpec(ctx.content.Targets[0], zone.Graveyard)
	if !ok {
		return game.AbilityContent{}, false
	}
	sacrifice, ok := lowerSequenceSacrificeInstruction(ctx)
	if !ok {
		return game.AbilityContent{}, false
	}
	sacrifice.PublishResult = sacrificeSucceededResultKey

	gate := opt.Val(game.InstructionResultGate{
		Key:       sacrificeSucceededResultKey,
		Succeeded: game.TriTrue,
	})
	sources := make([]game.BattlefieldSource, 0, target.MaxTargets)
	for i := range target.MaxTargets {
		sources = append(sources, game.CardBattlefieldSource(game.CardReference{
			Kind:        game.CardReferenceTarget,
			TargetIndex: i,
		}))
	}
	sequence := []game.Instruction{
		sacrifice,
		{
			Primitive: game.PutOnBattlefield{
				Sources:     sources,
				EntryTapped: true,
			},
			ResultGate: gate,
		},
	}
	return game.Mode{
		Targets:  []game.TargetSpec{target},
		Sequence: sequence,
	}.Ability(), true
}

func isSacrificeConditionedChosenCardsCategory(content compiler.AbilityContent) bool {
	for _, reference := range content.References {
		if reference.Kind == compiler.ReferenceChosenCards {
			return true
		}
	}
	return false
}

func matchesSacrificeConditionedReanimation(ctx contentCtx) bool {
	content := ctx.content
	if ctx.optional ||
		len(content.Effects) != 2 ||
		len(content.Targets) != 1 ||
		len(content.References) != 1 ||
		len(content.Conditions) != 1 ||
		len(content.Keywords) != 0 ||
		len(content.Modes) != 0 {
		return false
	}
	return matchesSacrificeConditionTarget(content.Targets[0]) &&
		matchesSacrificeConditionProducer(&content.Effects[0]) &&
		matchesSacrificeConditionConsumer(
			&content.Effects[1],
			content.References[0],
			content.Conditions[0],
		)
}

func matchesSacrificeConditionTarget(target compiler.CompiledTarget) bool {
	if !target.Exact ||
		target.Cardinality.Min != 2 ||
		target.Cardinality.Max != 2 ||
		!matchesPlainCreatureCardSelector(target.Selector, compiler.ControllerYou, zone.Graveyard) {
		return false
	}
	_, ok := cardInZoneTargetSpec(target, zone.Graveyard)
	return ok
}

func matchesSacrificeConditionProducer(effect *compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectSacrifice &&
		effect.Exact &&
		effect.Context == parser.EffectContextController &&
		!effect.Optional &&
		!effect.Negated &&
		effect.DelayedTiming == 0 &&
		effect.Amount.Known &&
		effect.Amount.Value == 1 &&
		matchesPlainCreatureCardSelector(effect.Selector, compiler.ControllerAny, zone.None) &&
		len(effect.Targets) == 0 &&
		len(effect.References) == 0
}

func matchesSacrificeConditionConsumer(
	effect *compiler.CompiledEffect,
	reference compiler.CompiledReference,
	condition compiler.CompiledCondition,
) bool {
	return effect.Kind == compiler.EffectReturn &&
		effect.Exact &&
		effect.Context == parser.EffectContextController &&
		effect.FromZone == zone.None &&
		effect.ToZone == zone.Battlefield &&
		effect.EntersTapped &&
		!effect.UnderYourControl &&
		!effect.Optional &&
		!effect.Negated &&
		effect.DelayedTiming == 0 &&
		len(effect.Targets) == 0 &&
		len(effect.References) == 1 &&
		effect.References[0].NodeID == reference.NodeID &&
		matchesChosenCardsTargetReference(reference) &&
		matchesSacrificeSucceededCondition(condition, effect)
}

func matchesChosenCardsTargetReference(reference compiler.CompiledReference) bool {
	return reference.Kind == compiler.ReferenceChosenCards &&
		reference.Binding == compiler.ReferenceBindingTarget &&
		reference.Occurrence == 0
}

func matchesSacrificeSucceededCondition(
	condition compiler.CompiledCondition,
	effect *compiler.CompiledEffect,
) bool {
	return condition.Kind == compiler.ConditionIf &&
		condition.Predicate == compiler.ConditionPredicatePriorInstructionAccepted &&
		!condition.Negated &&
		!condition.Intervening &&
		effect.Order.Contains(condition.Order)
}

func matchesPlainCreatureCardSelector(
	selector compiler.CompiledSelector,
	controller compiler.ControllerKind,
	cardZone zone.Type,
) bool {
	return selector.Kind == compiler.SelectorCreature &&
		selector.Controller == controller &&
		selector.Zone == cardZone &&
		selectorHasOnlyCreatureType(selector) &&
		!selectorHasScalarQualifiers(selector) &&
		!selectorHasListQualifiers(selector)
}

func selectorHasOnlyCreatureType(selector compiler.CompiledSelector) bool {
	return len(selector.RequiredTypesAny()) == 0 ||
		slices.Equal(selector.RequiredTypesAny(), []types.Card{types.Creature})
}

func selectorHasScalarQualifiers(selector compiler.CompiledSelector) bool {
	return selector.All ||
		selector.Another ||
		selector.Other ||
		selector.Attacking ||
		selector.Blocking ||
		selector.Tapped ||
		selector.Untapped ||
		selector.Keyword != parser.KeywordUnknown ||
		selector.ExcludedKeyword != parser.KeywordUnknown ||
		selector.MatchManaValue ||
		selector.MatchPower ||
		selector.MatchToughness ||
		selector.Colorless ||
		selector.Multicolored ||
		selector.BasicLandType ||
		selector.PlayerOrPlaneswalker
}

func selectorHasListQualifiers(selector compiler.CompiledSelector) bool {
	return len(selector.ExcludedTypes()) != 0 ||
		len(selector.Supertypes()) != 0 ||
		len(selector.ExcludedSupertypes()) != 0 ||
		len(selector.ColorsAny()) != 0 ||
		len(selector.ExcludedColors()) != 0 ||
		len(selector.SubtypesAny()) != 0 ||
		len(selector.SourceTypes()) != 0 ||
		len(selector.Alternatives) != 0
}

func lowerSequenceSacrificeInstruction(ctx contentCtx) (game.Instruction, bool) {
	sacrificeCtx := ctx
	sacrificeCtx.content = compiler.AbilityContent{
		Effects: []compiler.CompiledEffect{ctx.content.Effects[0]},
	}
	content, diagnostic := lowerSacrificeSpell(sacrificeCtx)
	if diagnostic != nil ||
		len(content.Modes) != 1 ||
		len(content.Modes[0].Targets) != 0 ||
		len(content.Modes[0].Sequence) != 1 {
		return game.Instruction{}, false
	}
	instruction := content.Modes[0].Sequence[0]
	if _, ok := instruction.Primitive.(game.SacrificePermanents); !ok {
		return game.Instruction{}, false
	}
	return instruction, true
}
