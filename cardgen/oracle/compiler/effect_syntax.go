package compiler

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
)

func compileEffectPayment(payment parser.EffectPaymentSyntax) CompiledEffectPayment {
	compiled := CompiledEffectPayment{
		Span:                   payment.Span,
		Form:                   payment.Form,
		Payer:                  payment.Payer,
		ManaCost:               slices.Clone(payment.ManaCost),
		GenericManaAmount:      compileTypedAmount(payment.GenericManaAmount),
		SuccessConditionNodeID: payment.SuccessConditionNodeID,
		FailureConditionNodeID: payment.FailureConditionNodeID,
		Order:                  payment.Order,
	}
	if payment.AdditionalCost != nil {
		additional := compileCost(*payment.AdditionalCost)
		compiled.AdditionalCost = &additional
	}
	return compiled
}

func applyEffectPaymentsToConditions(effects []CompiledEffect, conditions []CompiledCondition) {
	for i := range effects {
		effect := &effects[i]
		if len(effect.Payment.ManaCost) == 0 {
			continue
		}
		var predicate ConditionPredicate
		switch effect.Payment.Payer {
		case parser.EffectPaymentPayerTargetController:
			predicate = ConditionPredicateTargetControllerDoesNotPay
		case parser.EffectPaymentPayerEventPlayer:
			predicate = ConditionPredicateEventPlayerDoesNotPay
		default:
			continue
		}
		if effect.Payment.Form == parser.EffectPaymentFormMayPayThenIfDoesNot {
			continue
		}
		for i := range conditions {
			if conditions[i].Order.Contains(effect.Payment.Order) ||
				effect.Payment.GenericManaAmount.DynamicKind != DynamicAmountNone &&
					conditions[i].Order.Start == effect.Payment.Order.Start {
				conditions[i].Predicate = predicate
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
	for i := range syntaxes {
		syntax := &syntaxes[i]
		targets = append(targets, CompiledTarget{
			Span:       syntax.Span,
			ChoiceSpan: syntax.ChoiceSpan,
			Text:       syntax.Text,
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

// compileDamageRecipientSelectors compiles the two recipient groups of a
// dual-recipient group-damage effect onto the closed selector vocabulary. It
// returns nil when the parser found no recipient pair so single-recipient damage
// keeps using the lone Selector.
func compileDamageRecipientSelectors(pair []parser.SelectionSyntax) []CompiledSelector {
	if len(pair) == 0 {
		return nil
	}
	selectors := make([]CompiledSelector, 0, len(pair))
	for i := range pair {
		selectors = append(selectors, compileTypedSelection(pair[i]))
	}
	return selectors
}

// compileColorsAmongSelector compiles the permanent filter of a "one mana of any
// color among <permanents> you control" body, returning nil when the parser
// recorded no filter (a non-among-controlled mana body).
func compileColorsAmongSelector(syntax *parser.SelectionSyntax) *CompiledSelector {
	if syntax == nil {
		return nil
	}
	selector := compileTypedSelection(*syntax)
	return &selector
}

// compileTokenCopyForEachGroup compiles the controlled battlefield group a
// per-each copy-token create iterates, returning the zero CompiledSelector when
// the parser recorded no group (a non-per-each create).
func compileTokenCopyForEachGroup(syntax *parser.SelectionSyntax) CompiledSelector {
	if syntax == nil {
		return CompiledSelector{}
	}
	return compileTypedSelection(*syntax)
}

func compileTypedSelection(syntax parser.SelectionSyntax) CompiledSelector {
	selector := CompiledSelector{
		Kind:                   compileSelectionKind(syntax.Kind),
		Controller:             compileSelectionController(syntax.Controller),
		All:                    syntax.All,
		Another:                syntax.Another,
		Other:                  syntax.Other,
		Attacking:              syntax.Attacking,
		Blocking:               syntax.Blocking,
		Tapped:                 syntax.Tapped,
		Untapped:               syntax.Untapped,
		NonToken:               syntax.NonToken,
		TokenOnly:              syntax.TokenOnly,
		Keyword:                syntax.Keyword,
		ExcludedKeyword:        syntax.ExcludedKeyword,
		Zone:                   syntax.Zone,
		ManaValue:              syntax.ManaValue,
		MatchManaValue:         syntax.MatchManaValue,
		ManaValueX:             syntax.ManaValueX,
		Power:                  syntax.Power,
		MatchPower:             syntax.MatchPower,
		Toughness:              syntax.Toughness,
		MatchToughness:         syntax.MatchToughness,
		Colorless:              syntax.Colorless,
		Multicolored:           syntax.Multicolored,
		BasicLandType:          syntax.BasicLandType,
		MatchCounter:           syntax.CounterRequired,
		RequiredCounter:        syntax.CounterKind,
		PlayerOrPlaneswalker:   syntax.PlayerOrPlaneswalker,
		SubtypeFromEntryChoice: syntax.SubtypeFromEntryChoice,
		SubtypeFromChosenType:  syntax.SubtypeFromChosenType,
		ConjunctiveTypes:       syntax.ConjunctiveTypes,
		RequiredName:           syntax.RequiredName,
	}
	// A required card-type union is always kept. A single required card type is
	// kept for a spell selection ("counter target instant or sorcery spell") and
	// for a plain card selection (SelectionCard), where the only single types
	// that reach here are the spell types instant and sorcery, which have no
	// dedicated SelectionKind. Keeping that single type lets a library search for
	// "an instant card" or "a sorcery card" carry the type into its lowered spec
	// instead of silently dropping it. Typed card kinds (creature, artifact) keep
	// their single type in Kind, so this guard leaves their RequiredTypesAny
	// empty as before.
	if len(syntax.RequiredTypesAny) > 1 ||
		(len(syntax.RequiredTypesAny) == 1 &&
			(syntax.Kind == parser.SelectionSpell || syntax.Kind == parser.SelectionCard)) {
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
		if value, ok := runtimeSupertypeFromParser(supertype); ok {
			appendSelectorSupertype(&selector, value)
		}
	}
	for _, supertype := range syntax.ExcludedSupertypes {
		if value, ok := runtimeSupertypeFromParser(supertype); ok {
			appendSelectorExcludedSupertype(&selector, value)
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
	appendSelectorExcludedSubtypes(&selector, syntax.ExcludedSubtypes...)
	for i := range syntax.Alternatives {
		selector.Alternatives = append(selector.Alternatives, compileTypedSelection(syntax.Alternatives[i]))
	}
	for _, cardType := range syntax.SourceTypes {
		if value, ok := runtimeCardTypeFromParser(cardType); ok {
			appendSelectorSourceType(&selector, value)
		}
	}
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
	case parser.EffectStaticSubjectOtherAttackingCreatures:
		return StaticSubjectOtherAttackingCreatures
	case parser.EffectStaticSubjectBlockingCreatures:
		return StaticSubjectBlockingCreatures
	case parser.EffectStaticSubjectControlledPermanents:
		return StaticSubjectControlledPermanents
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
	case parser.EffectStaticSubjectOpponentControlledPermanents:
		return StaticSubjectOpponentControlledPermanents
	case parser.EffectStaticSubjectControlledCreatureSubtype:
		return StaticSubjectControlledCreatureSubtype
	case parser.EffectStaticSubjectOtherControlledCreatureSubtype:
		return StaticSubjectOtherControlledCreatureSubtype
	case parser.EffectStaticSubjectAllCreatureSubtype:
		return StaticSubjectAllCreatureSubtype
	case parser.EffectStaticSubjectOtherCreatureSubtype:
		return StaticSubjectOtherCreatureSubtype
	case parser.EffectStaticSubjectControlledAttackingCreatures:
		return StaticSubjectControlledAttackingCreatures
	case parser.EffectStaticSubjectControlledCreatureTokens:
		return StaticSubjectControlledCreatureTokens
	case parser.EffectStaticSubjectBattlefieldCreatureTokens:
		return StaticSubjectBattlefieldCreatureTokens
	case parser.EffectStaticSubjectControlledLegendaryCreatures:
		return StaticSubjectControlledLegendaryCreatures
	case parser.EffectStaticSubjectControlledUntappedCreatures:
		return StaticSubjectControlledUntappedCreatures
	case parser.EffectStaticSubjectOtherControlledTappedCreatures:
		return StaticSubjectOtherControlledTappedCreatures
	case parser.EffectStaticSubjectControlledArtifactCreatures:
		return StaticSubjectControlledArtifactCreatures
	case parser.EffectStaticSubjectOtherControlledArtifactCreatures:
		return StaticSubjectOtherControlledArtifactCreatures
	case parser.EffectStaticSubjectControlledNontokenCreatures:
		return StaticSubjectControlledNontokenCreatures
	case parser.EffectStaticSubjectOtherControlledNontokenCreatures:
		return StaticSubjectOtherControlledNontokenCreatures
	case parser.EffectStaticSubjectAllLands:
		return StaticSubjectAllLands
	case parser.EffectStaticSubjectControlledCreaturesChosenType:
		return StaticSubjectControlledCreaturesChosenType
	case parser.EffectStaticSubjectOtherControlledCreaturesChosenType:
		return StaticSubjectOtherControlledCreaturesChosenType
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
	case parser.SelectionTriggeredAbilityOrSpell:
		return SelectorTriggeredAbilityOrSpell
	case parser.SelectionPlaneswalker:
		return SelectorPlaneswalker
	case parser.SelectionBattle:
		return SelectorBattle
	case parser.SelectionCommander:
		return SelectorCommander
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
	case parser.EffectCantBeBlocked:
		return EffectCantBeBlocked
	case parser.EffectCast:
		return EffectCast
	case parser.EffectCounter:
		return EffectCounter
	case parser.EffectCopyStackObject:
		return EffectCopyStackObject
	case parser.EffectCreate:
		return EffectCreate
	case parser.EffectDealDamage:
		return EffectDealDamage
	case parser.EffectDestroy:
		return EffectDestroy
	case parser.EffectDig:
		return EffectDig
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
	case parser.EffectLifeTotalCantChange:
		return EffectLifeTotalCantChange
	case parser.EffectProtectionFromEverything:
		return EffectProtectionFromEverything
	case parser.EffectInvestigate:
		return EffectInvestigate
	case parser.EffectImpulseExile:
		return EffectImpulseExile
	case parser.EffectAdditionalLandPlays:
		return EffectAdditionalLandPlays
	case parser.EffectExplore:
		return EffectExplore
	case parser.EffectLose:
		return EffectLose
	case parser.EffectLoseGame:
		return EffectLoseGame
	case parser.EffectWinGame:
		return EffectWinGame
	case parser.EffectEnterAsCopy:
		return EffectEnterAsCopy
	case parser.EffectBecomeCopy:
		return EffectBecomeCopy
	case parser.EffectAmass:
		return EffectAmass
	case parser.EffectRenown:
		return EffectRenown
	case parser.EffectDevour:
		return EffectDevour
	case parser.EffectTribute:
		return EffectTribute
	case parser.EffectChooseCreatureType:
		return EffectChooseCreatureType
	case parser.EffectMassReanimationExchange:
		return EffectMassReanimationExchange
	case parser.EffectPunisherLoseLife:
		return EffectPunisherLoseLife
	case parser.EffectMustAttack:
		return EffectMustAttack
	case parser.EffectRepeatProcess:
		return EffectRepeatProcess
	case parser.EffectChooseNewTargets:
		return EffectChooseNewTargets
	case parser.EffectCastAsThoughFlash:
		return EffectCastAsThoughFlash
	case parser.EffectCantCastSpells:
		return EffectCantCastSpells
	case parser.EffectPreventDamage:
		return EffectPreventDamage
	case parser.EffectSpellsCantBeCountered:
		return EffectSpellsCantBeCountered
	case parser.EffectManifest:
		return EffectManifest
	case parser.EffectManifestDread:
		return EffectManifestDread
	case parser.EffectMill:
		return EffectMill
	case parser.EffectMoveCounters:
		return EffectMoveCounters
	case parser.EffectManaSpendRider:
		return EffectManaSpendRider
	case parser.EffectModifyPT:
		return EffectModifyPT
	case parser.EffectPut:
		return EffectPut
	case parser.EffectPhaseOut:
		return EffectPhaseOut
	case parser.EffectProliferate:
		return EffectProliferate
	case parser.EffectRegenerate:
		return EffectRegenerate
	case parser.EffectReorderLibraryTop:
		return EffectReorderLibraryTop
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
	case parser.EffectDurationUntilEndOfYourNextTurn:
		return DurationUntilEndOfYourNextTurn
	case parser.EffectDurationThisTurn:
		return DurationThisTurn
	case parser.EffectDurationThisCombat:
		return DurationThisCombat
	case parser.EffectDurationWhileSourceOnBattlefield:
		return DurationForAsLongAsSourceOnBattlefield
	case parser.EffectDurationWhileYouControlSource:
		return DurationForAsLongAsYouControlSource
	case parser.EffectDurationWhileControlledCreatureEnchanted:
		return DurationForAsLongAsControlledCreatureEnchanted
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
	case parser.DelayedTimingNextMain:
		return game.DelayedAtBeginningOfNextMainPhase
	default:
		return 0
	}
}

func compileTypedAmount(amount parser.EffectAmountSyntax) CompiledAmount {
	compiled := CompiledAmount{
		Value:         amount.Value,
		Known:         amount.Known,
		RangeKnown:    amount.RangeKnown,
		Minimum:       amount.Minimum,
		Maximum:       amount.Maximum,
		VariableX:     amount.VariableX,
		DynamicKind:   compileDynamicAmountKind(amount.DynamicKind),
		DynamicForm:   compileDynamicAmountForm(amount.DynamicForm),
		Multiplier:    amount.Multiplier,
		ReferenceSpan: amount.ReferenceSpan,
		CounterKind:   amount.CounterKind,
		Text:          amount.Text,
		Colors:        compileAmountColors(amount.Colors),
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
	case parser.EffectDynamicAmountSourceToughness:
		return DynamicAmountSourceToughness
	case parser.EffectDynamicAmountSourceManaValue:
		return DynamicAmountSourceManaValue
	case parser.EffectDynamicAmountSourceCounterCount:
		return DynamicAmountSourceCounterCount
	case parser.EffectDynamicAmountBasicLandTypes:
		return DynamicAmountBasicLandTypes
	case parser.EffectDynamicAmountEventCardCount:
		return DynamicAmountEventCardCount
	case parser.EffectDynamicAmountLifeLostThisWay:
		return DynamicAmountLifeLostThisWay
	case parser.EffectDynamicAmountGreatestPower:
		return DynamicAmountGreatestPower
	case parser.EffectDynamicAmountGreatestToughness:
		return DynamicAmountGreatestToughness
	case parser.EffectDynamicAmountGreatestManaValue:
		return DynamicAmountGreatestManaValue
	case parser.EffectDynamicAmountTotalPower:
		return DynamicAmountTotalPower
	case parser.EffectDynamicAmountTotalToughness:
		return DynamicAmountTotalToughness
	case parser.EffectDynamicAmountColorCount:
		return DynamicAmountColorCount
	case parser.EffectDynamicAmountDevotion:
		return DynamicAmountDevotion
	case parser.EffectDynamicAmountGreatestDiscardedThisWay:
		return DynamicAmountGreatestDiscardedThisWay
	case parser.EffectDynamicAmountSpellsCastThisTurn:
		return DynamicAmountSpellsCastThisTurn
	case parser.EffectDynamicAmountTriggeringLifeChange:
		return DynamicAmountTriggeringLifeChange
	case parser.EffectDynamicAmountSacrificedPower:
		return DynamicAmountSacrificedPower
	case parser.EffectDynamicAmountSacrificedToughness:
		return DynamicAmountSacrificedToughness
	case parser.EffectDynamicAmountSacrificedManaValue:
		return DynamicAmountSacrificedManaValue
	case parser.EffectDynamicAmountSharedCreatureTypeCount:
		return DynamicAmountSharedCreatureTypeCount
	case parser.EffectDynamicAmountTriggeringCombatDamage:
		return DynamicAmountTriggeringCombatDamage
	case parser.EffectDynamicAmountDestroyedThisWay:
		return DynamicAmountDestroyedThisWay
	default:
		return DynamicAmountNone
	}
}

// compileAmountColors maps the parser's recognized devotion colors to runtime
// colors. Unrecognized colors are dropped; the parser only emits the five
// recognized colors, so a complete devotion amount keeps all of its colors.
func compileAmountColors(colors []parser.Color) []color.Color {
	if len(colors) == 0 {
		return nil
	}
	mapped := make([]color.Color, 0, len(colors))
	for _, parserColor := range colors {
		runtimeColor, ok := runtimeColorFromParser(parserColor)
		if !ok {
			continue
		}
		mapped = append(mapped, runtimeColor)
	}
	return mapped
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
		Value:     amount.Value,
		Known:     amount.Known,
		Negative:  amount.Negative,
		VariableX: amount.VariableX,
	}
}
