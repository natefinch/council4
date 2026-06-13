// Package compiler lowers parsed Oracle syntax into semantic intermediate
// representation for card generation.
package compiler

import (
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
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
	if ability.Cost != nil {
		cost := compileCost(*ability.Cost, kind, ability.Atoms)
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
	if kind == AbilityTriggered &&
		len(tokens) >= 2 &&
		equalWord(tokens[0], "you") &&
		equalWord(tokens[1], "may") {
		compiled.Optional = true
		compiled.OptionalSpan = shared.Span{Start: tokens[0].Span.Start, End: tokens[1].Span.End}
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
		compiled.Content.Conditions = compileConditions(conditionTokens, kind == AbilityTriggered, ability.Atoms)
		if containsSequence(shared.NormalizedWords(tokens), "attacks", "each", "combat", "if", "able") {
			compiled.Content.Conditions = slices.DeleteFunc(compiled.Content.Conditions, func(condition CompiledCondition) bool {
				return strings.EqualFold(condition.Text, "if able")
			})
		}
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
				strings.Contains(strings.ToLower(component.Text), "from your graveyard") {
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
			Conditions: compileConditions(tokens, false, mode.Atoms),
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

func compileCost(phrase parser.Phrase, abilityKind AbilityKind, atoms parser.Atoms) CompiledCost {
	cost := CompiledCost{Span: phrase.Span, Text: phrase.Text}
	parts := splitTopLevel(phrase.Tokens, shared.Comma)
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}
		component := CostComponent{
			Kind: CostUnknown,
			Span: shared.SpanOf(part),
			Text: shared.SliceSpan(phrase.Text, relativeSpan(shared.SpanOf(part), phrase.Span.Start.Offset)),
		}
		if abilityKind == AbilityLoyalty {
			component.Kind = CostLoyalty
			component.Amount = joinedTokenText(part)
		} else {
			words := shared.NormalizedWords(part)
			switch {
			case len(part) == 1 && part[0].Kind == shared.Symbol && strings.EqualFold(part[0].Text, "{T}"):
				component.Kind = CostTap
				component.Symbol = part[0].Text
			case len(part) == 1 && part[0].Kind == shared.Symbol && strings.EqualFold(part[0].Text, "{Q}"):
				component.Kind = CostUntap
				component.Symbol = part[0].Text
			case startsWords(words, "sacrifice"):
				component.Kind = CostSacrifice
				component.Object = wordsAfterFirst(part)
			case startsWords(words, "discard"):
				component.Kind = CostDiscard
				component.Object = wordsAfterFirst(part)
			case startsWords(words, "pay") && shared.ContainsWord(words, "life"):
				component.Kind = CostPayLife
				component.Amount = firstInteger(part)
			case startsWords(words, "pay") && allEnergySymbols(part[1:]):
				component.Kind = CostEnergy
				component.Amount = strconv.Itoa(len(part) - 1)
				component.AmountValue = len(part) - 1
				component.AmountKnown = true
			case startsWords(words, "return") && shared.ContainsWord(words, "hand"):
				component.Kind = CostReturn
				component.Object = wordsAfterFirst(part)
			case startsWords(words, "reveal"):
				component.Kind = CostReveal
				component.Object = wordsAfterFirst(part)
			case startsWords(words, "exert"):
				component.Kind = CostExert
				component.Object = wordsAfterFirst(part)
			case startsWords(words, "mill"):
				component.Kind = CostMill
				component.Object = wordsAfterFirst(part)
			case startsWords(words, "put") && containsNoun(words, "counter"):
				component.Kind = CostPutCounter
				component.Object = wordsAfterFirst(part)
			case startsWords(words, "collect", "evidence") && len(part) == 3 && positiveIntegerWord(firstInteger(part)):
				component.Kind = CostCollectEvidence
				component.Amount = firstInteger(part)
			case startsWords(words, "exile"):
				component.Kind = CostExile
				component.Object = wordsAfterFirst(part)
			case startsWords(words, "remove") && (shared.ContainsWord(words, "counter") || shared.ContainsWord(words, "counters")):
				component.Kind = CostRemoveCounter
				component.Object = wordsAfterFirst(part)
			case startsWords(words, "tap"):
				component.Kind = CostTapPermanents
				component.Object = wordsAfterFirst(part)
			case allSymbols(part):
				component.Kind = CostMana
				component.Symbol = joinedTokenText(part)
			default:
			}
		}
		compileCostAtoms(&component, part, atoms)
		cost.Components = append(cost.Components, component)
	}
	return cost
}

func compileCostAtoms(component *CostComponent, tokens []shared.Token, atoms parser.Atoms) {
	if len(tokens) == 0 {
		return
	}
	object := tokens[1:]
	switch component.Kind {
	case CostPayLife, CostCollectEvidence:
		annotateIntegerCostAmount(component)
	case CostEnergy:
		component.AmountValue = len(tokens) - 1
		component.AmountKnown = component.AmountValue > 0
	case CostSacrifice:
		if costSelfReference(object, atoms, false) {
			component.SourceSelf = true
			return
		}
		annotateExactCostObject(component, object, atoms, false)
	case CostDiscard:
		annotateExactCostObject(component, object, atoms, true)
	case CostExile:
		if costSelfReference(object, atoms, false) {
			component.SourceSelf = true
			return
		}
		annotateExileCostObject(component, object, atoms)
	case CostExert:
		component.SourceSelf = costSelfReference(object, atoms, true)
	case CostMill:
		annotateMillCostObject(component, object, atoms)
	case CostPutCounter:
		annotatePutCounterCostObject(component, object, atoms)
	case CostRemoveCounter:
		annotateRemoveCounterCostObject(component, object, atoms)
	case CostReveal:
		annotateRevealCostObject(component, object, atoms)
	case CostReturn:
		annotateReturnCostObject(component, object, atoms)
	case CostTapPermanents:
		annotateTapPermanentsCostObject(component, object, atoms)
	default:
	}
}

func annotateIntegerCostAmount(component *CostComponent) {
	amount, err := strconv.Atoi(component.Amount)
	if err == nil && amount > 0 {
		component.AmountValue = amount
		component.AmountKnown = true
	}
}

func costAmountAt(component *CostComponent, token shared.Token, atoms parser.Atoms, allowX bool) bool {
	switch {
	case allowX && equalWord(token, "x"):
		component.AmountFromX = true
		return true
	case equalWord(token, "a"), equalWord(token, "an"):
		component.AmountValue = 1
		component.AmountKnown = true
		return true
	case token.Kind == shared.Integer:
		if value, err := strconv.Atoi(token.Text); err == nil && value > 0 {
			component.AmountValue = value
			component.AmountKnown = true
			return true
		}
	default:
		if value, ok := atoms.CardinalAt(token.Span); ok {
			component.AmountValue = value
			component.AmountKnown = true
			return true
		}
	}
	return false
}

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

func annotateExactCostObject(component *CostComponent, object []shared.Token, atoms parser.Atoms, cardObject bool) {
	words := object
	if cardObject {
		if len(words) < 2 || !costCardNoun(words[len(words)-1], atoms) {
			return
		}
		words = words[:len(words)-1]
	}
	if len(words) == 4 {
		if !equalWord(words[2], "you") || !equalWord(words[3], "control") {
			return
		}
		words = words[:2]
	}
	if cardObject && len(words) == 1 {
		if costAmountAt(component, words[0], atoms, false) {
			component.ObjectKind = SelectorCard
		}
		return
	}
	if len(words) != 2 || !costAmountAt(component, words[0], atoms, false) {
		return
	}
	noun, ok := atoms.ObjectNounAt(words[1].Span)
	if !ok {
		return
	}
	if cardObject && noun == parser.ObjectNounPermanent {
		return
	}
	if cardObject {
		component.ObjectKind = SelectorCard
		if typ, ok := costCardTypeFromNoun(noun); ok {
			component.ObjectType = typ
			component.ObjectTypeKnown = true
		}
		return
	}
	annotateCostObjectNoun(component, noun)
}

func costCardNoun(token shared.Token, atoms parser.Atoms) bool {
	noun, ok := atoms.ObjectNounAt(token.Span)
	return ok && noun == parser.ObjectNounCard
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

func costSelfReference(tokens []shared.Token, atoms parser.Atoms, allowIt bool) bool {
	if len(tokens) == 0 {
		return false
	}
	span := shared.SpanOf(tokens)
	for _, reference := range atoms.ReferencesIn(span) {
		if reference.Span != span {
			continue
		}
		switch reference.Kind {
		case parser.ReferenceSelfName:
			return true
		case parser.ReferencePronoun:
			if allowIt && len(tokens) == 1 && equalWord(tokens[0], "it") {
				return true
			}
		case parser.ReferenceThisObject:
			if len(tokens) != 2 {
				continue
			}
			noun, ok := atoms.ObjectNounAt(tokens[1].Span)
			if !ok {
				continue
			}
			switch noun {
			case parser.ObjectNounArtifact, parser.ObjectNounCreature, parser.ObjectNounEnchantment,
				parser.ObjectNounLand, parser.ObjectNounPermanent, parser.ObjectNounToken:
				return true
			default:
			}
		default:
		}
	}
	return false
}

func annotateMillCostObject(component *CostComponent, object []shared.Token, atoms parser.Atoms) {
	if len(object) != 2 || !costAmountAt(component, object[0], atoms, false) || !costCardNoun(object[1], atoms) {
		return
	}
	component.ObjectKind = SelectorCard
}

func annotatePutCounterCostObject(component *CostComponent, object []shared.Token, atoms parser.Atoms) {
	counterIndex := singleCounterWordIndex(object)
	if counterIndex <= 1 || counterIndex+2 >= len(object) || !equalWord(object[counterIndex+1], "on") {
		return
	}
	if !costAmountAt(component, object[0], atoms, false) ||
		!costSelfReference(object[counterIndex+2:], atoms, true) {
		return
	}
	kind, ok := exactCostCounterKind(object[1:counterIndex], atoms, putCounterCostKinds())
	if !ok {
		return
	}
	component.CounterKind = kind
	component.CounterKindKnown = true
	component.SourceSelf = true
}

func annotateRemoveCounterCostObject(component *CostComponent, object []shared.Token, atoms parser.Atoms) {
	counterIndex := singleCounterWordIndex(object)
	if counterIndex <= 1 || counterIndex+2 >= len(object) || !equalWord(object[counterIndex+1], "from") {
		return
	}
	if !costAmountAt(component, object[0], atoms, false) ||
		!costSelfReference(object[counterIndex+2:], atoms, true) {
		return
	}
	kind, ok := exactCostCounterKind(object[1:counterIndex], atoms, removeCounterCostKinds())
	if !ok {
		return
	}
	component.CounterKind = kind
	component.CounterKindKnown = true
	component.SourceSelf = true
}

func singleCounterWordIndex(tokens []shared.Token) int {
	index := -1
	for i, token := range tokens {
		if !equalWord(token, "counter") && !equalWord(token, "counters") {
			continue
		}
		if index >= 0 {
			return -1
		}
		index = i
	}
	return index
}

func exactCostCounterKind(tokens []shared.Token, atoms parser.Atoms, allowed []counter.Kind) (counter.Kind, bool) {
	if len(tokens) == 0 {
		return 0, false
	}
	kind, span, ok := atoms.CounterIn(shared.SpanOf(tokens))
	if !ok || span != shared.SpanOf(tokens) || !slices.Contains(allowed, kind) {
		return 0, false
	}
	return kind, true
}

func putCounterCostKinds() []counter.Kind {
	return []counter.Kind{counter.PlusOnePlusOne, counter.MinusOneMinusOne, counter.Charge, counter.Verse, counter.Blood}
}

func removeCounterCostKinds() []counter.Kind {
	return []counter.Kind{
		counter.PlusOnePlusOne, counter.MinusOneMinusOne, counter.Loyalty, counter.Charge,
		counter.Time, counter.Defense, counter.Lore, counter.Verse, counter.Shield,
		counter.Stun, counter.Finality, counter.Brick, counter.Page, counter.Enlightened,
		counter.Oil, counter.Blood, counter.Indestructible, counter.Deathtouch,
		counter.Flying, counter.FirstStrike, counter.Hexproof, counter.Lifelink,
		counter.Menace, counter.Reach, counter.Trample, counter.Vigilance,
	}
}

func annotateRevealCostObject(component *CostComponent, object []shared.Token, atoms parser.Atoms) {
	if len(object) >= 4 &&
		equalWord(object[len(object)-4], "that") &&
		equalWord(object[len(object)-3], "share") &&
		equalWord(object[len(object)-2], "a") &&
		equalWord(object[len(object)-1], "color") {
		object = object[:len(object)-4]
	}
	if len(object) < 5 ||
		!equalWord(object[len(object)-3], "from") ||
		!equalWord(object[len(object)-2], "your") ||
		!equalWord(object[len(object)-1], "hand") {
		return
	}
	if z, ok := atoms.ZoneIn(shared.SpanOf(object[len(object)-3:]), parser.ZoneRoleFrom); !ok || z != zone.Hand {
		return
	}
	prefix := object[:len(object)-3]
	if len(prefix) < 2 || len(prefix) > 3 || !costAmountAt(component, prefix[0], atoms, true) {
		return
	}
	if len(prefix) == 3 {
		colorAtom, ok := atoms.ColorAt(prefix[1].Span)
		if !ok {
			return
		}
		mapped, ok := compilerColor(colorAtom)
		if !ok {
			return
		}
		component.ObjectColor = mapped
		component.ObjectColorKnown = true
		prefix = append(prefix[:1], prefix[2])
	}
	if !costCardNoun(prefix[1], atoms) {
		return
	}
	component.ObjectKind = SelectorCard
	component.SourceZone = zone.Hand
}

func annotateReturnCostObject(component *CostComponent, object []shared.Token, atoms parser.Atoms) {
	if len(object) < 6 ||
		!equalWord(object[len(object)-6], "you") ||
		!equalWord(object[len(object)-5], "control") ||
		!equalWord(object[len(object)-4], "to") ||
		!strings.EqualFold(object[len(object)-2].Text, "owner's") ||
		!equalWord(object[len(object)-1], "hand") {
		return
	}
	pronoun, ok := atoms.PronounAt(object[len(object)-3].Span)
	if !ok || pronoun != parser.PronounIts && pronoun != parser.PronounTheir {
		return
	}
	if z, ok := atoms.ZoneIn(shared.SpanOf(object[len(object)-4:]), parser.ZoneRoleTo); !ok || z != zone.Hand {
		return
	}
	prefix := object[:len(object)-6]
	if len(prefix) < 2 || !costAmountAt(component, prefix[0], atoms, false) {
		return
	}
	prefix = prefix[1:]
	if len(prefix) > 0 && equalWord(prefix[0], "tapped") {
		component.RequireTapped = true
		prefix = prefix[1:]
	}
	if annotateCostPermanentObject(component, prefix, atoms, true, []types.Card{types.Land, types.Creature, types.Artifact, types.Enchantment}) {
		component.ObjectController = ControllerYou
		component.ToZone = zone.Hand
	}
}

func annotateTapPermanentsCostObject(component *CostComponent, object []shared.Token, atoms parser.Atoms) {
	if len(object) < 5 ||
		!costAmountAt(component, object[0], atoms, false) ||
		!equalWord(object[1], "untapped") ||
		!equalWord(object[len(object)-2], "you") ||
		!equalWord(object[len(object)-1], "control") {
		return
	}
	if annotateCostPermanentObject(component, object[2:len(object)-2], atoms, false, []types.Card{types.Creature, types.Artifact}) {
		component.RequireUntapped = true
		component.ObjectController = ControllerYou
	}
}

func annotateCostPermanentObject(component *CostComponent, object []shared.Token, atoms parser.Atoms, allowSnowLand bool, subtypeTypes []types.Card) bool {
	if len(object) == 0 {
		return false
	}
	if allowSnowLand && len(object) == 2 && equalWord(object[0], "snow") {
		noun, ok := atoms.ObjectNounAt(object[1].Span)
		supertype, superOK := atoms.SupertypeAt(object[0].Span)
		if !ok || noun != parser.ObjectNounLand || !superOK || supertype != parser.SupertypeSnow {
			return false
		}
		component.ObjectKind = SelectorLand
		component.ObjectType = types.Land
		component.ObjectTypeKnown = true
		component.ObjectSupertype = types.Snow
		component.SupertypeKnown = true
		return true
	}
	if len(object) == 1 {
		if noun, ok := atoms.ObjectNounAt(object[0].Span); ok {
			return annotateCostObjectNoun(component, noun)
		}
		if sub, ok := atoms.SubtypeAt(object[0].Span); ok {
			if !parser.SubtypeMatchesAnyRuntimeCardType(sub, subtypeTypes) {
				return false
			}
			component.SubtypesAny = []types.Sub{sub}
			return true
		}
	}
	return false
}

func annotateExileCostObject(component *CostComponent, object []shared.Token, atoms parser.Atoms) {
	if len(object) < 5 ||
		!equalWord(object[len(object)-3], "from") ||
		!equalWord(object[len(object)-2], "your") ||
		!equalWord(object[len(object)-1], "graveyard") {
		return
	}
	if z, ok := atoms.ZoneIn(shared.SpanOf(object[len(object)-3:]), parser.ZoneRoleFrom); !ok || z != zone.Graveyard {
		return
	}
	prefix := object[:len(object)-3]
	switch {
	case len(prefix) == 2 && exileCardAmount(component, prefix[0], atoms) && costCardNoun(prefix[1], atoms):
		component.ObjectKind = SelectorCard
	case len(prefix) == 3 && exileTypedCardAmount(component, prefix[0]) && costCardNoun(prefix[2], atoms):
		noun, ok := atoms.ObjectNounAt(prefix[1].Span)
		if !ok {
			return
		}
		typ, ok := costCardTypeFromNoun(noun)
		if !ok {
			return
		}
		component.ObjectKind = SelectorCard
		component.ObjectType = typ
		component.ObjectTypeKnown = true
	default:
		return
	}
	component.SourceZone = zone.Graveyard
}

func exileCardAmount(component *CostComponent, token shared.Token, atoms parser.Atoms) bool {
	if equalWord(token, "a") || equalWord(token, "an") {
		component.AmountValue = 1
		component.AmountKnown = true
		return true
	}
	if value, ok := atoms.CardinalAt(token.Span); ok && value == 2 {
		component.AmountValue = 2
		component.AmountKnown = true
		return true
	}
	return false
}

func exileTypedCardAmount(component *CostComponent, token shared.Token) bool {
	if !equalWord(token, "a") && !equalWord(token, "an") {
		return false
	}
	component.AmountValue = 1
	component.AmountKnown = true
	return true
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

func compileTrigger(ability parser.Ability, _ Context) CompiledTrigger {
	trigger := CompiledTrigger{
		Kind: TriggerUnknown,
	}
	if ability.Trigger == nil {
		return trigger
	}
	trigger.Span = ability.Trigger.Span
	trigger.Text = ability.Trigger.Text
	trigger.Event = ability.Trigger.Event.Text
	switch ability.Trigger.Introduction.Kind {
	case parser.TriggerIntroductionWhen:
		trigger.Kind = TriggerWhen
	case parser.TriggerIntroductionWhenever:
		trigger.Kind = TriggerWhenever
	case parser.TriggerIntroductionAt:
		trigger.Kind = TriggerAt
	default:
	}
	conditions := compileConditions(ability.Tokens, true, ability.Atoms)
	for i := range conditions {
		if conditions[i].Intervening {
			condition := conditions[i]
			trigger.Condition = &condition
			break
		}
	}
	switch {
	case ability.Trigger.PhaseStep != nil:
		trigger.Pattern = compilePhaseStepTriggerPattern(
			ability.Trigger.PhaseStep,
			trigger.Kind,
			trigger.Condition,
		)
	case ability.Trigger.PlayerEvent != nil:
		trigger.Pattern = compilePlayerEventTriggerPattern(
			ability.Trigger.PlayerEvent,
			trigger.Kind,
			trigger.Condition,
		)
	default:
		trigger.Pattern = compileTriggerPatternForSyntax(
			joinedSourceText(ability.Trigger.Event.Tokens),
			trigger.Kind,
			ability.Trigger.Event.Span,
			ability.Trigger.Event.Tokens,
			ability.Atoms,
			trigger.Condition,
		)
	}
	return trigger
}

func runtimeCardTypeFromParser(cardType parser.CardType) (types.Card, bool) {
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

func runtimeColorFromParser(colorValue parser.Color) (color.Color, bool) {
	switch colorValue {
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

func compileConditions(tokens []shared.Token, triggered bool, atoms parser.Atoms) []CompiledCondition {
	var conditions []CompiledCondition
	for i := 0; i < len(tokens); i++ {
		kind := conditionKindAt(tokens, i)
		if kind == ConditionUnknown {
			continue
		}
		// Skip source-tied duration phrases that are captured by compileDuration.
		// Only suppress "as long as you control" when it is preceded by "for"
		// (making it "for as long as you control"), which distinguishes a duration
		// suffix from a leading static-ability condition like "As long as you
		// control a Mountain, this creature has...".
		// Also skip "as long as this [type] remains on the battlefield" (but NOT
		// other "as long as this [type] is [state]" forms which are real conditions).
		if kind == ConditionAsLongAs {
			if i > 0 && equalWord(tokens[i-1], "for") {
				// "for as long as ..." — the "as long as" is a duration suffix.
				end := conditionEnd(tokens, i)
				i = end - 1
				continue
			}
			if isSourceOnBattlefieldPhrase(tokens, i) {
				// "as long as this [type] remains on the battlefield" — duration.
				end := conditionEnd(tokens, i)
				i = end - 1
				continue
			}
		}
		start := i
		end := conditionEnd(tokens, i)
		phrase := tokens[start:end]
		condition := CompiledCondition{
			Kind:        kind,
			Span:        shared.SpanOf(phrase),
			Text:        joinedSourceText(phrase),
			Intervening: triggered && kind == ConditionIf && isInterveningIf(tokens, start),
		}
		recognizeCondition(&condition, phrase, atoms)
		conditions = append(conditions, condition)
		i = end - 1
	}
	return conditions
}

func conditionEnd(tokens []shared.Token, start int) int {
	for i := start; i < len(tokens); i++ {
		if tokens[i].Kind == shared.Period || (i > start && tokens[i].Kind == shared.Comma) {
			return i
		}
	}
	return len(tokens)
}

func compileEffects(sentences []parser.Sentence) []CompiledEffect {
	var effects []CompiledEffect
	for _, sentence := range sentences {
		if sentence.StaticRule != nil {
			if effect, ok := compileStaticRuleEffect(sentence); ok {
				effects = append(effects, effect)
			}
			continue
		}
		for syntaxIndex := range sentence.Effects {
			syntax := &sentence.Effects[syntaxIndex]
			effects = append(effects, CompiledEffect{
				Kind:               compileEffectKind(syntax.Kind),
				Context:            syntax.Context,
				Connection:         syntax.Connection,
				ConnectionSpan:     syntax.ConnectionSpan,
				Span:               syntax.Span,
				ClauseSpan:         syntax.ClauseSpan,
				Text:               syntax.Text,
				VerbSpan:           syntax.VerbSpan,
				References:         compileTypedReferences(syntax.References),
				SubjectReferences:  compileTypedReferences(syntax.SubjectReferences),
				Targets:            compileTypedTargetList(syntax.Targets),
				SubjectTargets:     compileTypedTargetList(syntax.SubjectTargets),
				Duration:           compileEffectDuration(syntax.Duration),
				DelayedTiming:      compileDelayedTiming(syntax.DelayedTiming),
				Selector:           compileTypedSelection(syntax.Selection),
				Amount:             compileTypedAmount(syntax.Amount),
				PowerDelta:         compileSignedAmount(syntax.PowerDelta),
				ToughnessDelta:     compileSignedAmount(syntax.ToughnessDelta),
				StaticSubject:      compileStaticSubjectKind(syntax.StaticSubject.Kind),
				StaticSubjectSpan:  syntax.StaticSubject.Span,
				Details:            compiledEffectDetails(staticSubjectType(syntax.StaticSubject.SubtypeText, syntax.StaticSubject.Subtype, syntax.StaticSubject.SubtypeKnown), syntax.Symbol),
				CounterKind:        syntax.CounterKind,
				CounterKindKnown:   syntax.CounterKnown,
				FromZone:           syntax.FromZone,
				ToZone:             syntax.ToZone,
				Destination:        syntax.Destination,
				EntersTapped:       syntax.EntersTapped,
				EntersWithCounters: syntax.EntersWithCounters,
				UnderYourControl:   syntax.UnderYourControl,
				CastAsAdventure:    syntax.CastAsAdventure,
				Negated:            syntax.Negated,
				Optional:           syntax.Optional,
				OptionalSpan:       syntax.OptionalSpan,
				Mana: CompiledEffectMana{
					Span:            syntax.Mana.Span,
					Symbols:         slices.Clone(syntax.Mana.Symbols),
					Choice:          syntax.Mana.Choice,
					AnyColor:        syntax.Mana.AnyColor,
					LegacyBodyExact: syntax.Mana.LegacyBodyExact,
				},
				Replacement:             syntax.Replacement,
				Payment:                 compileEffectPayment(syntax.Payment),
				Exact:                   syntax.Exact,
				RequiresOrderedLowering: syntax.RequiresOrderedLowering,
				HasUnrecognizedSibling:  syntax.HasUnrecognizedSibling,
				UnsupportedDetail:       syntax.UnsupportedDetail,
			})
		}
	}
	return effects
}

func conditionKindAt(tokens []shared.Token, index int) ConditionKind {
	switch {
	case equalWord(tokens[index], "if"):
		return ConditionIf
	case equalWord(tokens[index], "unless"):
		return ConditionUnless
	case index+1 < len(tokens) &&
		equalWord(tokens[index], "only") &&
		equalWord(tokens[index+1], "if"):
		return ConditionOnlyIf
	case index+2 < len(tokens) &&
		equalWord(tokens[index], "as") &&
		equalWord(tokens[index+1], "long") &&
		equalWord(tokens[index+2], "as"):
		return ConditionAsLongAs
	default:
		return ConditionUnknown
	}
}

// isSourceOnBattlefieldPhrase reports whether tokens starting at index represent
// "as long as this [type] remains on the battlefield" or
// "as long as this [type] is on the battlefield". This is specifically the
// DurationForAsLongAsSourceOnBattlefield duration pattern — NOT other
// "as long as this [type] is [state]" conditions (which are real conditions).
func isSourceOnBattlefieldPhrase(tokens []shared.Token, index int) bool {
	words := shared.NormalizedWords(tokens[index:])
	return containsSequence(words, "as", "long", "as", "this") &&
		(containsSequence(words, "remains", "on", "the", "battlefield") ||
			containsSequence(words, "is", "on", "the", "battlefield"))
}

func compileStaticRuleEffect(sentence parser.Sentence) (CompiledEffect, bool) {
	rule, _, ok := semanticStaticRuleForSyntax(*sentence.StaticRule)
	if !ok {
		return CompiledEffect{}, false
	}
	kind := effectKindForStaticRule(rule)
	if kind == EffectUnknown {
		return CompiledEffect{}, false
	}
	selector := CompiledSelector{}
	if sentence.StaticRule.Subject.Kind == parser.StaticRuleSubjectSourceCreature {
		selector.Kind = SelectorCreature
	}
	return CompiledEffect{
		Kind:     kind,
		Span:     sentence.StaticRule.Span,
		Text:     sentence.Text,
		VerbSpan: sentence.StaticRule.Operation.Span,
		Selector: selector,
		Negated:  sentence.StaticRule.Constraint.Kind == parser.StaticRuleConstraintProhibition,
	}, true
}

func effectKindForStaticRule(rule StaticRuleKind) EffectKind {
	switch rule {
	case StaticRuleCantBlock:
		return EffectCantBlock
	case StaticRuleCantBeBlocked:
		return EffectCantBeBlocked
	case StaticRuleMustAttack:
		return EffectMustAttack
	case StaticRuleCantBeCountered:
		return EffectCantBeCountered
	default:
		return EffectUnknown
	}
}

func staticRuleSentencesOnly(sentences []parser.Sentence) bool {
	if len(sentences) == 0 {
		return false
	}
	for _, sentence := range sentences {
		if sentence.StaticRule == nil {
			return false
		}
	}
	return true
}

func compileStaticRuleReferences(sentences []parser.Sentence) []CompiledReference {
	references := make([]CompiledReference, 0, len(sentences))
	for i, sentence := range sentences {
		references = append(references, CompiledReference{
			Kind:       ReferenceThisObject,
			Span:       sentence.StaticRule.Subject.Span,
			Binding:    ReferenceBindingSource,
			Occurrence: i,
		})
	}
	return references
}

func abilityBodyTokens(ability parser.Ability) []shared.Token {
	tokens := ability.Tokens
	if ability.AbilityWord != nil {
		if dash := shared.TopLevelIndex(tokens, shared.EmDash); dash >= 0 {
			tokens = tokens[dash+1:]
		}
	}
	switch ability.Kind {
	case parser.AbilityActivated, parser.AbilityLoyalty:
		if colon := shared.TopLevelIndex(tokens, shared.Colon); colon >= 0 {
			return tokens[colon+1:]
		}
	case parser.AbilityTriggered:
		if comma := triggerBodyComma(tokens); comma >= 0 {
			return tokens[comma+1:]
		}
	default:
	}
	return tokens
}

func compileKeywords(tokens []shared.Token, atoms parser.Atoms) []CompiledKeyword {
	syntaxKeywords := atoms.KeywordsWithin(tokens)
	keywords := make([]CompiledKeyword, 0, len(syntaxKeywords))
	for i := range syntaxKeywords {
		keyword := &syntaxKeywords[i]
		compiled := CompiledKeyword{
			Kind:          keyword.Kind,
			Name:          keyword.Kind.String(),
			Span:          keyword.Span,
			Text:          keyword.Text,
			Parameter:     keyword.Parameter.Text,
			ParameterKind: keyword.Parameter.Kind,
			ManaCost:      keyword.Parameter.ManaCost(),
			Integer:       keyword.Parameter.Integer(),
			EnchantTarget: keyword.Parameter.EnchantTarget(),
		}
		if keyword.Parameter.Kind == parser.KeywordParameterProtection {
			compiled.Protection, compiled.ProtectionKnown = compileProtectionKeyword(keyword.Parameter.Protection())
		}
		keywords = append(keywords, compiled)
	}
	return keywords
}

func compileProtectionKeyword(parameter parser.ProtectionParameter) (game.ProtectionKeyword, bool) {
	families := 0
	for _, present := range []bool{
		parameter.Everything,
		parameter.EachColor,
		parameter.Multicolored,
		parameter.Monocolored,
		len(parameter.FromColors) > 0,
		len(parameter.FromTypes) > 0,
		len(parameter.FromSubtypes) > 0,
	} {
		if present {
			families++
		}
	}
	if families != 1 {
		return game.ProtectionKeyword{}, false
	}
	protection := game.ProtectionKeyword{
		Everything:   parameter.Everything,
		EachColor:    parameter.EachColor,
		Multicolored: parameter.Multicolored,
		Monocolored:  parameter.Monocolored,
		FromSubtypes: append([]types.Sub(nil), parameter.FromSubtypes...),
	}
	for _, value := range parameter.FromColors {
		compiled, ok := compilerColor(value)
		if !ok {
			return game.ProtectionKeyword{}, false
		}
		protection.FromColors = append(protection.FromColors, compiled)
	}
	for _, value := range parameter.FromTypes {
		compiled, ok := runtimeCardTypeFromParser(value)
		if !ok {
			return game.ProtectionKeyword{}, false
		}
		protection.FromTypes = append(protection.FromTypes, compiled)
	}
	return protection, true
}

func compileReferences(tokens []shared.Token, atoms parser.Atoms) []CompiledReference {
	recognized := atoms.ReferencesWithin(tokens)
	references := make([]CompiledReference, 0, len(recognized))
	for _, reference := range recognized {
		references = append(references, CompiledReference{
			Kind:    compileReferenceKind(reference.Kind),
			Pronoun: compileReferencePronoun(reference.Pronoun),
			Span:    reference.Span,
			Text:    joinedSourceText(reference.Tokens),
		})
	}

	return references
}

func compileTypedReferences(recognized []parser.Reference) []CompiledReference {
	references := make([]CompiledReference, 0, len(recognized))
	for _, reference := range recognized {
		references = append(references, CompiledReference{
			Kind:    compileReferenceKind(reference.Kind),
			Pronoun: compileReferencePronoun(reference.Pronoun),
			Span:    reference.Span,
			Text:    joinedSourceText(reference.Tokens),
		})
	}
	return references
}

func compileReferencePronoun(pronoun parser.PronounKind) ReferencePronounKind {
	switch pronoun {
	case parser.PronounIt:
		return ReferencePronounIt
	case parser.PronounIts:
		return ReferencePronounIts
	case parser.PronounThey:
		return ReferencePronounThey
	case parser.PronounTheir:
		return ReferencePronounTheir
	case parser.PronounThem:
		return ReferencePronounThem
	case parser.PronounThose:
		return ReferencePronounThose
	default:
		return ReferencePronounUnknown
	}
}

func compileReferenceKind(kind parser.ReferenceKind) ReferenceKind {
	switch kind {
	case parser.ReferenceSelfName:
		return ReferenceSelfName
	case parser.ReferenceThisObject:
		return ReferenceThisObject
	case parser.ReferenceThatObject:
		return ReferenceThatObject
	case parser.ReferenceThatPlayer:
		return ReferenceThatPlayer
	case parser.ReferencePronoun:
		return ReferencePronoun
	default:
		return ReferenceUnknown
	}
}

func semanticTokens(tokens []shared.Token, reminders, quoted []parser.Delimited) []shared.Token {
	excluded := append(append([]parser.Delimited(nil), reminders...), quoted...)
	result := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		var skip bool
		for _, delimiter := range excluded {
			if token.Span.Start.Offset >= delimiter.Span.Start.Offset &&
				token.Span.End.Offset <= delimiter.Span.End.Offset {
				skip = true
				break
			}
		}
		if !skip {
			result = append(result, token)
		}
	}
	return result
}

func splitTopLevel(tokens []shared.Token, separator shared.Kind) [][]shared.Token {
	var result [][]shared.Token
	start := 0
	depth := 0
	quoted := false
	for i, token := range tokens {
		switch token.Kind {
		case shared.LeftParen:
			if !quoted {
				depth++
			}
		case shared.RightParen:
			if !quoted && depth > 0 {
				depth--
			}
		case shared.Quote:
			quoted = !quoted
		default:
			if token.Kind == separator && depth == 0 && !quoted {
				result = append(result, tokens[start:i])
				start = i + 1
			}
		}
	}
	return append(result, tokens[start:])
}

func allSymbols(tokens []shared.Token) bool {
	if len(tokens) == 0 {
		return false
	}
	for _, token := range tokens {
		if token.Kind != shared.Symbol {
			return false
		}
	}
	return true
}

func allEnergySymbols(tokens []shared.Token) bool {
	if len(tokens) == 0 {
		return false
	}
	for _, token := range tokens {
		if token.Kind != shared.Symbol || !strings.EqualFold(token.Text, "{E}") {
			return false
		}
	}
	return true
}

func relativeSpan(span shared.Span, base int) shared.Span {
	span.Start.Offset -= base
	span.End.Offset -= base
	return span
}

func wordsAfterFirst(tokens []shared.Token) string {
	if len(tokens) < 2 {
		return ""
	}
	return joinedSourceText(tokens[1:])
}

func firstInteger(tokens []shared.Token) string {
	for _, token := range tokens {
		if token.Kind == shared.Integer {
			return token.Text
		}
	}
	return ""
}

func positiveIntegerWord(word string) bool {
	amount, err := strconv.Atoi(word)
	return err == nil && amount > 0
}

func joinedTokenText(tokens []shared.Token) string {
	var builder strings.Builder
	for _, token := range tokens {
		_, _ = builder.WriteString(token.Text)
	}
	return builder.String()
}

func joinedSourceText(tokens []shared.Token) string {
	if len(tokens) == 0 {
		return ""
	}
	var builder strings.Builder
	for i, token := range tokens {
		if i > 0 && needsSemanticSpace(tokens[i-1], token) {
			_ = builder.WriteByte(' ')
		}
		_, _ = builder.WriteString(token.Text)
	}
	return builder.String()
}

func needsSemanticSpace(previous, current shared.Token) bool {
	if current.Kind == shared.Comma || current.Kind == shared.Period || current.Kind == shared.Colon ||
		current.Kind == shared.Semicolon || current.Kind == shared.RightParen ||
		previous.Kind == shared.LeftParen || previous.Kind == shared.Quote || current.Kind == shared.Quote {
		return false
	}
	if previous.Kind == shared.Plus || previous.Kind == shared.Minus || previous.Kind == shared.Slash ||
		current.Kind == shared.Slash {
		return false
	}
	return previous.Kind != shared.Symbol && current.Kind != shared.Symbol
}

func joinWords(tokens []shared.Token) string {
	var words []string
	for _, token := range tokens {
		if token.Kind != shared.Word {
			return ""
		}
		words = append(words, token.Text)
	}
	return strings.Join(words, " ")
}

func startsWords(words []string, expected ...string) bool {
	if len(words) < len(expected) {
		return false
	}
	for i := range expected {
		if words[i] != expected[i] {
			return false
		}
	}
	return true
}

func containsSequence(words []string, expected ...string) bool {
	for i := 0; i+len(expected) <= len(words); i++ {
		if startsWords(words[i:], expected...) {
			return true
		}
	}
	return false
}

func equalWord(token shared.Token, word string) bool {
	return token.Kind == shared.Word && strings.EqualFold(token.Text, word)
}

func isInterveningIf(tokens []shared.Token, index int) bool {
	comma := triggerBodyComma(tokens)
	return comma >= 0 && index == comma+1
}

func triggerBodyComma(tokens []shared.Token) int {
	comma := shared.TopLevelIndex(tokens, shared.Comma)
	for comma > 0 &&
		comma+1 < len(tokens) &&
		strings.EqualFold(tokens[comma-1].Text, "noncreature") &&
		strings.EqualFold(tokens[comma+1].Text, "nonland") {
		next := shared.TopLevelIndex(tokens[comma+1:], shared.Comma)
		if next < 0 {
			return -1
		}
		comma += next + 1
	}
	return comma
}

func containsNoun(words []string, singular string) bool {
	return shared.ContainsWord(words, singular) || shared.ContainsWord(words, singular+"s")
}

func unsupportedDiagnostic(span shared.Span, text string) shared.Diagnostic {
	return shared.Diagnostic{
		Severity: shared.SeverityWarning,
		Summary:  "unsupported Oracle construct",
		Detail:   "the compiler preserved but did not confidently lower: " + text,
		Span:     span,
	}
}
