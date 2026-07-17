package compiler

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
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
	if payment.PerCreatureSelection != nil {
		compiled.PerCreatureSelector = compileTypedSelection(*payment.PerCreatureSelection)
	}
	return compiled
}

func applyEffectPaymentsToConditions(effects []CompiledEffect, conditions []CompiledCondition) {
	for i := range effects {
		effect := &effects[i]
		if len(effect.Payment.ManaCost) == 0 && effect.Payment.AdditionalCost == nil {
			continue
		}
		var predicate ConditionPredicate
		switch effect.Payment.Payer {
		case parser.EffectPaymentPayerTargetController, parser.EffectPaymentPayerTargetPlayer:
			predicate = ConditionPredicateTargetControllerDoesNotPay
		case parser.EffectPaymentPayerEventPlayer:
			predicate = ConditionPredicateEventPlayerDoesNotPay
		case parser.EffectPaymentPayerController:
			predicate = ConditionPredicateControllerDoesNotPay
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
				Min:       syntax.Cardinality.Min,
				Max:       syntax.Cardinality.Max,
				MaxEventX: syntax.Cardinality.MaxEventX,
			},
			Selector:          compileTypedSelection(syntax.Selection),
			Exact:             syntax.Exact,
			Order:             syntax.Order,
			KickerScaledCount: syntax.KickerScaledCount,
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

// compileRecipientControlsSelector compiles the optional "who controls
// <selection>" per-member qualifier of a group recipient ("Each player who
// controls an artifact or enchantment ...") into a CompiledSelector, or nil when
// the recipient carries no such qualifier. It reuses compileTypedSelection so the
// controlled-permanent characteristics compile through the same path as any other
// noun phrase.
func compileRecipientControlsSelector(syntax *parser.SelectionSyntax) *CompiledSelector {
	if syntax == nil {
		return nil
	}
	selector := compileTypedSelection(*syntax)
	return &selector
}

func compileTypedSelection(syntax parser.SelectionSyntax) CompiledSelector {
	selector := CompiledSelector{
		Kind:                            compileSelectionKind(syntax.Kind),
		Controller:                      compileSelectionController(syntax.Controller),
		All:                             syntax.All,
		Another:                         syntax.Another,
		Other:                           syntax.Other,
		Attacking:                       syntax.Attacking,
		Blocking:                        syntax.Blocking,
		Tapped:                          syntax.Tapped,
		Untapped:                        syntax.Untapped,
		NonToken:                        syntax.NonToken,
		TokenOnly:                       syntax.TokenOnly,
		Keyword:                         syntax.Keyword,
		ExcludedKeyword:                 syntax.ExcludedKeyword,
		Zone:                            syntax.Zone,
		ManaValue:                       syntax.ManaValue,
		MatchManaValue:                  syntax.MatchManaValue,
		MatchTotalManaValue:             syntax.MatchTotalManaValue,
		TotalManaValue:                  syntax.TotalManaValue,
		ManaValueX:                      syntax.ManaValueX,
		ManaValueDynamic:                compileDynamicAmountKind(syntax.ManaValueDynamic),
		Power:                           syntax.Power,
		MatchPower:                      syntax.MatchPower,
		Toughness:                       syntax.Toughness,
		MatchToughness:                  syntax.MatchToughness,
		Colorless:                       syntax.Colorless,
		Multicolored:                    syntax.Multicolored,
		Colored:                         syntax.Colored,
		BasicLandType:                   syntax.BasicLandType,
		Historic:                        syntax.Historic,
		MatchCounter:                    syntax.CounterRequired && !syntax.CounterAny,
		RequiredCounter:                 syntax.CounterKind,
		MatchAnyCounter:                 syntax.CounterAny,
		MatchNoCounters:                 syntax.CounterAbsent,
		MatchExcludedCounter:            syntax.CounterKindAbsent,
		ExcludedCounter:                 syntax.CounterKind,
		PlayerOrPlaneswalker:            syntax.PlayerOrPlaneswalker,
		SubtypeFromEntryChoice:          syntax.SubtypeFromEntryChoice,
		SubtypeFromChosenType:           syntax.SubtypeFromChosenType,
		SubtypeFromChosenTypeExcluded:   syntax.SubtypeFromChosenTypeExcluded,
		ConjunctiveTypes:                syntax.ConjunctiveTypes,
		RequiredName:                    syntax.RequiredName,
		EnteredThisTurn:                 syntax.EnteredThisTurn,
		DealtDamageThisTurn:             syntax.DealtDamageThisTurn,
		Modified:                        syntax.Modified,
		Enchanted:                       syntax.Enchanted,
		Equipped:                        syntax.Equipped,
		PowerLessThanSource:             syntax.PowerLessThanSource,
		PowerGreaterThanSource:          syntax.PowerGreaterThanSource,
		ManaValueLessThanEventPermanent: syntax.ManaValueLessThanEventPermanent,
		NameUniqueAmongControlled:       syntax.NameUniqueAmongControlled,
		InclusiveOneOfEach:              syntax.InclusiveOneOfEach,
		SingleGraveyard:                 syntax.SingleGraveyard,
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
	selector.SpellTargetRestrictions = compileSpellTargetRestrictions(syntax.SpellTargetRestrictions)
	selector.SameNameGroup = compileSameNameGroup(syntax.SameNameGroup)
	if syntax.ManaValueDynamicCount != nil {
		amount := compileTypedAmount(*syntax.ManaValueDynamicCount)
		selector.ManaValueDynamicCount = &amount
	}
	if syntax.ManaValueSacrificedCostAddend != nil {
		addend := *syntax.ManaValueSacrificedCostAddend
		selector.ManaValueSacrificedCost = &addend
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
	case parser.EffectStaticSubjectOtherControlledPermanents:
		return StaticSubjectOtherControlledPermanents
	case parser.EffectStaticSubjectControlledCreatures:
		return StaticSubjectControlledCreatures
	case parser.EffectStaticSubjectOtherControlledCreatures:
		return StaticSubjectOtherControlledCreatures
	case parser.EffectStaticSubjectControlledWalls:
		return StaticSubjectControlledWalls
	case parser.EffectStaticSubjectControlledArtifacts:
		return StaticSubjectControlledArtifacts
	case parser.EffectStaticSubjectControlledSagas:
		return StaticSubjectControlledSagas
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
	case parser.EffectStaticSubjectControlledAttackingCreatureSubtype:
		return StaticSubjectControlledAttackingCreatureSubtype
	case parser.EffectStaticSubjectControlledAttackingCreatureTokens:
		return StaticSubjectControlledAttackingCreatureTokens
	case parser.EffectStaticSubjectControlledCreatureTokens:
		return StaticSubjectControlledCreatureTokens
	case parser.EffectStaticSubjectBattlefieldCreatureTokens:
		return StaticSubjectBattlefieldCreatureTokens
	case parser.EffectStaticSubjectControlledLegendaryCreatures:
		return StaticSubjectControlledLegendaryCreatures
	case parser.EffectStaticSubjectControlledNonlegendaryCreatures:
		return StaticSubjectControlledNonlegendaryCreatures
	case parser.EffectStaticSubjectControlledCommanderCreatures:
		return StaticSubjectControlledCommanderCreatures
	case parser.EffectStaticSubjectOwnedCommanderCreatures:
		return StaticSubjectOwnedCommanderCreatures
	case parser.EffectStaticSubjectControlledCommanders:
		return StaticSubjectControlledCommanders
	case parser.EffectStaticSubjectControlledUntappedCreatures:
		return StaticSubjectControlledUntappedCreatures
	case parser.EffectStaticSubjectControlledModifiedCreatures:
		return StaticSubjectControlledModifiedCreatures
	case parser.EffectStaticSubjectOtherControlledTappedCreatures:
		return StaticSubjectOtherControlledTappedCreatures
	case parser.EffectStaticSubjectOtherControlledUntappedCreatures:
		return StaticSubjectOtherControlledUntappedCreatures
	case parser.EffectStaticSubjectControlledArtifactCreatures:
		return StaticSubjectControlledArtifactCreatures
	case parser.EffectStaticSubjectOtherControlledArtifactCreatures:
		return StaticSubjectOtherControlledArtifactCreatures
	case parser.EffectStaticSubjectControlledNontokenCreatures:
		return StaticSubjectControlledNontokenCreatures
	case parser.EffectStaticSubjectControlledNotOwnedCreatures:
		return StaticSubjectControlledNotOwnedCreatures
	case parser.EffectStaticSubjectControlledCreatureSubtypeTokens:
		return StaticSubjectControlledCreatureSubtypeTokens
	case parser.EffectStaticSubjectOtherControlledCreatureSubtypeTokens:
		return StaticSubjectOtherControlledCreatureSubtypeTokens
	case parser.EffectStaticSubjectOtherControlledNontokenCreatures:
		return StaticSubjectOtherControlledNontokenCreatures
	case parser.EffectStaticSubjectAllLands:
		return StaticSubjectAllLands
	case parser.EffectStaticSubjectControlledLands:
		return StaticSubjectControlledLands
	case parser.EffectStaticSubjectControlledCreaturesChosenType:
		return StaticSubjectControlledCreaturesChosenType
	case parser.EffectStaticSubjectOtherControlledCreaturesChosenType:
		return StaticSubjectOtherControlledCreaturesChosenType
	case parser.EffectStaticSubjectAllCreaturesChosenType:
		return StaticSubjectAllCreaturesChosenType
	case parser.EffectStaticSubjectOpponentControlledCreaturesChosenType:
		return StaticSubjectOpponentControlledCreaturesChosenType
	case parser.EffectStaticSubjectControlledPermanentSubtype:
		return StaticSubjectControlledPermanentSubtype
	case parser.EffectStaticSubjectOtherControlledPermanentSubtype:
		return StaticSubjectOtherControlledPermanentSubtype
	case parser.EffectStaticSubjectNonbasicLands:
		return StaticSubjectNonbasicLands
	case parser.EffectStaticSubjectNonlandPermanents:
		return StaticSubjectNonlandPermanents
	case parser.EffectStaticSubjectSnowPermanents:
		return StaticSubjectSnowPermanents
	case parser.EffectStaticSubjectAllPermanentSubtype:
		return StaticSubjectAllPermanentSubtype
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

// compileSpellTargetRestrictions lowers parser spell-target restriction
// alternatives ("Counter target spell that targets <X>") into their compiled
// form, dropping any unrecognized parser card type so the lowering can fail
// closed on an empty result.
func compileSpellTargetRestrictions(restrictions []parser.SpellTargetRestriction) []CompiledSpellTargetRestriction {
	if len(restrictions) == 0 {
		return nil
	}
	compiled := make([]CompiledSpellTargetRestriction, 0, len(restrictions))
	for _, restriction := range restrictions {
		var entry CompiledSpellTargetRestriction
		entry.Controller = compileSelectionController(restriction.Controller)
		switch restriction.Kind {
		case parser.SpellTargetRestrictionPlayer:
			entry.IsPlayer = true
		case parser.SpellTargetRestrictionPermanent:
			if restriction.PermanentType != parser.CardTypeUnknown {
				value, ok := runtimeCardTypeFromParser(restriction.PermanentType)
				if !ok {
					return nil
				}
				entry.PermanentTypes = []types.Card{value}
			}
		default:
			return nil
		}
		compiled = append(compiled, entry)
	}
	return compiled
}

// compileSameNameGroup lowers a parser same-name destroy group ("and all other
// <type> with the same name as that <noun>") into its compiled form, dropping
// any unrecognized parser card type so the lowering can fail closed. It returns
// nil when the parser recorded no group.
func compileSameNameGroup(group *parser.SameNameGroupSyntax) *CompiledSameNameGroup {
	if group == nil {
		return nil
	}
	compiled := &CompiledSameNameGroup{}
	for _, cardType := range group.GroupTypes {
		value, ok := runtimeCardTypeFromParser(cardType)
		if !ok {
			return nil
		}
		compiled.GroupTypes = append(compiled.GroupTypes, value)
	}
	return compiled
}

func compileSelectionController(controller parser.SelectionController) ControllerKind {
	switch controller {
	case parser.SelectionControllerYou:
		return ControllerYou
	case parser.SelectionControllerOpponent:
		return ControllerOpponent
	case parser.SelectionControllerNotYou:
		return ControllerNotYou
	case parser.SelectionControllerThatPlayer:
		return ControllerThatPlayer
	case parser.SelectionControllerDefendingPlayer:
		return ControllerDefendingPlayer
	case parser.SelectionControllerTargetedPlayers:
		return ControllerTargetedPlayers
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
	case parser.EffectCanAttackAsThoughDefender:
		return EffectCanAttackAsThoughDefender
	case parser.EffectCantAttack:
		return EffectCantAttack
	case parser.EffectCantAttackOrBlock:
		return EffectCantAttackOrBlock
	case parser.EffectCantBeBlocked:
		return EffectCantBeBlocked
	case parser.EffectCantBeSacrificed:
		return EffectCantBeSacrificed
	case parser.EffectCantBlock:
		return EffectCantBlock
	case parser.EffectCast:
		return EffectCast
	case parser.EffectChooseExiledCard:
		return EffectChooseExiledCard
	case parser.EffectReturnExiledCardsWithCounter:
		return EffectReturnExiledCardsWithCounter
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
	case parser.EffectLookAtHand:
		return EffectLookAtHand
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
	case parser.EffectExchange:
		return EffectExchange
	case parser.EffectFight:
		return EffectFight
	case parser.EffectGain:
		return EffectGain
	case parser.EffectGainControl:
		return EffectGainControl
	case parser.EffectBecomeMonarch:
		return EffectBecomeMonarch
	case parser.EffectCantBecomeMonarch:
		return EffectCantBecomeMonarch
	case parser.EffectVentureIntoDungeon:
		return EffectVentureIntoDungeon
	case parser.EffectVentureIntoUndercity:
		return EffectVentureIntoUndercity
	case parser.EffectTakeInitiative:
		return EffectTakeInitiative
	case parser.EffectRingTempts:
		return EffectRingTempts
	case parser.EffectGrantKeyword:
		return EffectGrantKeyword
	case parser.EffectLifeTotalCantChange:
		return EffectLifeTotalCantChange
	case parser.EffectProtectionFromEverything:
		return EffectProtectionFromEverything
	case parser.EffectInvestigate:
		return EffectInvestigate
	case parser.EffectGainPlayerCounter:
		return EffectGainPlayerCounter
	case parser.EffectImpulseExile:
		return EffectImpulseExile
	case parser.EffectCreateEmblem:
		return EffectCreateEmblem
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
	case parser.EffectBecomeType:
		return EffectBecomeType
	case parser.EffectBecomeColor:
		return EffectBecomeColor
	case parser.EffectPolymorph:
		return EffectPolymorph
	case parser.EffectTurnFaceDown:
		return EffectTurnFaceDown
	case parser.EffectSetBasePT:
		return EffectSetBasePT
	case parser.EffectSwitchPT:
		return EffectSwitchPT
	case parser.EffectDelayedTrigger:
		return EffectDelayedTrigger
	case parser.EffectPayRepeatedlyAnimate:
		return EffectPayRepeatedlyAnimate
	case parser.EffectAnimateSelf:
		return EffectAnimateSelf
	case parser.EffectAnimateTarget:
		return EffectAnimateTarget
	case parser.EffectAmass:
		return EffectAmass
	case parser.EffectIncubate:
		return EffectIncubate
	case parser.EffectBolster:
		return EffectBolster
	case parser.EffectRenown:
		return EffectRenown
	case parser.EffectAdapt:
		return EffectAdapt
	case parser.EffectMonstrosity:
		return EffectMonstrosity
	case parser.EffectConnive:
		return EffectConnive
	case parser.EffectDevour:
		return EffectDevour
	case parser.EffectTribute:
		return EffectTribute
	case parser.EffectChooseCreatureType:
		return EffectChooseCreatureType
	case parser.EffectChoosePermanent:
		return EffectChoosePermanent
	case parser.EffectNoMaximumHandSize:
		return EffectNoMaximumHandSize
	case parser.EffectAdditionalCombatPhase:
		return EffectAdditionalCombatPhase
	case parser.EffectAdditionalUpkeepStep:
		return EffectAdditionalUpkeepStep
	case parser.EffectRollDie:
		return EffectRollDie
	case parser.EffectExileIfLeaveBattlefield:
		return EffectExileIfLeaveBattlefield
	case parser.EffectExileIfWouldDieThisTurn:
		return EffectExileIfWouldDieThisTurn
	case parser.EffectMassReanimationExchange:
		return EffectMassReanimationExchange
	case parser.EffectChooseFromEachGraveyard:
		return EffectChooseFromEachGraveyard
	case parser.EffectPunisherLoseLife:
		return EffectPunisherLoseLife
	case parser.EffectMustAttack:
		return EffectMustAttack
	case parser.EffectDirectedMustAttack:
		return EffectDirectedMustAttack
	case parser.EffectAttackTax:
		return EffectAttackTax
	case parser.EffectRepeatProcess:
		return EffectRepeatProcess
	case parser.EffectChooseNewTargets:
		return EffectChooseNewTargets
	case parser.EffectCastAsThoughFlash:
		return EffectCastAsThoughFlash
	case parser.EffectPlayFromLibraryTop:
		return EffectPlayFromLibraryTop
	case parser.EffectPlay:
		return EffectPlay
	case parser.EffectCantCastSpells:
		return EffectCantCastSpells
	case parser.EffectSpellCostModifier:
		return EffectSpellCostModifier
	case parser.EffectPreventDamage:
		return EffectPreventDamage
	case parser.EffectSpellsCantBeCountered:
		return EffectSpellsCantBeCountered
	case parser.EffectGrantSpellKeyword:
		return EffectGrantSpellKeyword
	case parser.EffectManifest:
		return EffectManifest
	case parser.EffectManifestDread:
		return EffectManifestDread
	case parser.EffectCloak:
		return EffectCloak
	case parser.EffectLookAtLibraryTop:
		return EffectLookAtLibraryTop
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
	case parser.EffectPopulate:
		return EffectPopulate
	case parser.EffectProliferate:
		return EffectProliferate
	case parser.EffectRemoveCounter:
		return EffectRemoveCounter
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
	case parser.EffectGoad:
		return EffectGoad
	case parser.EffectTapOrUntap:
		return EffectTapOrUntap
	case parser.EffectUntap:
		return EffectUntap
	case parser.EffectRemoveFromCombat:
		return EffectRemoveFromCombat
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
	case parser.EffectDurationUntilYourNextEndStep:
		return DurationUntilYourNextEndStep
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
	case parser.EffectDurationWhileThatPlayerIsMonarch:
		return DurationForAsLongAsThatPlayerIsMonarch
	case parser.EffectDurationWhileExiled:
		return DurationForAsLongAsExiled
	default:
		return DurationNone
	}
}

func compileDelayedTiming(timing parser.DelayedTimingKind) game.DelayedTriggerTiming {
	switch timing {
	case parser.DelayedTimingNextEndStep:
		return game.DelayedAtBeginningOfNextEndStep
	case parser.DelayedTimingYourNextEndStep:
		return game.DelayedAtBeginningOfYourNextEndStep
	case parser.DelayedTimingNextUpkeep:
		return game.DelayedAtBeginningOfNextUpkeep
	case parser.DelayedTimingNextMain:
		return game.DelayedAtBeginningOfNextMainPhase
	case parser.DelayedTimingEndOfCombat:
		return game.DelayedAtEndOfCombat
	default:
		return 0
	}
}

func compileTypedAmount(amount parser.EffectAmountSyntax) CompiledAmount {
	compiled := CompiledAmount{
		Value:           amount.Value,
		Known:           amount.Known,
		RangeKnown:      amount.RangeKnown,
		Minimum:         amount.Minimum,
		Maximum:         amount.Maximum,
		VariableX:       amount.VariableX,
		AnyNumber:       amount.AnyNumber,
		DynamicKind:     compileDynamicAmountKind(amount.DynamicKind),
		DynamicForm:     compileDynamicAmountForm(amount.DynamicForm),
		Multiplier:      amount.Multiplier,
		ReferenceSpan:   amount.ReferenceSpan,
		ReferenceNodeID: amount.ReferenceNodeID,
		Addend:          amount.Addend,
		CounterKind:     amount.CounterKind,
		Text:            amount.Text,
		Colors:          compileAmountColors(amount.Colors),
		RoundUp:         amount.RoundUp,
	}
	if amount.Selection != nil {
		selection := compileTypedSelection(*amount.Selection)
		compiled.selector = &selection
	}
	if len(amount.Operands) != 0 {
		operands := make([]CompiledAmount, len(amount.Operands))
		for i := range amount.Operands {
			operands[i] = compileTypedAmount(amount.Operands[i])
		}
		compiled.Operands = operands
	}
	return compiled
}

func compileDynamicAmountKind(kind parser.EffectDynamicAmountKind) DynamicAmountKind {
	switch kind {
	case parser.EffectDynamicAmountCount:
		return DynamicAmountCount
	case parser.EffectDynamicAmountControllerLife:
		return DynamicAmountControllerLife
	case parser.EffectDynamicAmountControllerSpeed:
		return DynamicAmountControllerSpeed
	case parser.EffectDynamicAmountOpponentCount:
		return DynamicAmountOpponentCount
	case parser.EffectDynamicAmountOpponentControllingCount:
		return DynamicAmountOpponentControllingCount
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
	case parser.EffectDynamicAmountDamageDealtThisWay:
		return DynamicAmountDamageDealtThisWay
	case parser.EffectDynamicAmountExcessDamageDealtThisWay:
		return DynamicAmountExcessDamageDealtThisWay
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
	case parser.EffectDynamicAmountTotalManaValue:
		return DynamicAmountTotalManaValue
	case parser.EffectDynamicAmountReferencedCardsTotalManaValue:
		return DynamicAmountReferencedCardsTotalManaValue
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
	case parser.EffectDynamicAmountTriggeringEventTotalCombatDamage:
		return DynamicAmountTriggeringEventTotalCombatDamage
	case parser.EffectDynamicAmountDestroyedThisWay:
		return DynamicAmountDestroyedThisWay
	case parser.EffectDynamicAmountLifeLostThisTurn:
		return DynamicAmountLifeLostThisTurn
	case parser.EffectDynamicAmountLifeGainedThisTurn:
		return DynamicAmountLifeGainedThisTurn
	case parser.EffectDynamicAmountReferencedPlayerLifeLostThisTurn:
		return DynamicAmountReferencedPlayerLifeLostThisTurn
	case parser.EffectDynamicAmountReferencedPlayerLifeGainedThisTurn:
		return DynamicAmountReferencedPlayerLifeGainedThisTurn
	case parser.EffectDynamicAmountCreaturesBlockingSource:
		return DynamicAmountCreaturesBlockingSource
	case parser.EffectDynamicAmountPartySize:
		return DynamicAmountPartySize
	case parser.EffectDynamicAmountDamagePreventedThisWay:
		return DynamicAmountDamagePreventedThisWay
	case parser.EffectDynamicAmountTriggeringPlayerHandSize:
		return DynamicAmountTriggeringPlayerHandSize
	case parser.EffectDynamicAmountMaxOf:
		return DynamicAmountMaxOf
	case parser.EffectDynamicAmountTriggeringCounterCount:
		return DynamicAmountTriggeringCounterCount
	case parser.EffectDynamicAmountColorsOfManaSpent:
		return DynamicAmountColorsOfManaSpent
	case parser.EffectDynamicAmountDieRollResult:
		return DynamicAmountDieRollResult
	case parser.EffectDynamicAmountTimesKicked:
		return DynamicAmountTimesKicked
	case parser.EffectDynamicAmountOpponentsAttackedThisCombat:
		return DynamicAmountOpponentsAttackedThisCombat
	case parser.EffectDynamicAmountTriggeringEventAmount:
		return DynamicAmountTriggeringEventAmount
	case parser.EffectDynamicAmountCardsDrawnThisTurn:
		return DynamicAmountCardsDrawnThisTurn
	case parser.EffectDynamicAmountCardsNamedSelfInGraveyards:
		return DynamicAmountCardsNamedSelfInGraveyards
	case parser.EffectDynamicAmountCardsNamedSelfInControllerGraveyard:
		return DynamicAmountCardsNamedSelfInControllerGraveyard
	case parser.EffectDynamicAmountHalfPlayerLibrary:
		return DynamicAmountHalfPlayerLibrary
	case parser.EffectDynamicAmountHalfPlayerLife:
		return DynamicAmountHalfPlayerLife
	case parser.EffectDynamicAmountCommanderCastCount:
		return DynamicAmountCommanderCastCount
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
	case parser.EffectDynamicAmountFormHalfLibrary:
		return DynamicAmountFormHalfLibrary
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
