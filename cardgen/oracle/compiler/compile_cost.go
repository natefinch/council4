package compiler

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

// compileCost maps the parser's typed Cost onto the semantic cost IR. It reads
// typed cost components and never inspects retained cost text to derive meaning.
func compileCost(parserCost parser.Cost) CompiledCost {
	cost := CompiledCost{Span: parserCost.Span, Text: parserCost.Text, Order: parserCost.Order}
	for _, component := range parserCost.Components {
		cost.Components = append(cost.Components, compileCostComponent(component))
	}
	return cost
}

func compileCostComponent(component parser.CostComponent) CostComponent {
	compiled := CostComponent{
		Kind:                 compileCostKind(component.Kind),
		Span:                 component.Span,
		Text:                 component.Text,
		Symbol:               component.Symbol,
		Amount:               component.Amount,
		Object:               component.Object,
		AmountValue:          component.AmountValue,
		AmountKnown:          component.AmountKnown,
		AmountFromX:          component.AmountFromX,
		ObjectSupertype:      component.ObjectSupertype,
		SupertypeKnown:       component.SupertypeKnown,
		ObjectController:     compilerControllerRelation(component.ObjectController),
		RequireTapped:        component.RequireTapped,
		RequireUntapped:      component.RequireUntapped,
		SourceZone:           component.SourceZone,
		ToZone:               component.ToZone,
		SourceSelf:           component.SourceSelf,
		CounterKind:          component.CounterKind,
		CounterKindKnown:     component.CounterKindKnown,
		SubtypesAny:          append([]types.Sub(nil), component.SubtypesAny...),
		ExcludeSource:        component.ExcludeSource,
		DiscardWholeHand:     component.DiscardWholeHand,
		ChoiceGroup:          component.ChoiceGroup,
		PayLifeAmountDynamic: compilePayLifeDynamic(component.PayLifeDynamic),
		Order:                component.Order,
	}
	if component.ObjectColorKnown {
		if mapped, ok := compilerColor(component.ObjectColor); ok {
			compiled.ObjectColor = mapped
			compiled.ObjectColorKnown = true
		}
	}
	applyCostObjectNoun(&compiled, component)
	return compiled
}

// applyCostObjectNoun derives the selector kind and card type from the parser's
// typed object noun. A card object selects SelectorCard with an optional card
// type; a permanent object maps the noun onto its permanent selector.
func applyCostObjectNoun(compiled *CostComponent, component parser.CostComponent) {
	if component.ObjectIsCard {
		compiled.ObjectKind = SelectorCard
		if typ, ok := costCardTypeFromNoun(component.ObjectNoun); ok {
			compiled.ObjectType = typ
			compiled.ObjectTypeKnown = true
		}
		return
	}
	annotateCostObjectNoun(compiled, component.ObjectNoun)
	if typ, ok := costPermanentTypeFromNoun(component.SecondObjectNoun); ok {
		compiled.ObjectTypeAlt = typ
		compiled.ObjectTypeAltKnown = true
	}
}

// costPermanentTypeFromNoun maps a permanent-type object noun onto its runtime
// card type. It covers the two-type cost-union nouns, including planeswalker,
// which costCardTypeFromNoun omits.
func costPermanentTypeFromNoun(noun parser.ObjectNoun) (types.Card, bool) {
	switch noun {
	case parser.ObjectNounArtifact:
		return types.Artifact, true
	case parser.ObjectNounCreature:
		return types.Creature, true
	case parser.ObjectNounEnchantment:
		return types.Enchantment, true
	case parser.ObjectNounLand:
		return types.Land, true
	case parser.ObjectNounPlaneswalker:
		return types.Planeswalker, true
	default:
		return "", false
	}
}

func compilePayLifeDynamic(kind parser.PayLifeDynamicAmount) DynamicAmountKind {
	switch kind {
	case parser.PayLifeDynamicCommanderColorIdentityCount:
		return DynamicAmountCommanderColorCount
	default:
		return DynamicAmountNone
	}
}

func compileCostKind(kind parser.CostComponentKind) CostKind {
	switch kind {
	case parser.CostComponentMana:
		return CostMana
	case parser.CostComponentTap:
		return CostTap
	case parser.CostComponentUntap:
		return CostUntap
	case parser.CostComponentSacrifice:
		return CostSacrifice
	case parser.CostComponentDiscard:
		return CostDiscard
	case parser.CostComponentPayLife:
		return CostPayLife
	case parser.CostComponentExile:
		return CostExile
	case parser.CostComponentRemoveCounter:
		return CostRemoveCounter
	case parser.CostComponentReveal:
		return CostReveal
	case parser.CostComponentTapPermanents:
		return CostTapPermanents
	case parser.CostComponentEnergy:
		return CostEnergy
	case parser.CostComponentReturn:
		return CostReturn
	case parser.CostComponentExert:
		return CostExert
	case parser.CostComponentMill:
		return CostMill
	case parser.CostComponentPutCounter:
		return CostPutCounter
	case parser.CostComponentCollectEvidence:
		return CostCollectEvidence
	case parser.CostComponentLoyalty:
		return CostLoyalty
	default:
		return CostUnknown
	}
}

// annotateCostObjectNoun maps a typed parser object noun onto the semantic
// permanent selector and card type. It consumes the typed noun atom and reads
// no cost text.
func annotateCostObjectNoun(component *CostComponent, noun parser.ObjectNoun) bool {
	switch noun {
	case parser.ObjectNounArtifact:
		component.ObjectKind = SelectorArtifact
		component.ObjectType = types.Artifact
		component.ObjectTypeKnown = true
		return true
	case parser.ObjectNounCreature:
		component.ObjectKind = SelectorCreature
		component.ObjectType = types.Creature
		component.ObjectTypeKnown = true
		return true
	case parser.ObjectNounEnchantment:
		component.ObjectKind = SelectorEnchantment
		component.ObjectType = types.Enchantment
		component.ObjectTypeKnown = true
		return true
	case parser.ObjectNounLand:
		component.ObjectKind = SelectorLand
		component.ObjectType = types.Land
		component.ObjectTypeKnown = true
		return true
	case parser.ObjectNounPermanent:
		component.ObjectKind = SelectorPermanent
		component.PermanentModifier = true
		return true
	case parser.ObjectNounCard:
		component.ObjectKind = SelectorCard
		return true
	default:
		return false
	}
}

func costCardTypeFromNoun(noun parser.ObjectNoun) (types.Card, bool) {
	switch noun {
	case parser.ObjectNounArtifact:
		return types.Artifact, true
	case parser.ObjectNounCreature:
		return types.Creature, true
	case parser.ObjectNounEnchantment:
		return types.Enchantment, true
	case parser.ObjectNounLand:
		return types.Land, true
	default:
		return "", false
	}
}

func compilerCardType(cardType parser.CardType) (types.Card, bool) {
	switch cardType {
	case parser.CardTypeArtifact:
		return types.Artifact, true
	case parser.CardTypeBattle:
		return types.Battle, true
	case parser.CardTypeCreature:
		return types.Creature, true
	case parser.CardTypeEnchantment:
		return types.Enchantment, true
	case parser.CardTypeInstant:
		return types.Instant, true
	case parser.CardTypeLand:
		return types.Land, true
	case parser.CardTypePlaneswalker:
		return types.Planeswalker, true
	case parser.CardTypeSorcery:
		return types.Sorcery, true
	default:
		return "", false
	}
}

func compilerSupertype(supertype parser.Supertype) (types.Super, bool) {
	switch supertype {
	case parser.SupertypeLegendary:
		return types.Legendary, true
	case parser.SupertypeSnow:
		return types.Snow, true
	case parser.SupertypeBasic:
		return types.Basic, true
	case parser.SupertypeWorld:
		return types.World, true
	default:
		return "", false
	}
}

func compilerColor(value parser.Color) (color.Color, bool) {
	switch value {
	case parser.ColorWhite:
		return color.White, true
	case parser.ColorBlue:
		return color.Blue, true
	case parser.ColorBlack:
		return color.Black, true
	case parser.ColorRed:
		return color.Red, true
	case parser.ColorGreen:
		return color.Green, true
	default:
		return "", false
	}
}

func compilerControllerRelation(relation parser.ControllerRelation) ControllerKind {
	switch relation {
	case parser.ControllerRelationYouControl:
		return ControllerYou
	case parser.ControllerRelationYouDontControl:
		return ControllerNotYou
	case parser.ControllerRelationOpponentControls:
		return ControllerOpponent
	default:
		return ControllerAny
	}
}
