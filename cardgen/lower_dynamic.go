package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func lowerDynamicAmount(amount compiler.CompiledAmount, object game.ObjectReference) (game.DynamicAmount, bool) {
	if amount.Multiplier < 1 {
		return game.DynamicAmount{}, false
	}
	dynamic := game.DynamicAmount{Multiplier: amount.Multiplier}
	switch amount.DynamicKind {
	case compiler.DynamicAmountCount:
		if dynamic, ok := dynamicCardZoneAmount(amount.Selector(), amount.Multiplier); ok {
			return dynamic, true
		}
		selection, ok := dynamicAmountSelection(amount.Selector())
		if !ok {
			return game.DynamicAmount{}, false
		}
		dynamic.Kind = game.DynamicAmountCountSelector
		dynamic.Group = game.BattlefieldGroup(selection)
	case compiler.DynamicAmountControllerLife:
		dynamic.Kind = game.DynamicAmountControllerLife
	case compiler.DynamicAmountOpponentCount:
		dynamic.Kind = game.DynamicAmountOpponentCount
	case compiler.DynamicAmountBasicLandTypes:
		dynamic.Kind = game.DynamicAmountControllerBasicLandTypeCount
	case compiler.DynamicAmountSourcePower:
		if len(object.Validate()) != 0 {
			return game.DynamicAmount{}, false
		}
		dynamic.Kind = game.DynamicAmountObjectPower
		dynamic.Object = object
	default:
		return game.DynamicAmount{}, false
	}
	return dynamic, true
}

func dynamicAmountSelection(selector compiler.CompiledSelector) (game.Selection, bool) {
	if selector.Zone != zone.None {
		return game.Selection{}, false
	}
	selection, ok := dynamicCountCharacteristics(selector)
	if !ok {
		return game.Selection{}, false
	}
	requiredType, known := dynamicBattlefieldRequiredType(selector.Kind)
	switch {
	case known:
		if requiredType != "" {
			selection.RequiredTypes = []types.Card{requiredType}
		}
	case selector.Kind == compiler.SelectorUnknown && selectorHasCountCharacteristic(selector):
	default:
		return game.Selection{}, false
	}
	switch selector.Controller {
	case compiler.ControllerAny:
	case compiler.ControllerYou:
		selection.Controller = game.ControllerYou
	case compiler.ControllerOpponent:
		selection.Controller = game.ControllerOpponent
	default:
		return game.Selection{}, false
	}
	return selection, true
}

func dynamicBattlefieldRequiredType(kind compiler.SelectorKind) (types.Card, bool) {
	switch kind {
	case compiler.SelectorArtifact:
		return types.Artifact, true
	case compiler.SelectorCreature:
		return types.Creature, true
	case compiler.SelectorEnchantment:
		return types.Enchantment, true
	case compiler.SelectorLand:
		return types.Land, true
	case compiler.SelectorPermanent:
		return "", true
	default:
		return "", false
	}
}

func dynamicCardZoneAmount(selector compiler.CompiledSelector, multiplier int) (game.DynamicAmount, bool) {
	if selector.Controller != compiler.ControllerYou {
		return game.DynamicAmount{}, false
	}
	if selector.Zone != zone.Graveyard && selector.Zone != zone.Hand {
		return game.DynamicAmount{}, false
	}
	requiredType, known := dynamicZoneRequiredType(selector.Kind)
	if !known {
		return game.DynamicAmount{}, false
	}
	selection, ok := dynamicCountCharacteristics(selector)
	if !ok {
		return game.DynamicAmount{}, false
	}
	if requiredType != "" {
		selection.RequiredTypes = []types.Card{requiredType}
	}
	player := game.ControllerReference()
	return game.DynamicAmount{
		Kind:       game.DynamicAmountCountCardsInZone,
		Multiplier: multiplier,
		Player:     &player,
		CardZone:   selector.Zone,
		Selection:  &selection,
	}, true
}

func dynamicZoneRequiredType(kind compiler.SelectorKind) (types.Card, bool) {
	switch kind {
	case compiler.SelectorCard:
		return "", true
	case compiler.SelectorArtifact:
		return types.Artifact, true
	case compiler.SelectorCreature:
		return types.Creature, true
	case compiler.SelectorEnchantment:
		return types.Enchantment, true
	case compiler.SelectorLand:
		return types.Land, true
	default:
		return "", false
	}
}

// dynamicCountCharacteristics maps the characteristic filters of a compiled
// count selector onto a runtime Selection, returning false for any filter the
// executable backend cannot represent exactly so unsupported wordings stay
// rejected. It deliberately ignores the selector Kind and Controller, which
// callers translate per context (battlefield required type versus card zone).
func dynamicCountCharacteristics(selector compiler.CompiledSelector) (game.Selection, bool) {
	if selector.All || selector.Another || selector.Other ||
		selector.Attacking || selector.Blocking ||
		selector.Tapped || selector.Untapped ||
		selector.MatchManaValue || selector.MatchPower || selector.MatchToughness {
		return game.Selection{}, false
	}
	return selectorCharacteristics(selector)
}

// selectorCharacteristics maps the characteristic filters of a compiled selector
// (colors, colorless/multicolored, keyword, excluded types, supertypes,
// subtypes, excluded colors) onto a runtime Selection, returning false for any
// characteristic the executable backend cannot represent exactly. It ignores the
// selector Kind, Controller, combat, tapped, and "other" flags, which callers
// translate per context, and fails closed on a disjunctive required-type union.
func selectorCharacteristics(selector compiler.CompiledSelector) (game.Selection, bool) {
	selection := game.Selection{
		Colorless:    selector.Colorless,
		Multicolored: selector.Multicolored,
	}
	if selector.Keyword != parser.KeywordUnknown {
		keyword, ok := runtimeKeyword(selector.Keyword)
		if !ok {
			return game.Selection{}, false
		}
		selection.Keyword = keyword
	}
	if union := selector.RequiredTypesAny(); len(union) > 0 {
		return game.Selection{}, false
	}
	if excluded := selector.ExcludedTypes(); len(excluded) > 0 {
		selection.ExcludedTypes = append([]types.Card(nil), excluded...)
	}
	if supertypes := selector.Supertypes(); len(supertypes) > 0 {
		selection.Supertypes = append([]types.Super(nil), supertypes...)
	}
	if subtypes := selector.SubtypesAny(); len(subtypes) > 0 {
		selection.SubtypesAny = append([]types.Sub(nil), subtypes...)
	}
	if colors := selector.ColorsAny(); len(colors) > 0 {
		selection.ColorsAny = append([]color.Color(nil), colors...)
	}
	if excludedColors := selector.ExcludedColors(); len(excludedColors) > 0 {
		selection.ExcludedColors = append([]color.Color(nil), excludedColors...)
	}
	return selection, true
}

func selectorHasCountCharacteristic(selector compiler.CompiledSelector) bool {
	return selector.Colorless || selector.Multicolored ||
		selector.Keyword != parser.KeywordUnknown ||
		len(selector.SubtypesAny()) > 0 ||
		len(selector.Supertypes()) > 0 ||
		len(selector.ColorsAny()) > 0 ||
		len(selector.ExcludedColors()) > 0 ||
		len(selector.RequiredTypesAny()) > 0 ||
		len(selector.ExcludedTypes()) > 0
}

func runtimeKeyword(keyword parser.KeywordKind) (game.Keyword, bool) {
	switch keyword {
	case parser.KeywordCycling:
		return game.Cycling, true
	case parser.KeywordFlying:
		return game.Flying, true
	case parser.KeywordReach:
		return game.Reach, true
	case parser.KeywordTrample:
		return game.Trample, true
	case parser.KeywordLifelink:
		return game.Lifelink, true
	case parser.KeywordDeathtouch:
		return game.Deathtouch, true
	case parser.KeywordIndestructible:
		return game.Indestructible, true
	case parser.KeywordHaste:
		return game.Haste, true
	case parser.KeywordMenace:
		return game.Menace, true
	case parser.KeywordVigilance:
		return game.Vigilance, true
	default:
		return game.KeywordNone, false
	}
}

func exactDamageAmountReferences(amount compiler.CompiledAmount, references []compiler.CompiledReference) bool {
	if amount.DynamicKind != compiler.DynamicAmountSourcePower {
		_, ok := lowerDamageSourceReference(references)
		return ok
	}
	if len(references) != 2 ||
		references[1].Span != amount.ReferenceSpan {
		return false
	}
	_, ok := lowerDamageSourceReference(references[:1])
	return ok && references[1].Binding == references[0].Binding
}

func lowerDamageSourceReference(references []compiler.CompiledReference) (game.ObjectReference, bool) {
	if len(references) != 1 {
		return game.ObjectReference{}, false
	}
	return lowerObjectReference(references[0], referenceLoweringContext{
		AllowSource: true,
		AllowEvent:  true,
	})
}

func validModifyPTAmount(effect *compiler.CompiledEffect, referenceCount int) bool {
	if effect.Context != parser.EffectContextTarget && effect.Context != parser.EffectContextPriorSubject {
		return false
	}
	amount := effect.Amount
	if amount.DynamicKind == compiler.DynamicAmountNone {
		return true
	}
	if referenceCount != 0 || amount.DynamicKind == compiler.DynamicAmountSourcePower {
		return false
	}
	switch amount.DynamicForm {
	case compiler.DynamicAmountForEach:
		if !effect.PowerDelta.Known || !effect.ToughnessDelta.Known ||
			!dynamicPTMultiplierMatches(amount.Multiplier, effect.PowerDelta, effect.ToughnessDelta) {
			return false
		}
		return true
	case compiler.DynamicAmountWhereX:
		return !effect.PowerDelta.Known &&
			!effect.ToughnessDelta.Known
	default:
		return false
	}
}

func dynamicPTMultiplierMatches(
	multiplier int,
	power, toughness compiler.CompiledSignedAmount,
) bool {
	matches := func(amount compiler.CompiledSignedAmount) bool {
		return amount.Value == 0 || amount.Value == multiplier
	}
	return multiplier > 0 && matches(power) && matches(toughness)
}

func dynamicSignedQuantity(
	dynamic game.DynamicAmount,
	amount compiler.CompiledSignedAmount,
) game.Quantity {
	if amount.Value == 0 {
		return game.Fixed(0)
	}
	if amount.Negative {
		dynamic.Multiplier = -dynamic.Multiplier
	}
	return game.Dynamic(dynamic)
}

func fixedNumberSyntax(token shared.Token, atoms parser.Atoms, amount int) bool {
	if token.Kind == shared.Integer {
		return token.Text == fmt.Sprint(amount)
	}
	value, ok := atoms.CardinalAt(token.Span)
	return ok && value == amount
}

func singleSelfReference(references []compiler.CompiledReference) bool {
	return len(references) == 1 && references[0].Binding == compiler.ReferenceBindingSource
}

func damageTargetSpec(target compiler.CompiledTarget) (game.TargetSpec, bool) {
	if !target.Exact || !targetCardinalityIsOne(target) {
		return game.TargetSpec{}, false
	}
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: target.Text,
	}
	switch target.Selector.Kind {
	case compiler.SelectorAny:
		spec.Allow = game.TargetAllowPermanent | game.TargetAllowPlayer
	case compiler.SelectorCreature, compiler.SelectorPlaneswalker, compiler.SelectorBattle:
		permanent, ok := permanentTargetSpec(target)
		if !ok {
			return game.TargetSpec{}, false
		}
		return permanent, true
	case compiler.SelectorPlayer:
		spec.Allow = game.TargetAllowPlayer
	case compiler.SelectorOpponent:
		spec.Allow = game.TargetAllowPlayer
		spec.Predicate = game.TargetPredicate{Player: game.PlayerOpponent}
	default:
		return game.TargetSpec{}, false
	}
	return spec, true
}

func permanentTargetSpec(target compiler.CompiledTarget) (game.TargetSpec, bool) {
	if !target.Exact || !targetCardinalityIsOne(target) {
		return game.TargetSpec{}, false
	}
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowPermanent,
	}
	switch target.Selector.Kind {
	case compiler.SelectorArtifact:
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Artifact}}
	case compiler.SelectorCreature:
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Creature}}
	case compiler.SelectorEnchantment:
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Enchantment}}
	case compiler.SelectorLand:
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Land}}
	case compiler.SelectorPermanent:
	case compiler.SelectorPlaneswalker:
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Planeswalker}}
	case compiler.SelectorBattle:
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Battle}}
	default:
		return game.TargetSpec{}, false
	}
	if target.Selector.Another || target.Selector.Other ||
		(target.Selector.Tapped && target.Selector.Untapped) ||
		((target.Selector.Tapped || target.Selector.Untapped) &&
			(target.Selector.Attacking || target.Selector.Blocking)) ||
		selectorHasUnsupportedPermanentFilters(target.Selector) {
		return game.TargetSpec{}, false
	}
	if union := target.Selector.RequiredTypesAny(); len(union) > 0 {
		spec.Predicate.PermanentTypes = append([]types.Card(nil), union...)
	}
	if excludedTypes := target.Selector.ExcludedTypes(); len(excludedTypes) > 0 {
		spec.Predicate.ExcludedTypes = append([]types.Card(nil), excludedTypes...)
	}

	switch {
	case target.Selector.Attacking && target.Selector.Blocking:
		spec.Predicate.CombatState = game.CombatStateAttackingOrBlocking
	case target.Selector.Attacking:
		spec.Predicate.CombatState = game.CombatStateAttacking
	case target.Selector.Blocking:
		spec.Predicate.CombatState = game.CombatStateBlocking
	case target.Selector.Tapped:
		spec.Predicate.Tapped = game.TriTrue
	case target.Selector.Untapped:
		spec.Predicate.Tapped = game.TriFalse
	default:
	}
	switch target.Selector.Controller {
	case compiler.ControllerAny:
	case compiler.ControllerYou:
		spec.Predicate.Controller = game.ControllerYou
	case compiler.ControllerOpponent:
		spec.Predicate.Controller = game.ControllerOpponent
	case compiler.ControllerNotYou:
		spec.Predicate.Controller = game.ControllerNotYou
	default:
		return game.TargetSpec{}, false
	}
	spec.Constraint = lowerFirst(target.Text)
	return spec, true
}

func selectorHasUnsupportedPermanentFilters(selector compiler.CompiledSelector) bool {
	return len(selector.Supertypes()) != 0 ||
		len(selector.ColorsAny()) != 0 ||
		len(selector.ExcludedColors()) != 0 ||
		len(selector.SubtypesAny()) != 0 ||
		selector.Keyword != parser.KeywordUnknown ||
		selector.Zone != zone.None ||
		selector.MatchManaValue ||
		selector.MatchPower ||
		selector.MatchToughness
}

func stackSpellTargetSpec(target compiler.CompiledTarget) (game.TargetSpec, bool) {
	if !targetCardinalityIsOne(target) ||
		target.Selector.Another || target.Selector.Other ||
		target.Selector.Controller != compiler.ControllerAny ||
		target.Selector.Attacking || target.Selector.Blocking ||
		target.Selector.Tapped || target.Selector.Untapped ||
		len(target.Selector.Supertypes()) != 0 ||
		len(target.Selector.ColorsAny()) != 0 ||
		len(target.Selector.ExcludedColors()) != 0 ||
		len(target.Selector.SubtypesAny()) != 0 ||
		target.Selector.Keyword != parser.KeywordUnknown ||
		target.Selector.Zone != zone.None ||
		target.Selector.MatchManaValue ||
		target.Selector.MatchPower ||
		target.Selector.MatchToughness {
		return game.TargetSpec{}, false
	}
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowStackObject,
		Predicate: game.TargetPredicate{
			StackObjectKinds: []game.StackObjectKind{game.StackSpell},
		},
	}
	switch target.Selector.Kind {
	case compiler.SelectorSpell:
		required := target.Selector.RequiredTypesAny()
		excluded := target.Selector.ExcludedTypes()
		if len(required) > 1 || len(excluded) > 1 || len(required) > 0 && len(excluded) > 0 {
			return game.TargetSpec{}, false
		}
		spec.Predicate.SpellCardTypes = append([]types.Card(nil), required...)
		spec.Predicate.ExcludedSpellCardTypes = append([]types.Card(nil), excluded...)
	default:
		return game.TargetSpec{}, false
	}
	spec.Constraint = lowerFirst(target.Text)
	return spec, true
}

func counterAbilityTargetSpec(target compiler.CompiledTarget) (game.TargetSpec, bool) {
	if !targetCardinalityIsOne(target) ||
		target.Selector.Another || target.Selector.Other ||
		target.Selector.Controller != compiler.ControllerAny {
		return game.TargetSpec{}, false
	}
	var kinds []game.StackObjectKind
	switch target.Selector.Kind {
	case compiler.SelectorActivatedAbility:
		kinds = []game.StackObjectKind{game.StackActivatedAbility}
	case compiler.SelectorTriggeredAbility:
		kinds = []game.StackObjectKind{game.StackTriggeredAbility}
	case compiler.SelectorActivatedOrTriggeredAbility:
		kinds = []game.StackObjectKind{game.StackActivatedAbility, game.StackTriggeredAbility}
	case compiler.SelectorSpellActivatedOrTriggeredAbility:
		kinds = []game.StackObjectKind{game.StackSpell, game.StackActivatedAbility, game.StackTriggeredAbility}
	default:
		return game.TargetSpec{}, false
	}
	return game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: lowerFirst(target.Text),
		Allow:      game.TargetAllowStackObject,
		Predicate:  game.TargetPredicate{StackObjectKinds: kinds},
	}, true
}

func counterTargetSpec(target compiler.CompiledTarget) (game.TargetSpec, bool) {
	if spec, ok := stackSpellTargetSpec(target); ok {
		return spec, true
	}
	return counterAbilityTargetSpec(target)
}
