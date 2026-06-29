// Package compiler lowers parsed Oracle syntax into semantic intermediate
// representation for card generation.
package compiler

import (
	"math"
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Compile lowers a parsed syntax document into conservative semantic IR.
func Compile(document parser.Document, context Context) (Compilation, []shared.Diagnostic) {
	compilation := Compilation{Syntax: document}
	var diagnostics []shared.Diagnostic
	for i := range document.Abilities {
		compiled, abilityDiagnostics := compileAbility(&document.Abilities[i], context)
		compilation.Abilities = append(compilation.Abilities, compiled)
		diagnostics = append(diagnostics, abilityDiagnostics...)
	}
	return compilation, diagnostics
}

func compileAbility(
	ability *parser.Ability,
	context Context,
) (CompiledAbility, []shared.Diagnostic) {
	var diagnostics []shared.Diagnostic
	kind := compileAbilityKind(ability.Kind)
	compiled := CompiledAbility{
		Kind: kind,
		Span: ability.Span,
		Text: ability.Text,
	}
	if ability.AbilityWord != nil {
		compiled.AbilityWord = ability.AbilityWord.Label
	}
	compiled.Chapters = append([]int(nil), ability.Chapters...)
	compiled.ChapterSpan = ability.ChapterSpan
	if ability.CostSyntax != nil {
		cost := compileCost(*ability.CostSyntax)
		compiled.Cost = &cost
	}
	if ability.SourceAbilityCostReduction != nil {
		compiled.SourceAbilityCostReduction = &CompiledSourceAbilityCostReduction{
			Span:           ability.SourceAbilityCostReduction.Span,
			Amount:         ability.SourceAbilityCostReduction.Amount,
			CountSelection: compileTypedSelection(ability.SourceAbilityCostReduction.CountSelection),
		}
	}
	if ability.AlternativeCost != nil {
		compiled.AlternativeCost = &CompiledAlternativeCost{
			Kind:                  compileAlternativeCostKind(ability.AlternativeCost.Kind),
			Condition:             compileAlternativeCostCondition(ability.AlternativeCost.Condition),
			WithoutPayingManaCost: ability.AlternativeCost.WithoutPayingManaCost,
			ManaCost:              slices.Clone(ability.AlternativeCost.ManaCost),
			ReplaceTargetWithEach: ability.AlternativeCost.ReplaceTargetWithEach,
		}
	}
	if kind == AbilityTriggered {
		trigger := compileTrigger(ability, context)
		compiled.Trigger = &trigger
	}
	if ability.ExactSequence != nil {
		compiled.ExactSequence = compileExactSequenceKind(ability.ExactSequence.Kind)
		compiled.ExactSequenceBottom = ability.ExactSequence.Bottom
		if offset := ability.ExactSequence.DrawOffset; offset >= 0 && offset <= math.MaxUint8 {
			compiled.ExactSequenceDrawOffset = uint8(offset)
		}
		compiled.ExactSequenceLookAtTopTypes = compilerCardTypes(ability.ExactSequence.LookAtTopCardTypes)
		compiled.ExactSequenceLookAtTopEntersTapped = ability.ExactSequence.LookAtTopEntersTapped
		compiled.ExactSequenceLookAtTopElseHand = ability.ExactSequence.LookAtTopBattlefield == parser.LookAtTopBattlefieldElseHand
		compiled.ExactSequenceLookAtTopElseBottom = ability.ExactSequence.LookAtTopBattlefield == parser.LookAtTopBattlefieldElseBottom
		if n := ability.ExactSequence.DrawCount; n >= 0 && n <= math.MaxUint8 {
			compiled.ExactSequenceDrawCount = uint8(n)
		}
		if n := ability.ExactSequence.DiscardCount; n >= 0 && n <= math.MaxUint8 {
			compiled.ExactSequenceDiscardCount = uint8(n)
		}
	}
	compiled.ClassLevelGain = ability.ClassLevelGain
	compiled.LevelUpRecognized = ability.LevelUpRecognized
	compiled.LevelUpCost = slices.Clone(ability.LevelUpCost)
	if ability.LevelBand != nil {
		compiled.LevelBand = &CompiledLevelBand{
			Low:               ability.LevelBand.Low,
			High:              ability.LevelBand.High,
			Power:             ability.LevelBand.Power,
			Toughness:         ability.LevelBand.Toughness,
			HasPowerToughness: ability.LevelBand.HasPowerToughness,
		}
	}
	compiled.Companion = ability.Companion != nil
	compiled.PartnerWith = ability.PartnerWith != nil
	compiled.ChooseABackground = ability.ChooseABackground != nil
	compiled.Partner = ability.Partner != nil
	if ability.Modal != nil {
		for i := range ability.Modal.Options {
			compiledMode, modeDiagnostics := compileMode(&ability.Modal.Options[i], context)
			compiled.Content.Modes = append(compiled.Content.Modes, compiledMode)
			diagnostics = append(diagnostics, modeDiagnostics...)
		}
		if len(compiled.Content.Modes) > 0 {
			compiled.Content.Modes[0].Modal = &CompiledModalSemantics{
				MinModes:     ability.Modal.MinModes,
				MaxModes:     ability.Modal.MaxModes,
				Kind:         compileModalChoiceKind(ability.Modal.ChoiceKind),
				Bonus:        compileModeChoiceBonus(ability.Modal.ChoiceBonus),
				Spree:        ability.Modal.Spree,
				Escalate:     ability.Modal.Escalate,
				EscalateCost: slices.Clone(ability.Modal.EscalateCost),
			}
		}
	}

	timing, timingSpan := compileActivationTiming(kind, ability.ActivationRestrictions)
	if timing != ActivationTimingNone {
		compiled.ActivationTiming = timing
		compiled.ActivationTimingSpan = timingSpan
	}
	if kind == AbilityTriggered && ability.Optional {
		compiled.Optional = true
		compiled.OptionalSpan = ability.OptionalSpan
	}
	if kind != AbilitySpellAlternativeCost && ability.ExactSequence == nil {
		if kind == AbilityStatic && staticRuleSentencesOnly(ability.Sentences) {
			if staticRuleSentencesHaveGuard(ability.Sentences) {
				compiled.Content.Conditions = compileConditions(
					ability.ConditionSegments,
					ability.ConditionClauses,
					ability.EventHistoryConditions,
				)
			}
			compiled.Content.Effects = compileEffects(ability.Sentences)
			applyEffectPaymentsToConditions(compiled.Content.Effects, compiled.Content.Conditions)
			compiled.Content.References = compileStaticRuleReferences(ability.Sentences)
		} else {
			compiled.Content.Keywords = compileKeywords(ability.SemanticKeywords)
			compiled.Content.Targets = compileTypedTargets(ability.Sentences)
			compiled.Content.Conditions = compileConditions(
				ability.ConditionSegments,
				ability.ConditionClauses,
				ability.EventHistoryConditions,
			)
			compiled.Content.Effects = compileEffects(ability.Sentences)
			applyEffectPaymentsToConditions(compiled.Content.Effects, compiled.Content.Conditions)
			compiled.Content.References = compileTypedReferences(ability.SemanticReferences)
			compiled.Content.References = bindReferences(
				compiled.Content.References,
				compiled.Content.Targets,
				compiled.Content.Effects,
				compiled.Trigger,
			)
		}
		compiled.Content.Effects = appendDiceTableEffects(compiled.Content.Effects, ability.DiceTable)
		compiled.Content.Effects = appendCoinFlipEffects(compiled.Content.Effects, ability.CoinFlip)
		compiled.Content.Effects = appendVoteEffects(compiled.Content.Effects, ability.Vote)
	}
	compiled.Content.References = bindActivationCostReferences(compiled.Kind, compiled.Cost, compiled.Content.References)
	bindConditionReferences(compiled.Content.Conditions, compiled.Content.References, compiled.Trigger)
	applyEffectReferenceBindings(compiled.Content.Effects, compiled.Content.References)
	recognizeActivationZone(&compiled)
	if compiled.Trigger != nil && compiled.Trigger.Condition != nil {
		for i := range compiled.Content.Conditions {
			if compiled.Content.Conditions[i].NodeID == compiled.Trigger.Condition.NodeID {
				condition := compiled.Content.Conditions[i]
				compiled.Trigger.Condition = &condition
				compiled.Trigger.Pattern.InterveningCondition = &condition
				break
			}

		}
	}
	recognizeStaticDeclarations(&compiled, ability)

	for _, mode := range compiled.Content.Modes {
		if len(mode.Content.Effects) == 0 && len(mode.Content.Keywords) == 0 {
			diagnostics = append(diagnostics, unsupportedDiagnostic(mode.Span, mode.Text))
		}
	}
	if kind != AbilityReminder && kind != AbilitySpellAdditionalCost && kind != AbilitySpellAlternativeCost && kind != AbilityLevelBand && ability.Modal == nil &&
		compiled.ExactSequence == ExactSequenceUnknown &&
		compiled.ClassLevelGain == 0 &&
		!compiled.LevelUpRecognized &&
		!compiled.Companion &&
		!compiled.PartnerWith &&
		!compiled.ChooseABackground &&
		!compiled.Partner &&
		len(compiled.Content.Effects) == 0 && len(compiled.Content.Keywords) == 0 &&
		!legacyEffectsPresent(ability.Sentences) &&
		(compiled.Static == nil || len(compiled.Static.Declarations) == 0) {
		diagnostics = append(diagnostics, unsupportedDiagnostic(ability.Span, ability.Text))
	}

	// Content.Span is the parser-emitted span of the ability's resolving content
	// after shell/timing extraction. It is non-zero even for unrecognized
	// content, excludes the cost span for activated/loyalty abilities, and begins
	// after "you may" for an optional triggered ability.
	compiled.Content.Span = ability.ContentSpan
	if compiled.Cost != nil {
		for _, component := range compiled.Cost.Components {
			if component.Kind == CostUnknown {
				diagnostics = append(diagnostics, shared.Diagnostic{
					Severity: shared.SeverityWarning,
					Summary:  "unsupported cost",
					Detail:   "the compiler preserved this cost component but did not assign executable semantics",
					Span:     component.Span,
				})
			}
		}
	}
	return compiled, diagnostics
}

func compileModalChoiceKind(kind parser.ModalChoiceKind) CompiledModalChoiceKind {
	switch kind {
	case parser.ModalChoiceKindOneOrMore:
		return CompiledModalChoiceOneOrMore
	case parser.ModalChoiceKindOneAtRandom:
		return CompiledModalChoiceOneAtRandom
	default:
		return CompiledModalChoiceUnknown
	}
}

func compileModeChoiceBonus(bonus parser.ModalChoiceBonusSyntax) CompiledModeChoiceBonus {
	compiled := CompiledModeChoiceBonus{AdditionalMaxModes: bonus.AdditionalMaxModes}
	if bonus.Condition == parser.ModalChoiceBonusConditionControlsCommander {
		compiled.Condition = ModeChoiceBonusConditionControlsCommander
	}
	return compiled
}

func legacyEffectsPresent(sentences []parser.Sentence) bool {
	return slices.ContainsFunc(sentences, func(sentence parser.Sentence) bool {
		return sentence.LegacyEffects
	})
}

func applyEffectReferenceBindings(effects []CompiledEffect, references []CompiledReference) {
	for effectIndex := range effects {
		applyReferenceBindings(effects[effectIndex].References, references)
		applyReferenceBindings(effects[effectIndex].SubjectReferences, references)
	}
}

func applyReferenceBindings(ownedReferences, references []CompiledReference) {
	for referenceIndex, owned := range ownedReferences {
		for _, reference := range references {
			if reference.NodeID == owned.NodeID {
				ownedReferences[referenceIndex] = reference
				break
			}
		}
	}
}

func compileAbilityKind(kind parser.AbilityKind) AbilityKind {
	switch kind {
	case parser.AbilitySpell:
		return AbilitySpell
	case parser.AbilityActivated:
		return AbilityActivated
	case parser.AbilityLoyalty:
		return AbilityLoyalty
	case parser.AbilityChapter:
		return AbilityChapter
	case parser.AbilityTriggered:
		return AbilityTriggered
	case parser.AbilityReplacement:
		return AbilityReplacement
	case parser.AbilityStatic:
		return AbilityStatic
	case parser.AbilityReminder:
		return AbilityReminder
	case parser.AbilitySpellAdditionalCost:
		return AbilitySpellAdditionalCost
	case parser.AbilitySpellAlternativeCost:
		return AbilitySpellAlternativeCost
	case parser.AbilityLevelBand:
		return AbilityLevelBand
	default:
		return AbilityUnknown
	}
}

func compileAlternativeCostKind(kind parser.SpellAlternativeCostKind) AlternativeCostKind {
	switch kind {
	case parser.SpellAlternativeCostCommander:
		return AlternativeCostCommander
	case parser.SpellAlternativeCostOverload:
		return AlternativeCostOverload
	case parser.SpellAlternativeCostPitch:
		return AlternativeCostPitch
	case parser.SpellAlternativeCostFlashback:
		return AlternativeCostFlashback
	case parser.SpellAlternativeCostEscape:
		return AlternativeCostEscape
	case parser.SpellAlternativeCostDiscard:
		return AlternativeCostDiscard
	default:
		return AlternativeCostUnknown
	}
}

func compileAlternativeCostCondition(condition parser.SpellAlternativeCostCondition) AlternativeCostCondition {
	switch condition {
	case parser.SpellAlternativeCostConditionControlsCommander:
		return AlternativeCostConditionControlsCommander
	case parser.SpellAlternativeCostConditionNotYourTurn:
		return AlternativeCostConditionNotYourTurn
	default:
		return AlternativeCostConditionUnknown
	}
}

func compileActivationTiming(kind AbilityKind, restrictions []parser.ActivationRestriction) (ActivationTimingKind, shared.Span) {
	if kind != AbilityActivated || len(restrictions) == 0 {
		return ActivationTimingNone, shared.Span{}
	}
	span := shared.Span{
		Start: restrictions[0].Span.Start,
		End:   restrictions[len(restrictions)-1].Span.End,
	}
	compiled := make([]ActivationTimingKind, 0, len(restrictions))
	for i := range restrictions {
		compiled = append(compiled, compileActivationRestriction(&restrictions[i]))
	}
	if len(compiled) == 1 {
		return compiled[0], span
	}
	if len(compiled) == 2 &&
		(compiled[0] == ActivationTimingSorcery && compiled[1] == ActivationTimingOncePerTurn ||
			compiled[0] == ActivationTimingOncePerTurn && compiled[1] == ActivationTimingSorcery) {
		return ActivationTimingSorceryOncePerTurn, span
	}
	return ActivationTimingUnsupported, span
}

func compileActivationRestriction(restriction *parser.ActivationRestriction) ActivationTimingKind {
	switch restriction.Kind {
	case parser.ActivationRestrictionSorceryTiming:
		return ActivationTimingSorcery
	case parser.ActivationRestrictionInstantTiming:
		return ActivationTimingInstant
	case parser.ActivationRestrictionFrequency:
		if restriction.Frequency.Count.Kind == parser.ActivationFrequencyCountOnce &&
			restriction.Frequency.Period.Kind == parser.ActivationFrequencyPeriodTurn {
			return ActivationTimingOncePerTurn
		}
	case parser.ActivationRestrictionPhaseStep:
		if restriction.PhaseStep.Name.Kind == parser.PhaseStepNameCombat &&
			restriction.PhaseStep.Player.Kind == parser.TriggerPlayerSelectorAny &&
			(restriction.PhaseStep.Quantifier.Kind == parser.PhaseStepQuantifierNone ||
				restriction.PhaseStep.Quantifier.Kind == parser.PhaseStepQuantifierEach) {
			return ActivationTimingDuringCombat
		}
		if restriction.PhaseStep.Name.Kind == parser.PhaseStepNameUpkeep &&
			restriction.PhaseStep.Player.Kind == parser.TriggerPlayerSelectorYou &&
			(restriction.PhaseStep.Quantifier.Kind == parser.PhaseStepQuantifierSingle ||
				restriction.PhaseStep.Quantifier.Kind == parser.PhaseStepQuantifierEachOf) {
			return ActivationTimingDuringUpkeep
		}
	case parser.ActivationRestrictionPlayerTurn:
		if restriction.PlayerTurn.Player.Kind == parser.TriggerPlayerSelectorYou {
			return ActivationTimingDuringYourTurn
		}
	case parser.ActivationRestrictionTurnBeforeAttackers:
		if restriction.PlayerTurn.Player.Kind == parser.TriggerPlayerSelectorYou {
			return ActivationTimingDuringYourTurnBeforeAttackers
		}
	default:
	}
	return ActivationTimingUnsupported
}

func recognizeActivationZone(ability *CompiledAbility) {
	if ability.Kind != AbilityActivated {
		return
	}
	ability.ActivationZone = zone.Battlefield
	if activationCostUsesSourceFromZone(*ability, zone.Hand) {
		ability.ActivationZone = zone.Hand
		return
	}
	if activationCostUsesSourceFromGraveyard(*ability) ||
		contentReturnsSourceFromGraveyard(ability.Content) {
		ability.ActivationZone = zone.Graveyard
	}
}

func activationCostUsesSourceFromGraveyard(ability CompiledAbility) bool {
	return activationCostUsesSourceFromZone(ability, zone.Graveyard)
}

func activationCostUsesSourceFromZone(ability CompiledAbility, sourceZone zone.Type) bool {
	if ability.Cost == nil {
		return false
	}
	for _, component := range ability.Cost.Components {
		if component.SourceSelf && component.SourceZone == sourceZone {
			return true
		}
	}
	for _, reference := range ability.Content.References {
		if reference.Binding != ReferenceBindingSource || !ability.Cost.Order.Contains(reference.Order) {
			continue
		}
		for _, component := range ability.Cost.Components {
			if component.Order.Contains(reference.Order) &&
				component.SourceZone == sourceZone {
				return true
			}
		}
	}
	return false
}

func contentReturnsSourceFromGraveyard(content AbilityContent) bool {
	for effectIndex := range content.Effects {
		effect := &content.Effects[effectIndex]
		if effect.Kind != EffectReturn || effect.FromZone != zone.Graveyard {
			continue
		}
		for _, reference := range content.References {
			if reference.Binding == ReferenceBindingSource &&
				referenceFollowsEffectVerbInClause(effectIndex, content.Effects, reference.Order) {
				return true
			}
		}
	}
	for _, mode := range content.Modes {
		if contentReturnsSourceFromGraveyard(mode.Content) {
			return true
		}
	}
	return false
}

func referenceFollowsEffectVerbInClause(effectIndex int, effects []CompiledEffect, reference shared.SourceOrder) bool {
	effect := effects[effectIndex]
	if reference.Start < effect.VerbOrder.End || reference.End > effect.Order.End {
		return false
	}
	for i := effectIndex + 1; i < len(effects); i++ {
		next := effects[i]
		if next.Order != effect.Order {
			continue
		}
		if next.VerbOrder.Start < reference.End {
			return false
		}
		break
	}
	return true
}

func compileMode(
	mode *parser.Mode,
	context Context,
) (CompiledMode, []shared.Diagnostic) {
	targets := compileTypedTargets(mode.Sentences)
	effects := compileEffects(mode.Sentences)
	references := bindReferences(compileTypedReferences(mode.SemanticReferences), targets, effects, nil)
	applyEffectReferenceBindings(effects, references)
	compiled := CompiledMode{
		Span:  mode.Span,
		Text:  mode.Text,
		Label: compileModeLabel(mode.Label),
		Content: AbilityContent{
			Targets:    targets,
			Conditions: compileConditions(mode.ConditionSegments, mode.ConditionClauses, mode.EventHistoryConditions),
			Effects:    effects,
			Keywords:   compileKeywords(mode.SemanticKeywords),
			References: references,
		},
	}
	if mode.SpreeCost != nil {
		compiled.SpreeCost = slices.Clone(mode.SpreeCost.Cost)
	}
	applyEffectPaymentsToConditions(compiled.Content.Effects, compiled.Content.Conditions)
	compiled.Content.Span = mode.Body.Span
	return compiled, nil
}

func compileModeLabel(label *parser.ModeLabelClause) CompiledModeLabel {
	if label == nil {
		return CompiledModeLabelNone
	}
	switch label.Kind {
	case parser.ModeLabelSellContraband:
		return CompiledModeLabelSellContraband
	case parser.ModeLabelBuyInformation:
		return CompiledModeLabelBuyInformation
	case parser.ModeLabelHireMercenary:
		return CompiledModeLabelHireMercenary
	default:
		return CompiledModeLabelNone
	}
}
