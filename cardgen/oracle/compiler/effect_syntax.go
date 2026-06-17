package compiler

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func compileEffectPayment(payment parser.EffectPaymentSyntax) CompiledEffectPayment {
	return CompiledEffectPayment{
		Span:     payment.Span,
		Payer:    payment.Payer,
		ManaCost: slices.Clone(payment.ManaCost),
		Order:    payment.Order,
	}
}

func applyEffectPaymentsToConditions(effects []CompiledEffect, conditions []CompiledCondition) {
	for i := range effects {
		effect := &effects[i]
		if effect.Payment.Payer != parser.EffectPaymentPayerTargetController || len(effect.Payment.ManaCost) == 0 {
			continue
		}
		for i := range conditions {
			if conditions[i].Order.Contains(effect.Payment.Order) {
				conditions[i].Predicate = ConditionPredicateTargetControllerDoesNotPay
			}
		}
	}
}

func compileTypedTargets(sentences []parser.Sentence) []CompiledTarget {
	var targets []CompiledTarget
	for _, sentence := range sentences {
		targets = append(targets, compileTypedTargetList(sentence.Targets)...)
	}
	return targets
}

func compileTypedTargetList(syntaxes []parser.TargetSyntax) []CompiledTarget {
	targets := make([]CompiledTarget, 0, len(syntaxes))
	for _, syntax := range syntaxes {
		targets = append(targets, CompiledTarget{
			Span: syntax.Span,
			Text: syntax.Text,
			Cardinality: TargetCardinality{
				Min: syntax.Cardinality.Min,
				Max: syntax.Cardinality.Max,
			},
			Selector: compileTypedSelection(syntax.Selection),
			Exact:    syntax.Exact,
			Order:    syntax.Order,
		})
	}
	return targets
}

func compileTypedSelection(syntax parser.SelectionSyntax) CompiledSelector {
	selector := CompiledSelector{
		Kind:           compileSelectionKind(syntax.Kind),
		Controller:     compileSelectionController(syntax.Controller),
		All:            syntax.All,
		Another:        syntax.Another,
		Other:          syntax.Other,
		Attacking:      syntax.Attacking,
		Blocking:       syntax.Blocking,
		Tapped:         syntax.Tapped,
		Untapped:       syntax.Untapped,
		Keyword:        syntax.Keyword,
		Zone:           syntax.Zone,
		ManaValue:      syntax.ManaValue,
		MatchManaValue: syntax.MatchManaValue,
		Power:          syntax.Power,
		MatchPower:     syntax.MatchPower,
		Toughness:      syntax.Toughness,
		MatchToughness: syntax.MatchToughness,
		Colorless:      syntax.Colorless,
		Multicolored:   syntax.Multicolored,
	}
	if len(syntax.RequiredTypesAny) > 1 ||
		syntax.Kind == parser.SelectionSpell && len(syntax.RequiredTypesAny) == 1 {
		for _, cardType := range syntax.RequiredTypesAny {
			if value, ok := runtimeCardTypeFromParser(cardType); ok {
				setSelectorRequiredTypesAny(&selector, append(selector.RequiredTypesAny(), value))
			}
		}
	}

	for _, cardType := range syntax.ExcludedTypes {
		if value, ok := runtimeCardTypeFromParser(cardType); ok {
			appendSelectorExcludedType(&selector, value)
		}
	}
	for _, supertype := range syntax.Supertypes {
		switch supertype {
		case parser.SupertypeBasic:
			appendSelectorSupertype(&selector, types.Basic)
		case parser.SupertypeLegendary:
			appendSelectorSupertype(&selector, types.Legendary)
		case parser.SupertypeSnow:
			appendSelectorSupertype(&selector, types.Snow)
		case parser.SupertypeWorld:
			appendSelectorSupertype(&selector, types.World)
		default:
		}
	}
	for _, colorValue := range syntax.ColorsAny {
		if value, ok := runtimeColorFromParser(colorValue); ok {
			appendSelectorColorAny(&selector, value)
		}
	}
	for _, colorValue := range syntax.ExcludedColors {
		if value, ok := runtimeColorFromParser(colorValue); ok {
			appendSelectorExcludedColor(&selector, value)
		}
	}
	appendSelectorSubtypesAny(&selector, syntax.SubtypesAny...)
	return selector
}

func compileStaticSubjectKind(kind parser.EffectStaticSubjectKind) StaticSubjectKind {
	switch kind {
	case parser.EffectStaticSubjectAttachedObject:
		return StaticSubjectAttachedObject
	case parser.EffectStaticSubjectAllCreatures:
		return StaticSubjectAllCreatures
	case parser.EffectStaticSubjectAllOtherCreatures:
		return StaticSubjectAllOtherCreatures
	case parser.EffectStaticSubjectAttackingCreatures:
		return StaticSubjectAttackingCreatures
	case parser.EffectStaticSubjectBlockingCreatures:
		return StaticSubjectBlockingCreatures
	case parser.EffectStaticSubjectControlledCreatures:
		return StaticSubjectControlledCreatures
	case parser.EffectStaticSubjectOtherControlledCreatures:
		return StaticSubjectOtherControlledCreatures
	case parser.EffectStaticSubjectControlledWalls:
		return StaticSubjectControlledWalls
	case parser.EffectStaticSubjectControlledArtifacts:
		return StaticSubjectControlledArtifacts
	case parser.EffectStaticSubjectControlledTokens:
		return StaticSubjectControlledTokens
	case parser.EffectStaticSubjectOpponentControlledCreatures:
		return StaticSubjectOpponentControlledCreatures
	case parser.EffectStaticSubjectControlledCreatureSubtype:
		return StaticSubjectControlledCreatureSubtype
	case parser.EffectStaticSubjectOtherControlledCreatureSubtype:
		return StaticSubjectOtherControlledCreatureSubtype
	default:
		return StaticSubjectNone
	}
}

func compileSelectionKind(kind parser.SelectionKind) SelectorKind {
	switch kind {
	case parser.SelectionAny:
		return SelectorAny
	case parser.SelectionPlayer:
		return SelectorPlayer
	case parser.SelectionOpponent:
		return SelectorOpponent
	case parser.SelectionArtifact:
		return SelectorArtifact
	case parser.SelectionCreature:
		return SelectorCreature
	case parser.SelectionEnchantment:
		return SelectorEnchantment
	case parser.SelectionLand:
		return SelectorLand
	case parser.SelectionPermanent:
		return SelectorPermanent
	case parser.SelectionCard:
		return SelectorCard
	case parser.SelectionSpell:
		return SelectorSpell
	case parser.SelectionActivatedAbility:
		return SelectorActivatedAbility
	case parser.SelectionTriggeredAbility:
		return SelectorTriggeredAbility
	case parser.SelectionActivatedOrTriggeredAbility:
		return SelectorActivatedOrTriggeredAbility
	case parser.SelectionSpellActivatedOrTriggeredAbility:
		return SelectorSpellActivatedOrTriggeredAbility
	case parser.SelectionPlaneswalker:
		return SelectorPlaneswalker
	case parser.SelectionBattle:
		return SelectorBattle
	default:
		return SelectorUnknown
	}
}

func compileSelectionController(controller parser.SelectionController) ControllerKind {
	switch controller {
	case parser.SelectionControllerYou:
		return ControllerYou
	case parser.SelectionControllerOpponent:
		return ControllerOpponent
	case parser.SelectionControllerNotYou:
		return ControllerNotYou
	default:
		return ControllerAny
	}
}

func compileEffectKind(kind parser.EffectKind) EffectKind {
	switch kind {
	case parser.EffectAddMana:
		return EffectAddMana
	case parser.EffectAttach:
		return EffectAttach
	case parser.EffectCast:
		return EffectCast
	case parser.EffectCounter:
		return EffectCounter
	case parser.EffectCreate:
		return EffectCreate
	case parser.EffectDealDamage:
		return EffectDealDamage
	case parser.EffectDestroy:
		return EffectDestroy
	case parser.EffectDiscard:
		return EffectDiscard
	case parser.EffectDiscover:
		return EffectDiscover
	case parser.EffectDouble:
		return EffectDouble
	case parser.EffectDraw:
		return EffectDraw
	case parser.EffectEnterTapped:
		return EffectEnterTapped
	case parser.EffectEnterPrepared:
		return EffectEnterPrepared
	case parser.EffectExile:
		return EffectExile
	case parser.EffectFight:
		return EffectFight
	case parser.EffectGain:
		return EffectGain
	case parser.EffectGainControl:
		return EffectGainControl
	case parser.EffectGrantKeyword:
		return EffectGrantKeyword
	case parser.EffectInvestigate:
		return EffectInvestigate
	case parser.EffectExplore:
		return EffectExplore
	case parser.EffectLose:
		return EffectLose
	case parser.EffectManifest:
		return EffectManifest
	case parser.EffectManifestDread:
		return EffectManifestDread
	case parser.EffectMill:
		return EffectMill
	case parser.EffectModifyPT:
		return EffectModifyPT
	case parser.EffectPut:
		return EffectPut
	case parser.EffectProliferate:
		return EffectProliferate
	case parser.EffectRegenerate:
		return EffectRegenerate
	case parser.EffectReturn:
		return EffectReturn
	case parser.EffectReveal:
		return EffectReveal
	case parser.EffectSacrifice:
		return EffectSacrifice
	case parser.EffectScry:
		return EffectScry
	case parser.EffectSurveil:
		return EffectSurveil
	case parser.EffectSearch:
		return EffectSearch
	case parser.EffectShuffle:
		return EffectShuffle
	case parser.EffectTap:
		return EffectTap
	case parser.EffectUntap:
		return EffectUntap
	case parser.EffectTransform:
		return EffectTransform
	default:
		return EffectUnknown
	}
}

func compileEffectDuration(duration parser.EffectDurationKind) DurationKind {
	switch duration {
	case parser.EffectDurationUntilEndOfTurn:
		return DurationUntilEndOfTurn
	case parser.EffectDurationUntilYourNextTurn:
		return DurationUntilYourNextTurn
	case parser.EffectDurationThisTurn:
		return DurationThisTurn
	case parser.EffectDurationThisCombat:
		return DurationThisCombat
	case parser.EffectDurationWhileSourceOnBattlefield:
		return DurationForAsLongAsSourceOnBattlefield
	case parser.EffectDurationWhileYouControlSource:
		return DurationForAsLongAsYouControlSource
	default:
		return DurationNone
	}
}

func compileDelayedTiming(timing parser.DelayedTimingKind) game.DelayedTriggerTiming {
	switch timing {
	case parser.DelayedTimingNextEndStep:
		return game.DelayedAtBeginningOfNextEndStep
	case parser.DelayedTimingNextUpkeep:
		return game.DelayedAtBeginningOfNextUpkeep
	default:
		return 0
	}
}

func compileTypedAmount(amount parser.EffectAmountSyntax) CompiledAmount {
	compiled := CompiledAmount{
		Value:         amount.Value,
		Known:         amount.Known,
		VariableX:     amount.VariableX,
		DynamicKind:   compileDynamicAmountKind(amount.DynamicKind),
		DynamicForm:   compileDynamicAmountForm(amount.DynamicForm),
		Multiplier:    amount.Multiplier,
		ReferenceSpan: amount.ReferenceSpan,
		Text:          amount.Text,
	}
	if amount.Selection != nil {
		selection := compileTypedSelection(*amount.Selection)
		compiled.selector = &selection
	}
	return compiled
}

func compileDynamicAmountKind(kind parser.EffectDynamicAmountKind) DynamicAmountKind {
	switch kind {
	case parser.EffectDynamicAmountCount:
		return DynamicAmountCount
	case parser.EffectDynamicAmountControllerLife:
		return DynamicAmountControllerLife
	case parser.EffectDynamicAmountOpponentCount:
		return DynamicAmountOpponentCount
	case parser.EffectDynamicAmountSourcePower:
		return DynamicAmountSourcePower
	case parser.EffectDynamicAmountBasicLandTypes:
		return DynamicAmountBasicLandTypes
	default:
		return DynamicAmountNone
	}
}

func compileDynamicAmountForm(form parser.EffectDynamicAmountForm) DynamicAmountForm {
	switch form {
	case parser.EffectDynamicAmountFormEqual:
		return DynamicAmountEqual
	case parser.EffectDynamicAmountFormForEach:
		return DynamicAmountForEach
	case parser.EffectDynamicAmountFormWhereX:
		return DynamicAmountWhereX
	default:
		return DynamicAmountFormNone
	}
}

func compileSignedAmount(amount parser.SignedAmountSyntax) CompiledSignedAmount {
	return CompiledSignedAmount{
		Value:    amount.Value,
		Known:    amount.Known,
		Negative: amount.Negative,
	}
}
