package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func stackSpellTargetSpec(target compiler.CompiledTarget) (game.TargetSpec, bool) {
	if !targetCardinalityIsOne(target) {
		return game.TargetSpec{}, false
	}
	return stackSpellTargetSpecWithBounds(target, 1, 1)
}

func stackSpellTargetSpecUnbounded(target compiler.CompiledTarget) (game.TargetSpec, bool) {
	if !targetCardinalityIsUnbounded(target) {
		return game.TargetSpec{}, false
	}
	return stackSpellTargetSpecWithBounds(target, target.Cardinality.Min, target.Cardinality.Max)
}

// The parser owns qualifier scope and consumption; this projection uses only
// typed filters, also shared by constructed condition subjects. Each caller
// still owns its payload, cardinality and reference policy.
func stackSpellTargetSpecWithBounds(target compiler.CompiledTarget, minTargets, maxTargets int) (game.TargetSpec, bool) {
	selector := target.Selector
	if selector.Kind != compiler.SelectorSpell ||
		selector.All || selector.Another || selector.Other ||
		selector.Attacking || selector.Blocking || selector.Tapped || selector.Untapped ||
		selector.NonToken || selector.TokenOnly || selector.Colored || selector.Historic ||
		selector.BasicLandType || selector.ConjunctiveTypes || selector.PlayerOrPlaneswalker ||
		len(selector.Supertypes()) != 0 || len(selector.ExcludedSupertypes()) != 0 ||
		len(selector.SubtypesAny()) != 0 || len(selector.ExcludedSubtypes()) != 0 ||
		len(selector.SourceTypes()) != 0 || len(selector.Alternatives) != 0 ||
		selector.Keyword != parser.KeywordUnknown || selector.ExcludedKeyword != parser.KeywordUnknown ||
		selector.Zone != zone.None ||
		selectorHasCounterQualifier(selector) || selectorHasAttachmentQualifier(selector) ||
		selector.MatchPower || selector.MatchToughness ||
		selector.PowerLessThanSource || selector.PowerGreaterThanSource ||
		selector.MatchTotalManaValue || selector.ManaValueX ||
		selector.ManaValueDynamic != compiler.DynamicAmountNone || selector.ManaValueDynamicCount != nil ||
		selector.ManaValueSacrificedCost != nil ||
		selector.ManaValueLessThanEventPermanent || selector.ManaValueLessOrEqualEventPermanent ||
		selector.SubtypeFromEntryChoice || selector.ColorFromEntryChoice ||
		selector.SubtypeFromChosenType || selector.SubtypeFromChosenTypeExcluded ||
		selector.RequiredName != "" || selector.EnteredThisTurn || selector.DealtDamageThisTurn ||
		selector.NameUniqueAmongControlled || selector.SingleGraveyard || selector.SameNameGroup != nil {
		return game.TargetSpec{}, false
	}
	controller, ok := counterAbilityController(selector.Controller)
	if !ok {
		return game.TargetSpec{}, false
	}
	required := selector.RequiredTypesAny()
	excluded := selector.ExcludedTypes()
	colors := selector.ColorsAny()
	excludedColors := selector.ExcludedColors()
	colorShapes := len(colors) + len(excludedColors)
	if selector.Colorless {
		colorShapes++
	}
	if selector.Multicolored {
		colorShapes++
	}
	if colorShapes > 1 || len(excluded) > 1 ||
		len(required) > 1 && (len(excluded) > 0 || colorShapes > 0) {
		return game.TargetSpec{}, false
	}
	predicate := game.TargetPredicate{
		StackObjectKinds:       []game.StackObjectKind{game.StackSpell},
		Controller:             controller,
		ExcludedSpellCardTypes: append([]types.Card(nil), excluded...),
		SpellColors:            append([]color.Color(nil), colors...),
		SpellExcludedColors:    append([]color.Color(nil), excludedColors...),
		SpellColorless:         selector.Colorless,
		SpellMulticolored:      selector.Multicolored,
	}
	if selector.MatchManaValue {
		predicate.ManaValue = opt.Val(selector.ManaValue)
	}
	if len(required) == 1 {
		predicate.SpellCardTypes = append([]types.Card(nil), required...)
	} else if len(required) > 1 {
		predicate.SpellCardTypesAny = append([]types.Card(nil), required...)
	}
	restrictions, ok := lowerSpellTargetRestrictions(selector.SpellTargetRestrictions)
	if !ok {
		return game.TargetSpec{}, false
	}
	predicate.SpellTargets = restrictions
	return game.TargetSpec{
		MinTargets: minTargets,
		MaxTargets: maxTargets,
		Allow:      game.TargetAllowStackObject,
		Predicate:  predicate,
		Constraint: lowerFirst(target.Text),
	}, true
}
