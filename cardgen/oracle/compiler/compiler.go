// Package compiler lowers parsed Oracle syntax into semantic intermediate
// representation for card generation.
package compiler

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Compile lowers a parsed syntax document into conservative semantic IR.
func Compile(document parser.Document, context Context) (Compilation, []shared.Diagnostic) {
	compilation := Compilation{Syntax: document}
	var diagnostics []shared.Diagnostic
	for _, ability := range document.Abilities {
		compiled, abilityDiagnostics := compileAbility(ability, context)
		compilation.Abilities = append(compilation.Abilities, compiled)
		diagnostics = append(diagnostics, abilityDiagnostics...)
	}
	return compilation, diagnostics
}

func compileAbility(
	ability parser.Ability,
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
		compiled.AbilityWord = ability.AbilityWord.Text
	}
	compiled.Chapters = append([]int(nil), ability.Chapters...)
	compiled.ChapterSpan = ability.ChapterSpan
	if ability.CostSyntax() != nil {
		cost := compileCost(*ability.CostSyntax())
		compiled.Cost = &cost
	}
	if kind == AbilityTriggered {
		trigger := compileTrigger(ability, context)
		compiled.Trigger = &trigger
	}
	if ability.Modal != nil {
		for _, mode := range ability.Modal.Options {
			compiledMode, modeDiagnostics := compileMode(mode, context)
			compiled.Content.Modes = append(compiled.Content.Modes, compiledMode)
			diagnostics = append(diagnostics, modeDiagnostics...)
		}
	}

	body := abilityBodyTokens(ability)
	timing, timingSpan := compileActivationTiming(kind, ability.ActivationRestrictions)
	if timing != ActivationTimingNone {
		body = tokensOutsideSpan(body, timingSpan)
		compiled.ActivationTiming = timing
		compiled.ActivationTimingSpan = timingSpan
	}
	tokens := semanticTokens(body, ability.Reminders, ability.Quoted)
	if kind == AbilityTriggered && ability.Optional() {
		compiled.Optional = true
		compiled.OptionalSpan = ability.OptionalSpan()
	}
	if kind == AbilityStatic && staticRuleSentencesOnly(ability.Sentences) {
		compiled.Content.Effects = compileEffects(ability.Sentences)
		applyEffectPaymentsToConditions(compiled.Content.Effects, compiled.Content.Conditions)
		compiled.Content.References = compileStaticRuleReferences(ability.Sentences)
	} else {
		compiled.Content.Keywords = compileKeywords(tokens, ability.Atoms)
		compiled.Content.Targets = compileTypedTargets(ability.Sentences)
		conditionTokens := tokens
		if kind == AbilityTriggered {
			conditionTokens = semanticTokens(ability.Tokens, ability.Reminders, ability.Quoted)
		}
		compiled.Content.Conditions = compileConditions(
			conditionTokens,
			kind == AbilityTriggered,
			ability.ConditionBoundaries(),
			ability.ConditionClauses(),
			ability.EventHistoryConditions(),
		)
		compiled.Content.Effects = compileEffects(ability.Sentences)
		applyEffectPaymentsToConditions(compiled.Content.Effects, compiled.Content.Conditions)
		referenceTokens := semanticTokens(ability.Tokens, ability.Reminders, ability.Quoted)
		if timing != ActivationTimingNone {
			referenceTokens = tokensOutsideSpan(referenceTokens, timingSpan)
		}
		compiled.Content.References = compileReferences(
			referenceTokens,
			ability.Atoms,
		)
		compiled.Content.References = bindReferences(
			compiled.Content.References,
			compiled.Content.Targets,
			compiled.Content.Effects,
			compiled.Trigger,
		)
	}
	compiled.Content.References = bindActivationCostReferences(compiled.Kind, compiled.Cost, compiled.Content.References)
	bindConditionReferences(compiled.Content.Conditions, compiled.Content.References, compiled.Trigger)
	applyEffectReferenceBindings(compiled.Content.Effects, compiled.Content.References)
	recognizeActivationZone(&compiled)
	if compiled.Trigger != nil && compiled.Trigger.Condition != nil {
		for i := range compiled.Content.Conditions {
			if compiled.Content.Conditions[i].Span == compiled.Trigger.Condition.Span {
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
	if kind != AbilityReminder && ability.Modal == nil &&
		len(compiled.Content.Effects) == 0 && len(compiled.Content.Keywords) == 0 &&
		!legacyEffectsPresent(ability.Sentences) &&
		(compiled.Static == nil || len(compiled.Static.Declarations) == 0) {
		diagnostics = append(diagnostics, unsupportedDiagnostic(ability.Span, ability.Text))
	}

	// Set Content.Span from the body token range after shell/timing extraction.
	// This is non-zero even for unrecognized content, and for activated/loyalty
	// abilities it excludes the cost span. For triggered abilities with an
	// optional prefix, advance past "you may" since that is shell semantics.
	compiled.Content.Span = shared.SpanOf(body)
	if compiled.Optional && len(tokens) >= 3 {
		compiled.Content.Span.Start = tokens[2].Span.Start
	}
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
			if reference.Span == owned.Span {
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
	default:
		return AbilityUnknown
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
	default:
	}
	return ActivationTimingUnsupported
}

func recognizeActivationZone(ability *CompiledAbility) {
	if ability.Kind != AbilityActivated {
		return
	}
	ability.ActivationZone = zone.Battlefield
	if activationCostUsesSourceFromGraveyard(*ability) ||
		contentReturnsSourceFromGraveyard(ability.Content) {
		ability.ActivationZone = zone.Graveyard
	}
}

func activationCostUsesSourceFromGraveyard(ability CompiledAbility) bool {
	if ability.Cost == nil {
		return false
	}
	for _, reference := range ability.Content.References {
		if reference.Binding != ReferenceBindingSource || !spanContains(ability.Cost.Span, reference.Span) {
			continue
		}
		for _, component := range ability.Cost.Components {
			if spanContains(component.Span, reference.Span) &&
				component.SourceZone == zone.Graveyard {
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
				referenceFollowsEffectVerbInClause(effectIndex, content.Effects, reference.Span) {
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

func referenceFollowsEffectVerbInClause(effectIndex int, effects []CompiledEffect, reference shared.Span) bool {
	effect := effects[effectIndex]
	if reference.Start.Offset < effect.VerbSpan.End.Offset || reference.End.Offset > effect.Span.End.Offset {
		return false
	}
	for i := effectIndex + 1; i < len(effects); i++ {
		next := effects[i]
		if next.Span != effect.Span {
			continue
		}
		if next.VerbSpan.Start.Offset < reference.End.Offset {
			return false
		}
		break
	}
	return true
}

func tokensOutsideSpan(tokens []shared.Token, span shared.Span) []shared.Token {
	return slices.DeleteFunc(append([]shared.Token(nil), tokens...), func(token shared.Token) bool {
		return span.Start.Offset <= token.Span.Start.Offset &&
			span.End.Offset >= token.Span.End.Offset
	})
}

// tokensWithinSpan returns the contiguous run of tokens that lie within span.
// An empty span selects no tokens.
func tokensWithinSpan(tokens []shared.Token, span shared.Span) []shared.Token {
	var result []shared.Token
	for _, token := range tokens {
		if token.Span.Start.Offset >= span.Start.Offset && token.Span.End.Offset <= span.End.Offset {
			result = append(result, token)
		}
	}
	return result
}

func compileMode(
	mode parser.Mode,
	context Context,
) (CompiledMode, []shared.Diagnostic) {
	tokens := semanticTokens(mode.Tokens, mode.Reminders, mode.Quoted)
	targets := compileTypedTargets(mode.Sentences)
	effects := compileEffects(mode.Sentences)
	references := bindReferences(compileReferences(tokens, mode.Atoms), targets, effects, nil)
	applyEffectReferenceBindings(effects, references)
	compiled := CompiledMode{
		Span: mode.Span,
		Text: mode.Text,
		Content: AbilityContent{
			Targets:    targets,
			Conditions: compileConditions(tokens, false, mode.ConditionBoundaries, mode.ConditionClauses(), mode.EventHistoryConditions()),
			Effects:    effects,
			Keywords:   compileKeywords(tokens, mode.Atoms),
			References: references,
		},
	}
	applyEffectPaymentsToConditions(compiled.Content.Effects, compiled.Content.Conditions)
	// Set Content.Span to the mode's full source span: a modal option has no
	// shell cost or trigger, so the mode body IS the content.
	compiled.Content.Span = mode.Span
	return compiled, nil
}
