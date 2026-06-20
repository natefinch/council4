package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func lowerActivatedAdditionalCost(cardName string, component compiler.CostComponent) (cost.Additional, bool) {
	switch component.Kind {
	case compiler.CostSacrifice:
		return lowerSacrificeCost(cardName, component)
	case compiler.CostDiscard:
		return lowerDiscardCost(component)
	case compiler.CostPayLife:
		if component.AmountFromX {
			return cost.Additional{
				Kind:        cost.AdditionalPayLife,
				Text:        component.Text,
				AmountFromX: true,
			}, true
		}
		if !component.AmountKnown || component.AmountValue <= 0 {
			return cost.Additional{}, false
		}
		return cost.Additional{
			Kind:   cost.AdditionalPayLife,
			Text:   component.Text,
			Amount: component.AmountValue,
		}, true
	case compiler.CostExile:
		if component.SourceSelf {
			source := zone.Battlefield
			if component.SourceZone != zone.None {
				source = component.SourceZone
			}
			return cost.Additional{
				Kind:   cost.AdditionalExileSource,
				Text:   component.Text,
				Amount: 1,
				Source: source,
			}, true
		}
		return lowerExileCost(component)
	case compiler.CostReveal:
		return lowerRevealCost(component)
	case compiler.CostRemoveCounter:
		return lowerRemoveCounterCost(cardName, component)
	case compiler.CostTapPermanents:
		return lowerTapPermanentsCost(component)
	case compiler.CostEnergy:
		if !component.AmountKnown || component.AmountValue <= 0 {
			return cost.Additional{}, false
		}
		return cost.Additional{
			Kind:   cost.AdditionalEnergy,
			Text:   component.Text,
			Amount: component.AmountValue,
		}, true
	case compiler.CostReturn:
		return lowerReturnToHandCost(component)
	case compiler.CostExert:
		return lowerExertCost(cardName, component)
	case compiler.CostMill:
		return lowerMillCost(component)
	case compiler.CostPutCounter:
		return lowerPutCounterCost(cardName, component)
	case compiler.CostCollectEvidence:
		return lowerCollectEvidenceCost(component)
	default:
		return cost.Additional{}, false
	}
}

// lowerActivationCostComponents is the shared cost-parsing kernel used by both
// lowerActivatedAbility and lowerManaAbility. It iterates the compiled cost
// components and produces (manaCost, additionalCosts):
//
//   - CostMana must be the first component and may appear at most once.
//   - CostTap and CostUntap each may appear at most once.
//   - All other cost kinds are delegated to lowerActivatedAdditionalCost,
//     which covers sacrifice, discard, pay-life, exile, reveal, remove-counter,
//     tap-permanents, energy, return, exert, mill, put-counter, and
//     collect-evidence.
//
// Returns nil, nil, false if any component is unsupported or ordering rules are
// violated. The caller must check that ability.Cost is non-nil and non-empty
// before calling.
func lowerActivationCostComponents(
	cardName string,
	compiled *compiler.CompiledCost,
) (cost.Mana, []cost.Additional, bool) {
	var manaCost cost.Mana
	var additionalCosts []cost.Additional
	for i, component := range compiled.Components {
		switch component.Kind {
		case compiler.CostMana:
			if i != 0 || manaCost != nil {
				return nil, nil, false
			}
			parsed, err := parseManaCostValue(component.Symbol)
			if err != nil || len(parsed) == 0 {
				return nil, nil, false
			}
			manaCost = parsed
		case compiler.CostTap:
			if slices.ContainsFunc(additionalCosts, func(a cost.Additional) bool {
				return a.Kind == cost.AdditionalTap
			}) {
				return nil, nil, false
			}
			additionalCosts = append(additionalCosts, cost.T)
		case compiler.CostUntap:
			if slices.ContainsFunc(additionalCosts, func(a cost.Additional) bool {
				return a.Kind == cost.AdditionalUntap
			}) {
				return nil, nil, false
			}
			additionalCosts = append(additionalCosts, cost.Additional{
				Kind: cost.AdditionalUntap,
				Text: component.Text,
			})
		default:
			additional, ok := lowerActivatedAdditionalCost(cardName, component)
			if !ok {
				return nil, nil, false
			}
			additionalCosts = append(additionalCosts, additional)
		}
	}
	return manaCost, additionalCosts, true
}

func lowerCollectEvidenceCost(component compiler.CostComponent) (cost.Additional, bool) {
	if !component.AmountKnown || component.AmountValue <= 0 {
		return cost.Additional{}, false
	}
	return cost.Additional{
		Kind:   cost.AdditionalCollectEvidence,
		Text:   component.Text,
		Amount: component.AmountValue,
		Source: zone.Graveyard,
	}, true
}

func lowerExertCost(_ string, component compiler.CostComponent) (cost.Additional, bool) {
	if !component.SourceSelf {
		return cost.Additional{}, false
	}
	return cost.Additional{
		Kind: cost.AdditionalExert,
		Text: component.Text,
	}, true
}

func lowerMillCost(component compiler.CostComponent) (cost.Additional, bool) {
	if !component.AmountKnown || component.ObjectKind != compiler.SelectorCard {
		return cost.Additional{}, false
	}
	return cost.Additional{
		Kind:   cost.AdditionalMill,
		Text:   component.Text,
		Amount: component.AmountValue,
	}, true
}

func lowerPutCounterCost(_ string, component compiler.CostComponent) (cost.Additional, bool) {
	if !component.AmountKnown || !component.CounterKindKnown || !component.SourceSelf {
		return cost.Additional{}, false
	}
	return cost.Additional{
		Kind:        cost.AdditionalPutCounter,
		Text:        component.Text,
		Amount:      component.AmountValue,
		CounterKind: component.CounterKind,
	}, true
}

func lowerRevealCost(component compiler.CostComponent) (cost.Additional, bool) {
	if component.SourceZone != zone.Hand || component.ObjectKind != compiler.SelectorCard {
		return cost.Additional{}, false
	}
	if !component.AmountKnown && !component.AmountFromX {
		return cost.Additional{}, false
	}
	additional := cost.Additional{
		Kind:   cost.AdditionalReveal,
		Text:   component.Text,
		Source: zone.Hand,
	}
	if component.AmountFromX {
		additional.AmountFromX = true
	} else {
		additional.Amount = component.AmountValue
	}
	if component.ObjectColorKnown {
		additional.MatchCardColor = true
		additional.CardColor = component.ObjectColor
	}
	return additional, true
}

func lowerReturnToHandCost(component compiler.CostComponent) (cost.Additional, bool) {
	if !component.AmountKnown ||
		component.ObjectController != compiler.ControllerYou ||
		component.ToZone != zone.Hand {
		return cost.Additional{}, false
	}
	additional := cost.Additional{
		Kind:          cost.AdditionalReturnToHand,
		Text:          component.Text,
		Amount:        component.AmountValue,
		RequireTapped: component.RequireTapped,
	}
	if lowerCostPermanentObject(component, &additional, true) {
		return additional, true
	}
	return cost.Additional{}, false
}

func lowerCostPermanentObject(component compiler.CostComponent, additional *cost.Additional, allowSnowLand bool) bool {
	if component.ObjectColorKnown || component.ObjectNonToken {
		return false
	}
	switch component.ObjectKind {
	case compiler.SelectorPermanent:
		return true
	case compiler.SelectorArtifact, compiler.SelectorCreature, compiler.SelectorEnchantment, compiler.SelectorLand:
		additional.MatchPermanentType = true
		additional.PermanentType = component.ObjectType
		if component.ObjectTypeAltKnown {
			additional.PermanentTypeAlt = component.ObjectTypeAlt
		}
		if allowSnowLand && component.SupertypeKnown && component.ObjectSupertype == types.Snow && component.ObjectType == types.Land {
			additional.RequireSupertype = types.Snow
		}
		return true
	default:
	}
	if len(component.SubtypesAny) == 1 {
		additional.SubtypesAny = cost.SubtypeSet{component.SubtypesAny[0]}
		return true
	}
	if len(component.SubtypesAny) == 2 {
		additional.SubtypesAny = cost.SubtypeSet{component.SubtypesAny[0], component.SubtypesAny[1]}
		return true
	}
	return false
}

func lowerTapPermanentsCost(component compiler.CostComponent) (cost.Additional, bool) {
	if !component.AmountKnown ||
		!component.RequireUntapped ||
		component.ObjectController != compiler.ControllerYou {
		return cost.Additional{}, false
	}
	additional := cost.Additional{
		Kind:   cost.AdditionalTapPermanents,
		Text:   component.Text,
		Amount: component.AmountValue,
	}
	if lowerCostPermanentObject(component, &additional, false) {
		return additional, true
	}
	return cost.Additional{}, false
}

func lowerRemoveCounterCost(
	_ string,
	component compiler.CostComponent,
) (cost.Additional, bool) {
	if !component.AmountKnown || !component.CounterKindKnown || !component.SourceSelf {
		return cost.Additional{}, false
	}
	return cost.Additional{
		Kind:        cost.AdditionalRemoveCounter,
		Text:        component.Text,
		Amount:      component.AmountValue,
		CounterKind: component.CounterKind,
	}, true
}

func lowerExileCost(component compiler.CostComponent) (cost.Additional, bool) {
	if component.SourceZone != zone.Graveyard ||
		component.ObjectKind != compiler.SelectorCard {
		return cost.Additional{}, false
	}
	if !component.AmountKnown && !component.AmountFromX {
		return cost.Additional{}, false
	}
	additional := cost.Additional{
		Kind:   cost.AdditionalExile,
		Text:   component.Text,
		Source: zone.Graveyard,
	}
	if component.AmountFromX {
		additional.AmountFromX = true
	} else {
		additional.Amount = component.AmountValue
	}
	if component.ObjectTypeKnown {
		additional.MatchCardType = true
		additional.CardType = component.ObjectType
	}
	if len(component.SubtypesAny) == 1 {
		additional.SubtypesAny = cost.SubtypeSet{component.SubtypesAny[0]}
	}
	return additional, true
}

func lowerSacrificeCost(_ string, component compiler.CostComponent) (cost.Additional, bool) {
	if component.SourceSelf {
		if component.ExcludeSource {
			return cost.Additional{}, false
		}
		return cost.Additional{
			Kind:   cost.AdditionalSacrificeSource,
			Text:   component.Text,
			Amount: 1,
		}, true
	}
	if !component.AmountKnown {
		return cost.Additional{}, false
	}
	additional := cost.Additional{
		Kind:          cost.AdditionalSacrifice,
		Text:          component.Text,
		Amount:        component.AmountValue,
		ExcludeSource: component.ExcludeSource,
	}
	if !lowerCostPermanentObject(component, &additional, false) {
		return cost.Additional{}, false
	}
	return additional, true
}

func lowerDiscardCost(component compiler.CostComponent) (cost.Additional, bool) {
	if component.SourceSelf {
		if !component.AmountKnown || component.AmountValue != 1 || component.SourceZone != zone.Hand {
			return cost.Additional{}, false
		}
		return cost.Additional{
			Kind:       cost.AdditionalDiscard,
			Text:       component.Text,
			Amount:     1,
			Source:     zone.Hand,
			SourceSelf: true,
		}, true
	}
	if !component.AmountKnown ||
		component.ObjectKind != compiler.SelectorCard ||
		component.ObjectColorKnown ||
		component.ObjectNonToken ||
		component.PermanentModifier {
		return cost.Additional{}, false
	}
	additional := cost.Additional{
		Kind:   cost.AdditionalDiscard,
		Text:   component.Text,
		Amount: component.AmountValue,
		Source: zone.Hand,
	}
	if component.ObjectTypeKnown {
		additional.MatchCardType = true
		additional.CardType = component.ObjectType
	}
	return additional, true
}
