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
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Compile lowers a parsed syntax document into conservative semantic IR.
func Compile(document parser.Document, context Context) (Compilation, []shared.Diagnostic) {
	compilation := Compilation{Syntax: document}
	var diagnostics []shared.Diagnostic
	for _, ability := range document.Abilities {
		compiled, abilityDiagnostics := compileAbility(document.Source, ability, context)
		compilation.Abilities = append(compilation.Abilities, compiled)
		diagnostics = append(diagnostics, abilityDiagnostics...)
	}
	return compilation, diagnostics
}

func compileAbility(
	source string,
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
			compiledMode, modeDiagnostics := compileMode(source, mode, context)
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
		compiled.Content.Effects = compileEffects(ability.Sentences, nil, nil, ability.Atoms)
		compiled.Content.References = compileStaticRuleReferences(ability.Sentences)
	} else {
		compiled.Content.Keywords = compileKeywords(tokens, ability.Atoms)
		compiled.Content.Targets = compileTargets(tokens, ability.Atoms)
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
		compiled.Content.Effects = compileEffects(
			parser.ParseSentences(source, body),
			ability.Reminders,
			ability.Quoted,
			ability.Atoms,
		)
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
	for effectIndex, effect := range content.Effects {
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
	source string,
	mode parser.Mode,
	context Context,
) (CompiledMode, []shared.Diagnostic) {
	tokens := semanticTokens(mode.Tokens, mode.Reminders, mode.Quoted)
	compiled := CompiledMode{
		Span: mode.Span,
		Text: mode.Text,
		Content: AbilityContent{
			Targets:    compileTargets(tokens, mode.Atoms),
			Conditions: compileConditions(tokens, false, mode.Atoms),
			Effects: compileEffects(
				parser.ParseSentences(source, mode.Tokens),
				mode.Reminders,
				mode.Quoted,
				mode.Atoms,
			),
			Keywords: compileKeywords(tokens, mode.Atoms),
			References: bindReferences(
				compileReferences(tokens, mode.Atoms),
				compileTargets(tokens, mode.Atoms),
				compileEffects(
					parser.ParseSentences(source, mode.Tokens),
					mode.Reminders,
					mode.Quoted,
					mode.Atoms,
				),
				nil,
			),
		},
	}
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

func compileTargets(tokens []shared.Token, atoms parser.Atoms) []CompiledTarget {
	var targets []CompiledTarget
	for i, token := range tokens {
		if token.Kind != shared.Word || !strings.EqualFold(token.Text, "target") {
			continue
		}
		start := i
		cardinality := TargetCardinality{Min: 1, Max: 1}
		switch {
		case i >= 3 && equalWord(tokens[i-3], "any") && equalWord(tokens[i-2], "number") && equalWord(tokens[i-1], "of"):
			start = i - 3
			cardinality.Min = 0
			cardinality.Max = 99
		case i >= 3 && equalWord(tokens[i-3], "up") && equalWord(tokens[i-2], "to"):
			start = i - 3
			cardinality.Min = 0
			cardinality.Max = numberWord(tokens[i-1], atoms)
			if cardinality.Max == 0 {
				cardinality.Max = 1
			}
		case i >= 1:
			if count := numberWord(tokens[i-1], atoms); count > 0 {
				start = i - 1
				cardinality.Min = count
				cardinality.Max = count
			} else if equalWord(tokens[i-1], "any") {
				start = i - 1
			}
		default:
		}
		end := targetPhraseEnd(tokens, i+1)
		phraseTokens := tokens[start:end]
		selectorTokens := append([]shared.Token(nil), tokens[start:i]...)
		selectorTokens = append(selectorTokens, tokens[i+1:end]...)
		targets = append(targets, CompiledTarget{
			Span:        shared.SpanOf(phraseTokens),
			Text:        joinedSourceText(phraseTokens),
			Cardinality: cardinality,
			Selector:    compileSelector(selectorTokens, atoms),
		})
	}
	return targets
}

func targetPhraseEnd(tokens []shared.Token, start int) int {
	const spellOrAbility = "spell, activated ability, or triggered ability"
	for end := start + 1; end <= len(tokens); end++ {
		if joinedSourceText(tokens[start:end]) == spellOrAbility {
			return end
		}
	}
	end := start
	for end < len(tokens) {
		token := tokens[end]
		if token.Kind == shared.Comma || token.Kind == shared.Period || token.Kind == shared.Semicolon ||
			(equalWord(token, "and") && end+2 < len(tokens) && equalWord(tokens[end+1], "you") && isEffectVerb(tokens[end+2])) ||
			(equalWord(token, "and") && end+1 < len(tokens) && isEffectVerb(tokens[end+1])) ||
			(end > start && isEffectVerb(token)) ||
			(equalWord(token, "until") && end+1 < len(tokens) && equalWord(tokens[end+1], "end")) ||
			(equalWord(token, "until") && end+2 < len(tokens) && equalWord(tokens[end+1], "your") && equalWord(tokens[end+2], "next")) ||
			// "for as long as" marks a source-tied duration, not part of the target.
			(equalWord(token, "for") && end+3 < len(tokens) &&
				equalWord(tokens[end+1], "as") && equalWord(tokens[end+2], "long") && equalWord(tokens[end+3], "as")) ||
			// "as long as this" marks a source-on-battlefield duration.
			(equalWord(token, "as") && end+3 < len(tokens) &&
				equalWord(tokens[end+1], "long") && equalWord(tokens[end+2], "as") && equalWord(tokens[end+3], "this")) {
			break
		}
		end++
	}
	return end
}

func compileSelector(tokens []shared.Token, atoms parser.Atoms) CompiledSelector {
	selector := CompiledSelector{Raw: joinedSourceText(tokens)}
	span := shared.SpanOf(tokens)
	if selector.Raw == "activated ability" {
		selector.Kind = SelectorActivatedAbility
		return selector
	}
	if selector.Raw == "triggered ability" {
		selector.Kind = SelectorTriggeredAbility
		return selector
	}
	if selector.Raw == "activated or triggered ability" {
		selector.Kind = SelectorActivatedOrTriggeredAbility
		return selector
	}
	if selector.Raw == "spell, activated ability, or triggered ability" {
		selector.Kind = SelectorSpellActivatedOrTriggeredAbility
		return selector
	}
	for _, token := range tokens {
		noun, ok := atoms.ObjectNounAt(token.Span)
		if !ok {
			continue
		}
		if kind := selectorKindFromAtom(noun); kind != SelectorUnknown && selector.Kind == SelectorUnknown {
			selector.Kind = kind
		}
	}
	var requiredTypes []types.Card
	for _, token := range tokens {
		_, excludedCardType := atoms.ExcludedCardTypeAt(token.Span)
		if cardType, ok := atoms.CardTypeAt(token.Span); ok && !excludedCardType {
			if runtimeType, typeOK := runtimeCardTypeFromParser(cardType); typeOK && !slices.Contains(requiredTypes, runtimeType) {
				requiredTypes = append(requiredTypes, runtimeType)
			}
			if cardType == parser.CardTypeBattle && selector.Kind == SelectorUnknown {
				selector.Kind = SelectorBattle
			}
		}
		if cardType, ok := atoms.ExcludedCardTypeAt(token.Span); ok {
			if runtimeType, typeOK := runtimeCardTypeFromParser(cardType); typeOK && !slices.Contains(selector.ExcludedTypes(), runtimeType) {
				appendSelectorExcludedType(&selector, runtimeType)
			}
		}
		if colorValue, ok := atoms.ColorAt(token.Span); ok {
			if runtimeColor, colorOK := runtimeColorFromParser(colorValue); colorOK && !slices.Contains(selector.ColorsAny(), runtimeColor) {
				appendSelectorColorAny(&selector, runtimeColor)
			}
		}
		if colorValue, ok := atoms.ExcludedColorAt(token.Span); ok {
			if runtimeColor, colorOK := runtimeColorFromParser(colorValue); colorOK && !slices.Contains(selector.ExcludedColors(), runtimeColor) {
				appendSelectorExcludedColor(&selector, runtimeColor)
			}
		}
	}
	if len(requiredTypes) > 1 {
		setSelectorRequiredTypesAny(&selector, requiredTypes)
	}
	if shared.ContainsWord(shared.NormalizedWords(tokens), "any") && selector.Kind == SelectorUnknown {
		selector.Kind = SelectorAny
	}
	appendSelectorSubtypesAny(&selector, atoms.SubtypesIn(span)...)
	if relation, ok := atoms.ControllerIn(span); ok {
		switch relation {
		case parser.ControllerRelationYouDontControl:
			selector.Controller = ControllerNotYou
		case parser.ControllerRelationYouControl:
			selector.Controller = ControllerYou
		case parser.ControllerRelationOpponentControls:
			selector.Controller = ControllerOpponent
		default:
		}
	}
	if selector.Controller == ControllerAny {
		for _, token := range tokens {
			if noun, ok := atoms.ObjectNounAt(token.Span); ok && noun == parser.ObjectNounOpponent {
				selector.Controller = ControllerOpponent
			}
		}
	}
	selector.Another = atoms.SelectionFlagIn(span, parser.SelectionFlagAnother)
	selector.Other = atoms.SelectionFlagIn(span, parser.SelectionFlagOther)
	selector.Attacking = atoms.SelectionFlagIn(span, parser.SelectionFlagAttacking)
	selector.Blocking = atoms.SelectionFlagIn(span, parser.SelectionFlagBlocking)
	selector.Tapped = atoms.SelectionFlagIn(span, parser.SelectionFlagTapped)
	selector.Untapped = atoms.SelectionFlagIn(span, parser.SelectionFlagUntapped)
	selector.Keyword = selectorKeyword(tokens, atoms)
	compileSelectorNumberFilters(tokens, atoms, &selector)
	return selector
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

func compileSelectorNumberFilters(tokens []shared.Token, atoms parser.Atoms, selector *CompiledSelector) {
	for i := range len(tokens) {
		switch {
		case i+2 < len(tokens) && equalWord(tokens[i], "mana") && equalWord(tokens[i+1], "value"):
			if comparison, ok := selectorNumberComparison(tokens[i+2:], atoms); ok {
				selector.ManaValue = comparison
				selector.MatchManaValue = true
			}
		case equalWord(tokens[i], "power"):
			if comparison, ok := selectorNumberComparison(tokens[i+1:], atoms); ok {
				selector.Power = comparison
				selector.MatchPower = true
			}
		case equalWord(tokens[i], "toughness"):
			if comparison, ok := selectorNumberComparison(tokens[i+1:], atoms); ok {
				selector.Toughness = comparison
				selector.MatchToughness = true
			}
		default:
		}
	}
}

func selectorNumberComparison(tokens []shared.Token, atoms parser.Atoms) (compare.Int, bool) {
	if len(tokens) == 0 {
		return compare.Int{}, false
	}
	if value, ok := selectorNumberValue(tokens[0], atoms); ok {
		if len(tokens) >= 3 && equalWord(tokens[1], "or") {
			switch {
			case equalWord(tokens[2], "less"):
				return compare.Int{Op: compare.LessOrEqual, Value: value}, true
			case equalWord(tokens[2], "greater"):
				return compare.Int{Op: compare.GreaterOrEqual, Value: value}, true
			default:
				return compare.Int{}, false
			}
		}
		return compare.Int{Op: compare.Equal, Value: value}, true
	}
	if len(tokens) >= 3 && equalWord(tokens[0], "equal") && equalWord(tokens[1], "to") {
		if value, ok := selectorNumberValue(tokens[2], atoms); ok {
			return compare.Int{Op: compare.Equal, Value: value}, true
		}
	}
	return compare.Int{}, false
}

func selectorNumberValue(token shared.Token, atoms parser.Atoms) (int, bool) {
	if token.Kind == shared.Integer {
		value, err := strconv.Atoi(token.Text)
		return value, err == nil
	}
	value, ok := atoms.CardinalAt(token.Span)
	return value, ok
}

func selectorKindFromAtom(noun parser.ObjectNoun) SelectorKind {
	switch noun {
	case parser.ObjectNounArtifact:
		return SelectorArtifact
	case parser.ObjectNounCard:
		return SelectorCard
	case parser.ObjectNounCreature:
		return SelectorCreature
	case parser.ObjectNounEnchantment:
		return SelectorEnchantment
	case parser.ObjectNounLand:
		return SelectorLand
	case parser.ObjectNounOpponent:
		return SelectorOpponent
	case parser.ObjectNounPermanent:
		return SelectorPermanent
	case parser.ObjectNounPlaneswalker:
		return SelectorPlaneswalker
	case parser.ObjectNounPlayer:
		return SelectorPlayer
	case parser.ObjectNounSpell:
		return SelectorSpell
	default:
		return SelectorUnknown
	}
}

func selectorKeyword(tokens []shared.Token, atoms parser.Atoms) parser.KeywordKind {
	span := shared.SpanOf(tokens)
	selector, ok := atoms.KeywordSelectorIn(span, false)
	if !ok ||
		selector.Form == parser.KeywordSelectorFormUnknown ||
		selector.Keyword != parser.KeywordCycling {
		return parser.KeywordUnknown
	}
	return selector.Keyword
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

func compileEffects(
	sentences []parser.Sentence,
	reminders, quoted []parser.Delimited,
	atoms parser.Atoms,
) []CompiledEffect {
	var effects []CompiledEffect
	for _, sentence := range sentences {
		if sentence.StaticRule != nil {
			if effect, ok := compileStaticRuleEffect(sentence); ok {
				effects = append(effects, effect)
			}
			continue
		}
		tokens := semanticTokens(sentence.Tokens, reminders, quoted)
		duration := compileDuration(tokens, atoms)
		staticSubject := compileStaticSubject(tokens, atoms)
		effectIndices := effectTokenIndices(tokens, atoms)
		for effectIndex, tokenIndex := range effectIndices {
			token := tokens[tokenIndex]
			kind := effectKindAt(tokens, tokenIndex)
			clauseEnd := effectClauseEnd(tokens, effectIndices, effectIndex)
			clauseTokens := tokens[tokenIndex+1 : clauseEnd]
			clauseTokens, delayedTiming := stripDelayedTimingSuffix(clauseTokens)
			powerDelta, toughnessDelta := compilePTChange(clauseTokens)
			counterKind, counterKindKnown := counterKindWord(clauseTokens, atoms)
			if !counterKindKnown && kind == EffectReturn {
				counterKind, _, counterKindKnown = atoms.CounterIn(shared.SpanOf(clauseTokens))
			}
			effects = append(effects, CompiledEffect{
				Kind:              kind,
				Span:              sentence.Span,
				Text:              sentence.Text,
				VerbSpan:          token.Span,
				Duration:          duration,
				DelayedTiming:     delayedTiming,
				Selector:          compileSelector(clauseTokens, atoms),
				Amount:            compileEffectAmount(clauseTokens, atoms),
				PowerDelta:        powerDelta,
				ToughnessDelta:    toughnessDelta,
				StaticSubject:     staticSubject.kind,
				StaticSubjectSpan: staticSubject.span,
				Details:           compiledEffectDetails(staticSubjectType(staticSubject.subtype, staticSubject.sub, staticSubject.subKnown), firstSymbol(clauseTokens)),
				CounterKind:       counterKind,
				CounterKindKnown:  counterKindKnown,
				FromZone:          compileFromZone(clauseTokens, atoms),
				ToZone:            compileToZone(clauseTokens, atoms),
				Negated:           effectNegated(tokens, tokenIndex),
			})
		}

	}
	return effects
}

func stripDelayedTimingSuffix(tokens []shared.Token) ([]shared.Token, game.DelayedTriggerTiming) {
	end := len(tokens)
	if end > 0 && tokens[end-1].Kind == shared.Period {
		end--
	}
	suffixes := []struct {
		timing game.DelayedTriggerTiming
		text   []string
	}{
		{
			timing: game.DelayedAtBeginningOfNextEndStep,
			text:   []string{"at", "the", "beginning", "of", "the", "next", "end", "step"},
		},
		{
			timing: game.DelayedAtBeginningOfNextUpkeep,
			text:   []string{"at", "the", "beginning", "of", "the", "next", "turn's", "upkeep"},
		},
	}
	for _, suffix := range suffixes {
		start := end - len(suffix.text)
		if start < 0 || !tokenTextsEqual(tokens[start:end], suffix.text) {
			continue
		}
		stripped := make([]shared.Token, 0, len(tokens)-len(suffix.text))
		stripped = append(stripped, tokens[:start]...)
		stripped = append(stripped, tokens[end:]...)
		return stripped, suffix.timing
	}
	return tokens, 0
}

func tokenTextsEqual(tokens []shared.Token, text []string) bool {
	if len(tokens) != len(text) {
		return false
	}
	for i := range tokens {
		if !strings.EqualFold(tokens[i].Text, text[i]) {
			return false
		}
	}
	return true
}

func compileFromZone(tokens []shared.Token, atoms parser.Atoms) zone.Type {
	for i := 0; i+2 < len(tokens); i++ {
		if !equalWord(tokens[i], "from") || !legacyGraveyardZonePhrase(tokens[i+1:]) {
			continue
		}
		if z, ok := atoms.ZoneIn(tokens[i].Span, parser.ZoneRoleFrom); ok {
			return z
		}
	}
	return zone.None
}

func compileToZone(tokens []shared.Token, atoms parser.Atoms) zone.Type {
	for i := range len(tokens) {
		switch {
		case equalWord(tokens[i], "to") && i+2 < len(tokens) && legacyHandZonePhrase(tokens[i+1:]):
			if z, ok := atoms.ZoneIn(tokens[i].Span, parser.ZoneRoleTo); ok {
				return z
			}
		case equalWord(tokens[i], "to") && i+2 < len(tokens) && legacyBattlefieldZonePhrase(tokens[i+1:]):
			if z, ok := atoms.ZoneIn(tokens[i].Span, parser.ZoneRoleTo); ok {
				return z
			}
		case equalWord(tokens[i], "onto") && i+2 < len(tokens) && legacyBattlefieldZonePhrase(tokens[i+1:]):
			if z, ok := atoms.ZoneIn(tokens[i].Span, parser.ZoneRoleTo); ok {
				return z
			}
		case equalWord(tokens[i], "on") && i+4 < len(tokens) &&
			(equalWord(tokens[i+1], "top") || equalWord(tokens[i+1], "bottom")) &&
			equalWord(tokens[i+2], "of") &&
			legacyLibraryZonePhrase(tokens[i+3:]):
			if z, ok := atoms.ZoneIn(tokens[i].Span, parser.ZoneRoleTo); ok {
				return z
			}
		case equalWord(tokens[i], "on") && i+5 < len(tokens) &&
			equalWord(tokens[i+1], "the") &&
			(equalWord(tokens[i+2], "top") || equalWord(tokens[i+2], "bottom")) &&
			equalWord(tokens[i+3], "of") &&
			legacyLibraryZonePhrase(tokens[i+4:]):
			if z, ok := atoms.ZoneIn(tokens[i].Span, parser.ZoneRoleTo); ok {
				return z
			}
		case equalWord(tokens[i], "into") && i+2 < len(tokens) && legacyLibraryZonePhrase(tokens[i+1:]):
			if z, ok := atoms.ZoneIn(tokens[i].Span, parser.ZoneRoleTo); ok {
				return z
			}
		default:
		}
	}
	return zone.None
}

func legacyGraveyardZonePhrase(tokens []shared.Token) bool {
	switch {
	case len(tokens) >= 2 &&
		(equalWord(tokens[0], "your") || equalWord(tokens[0], "a") || equalWord(tokens[0], "an")) &&
		equalWord(tokens[1], "graveyard"):
		return true
	case len(tokens) >= 3 &&
		equalWord(tokens[0], "an") &&
		strings.EqualFold(tokens[1].Text, "opponent's") &&
		equalWord(tokens[2], "graveyard"):
		return true
	default:
		return false
	}
}

func legacyHandZonePhrase(tokens []shared.Token) bool {
	return len(tokens) >= 2 &&
		(equalWord(tokens[0], "your") || equalWord(tokens[0], "their")) &&
		equalWord(tokens[1], "hand")
}

func legacyBattlefieldZonePhrase(tokens []shared.Token) bool {
	return len(tokens) >= 2 && equalWord(tokens[0], "the") && equalWord(tokens[1], "battlefield")
}

func legacyLibraryZonePhrase(tokens []shared.Token) bool {
	return len(tokens) >= 2 && equalWord(tokens[0], "your") && equalWord(tokens[1], "library")
}

func counterKindWord(tokens []shared.Token, atoms parser.Atoms) (counter.Kind, bool) {
	counterIndex := -1
	for index, token := range tokens {
		if equalWord(token, "counter") || equalWord(token, "counters") {
			if counterIndex >= 0 {
				return 0, false
			}
			counterIndex = index
		}
	}
	if counterIndex <= 0 {
		return 0, false
	}
	kind, nameSpan, ok := atoms.CounterIn(shared.SpanOf(tokens))
	if !ok {
		return 0, false
	}
	nameStart := counterIndex
	for nameStart > 0 && tokens[nameStart-1].Span.Start.Offset >= nameSpan.Start.Offset {
		nameStart--
	}
	amountTokens := tokens[:nameStart]
	if len(amountTokens) > 0 && equalWord(amountTokens[0], "with") {
		amountTokens = amountTokens[1:]
	}
	if len(amountTokens) != 1 {
		return 0, false
	}
	amount := amountTokens[0]
	switch {
	case equalWord(amount, "a"), equalWord(amount, "an"), equalWord(amount, "x"):
		return kind, CounterKindPlacementSupported(kind)
	}
	if value, ok := atoms.CardinalAt(amount.Span); ok && value <= 4 {
		return kind, CounterKindPlacementSupported(kind)
	}
	if amount.Kind == shared.Integer {
		if value, err := strconv.Atoi(amount.Text); err == nil && value > 0 {
			return kind, CounterKindPlacementSupported(kind)
		}
	}
	return 0, false
}

func effectTokenIndices(tokens []shared.Token, atoms parser.Atoms) []int {
	var indices []int
	for index := range tokens {
		if effectKindAt(tokens, index) != EffectUnknown &&
			!atoms.SelfNameAt(tokens[index].Span) {
			indices = append(indices, index)
		}
	}
	return indices
}

func effectClauseEnd(tokens []shared.Token, effectIndices []int, effectIndex int) int {
	start := effectIndices[effectIndex] + 1
	end := len(tokens)
	for _, nextEffect := range effectIndices[effectIndex+1:] {
		if coordination := effectCoordinationStart(tokens, start, nextEffect); coordination >= 0 {
			end = coordination
			break
		}
	}
	for index := start; index < end; index++ {
		if conditionKindAt(tokens, index) != ConditionUnknown {
			return index
		}
	}
	return end
}

func effectCoordinationStart(tokens []shared.Token, start, effectIndex int) int {
	for index := effectIndex - 1; index >= start; index-- {
		if tokens[index].Kind == shared.Comma || tokens[index].Kind == shared.Semicolon {
			return -1
		}
		if equalWord(tokens[index], "then") || equalWord(tokens[index], "and") {
			return index
		}
	}
	return -1
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

type compiledStaticSubject struct {
	kind     StaticSubjectKind
	span     shared.Span
	subtype  string
	sub      types.Sub
	subKnown bool
}

func compileStaticSubject(tokens []shared.Token, atoms parser.Atoms) compiledStaticSubject {
	subtypeAt := func(index int) (types.Sub, bool) {
		if index < len(tokens) {
			if sub, ok := atoms.SubtypeAt(tokens[index].Span); ok {
				if parser.SubtypeMatchesAnyRuntimeCardType(sub, []types.Card{types.Creature, types.Kindred}) {
					return sub, true
				}
			}
		}
		return "", false
	}
	switch {
	case len(tokens) >= 3 &&
		(equalWord(tokens[0], "enchanted") || equalWord(tokens[0], "equipped")) &&
		equalWord(tokens[1], "creature") &&
		(equalWord(tokens[2], "gets") || equalWord(tokens[2], "has")):
		return compiledStaticSubject{kind: StaticSubjectAttachedObject, span: shared.SpanOf(tokens[:2])}
	case len(tokens) >= 5 &&
		equalWord(tokens[0], "other") &&
		equalWord(tokens[1], "creatures") &&
		equalWord(tokens[2], "you") &&
		equalWord(tokens[3], "control") &&
		(equalWord(tokens[4], "get") || equalWord(tokens[4], "have")):
		return compiledStaticSubject{kind: StaticSubjectOtherControlledCreatures, span: shared.SpanOf(tokens[:4])}
	case len(tokens) >= 4 &&
		equalWord(tokens[0], "creatures") &&
		equalWord(tokens[1], "you") &&
		equalWord(tokens[2], "control") &&
		(equalWord(tokens[3], "get") || equalWord(tokens[3], "have")):
		return compiledStaticSubject{kind: StaticSubjectControlledCreatures, span: shared.SpanOf(tokens[:3])}
	case len(tokens) >= 6 &&
		equalWord(tokens[0], "creatures") &&
		equalWord(tokens[1], "your") &&
		equalWord(tokens[2], "opponents") &&
		equalWord(tokens[3], "control") &&
		(equalWord(tokens[4], "get") || equalWord(tokens[4], "have")):
		return compiledStaticSubject{kind: StaticSubjectOpponentControlledCreatures, span: shared.SpanOf(tokens[:4])}
	case len(tokens) >= 5 &&
		equalWord(tokens[0], "each") &&
		equalWord(tokens[1], "wall") &&
		equalWord(tokens[2], "you") &&
		equalWord(tokens[3], "control") &&
		(equalWord(tokens[4], "gets") || equalWord(tokens[4], "has")):
		return compiledStaticSubject{kind: StaticSubjectControlledWalls, span: shared.SpanOf(tokens[:4]), sub: types.Wall, subKnown: true}
	case len(tokens) >= 4 &&
		equalWord(tokens[0], "walls") &&
		equalWord(tokens[1], "you") &&
		equalWord(tokens[2], "control") &&
		(equalWord(tokens[3], "get") || equalWord(tokens[3], "have")):
		return compiledStaticSubject{kind: StaticSubjectControlledWalls, span: shared.SpanOf(tokens[:3]), sub: types.Wall, subKnown: true}
	case len(tokens) >= 4 &&
		equalWord(tokens[0], "artifacts") &&
		equalWord(tokens[1], "you") &&
		equalWord(tokens[2], "control") &&
		(equalWord(tokens[3], "get") || equalWord(tokens[3], "have")):
		return compiledStaticSubject{kind: StaticSubjectControlledArtifacts, span: shared.SpanOf(tokens[:3])}
	case len(tokens) >= 4 &&
		equalWord(tokens[0], "tokens") &&
		equalWord(tokens[1], "you") &&
		equalWord(tokens[2], "control") &&
		(equalWord(tokens[3], "get") || equalWord(tokens[3], "have")):
		return compiledStaticSubject{kind: StaticSubjectControlledTokens, span: shared.SpanOf(tokens[:3])}
	case len(tokens) >= 5 &&
		equalWord(tokens[0], "other") &&
		tokens[1].Kind == shared.Word &&
		equalWord(tokens[2], "you") &&
		equalWord(tokens[3], "control") &&
		equalWord(tokens[4], "have"):
		sub, ok := subtypeAt(1)
		return compiledStaticSubject{
			kind:     StaticSubjectOtherControlledCreatureSubtype,
			span:     shared.SpanOf(tokens[:4]),
			subtype:  tokens[1].Text,
			sub:      sub,
			subKnown: ok,
		}
	case len(tokens) >= 4 &&
		tokens[0].Kind == shared.Word &&
		equalWord(tokens[1], "you") &&
		equalWord(tokens[2], "control") &&
		equalWord(tokens[3], "have"):
		sub, ok := subtypeAt(0)
		return compiledStaticSubject{
			kind:     StaticSubjectControlledCreatureSubtype,
			span:     shared.SpanOf(tokens[:3]),
			subtype:  tokens[0].Text,
			sub:      sub,
			subKnown: ok,
		}
	default:
		return compiledStaticSubject{}
	}
}

func compilePTChange(tokens []shared.Token) (power, toughness CompiledSignedAmount) {
	for i := 0; i+4 < len(tokens); i++ {
		power, powerOK := signedAmount(tokens[i], tokens[i+1])
		toughness, toughnessOK := signedAmount(tokens[i+3], tokens[i+4])
		if powerOK && tokens[i+2].Kind == shared.Slash && toughnessOK {
			return power, toughness
		}
	}
	return CompiledSignedAmount{}, CompiledSignedAmount{}
}

func signedAmount(sign, amount shared.Token) (CompiledSignedAmount, bool) {
	if amount.Kind != shared.Integer || (sign.Kind != shared.Plus && sign.Kind != shared.Minus) {
		return CompiledSignedAmount{}, false
	}
	value, err := strconv.Atoi(amount.Text)
	if err != nil {
		return CompiledSignedAmount{}, false
	}
	negative := sign.Kind == shared.Minus
	return CompiledSignedAmount{Value: value, Known: true, Negative: negative}, true
}

func compileEffectAmount(tokens []shared.Token, atoms parser.Atoms) CompiledAmount {
	dynamic := compileDynamicEffectAmount(tokens, atoms)
	if dynamic.matched {
		return dynamic.amount
	}
	if dynamic.attempted {
		return CompiledAmount{}
	}
	for _, token := range tokens {
		if value := numberWord(token, atoms); value > 0 {
			return CompiledAmount{Value: value, Known: true}
		}
		if equalWord(token, "a") || equalWord(token, "an") {
			return CompiledAmount{Value: 1, Known: true}
		}
	}
	for _, token := range tokens {
		if token.Kind == shared.Symbol {
			return CompiledAmount{Value: 1, Known: true}
		}
	}
	return CompiledAmount{}
}

type compiledDynamicAmount struct {
	amount    CompiledAmount
	matched   bool
	attempted bool
}

func compileDynamicEffectAmount(tokens []shared.Token, atoms parser.Atoms) compiledDynamicAmount {
	var matches []CompiledAmount
	attempted := false
	for i := range tokens {
		prefix, ok := dynamicAmountPrefix(tokens, i, atoms)
		if !ok {
			continue
		}
		attempted = true
		if subject, matched := dynamicAmountSubject(tokens, prefix.subjectStart, atoms); matched {
			if !prefix.allows(subject) {
				continue
			}
			amount := subject.amount
			amount.DynamicForm = prefix.form
			if amount.Multiplier == 0 {
				amount.Multiplier = prefix.multiplier
			}
			amount.Text = joinedSourceText(tokens[i:subject.end])
			matches = append(matches, amount)
		}
	}
	if len(matches) != 1 {
		return compiledDynamicAmount{attempted: attempted}
	}
	return compiledDynamicAmount{
		amount:    matches[0],
		matched:   true,
		attempted: true,
	}
}

type dynamicAmountPrefixMatch struct {
	form          DynamicAmountForm
	subjectStart  int
	multiplier    int
	subjectClass  dynamicAmountSubjectClass
	subjectNumber dynamicSubjectNumber
}

type dynamicAmountSubjectClass uint8

const (
	dynamicAmountCountSubject dynamicAmountSubjectClass = iota
	dynamicAmountValueSubject
)

type dynamicSubjectNumber uint8

const (
	dynamicSubjectNumberNone dynamicSubjectNumber = iota
	dynamicSubjectSingular
	dynamicSubjectPlural
	dynamicSubjectInvariant
)

type dynamicAmountSubjectMatch struct {
	amount CompiledAmount
	end    int
	number dynamicSubjectNumber
}

func (m dynamicAmountPrefixMatch) allows(subject dynamicAmountSubjectMatch) bool {
	switch m.subjectClass {
	case dynamicAmountCountSubject:
		if subject.amount.DynamicKind != DynamicAmountCount &&
			subject.amount.DynamicKind != DynamicAmountOpponentCount &&
			subject.amount.DynamicKind != DynamicAmountBasicLandTypes {
			return false
		}
		return subject.number == dynamicSubjectInvariant || subject.number == m.subjectNumber
	case dynamicAmountValueSubject:
		return subject.number == dynamicSubjectNumberNone &&
			(subject.amount.DynamicKind == DynamicAmountControllerLife ||
				subject.amount.DynamicKind == DynamicAmountSourcePower)
	default:
		return false
	}
}

func dynamicAmountPrefix(tokens []shared.Token, index int, atoms parser.Atoms) (dynamicAmountPrefixMatch, bool) {
	switch {
	case wordsAt(tokens, index, "equal", "to", "twice", "the", "number", "of"):
		return dynamicAmountPrefixMatch{DynamicAmountEqual, index + 6, 2, dynamicAmountCountSubject, dynamicSubjectPlural}, true
	case wordsAt(tokens, index, "equal", "to", "the", "number", "of"):
		return dynamicAmountPrefixMatch{DynamicAmountEqual, index + 5, 1, dynamicAmountCountSubject, dynamicSubjectPlural}, true
	case wordsAt(tokens, index, "for", "each"):
		multiplier := precedingAmountMultiplier(tokens[:index], atoms)
		return dynamicAmountPrefixMatch{DynamicAmountForEach, index + 2, multiplier, dynamicAmountCountSubject, dynamicSubjectSingular}, multiplier > 0
	case wordsAt(tokens, index, "equal", "to"):
		return dynamicAmountPrefixMatch{DynamicAmountEqual, index + 2, 1, dynamicAmountValueSubject, dynamicSubjectNumberNone}, true
	case wordsAt(tokens, index, "where", "X", "is", "twice", "the", "number", "of"):
		return dynamicAmountPrefixMatch{DynamicAmountWhereX, index + 7, 2, dynamicAmountCountSubject, dynamicSubjectPlural}, true
	case wordsAt(tokens, index, "where", "X", "is", "the", "number", "of"):
		return dynamicAmountPrefixMatch{DynamicAmountWhereX, index + 6, 1, dynamicAmountCountSubject, dynamicSubjectPlural}, true
	case wordsAt(tokens, index, "where", "X", "is"):
		return dynamicAmountPrefixMatch{DynamicAmountWhereX, index + 3, 1, dynamicAmountValueSubject, dynamicSubjectNumberNone}, true
	default:
		return dynamicAmountPrefixMatch{}, false
	}
}

type dynamicCountNoun struct {
	singular  string
	plural    string
	invariant string
	kind      DynamicAmountKind
	selector  SelectorKind
}

var dynamicCountNouns = []dynamicCountNoun{
	{singular: "creature", plural: "creatures", kind: DynamicAmountCount, selector: SelectorCreature},
	{singular: "artifact", plural: "artifacts", kind: DynamicAmountCount, selector: SelectorArtifact},
	{singular: "enchantment", plural: "enchantments", kind: DynamicAmountCount, selector: SelectorEnchantment},
	{singular: "land", plural: "lands", kind: DynamicAmountCount, selector: SelectorLand},
	{singular: "permanent", plural: "permanents", kind: DynamicAmountCount, selector: SelectorPermanent},
	{singular: "opponent", plural: "opponents", kind: DynamicAmountOpponentCount},
}

func (n dynamicCountNoun) numberAt(tokens []shared.Token, start int) (dynamicSubjectNumber, bool) {
	switch {
	case equalWord(tokens[start], n.singular):
		return dynamicSubjectSingular, true
	case equalWord(tokens[start], n.plural):
		return dynamicSubjectPlural, true
	case n.invariant != "" && equalWord(tokens[start], n.invariant):
		return dynamicSubjectInvariant, true
	default:
		return dynamicSubjectNumberNone, false
	}
}

func dynamicAmountSubject(tokens []shared.Token, start int, atoms parser.Atoms) (dynamicAmountSubjectMatch, bool) {
	if start >= len(tokens) {
		return dynamicAmountSubjectMatch{}, false
	}
	if subject, ok := dynamicBasicLandTypeAmountSubject(tokens, start); ok {
		return subject, true
	}
	if subject, ok := dynamicCountAmountSubject(tokens, start, atoms); ok {
		return subject, true
	}

	switch {
	case wordsAt(tokens, start, "your", "life", "total") &&
		dynamicSubjectBoundary(tokens, start+3):
		return dynamicAmountSubjectMatch{
			amount: CompiledAmount{DynamicKind: DynamicAmountControllerLife},
			end:    start + 3,
		}, true
	case wordsAt(tokens, start, "its", "power") &&
		dynamicSubjectBoundary(tokens, start+2):
		return dynamicAmountSubjectMatch{
			amount: CompiledAmount{
				DynamicKind:   DynamicAmountSourcePower,
				ReferenceSpan: tokens[start].Span,
			},
			end: start + 2,
		}, true
	case wordsAt(tokens, start, "this", "creature") &&
		start+4 < len(tokens) &&
		tokens[start+2].Kind == shared.Apostrophe &&
		equalWord(tokens[start+3], "s") &&
		equalWord(tokens[start+4], "power") &&
		dynamicSubjectBoundary(tokens, start+5):
		return dynamicAmountSubjectMatch{
			amount: CompiledAmount{
				DynamicKind:   DynamicAmountSourcePower,
				ReferenceSpan: shared.SpanOf(tokens[start : start+2]),
			},
			end: start + 5,
		}, true
	case start+2 < len(tokens) &&
		equalWord(tokens[start], "this") &&
		strings.EqualFold(tokens[start+1].Text, "creature's") &&
		equalWord(tokens[start+2], "power") &&
		dynamicSubjectBoundary(tokens, start+3):
		return dynamicAmountSubjectMatch{
			amount: CompiledAmount{
				DynamicKind:   DynamicAmountSourcePower,
				ReferenceSpan: shared.SpanOf(tokens[start : start+2]),
			},
			end: start + 3,
		}, true
	default:
		// The card's own name as a power source is recognized from the
		// parser-emitted self-name span, not from name spelling.
		nameSpan, ok := atoms.SelfNameSpanStartingAt(tokens[start].Span)
		if !ok {
			return dynamicAmountSubjectMatch{}, false
		}
		end := start
		for end < len(tokens) && tokens[end].Span.End.Offset <= nameSpan.End.Offset {
			end++
		}
		// Possessive fused into the name's final token ("[Name]'s power").
		if end < len(tokens) &&
			equalWord(tokens[end], "power") &&
			dynamicSubjectBoundary(tokens, end+1) {
			return dynamicAmountSubjectMatch{
				amount: CompiledAmount{
					DynamicKind:   DynamicAmountSourcePower,
					ReferenceSpan: shared.SpanOf(tokens[start:end]),
				},
				end: end + 1,
			}, true
		}
		// Plain name followed by a separate possessive marker.
		if end+2 < len(tokens) &&
			tokens[end].Kind == shared.Apostrophe &&
			equalWord(tokens[end+1], "s") &&
			equalWord(tokens[end+2], "power") &&
			dynamicSubjectBoundary(tokens, end+3) {
			return dynamicAmountSubjectMatch{
				amount: CompiledAmount{
					DynamicKind:   DynamicAmountSourcePower,
					ReferenceSpan: shared.SpanOf(tokens[start:end]),
				},
				end: end + 3,
			}, true
		}
		return dynamicAmountSubjectMatch{}, false
	}
}

func dynamicBasicLandTypeAmountSubject(tokens []shared.Token, start int) (dynamicAmountSubjectMatch, bool) {
	var number dynamicSubjectNumber
	var end int
	switch {
	case wordsAt(tokens, start, "basic", "land", "type", "among", "lands", "you", "control"):
		number = dynamicSubjectSingular
		end = start + 7
	case wordsAt(tokens, start, "basic", "land", "types", "among", "lands", "you", "control"):
		number = dynamicSubjectPlural
		end = start + 7
	default:
		return dynamicAmountSubjectMatch{}, false
	}
	if !dynamicSubjectBoundary(tokens, end) {
		return dynamicAmountSubjectMatch{}, false
	}
	return dynamicAmountSubjectMatch{
		amount: CompiledAmount{DynamicKind: DynamicAmountBasicLandTypes},
		end:    end,
		number: number,
	}, true
}

func dynamicCountAmountSubject(tokens []shared.Token, start int, atoms parser.Atoms) (dynamicAmountSubjectMatch, bool) {
	if subject, ok := dynamicCardCountAmountSubject(tokens, start, atoms); ok {
		return subject, true
	}
	suffixes := []struct {
		words      []string
		controller ControllerKind
	}{
		{words: []string{"you", "control"}, controller: ControllerYou},
		{words: []string{"your", "opponents", "control"}, controller: ControllerOpponent},
		{words: []string{"on", "the", "battlefield"}, controller: ControllerAny},
	}
	for _, noun := range dynamicCountNouns {
		number, ok := noun.numberAt(tokens, start)
		if !ok {
			continue
		}
		end := start + 1
		if noun.kind == DynamicAmountOpponentCount {
			if wordsAt(tokens, end, "you", "have") {
				end += 2
			}
			if dynamicSubjectBoundary(tokens, end) {
				return dynamicAmountSubjectMatch{
					amount: CompiledAmount{DynamicKind: noun.kind},
					end:    end,
					number: number,
				}, true
			}
			continue
		}
		for _, suffix := range suffixes {
			subjectEnd := end + len(suffix.words)
			if !wordsAt(tokens, end, suffix.words...) ||
				!dynamicSubjectBoundary(tokens, subjectEnd) {
				continue
			}
			return dynamicAmountSubjectMatch{
				amount: CompiledAmount{
					DynamicKind: noun.kind,
					selector: &CompiledSelector{
						Kind:       noun.selector,
						Controller: suffix.controller,
						Raw:        joinedSourceText(tokens[start:subjectEnd]),
					},
				},
				end:    subjectEnd,
				number: number,
			}, true
		}
	}
	return dynamicAmountSubjectMatch{}, false
}

func dynamicCardCountAmountSubject(tokens []shared.Token, start int, atoms parser.Atoms) (dynamicAmountSubjectMatch, bool) {
	if start >= len(tokens) ||
		(!equalWord(tokens[start], "card") && !equalWord(tokens[start], "cards")) {
		return dynamicAmountSubjectMatch{}, false
	}
	number := dynamicSubjectPlural
	if equalWord(tokens[start], "card") {
		number = dynamicSubjectSingular
	}
	end := start + 1
	selector := CompiledSelector{
		Kind: SelectorCard,
		Raw:  joinedSourceText(tokens[start:]),
	}
	if end >= len(tokens) {
		return dynamicAmountSubjectMatch{}, false
	}
	keywordSelector, ok := atoms.KeywordSelectorStartingAt(tokens[end].Span)
	if !ok ||
		keywordSelector.Excluded ||
		keywordSelector.Form == parser.KeywordSelectorFormUnknown ||
		keywordSelector.Keyword != parser.KeywordCycling {
		return dynamicAmountSubjectMatch{}, false
	}
	for end < len(tokens) && tokens[end].Span.End.Offset <= keywordSelector.Span.End.Offset {
		end++
	}
	selector.Keyword = keywordSelector.Keyword
	switch {
	case wordsAt(tokens, end, "in", "your", "graveyard"):
		selector.Controller = ControllerYou
		selector.Zone = zone.Graveyard
		end += 3
	default:
		return dynamicAmountSubjectMatch{}, false
	}
	if !dynamicSubjectBoundary(tokens, end) {
		return dynamicAmountSubjectMatch{}, false
	}
	selector.Raw = joinedSourceText(tokens[start:end])
	return dynamicAmountSubjectMatch{
		amount: CompiledAmount{
			DynamicKind: DynamicAmountCount,
			selector:    &selector,
		},
		end:    end,
		number: number,
	}, true
}

func dynamicSubjectBoundary(tokens []shared.Token, end int) bool {
	if end >= len(tokens) {
		return true
	}
	switch tokens[end].Kind {
	case shared.Comma, shared.Period:
		return true
	default:
		return equalWord(tokens[end], "to") || equalWord(tokens[end], "until")
	}
}

func precedingAmountMultiplier(tokens []shared.Token, atoms parser.Atoms) int {
	multiplier := 0
	for _, token := range tokens {
		value := numberWord(token, atoms)
		if value == 0 {
			continue
		}
		if multiplier != 0 && multiplier != value {
			return 0
		}
		multiplier = value
	}
	if multiplier == 0 {
		return 1
	}
	return multiplier
}

func wordsAt(tokens []shared.Token, start int, words ...string) bool {
	if start < 0 || start+len(words) > len(tokens) {
		return false
	}
	for i, word := range words {
		if !equalWord(tokens[start+i], word) {
			return false
		}
	}
	return true
}

func firstSymbol(tokens []shared.Token) string {
	for _, token := range tokens {
		if token.Kind == shared.Symbol {
			return token.Text
		}
	}
	return ""
}

func effectKindAt(tokens []shared.Token, index int) EffectKind {
	kind := effectKind(tokens[index])
	if kind == EffectGrantKeyword &&
		index >= 2 &&
		(equalWord(tokens[index-2], "opponent") || equalWord(tokens[index-2], "opponents")) &&
		equalWord(tokens[index-1], "you") {
		return EffectUnknown
	}
	if kind == EffectEnterTapped && index+1 < len(tokens) && equalWord(tokens[index+1], "prepared") {
		return EffectEnterPrepared
	}
	if kind == EffectCast && index > 0 &&
		(equalWord(tokens[index-1], "was") || equalWord(tokens[index-1], "were")) {
		return EffectUnknown
	}
	if kind == EffectCounter && !counterIsVerb(tokens, index) {
		return EffectUnknown
	}
	if kind == EffectGain && index+1 < len(tokens) && equalWord(tokens[index+1], "control") {
		return EffectGainControl
	}
	if kind == EffectDouble && index+1 < len(tokens) && equalWord(tokens[index+1], "strike") {
		return EffectUnknown
	}
	if kind == EffectGrantKeyword && keywordGrantContinuesPTBuff(tokens, index) {
		return EffectUnknown
	}
	return kind
}

func keywordGrantContinuesPTBuff(tokens []shared.Token, index int) bool {
	for i := range index {
		if !equalWord(tokens[i], "get") && !equalWord(tokens[i], "gets") {
			continue
		}
		power, toughness := compilePTChange(tokens[i+1 : index])
		return power.Known && toughness.Known
	}
	return false
}

func counterIsVerb(tokens []shared.Token, index int) bool {
	if index == 0 {
		return true
	}
	previous := tokens[index-1]
	if previous.Kind == shared.Comma || previous.Kind == shared.Period || previous.Kind == shared.Semicolon {
		return true
	}
	if equalWord(previous, "then") || equalWord(previous, "may") ||
		equalWord(previous, "can") {
		return true
	}
	if index+1 >= len(tokens) {
		return false
	}
	return equalWord(tokens[index+1], "target") || equalWord(tokens[index+1], "it") ||
		equalWord(tokens[index+1], "that")
}

func effectNegated(tokens []shared.Token, verbIndex int) bool {
	start := max(0, verbIndex-3)
	for _, token := range tokens[start:verbIndex] {
		if equalWord(token, "can't") || equalWord(token, "cannot") {
			return true
		}
	}
	return false
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

func effectKind(token shared.Token) EffectKind {
	if token.Kind != shared.Word {
		return EffectUnknown
	}
	switch strings.ToLower(token.Text) {
	case "add", "adds":
		return EffectAddMana
	case "attach", "attaches":
		return EffectAttach
	case "cast", "casts":
		return EffectCast
	case "counter", "counters":
		return EffectCounter
	case "create", "creates":
		return EffectCreate
	case "deal", "deals":
		return EffectDealDamage
	case "destroy", "destroys":
		return EffectDestroy
	case "discard", "discards":
		return EffectDiscard
	case "discover", "discovers":
		return EffectDiscover
	case "double", "doubles":
		return EffectDouble
	case "draw", "draws":
		return EffectDraw
	case "enters":
		return EffectEnterTapped
	case "exile", "exiles":
		return EffectExile
	case "fight", "fights":
		return EffectFight
	case "gain", "gains":
		return EffectGain
	case "has", "have":
		return EffectGrantKeyword
	case "investigate", "investigates":
		return EffectInvestigate
	case "explore", "explores":
		return EffectExplore
	case "lose", "loses":
		return EffectLose
	case "manifest":
		return EffectManifest
	case "look":
		return EffectManifestDread
	case "mill", "mills":
		return EffectMill
	case "get", "gets":
		return EffectModifyPT
	case "put", "puts":
		return EffectPut
	case "proliferate", "proliferates":
		return EffectProliferate
	case "regenerate", "regenerates":
		return EffectRegenerate
	case "return", "returns":
		return EffectReturn
	case "reveal", "reveals":
		return EffectReveal
	case "sacrifice", "sacrifices":
		return EffectSacrifice
	case "scry", "scries":
		return EffectScry
	case "surveil", "surveils":
		return EffectSurveil
	case "search", "searches":
		return EffectSearch
	case "shuffle", "shuffles":
		return EffectShuffle
	case "tap", "taps":
		return EffectTap
	case "untap", "untaps":
		return EffectUntap
	case "transform", "transforms":
		return EffectTransform
	default:
		return EffectUnknown
	}
}

func isEffectVerb(token shared.Token) bool {
	return effectKind(token) != EffectUnknown
}

func compileDuration(tokens []shared.Token, atoms parser.Atoms) DurationKind {
	words := shared.NormalizedWords(tokens)
	switch {
	case containsSequence(words, "until", "end", "of", "turn"):
		return DurationUntilEndOfTurn
	case containsSequence(words, "until", "your", "next", "turn"):
		return DurationUntilYourNextTurn
	case containsSequence(words, "this", "combat"):
		return DurationThisCombat
	case containsSequence(words, "this", "turn"):
		return DurationThisTurn
	}
	// Source-tied control durations: "as long as this [type] remains on the
	// battlefield" and "for as long as this [type] remains on the battlefield".
	if containsSequence(words, "as", "long", "as", "this") &&
		(containsSequence(words, "remains", "on", "the", "battlefield") ||
			containsSequence(words, "is", "on", "the", "battlefield")) {
		return DurationForAsLongAsSourceOnBattlefield
	}
	// "for as long as you control this [type]" — self-referential.
	if containsSequence(words, "for", "as", "long", "as", "you", "control", "this") {
		return DurationForAsLongAsYouControlSource
	}
	// "for as long as you control [CardName]" — explicit source name, recognized
	// from the parser-emitted self-name span rather than name spelling.
	for index := 0; index+6 < len(tokens); index++ {
		if wordsAt(tokens, index, "for", "as", "long", "as", "you", "control") &&
			atoms.SelfNameStartingAt(tokens[index+6].Span) {
			return DurationForAsLongAsYouControlSource
		}
	}
	return DurationNone
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

func numberWord(token shared.Token, atoms parser.Atoms) int {
	if token.Kind == shared.Integer {
		value, _ := strconv.Atoi(token.Text)
		return value
	}
	// The parser owns the cardinal vocabulary and emits a typed value per
	// cardinal word; the compiler keeps its conservative numeric range policy of
	// recognizing only "one" through "four".
	if value, ok := atoms.CardinalAt(token.Span); ok && value <= 4 {
		return value
	}
	return 0
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
