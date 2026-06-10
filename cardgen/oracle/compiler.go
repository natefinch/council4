package oracle

import (
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/mtg/game/counter"
)

// Compile parses and semantically lowers one card face's Oracle text.
func Compile(source string, context ParseContext) (Compilation, []Diagnostic) {
	document, diagnostics := Parse(source, context)
	compilation, compilerDiagnostics := CompileDocument(document, context)
	return compilation, append(diagnostics, compilerDiagnostics...)
}

// CompileDocument lowers a syntax document into conservative semantic IR.
func CompileDocument(document Document, context ParseContext) (Compilation, []Diagnostic) {
	compilation := Compilation{Syntax: document}
	var diagnostics []Diagnostic
	for _, ability := range document.Abilities {
		compiled, abilityDiagnostics := compileAbility(document.Source, ability, context)
		compilation.Abilities = append(compilation.Abilities, compiled)
		diagnostics = append(diagnostics, abilityDiagnostics...)
	}
	return compilation, diagnostics
}

func compileAbility(
	source string,
	ability Ability,
	context ParseContext,
) (CompiledAbility, []Diagnostic) {
	var diagnostics []Diagnostic
	compiled := CompiledAbility{
		Kind: ability.Kind,
		Span: ability.Span,
		Text: ability.Text,
	}
	if ability.AbilityWord != nil {
		compiled.AbilityWord = ability.AbilityWord.Text
	}
	compiled.Chapters = append([]int(nil), ability.Chapters...)
	compiled.ChapterSpan = ability.ChapterSpan
	if ability.Cost != nil {
		cost := compileCost(*ability.Cost, ability.Kind)
		compiled.Cost = &cost
	}
	if ability.Kind == AbilityTriggered {
		trigger := compileTrigger(ability)
		compiled.Trigger = &trigger
	}
	if ability.Modal != nil {
		for _, mode := range ability.Modal.Options {
			compiledMode, modeDiagnostics := compileMode(source, mode, context)
			compiled.Modes = append(compiled.Modes, compiledMode)
			diagnostics = append(diagnostics, modeDiagnostics...)
		}
	}

	body := abilityBodyTokens(ability)
	timing, timingSpan := compileActivationTiming(ability.Kind, parseSentences(source, body))
	if timing != ActivationTimingNone {
		body = tokensOutsideSpan(body, timingSpan)
		compiled.ActivationTiming = timing
		compiled.ActivationTimingSpan = timingSpan
	}
	tokens := semanticTokens(body, ability.Reminders, ability.Quoted)
	if ability.Kind == AbilityTriggered &&
		len(tokens) >= 2 &&
		equalWord(tokens[0], "you") &&
		equalWord(tokens[1], "may") {
		compiled.Optional = true
		compiled.OptionalSpan = Span{Start: tokens[0].Span.Start, End: tokens[1].Span.End}
	}
	compiled.Keywords = compileKeywords(tokens)
	compiled.Targets = compileTargets(tokens)
	conditionTokens := tokens
	if ability.Kind == AbilityTriggered {
		conditionTokens = semanticTokens(ability.Tokens, ability.Reminders, ability.Quoted)
	}
	compiled.Conditions = compileConditions(conditionTokens, ability.Kind == AbilityTriggered)
	if ability.Text == "This creature attacks each combat if able." {
		compiled.Conditions = nil
	}
	compiled.Effects = compileEffects(
		parseSentences(source, body),
		ability.Reminders,
		ability.Quoted,
		context.CardName,
	)
	referenceTokens := semanticTokens(ability.Tokens, ability.Reminders, ability.Quoted)
	if timing != ActivationTimingNone {
		referenceTokens = tokensOutsideSpan(referenceTokens, timingSpan)
	}
	compiled.References = compileReferences(
		referenceTokens,
		context.CardName,
	)

	for _, mode := range compiled.Modes {
		if len(mode.Effects) == 0 && len(mode.Keywords) == 0 {
			diagnostics = append(diagnostics, unsupportedDiagnostic(mode.Span, mode.Text))
		}
	}
	if ability.Kind != AbilityReminder && ability.Modal == nil &&
		len(compiled.Effects) == 0 && len(compiled.Keywords) == 0 {
		diagnostics = append(diagnostics, unsupportedDiagnostic(ability.Span, ability.Text))
	}
	if compiled.Cost != nil {
		for _, component := range compiled.Cost.Components {
			if component.Kind == CostUnknown {
				diagnostics = append(diagnostics, Diagnostic{
					Severity: SeverityWarning,
					Summary:  "unsupported cost",
					Detail:   "the compiler preserved this cost component but did not assign executable semantics",
					Span:     component.Span,
				})
			}
		}
	}
	return compiled, diagnostics
}

func compileActivationTiming(kind AbilityKind, sentences []Sentence) (ActivationTimingKind, Span) {
	if kind != AbilityActivated || len(sentences) == 0 {
		return ActivationTimingNone, Span{}
	}
	last := len(sentences) - 1
	lastKind := activationTimingSentence(sentences[last].Text)
	if lastKind == ActivationTimingNone {
		return ActivationTimingNone, Span{}
	}
	if last > 0 {
		previousKind := activationTimingSentence(sentences[last-1].Text)
		if previousKind != ActivationTimingNone {
			if (previousKind == ActivationTimingSorcery && lastKind == ActivationTimingOncePerTurn) ||
				(previousKind == ActivationTimingOncePerTurn && lastKind == ActivationTimingSorcery) {
				return ActivationTimingSorceryOncePerTurn, Span{
					Start: sentences[last-1].Span.Start,
					End:   sentences[last].Span.End,
				}
			}
			return ActivationTimingNone, Span{}
		}
	}
	return lastKind, sentences[last].Span
}

func activationTimingSentence(text string) ActivationTimingKind {
	switch text {
	case "Activate only as a sorcery.":
		return ActivationTimingSorcery
	case "Activate only once each turn.":
		return ActivationTimingOncePerTurn
	case "Activate only during combat.":
		return ActivationTimingDuringCombat
	case "Activate only during your upkeep.":
		return ActivationTimingDuringUpkeep
	default:
		return ActivationTimingNone
	}
}

func tokensOutsideSpan(tokens []Token, span Span) []Token {
	return slices.DeleteFunc(append([]Token(nil), tokens...), func(token Token) bool {
		return span.Start.Offset <= token.Span.Start.Offset &&
			span.End.Offset >= token.Span.End.Offset
	})
}

func compileMode(
	source string,
	mode Mode,
	context ParseContext,
) (CompiledMode, []Diagnostic) {
	tokens := semanticTokens(mode.Tokens, mode.Reminders, mode.Quoted)
	compiled := CompiledMode{
		Span:       mode.Span,
		Text:       mode.Text,
		Targets:    compileTargets(tokens),
		Conditions: compileConditions(tokens, false),
		Effects: compileEffects(
			parseSentences(source, mode.Tokens),
			mode.Reminders,
			mode.Quoted,
			context.CardName,
		),
		Keywords:   compileKeywords(tokens),
		References: compileReferences(tokens, context.CardName),
	}
	return compiled, nil
}

func compileCost(phrase Phrase, abilityKind AbilityKind) CompiledCost {
	cost := CompiledCost{Span: phrase.Span, Text: phrase.Text}
	parts := splitTopLevel(phrase.Tokens, Comma)
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}
		component := CostComponent{
			Kind: CostUnknown,
			Span: spanOf(part),
			Text: sliceSpan(phrase.Text, relativeSpan(spanOf(part), phrase.Span.Start.Offset)),
		}
		if abilityKind == AbilityLoyalty {
			component.Kind = CostLoyalty
			component.Amount = joinedTokenText(part)
		} else {
			words := normalizedWords(part)
			switch {
			case len(part) == 1 && part[0].Kind == Symbol && strings.EqualFold(part[0].Text, "{T}"):
				component.Kind = CostTap
				component.Symbol = part[0].Text
			case len(part) == 1 && part[0].Kind == Symbol && strings.EqualFold(part[0].Text, "{Q}"):
				component.Kind = CostUntap
				component.Symbol = part[0].Text
			case startsWords(words, "sacrifice"):
				component.Kind = CostSacrifice
				component.Object = wordsAfterFirst(part)
			case startsWords(words, "discard"):
				component.Kind = CostDiscard
				component.Object = wordsAfterFirst(part)
			case startsWords(words, "pay") && containsWord(words, "life"):
				component.Kind = CostPayLife
				component.Amount = firstInteger(part)
			case startsWords(words, "exile"):
				component.Kind = CostExile
				component.Object = wordsAfterFirst(part)
			case startsWords(words, "remove") && containsWord(words, "counter"):
				component.Kind = CostRemoveCounter
				component.Object = wordsAfterFirst(part)
			case allSymbols(part):
				component.Kind = CostMana
				component.Symbol = joinedTokenText(part)
			default:
			}
		}
		cost.Components = append(cost.Components, component)
	}
	return cost
}

func compileTrigger(ability Ability) CompiledTrigger {
	tokens := ability.Tokens
	if ability.AbilityWord != nil || len(ability.Chapters) > 0 {
		if dash := topLevelIndex(tokens, EmDash); dash >= 0 {
			tokens = tokens[dash+1:]
		}
	}
	end := triggerBodyComma(tokens)
	if end < 0 {
		end = len(tokens)
	}
	triggerTokens := tokens[:end]
	trigger := CompiledTrigger{
		Kind: TriggerUnknown,
		Span: spanOf(triggerTokens),
		Text: joinedSourceText(triggerTokens),
	}
	if len(triggerTokens) == 0 {
		return trigger
	}
	switch strings.ToLower(triggerTokens[0].Text) {
	case "when":
		trigger.Kind = TriggerWhen
	case "whenever":
		trigger.Kind = TriggerWhenever
	case "at":
		trigger.Kind = TriggerAt
	default:
	}
	if len(triggerTokens) > 1 {
		trigger.Event = joinedSourceText(triggerTokens[1:])
	}
	conditions := compileConditions(ability.Tokens, true)
	for i := range conditions {
		if conditions[i].Intervening {
			condition := conditions[i]
			trigger.Condition = &condition
			break
		}
	}
	return trigger
}

func compileTargets(tokens []Token) []CompiledTarget {
	var targets []CompiledTarget
	for i, token := range tokens {
		if token.Kind != Word || !strings.EqualFold(token.Text, "target") {
			continue
		}
		start := i
		cardinality := TargetCardinality{Min: 1, Max: 1}
		if i >= 3 && equalWord(tokens[i-3], "up") && equalWord(tokens[i-2], "to") {
			start = i - 3
			cardinality.Min = 0
			cardinality.Max = numberWord(tokens[i-1])
			if cardinality.Max == 0 {
				cardinality.Max = 1
			}
		} else if i >= 1 {
			if count := numberWord(tokens[i-1]); count > 0 {
				start = i - 1
				cardinality.Min = count
				cardinality.Max = count
			} else if equalWord(tokens[i-1], "any") {
				start = i - 1
			}
		}
		end := targetPhraseEnd(tokens, i+1)
		phraseTokens := tokens[start:end]
		selectorTokens := append([]Token(nil), tokens[start:i]...)
		selectorTokens = append(selectorTokens, tokens[i+1:end]...)
		targets = append(targets, CompiledTarget{
			Span:        spanOf(phraseTokens),
			Text:        joinedSourceText(phraseTokens),
			Cardinality: cardinality,
			Selector:    compileSelector(selectorTokens),
		})
	}
	return targets
}

func targetPhraseEnd(tokens []Token, start int) int {
	end := start
	for end < len(tokens) {
		token := tokens[end]
		if token.Kind == Comma || token.Kind == Period || token.Kind == Semicolon ||
			(end > start && isEffectVerb(token)) {
			break
		}
		end++
	}
	return end
}

func compileSelector(tokens []Token) CompiledSelector {
	selector := CompiledSelector{Raw: joinedSourceText(tokens)}
	words := normalizedWords(tokens)
	switch {
	case containsNoun(words, "artifact"):
		selector.Kind = SelectorArtifact
	case containsNoun(words, "creature"):
		selector.Kind = SelectorCreature
	case containsNoun(words, "enchantment"):
		selector.Kind = SelectorEnchantment
	case containsNoun(words, "land"):
		selector.Kind = SelectorLand
	case containsNoun(words, "planeswalker"):
		selector.Kind = SelectorPlaneswalker
	case containsNoun(words, "battle"):
		selector.Kind = SelectorBattle
	case containsNoun(words, "permanent"):
		selector.Kind = SelectorPermanent
	case containsNoun(words, "opponent"):
		selector.Kind = SelectorOpponent
	case containsNoun(words, "player"):
		selector.Kind = SelectorPlayer
	case containsNoun(words, "spell"):
		selector.Kind = SelectorSpell
	case containsNoun(words, "card"):
		selector.Kind = SelectorCard
	case containsWord(words, "any"):
		selector.Kind = SelectorAny
	default:
	}
	switch {
	case containsSequence(words, "you", "don't", "control"):
		selector.Controller = ControllerNotYou
	case containsSequence(words, "you", "control"):
		selector.Controller = ControllerYou
	case containsNoun(words, "opponent"):
		selector.Controller = ControllerOpponent
	default:
	}
	selector.Another = containsWord(words, "another")
	selector.Other = containsWord(words, "other")
	selector.Attacking = containsWord(words, "attacking")
	selector.Blocking = containsWord(words, "blocking")
	selector.Tapped = containsWord(words, "tapped")
	selector.Untapped = containsWord(words, "untapped")
	return selector
}

func compileConditions(tokens []Token, triggered bool) []CompiledCondition {
	var conditions []CompiledCondition
	for i := 0; i < len(tokens); i++ {
		kind := conditionKindAt(tokens, i)
		if kind == ConditionUnknown {
			continue
		}
		start := i
		end := conditionEnd(tokens, i)
		phrase := tokens[start:end]
		conditions = append(conditions, CompiledCondition{
			Kind:        kind,
			Span:        spanOf(phrase),
			Text:        joinedSourceText(phrase),
			Intervening: triggered && kind == ConditionIf && isInterveningIf(tokens, start),
		})
		i = end - 1
	}
	return conditions
}

func conditionEnd(tokens []Token, start int) int {
	for i := start; i < len(tokens); i++ {
		if tokens[i].Kind == Period || (i > start && tokens[i].Kind == Comma) {
			return i
		}
	}
	return len(tokens)
}

func compileEffects(
	sentences []Sentence,
	reminders, quoted []Delimited,
	cardName string,
) []CompiledEffect {
	var effects []CompiledEffect
	for _, sentence := range sentences {
		tokens := semanticTokens(sentence.Tokens, reminders, quoted)
		if effect, ok := compileStaticRuleEffect(sentence, tokens); ok {
			effects = append(effects, effect)
			continue
		}
		duration := compileDuration(tokens)
		staticSubject, staticSubjectSpan, staticSubjectSubtype := compileStaticSubject(tokens)
		effectIndices := effectTokenIndices(tokens, cardName)
		for effectIndex, tokenIndex := range effectIndices {
			token := tokens[tokenIndex]
			kind := effectKindAt(tokens, tokenIndex)
			clauseEnd := effectClauseEnd(tokens, effectIndices, effectIndex)
			clauseTokens := tokens[tokenIndex+1 : clauseEnd]
			powerDelta, toughnessDelta := compilePTChange(clauseTokens)
			counterKind, counterKindKnown := counterKindWord(clauseTokens)
			effects = append(effects, CompiledEffect{
				Kind:                 kind,
				Span:                 sentence.Span,
				Text:                 sentence.Text,
				VerbSpan:             token.Span,
				Duration:             duration,
				Selector:             compileSelector(clauseTokens),
				Amount:               compileEffectAmount(clauseTokens, cardName),
				PowerDelta:           powerDelta,
				ToughnessDelta:       toughnessDelta,
				StaticSubject:        staticSubject,
				StaticSubjectSpan:    staticSubjectSpan,
				StaticSubjectSubtype: staticSubjectSubtype,
				Symbol:               firstSymbol(clauseTokens),
				CounterKind:          counterKind,
				CounterKindKnown:     (kind == EffectPut || kind == EffectEnterTapped) && counterKindKnown,
				Negated:              effectNegated(tokens, tokenIndex),
			})
		}

	}
	return effects
}

func counterKindWord(tokens []Token) (counter.Kind, bool) {
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
	prefix := strings.TrimPrefix(strings.ToLower(joinedSourceText(tokens[:counterIndex])), "with ")
	names := []counter.Kind{
		counter.PlusOnePlusOne,
		counter.MinusOneMinusOne,
		counter.Loyalty,
		counter.Charge,
		counter.Time,
		counter.Defense,
		counter.Poison,
		counter.Lore,
		counter.Verse,
		counter.Shield,
		counter.Stun,
		counter.Finality,
		counter.Brick,
		counter.Page,
		counter.Enlightened,
		counter.Indestructible,
		counter.Deathtouch,
		counter.Flying,
		counter.FirstStrike,
		counter.Hexproof,
		counter.Lifelink,
		counter.Menace,
		counter.Reach,
		counter.Trample,
		counter.Vigilance,
		counter.Energy,
		counter.Experience,
	}
	for _, kind := range names {
		name := kind.String()
		if !strings.HasSuffix(prefix, name) {
			continue
		}
		amount := strings.TrimSpace(strings.TrimSuffix(prefix, name))
		switch amount {
		case "a", "an", "one", "two", "three", "four", "x":
			return kind, CounterKindPlacementSupported(kind)
		default:
			if value, err := strconv.Atoi(amount); err == nil && value > 0 {
				return kind, CounterKindPlacementSupported(kind)
			}
		}
	}
	return 0, false
}

func effectTokenIndices(tokens []Token, cardName string) []int {
	var indices []int
	for index := range tokens {
		if effectKindAt(tokens, index) != EffectUnknown &&
			!tokenInCardName(tokens, index, cardName) {
			indices = append(indices, index)
		}
	}
	return indices
}

func effectClauseEnd(tokens []Token, effectIndices []int, effectIndex int) int {
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

func effectCoordinationStart(tokens []Token, start, effectIndex int) int {
	for index := effectIndex - 1; index >= start; index-- {
		if tokens[index].Kind == Comma || tokens[index].Kind == Semicolon {
			return -1
		}
		if equalWord(tokens[index], "then") || equalWord(tokens[index], "and") {
			return index
		}
	}
	return -1
}

func tokenInCardName(tokens []Token, index int, cardName string) bool {
	nameWords := strings.Fields(cardName)
	for start := 0; start <= index; start++ {
		end := start + len(nameWords)
		if index >= end {
			continue
		}
		if wordsAt(tokens, start, nameWords...) ||
			possessiveNameAt(tokens, start, nameWords) {
			return true
		}
	}
	return false
}

func conditionKindAt(tokens []Token, index int) ConditionKind {
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

func compileStaticRuleEffect(sentence Sentence, tokens []Token) (CompiledEffect, bool) {
	var kind EffectKind
	var verb string
	var selector CompiledSelector
	switch sentence.Text {
	case "This creature can't block.":
		kind = EffectCantBlock
		verb = "block"
		selector = CompiledSelector{
			Kind: SelectorCreature,
			Raw:  "this creature",
		}
	case "This creature can't be blocked.":
		kind = EffectCantBeBlocked
		verb = "blocked"
		selector = CompiledSelector{
			Kind: SelectorCreature,
			Raw:  "this creature",
		}
	case "This creature attacks each combat if able.":
		kind = EffectMustAttack
		verb = "attacks"
		selector = CompiledSelector{
			Kind: SelectorCreature,
			Raw:  "this creature",
		}
	case "This spell can't be countered.":
		kind = EffectCantBeCountered
		verb = "countered"
		selector = CompiledSelector{Raw: "this spell"}
	default:
		return CompiledEffect{}, false
	}
	for _, token := range tokens {
		if equalWord(token, verb) {
			return CompiledEffect{
				Kind:     kind,
				Span:     sentence.Span,
				Text:     sentence.Text,
				VerbSpan: token.Span,
				Selector: selector,
				Negated:  kind != EffectMustAttack,
			}, true
		}
	}
	return CompiledEffect{}, false
}

func compileStaticSubject(tokens []Token) (StaticSubjectKind, Span, string) {
	switch {
	case len(tokens) >= 3 &&
		(equalWord(tokens[0], "enchanted") || equalWord(tokens[0], "equipped")) &&
		equalWord(tokens[1], "creature") &&
		(equalWord(tokens[2], "gets") || equalWord(tokens[2], "has")):
		return StaticSubjectAttachedObject, spanOf(tokens[:2]), ""
	case len(tokens) >= 5 &&
		equalWord(tokens[0], "other") &&
		equalWord(tokens[1], "creatures") &&
		equalWord(tokens[2], "you") &&
		equalWord(tokens[3], "control") &&
		(equalWord(tokens[4], "get") || equalWord(tokens[4], "have")):
		return StaticSubjectOtherControlledCreatures, spanOf(tokens[:4]), ""
	case len(tokens) >= 4 &&
		equalWord(tokens[0], "creatures") &&
		equalWord(tokens[1], "you") &&
		equalWord(tokens[2], "control") &&
		(equalWord(tokens[3], "get") || equalWord(tokens[3], "have")):
		return StaticSubjectControlledCreatures, spanOf(tokens[:3]), ""
	case len(tokens) >= 6 &&
		equalWord(tokens[0], "creatures") &&
		equalWord(tokens[1], "your") &&
		equalWord(tokens[2], "opponents") &&
		equalWord(tokens[3], "control") &&
		(equalWord(tokens[4], "get") || equalWord(tokens[4], "have")):
		return StaticSubjectOpponentControlledCreatures, spanOf(tokens[:4]), ""
	case len(tokens) >= 5 &&
		equalWord(tokens[0], "each") &&
		equalWord(tokens[1], "wall") &&
		equalWord(tokens[2], "you") &&
		equalWord(tokens[3], "control") &&
		(equalWord(tokens[4], "gets") || equalWord(tokens[4], "has")):
		return StaticSubjectControlledWalls, spanOf(tokens[:4]), ""
	case len(tokens) >= 4 &&
		equalWord(tokens[0], "walls") &&
		equalWord(tokens[1], "you") &&
		equalWord(tokens[2], "control") &&
		(equalWord(tokens[3], "get") || equalWord(tokens[3], "have")):
		return StaticSubjectControlledWalls, spanOf(tokens[:3]), ""
	case len(tokens) >= 4 &&
		equalWord(tokens[0], "artifacts") &&
		equalWord(tokens[1], "you") &&
		equalWord(tokens[2], "control") &&
		(equalWord(tokens[3], "get") || equalWord(tokens[3], "have")):
		return StaticSubjectControlledArtifacts, spanOf(tokens[:3]), ""
	case len(tokens) >= 4 &&
		equalWord(tokens[0], "tokens") &&
		equalWord(tokens[1], "you") &&
		equalWord(tokens[2], "control") &&
		(equalWord(tokens[3], "get") || equalWord(tokens[3], "have")):
		return StaticSubjectControlledTokens, spanOf(tokens[:3]), ""
	case len(tokens) >= 5 &&
		equalWord(tokens[0], "other") &&
		tokens[1].Kind == Word &&
		equalWord(tokens[2], "you") &&
		equalWord(tokens[3], "control") &&
		equalWord(tokens[4], "have"):
		return StaticSubjectOtherControlledCreatureSubtype, spanOf(tokens[:4]), tokens[1].Text
	case len(tokens) >= 4 &&
		tokens[0].Kind == Word &&
		equalWord(tokens[1], "you") &&
		equalWord(tokens[2], "control") &&
		equalWord(tokens[3], "have"):
		return StaticSubjectControlledCreatureSubtype, spanOf(tokens[:3]), tokens[0].Text
	default:
		return StaticSubjectNone, Span{}, ""
	}
}

func compilePTChange(tokens []Token) (power, toughness CompiledSignedAmount) {
	for i := 0; i+4 < len(tokens); i++ {
		power, powerOK := signedAmount(tokens[i], tokens[i+1])
		toughness, toughnessOK := signedAmount(tokens[i+3], tokens[i+4])
		if powerOK && tokens[i+2].Kind == Slash && toughnessOK {
			return power, toughness
		}
	}
	return CompiledSignedAmount{}, CompiledSignedAmount{}
}

func signedAmount(sign, amount Token) (CompiledSignedAmount, bool) {
	if amount.Kind != Integer || (sign.Kind != Plus && sign.Kind != Minus) {
		return CompiledSignedAmount{}, false
	}
	value, err := strconv.Atoi(amount.Text)
	if err != nil {
		return CompiledSignedAmount{}, false
	}
	negative := sign.Kind == Minus
	return CompiledSignedAmount{Value: value, Known: true, Negative: negative}, true
}

func compileEffectAmount(tokens []Token, cardName string) CompiledAmount {
	dynamic := compileDynamicEffectAmount(tokens, cardName)
	if dynamic.matched {
		return dynamic.amount
	}
	if dynamic.attempted {
		return CompiledAmount{}
	}
	for _, token := range tokens {
		if value := numberWord(token); value > 0 {
			return CompiledAmount{Value: value, Known: true}
		}
		if equalWord(token, "a") || equalWord(token, "an") {
			return CompiledAmount{Value: 1, Known: true}
		}
	}
	for _, token := range tokens {
		if token.Kind == Symbol {
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

func compileDynamicEffectAmount(tokens []Token, cardName string) compiledDynamicAmount {
	var matches []CompiledAmount
	attempted := false
	for i := range tokens {
		prefix, ok := dynamicAmountPrefix(tokens, i)
		if !ok {
			continue
		}
		attempted = true
		if subject, matched := dynamicAmountSubject(tokens, prefix.subjectStart, cardName); matched {
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
			subject.amount.DynamicKind != DynamicAmountOpponentCount {
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

func dynamicAmountPrefix(tokens []Token, index int) (dynamicAmountPrefixMatch, bool) {
	switch {
	case wordsAt(tokens, index, "equal", "to", "twice", "the", "number", "of"):
		return dynamicAmountPrefixMatch{DynamicAmountEqual, index + 6, 2, dynamicAmountCountSubject, dynamicSubjectPlural}, true
	case wordsAt(tokens, index, "equal", "to", "the", "number", "of"):
		return dynamicAmountPrefixMatch{DynamicAmountEqual, index + 5, 1, dynamicAmountCountSubject, dynamicSubjectPlural}, true
	case wordsAt(tokens, index, "for", "each"):
		multiplier := precedingAmountMultiplier(tokens[:index])
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

func (n dynamicCountNoun) numberAt(tokens []Token, start int) (dynamicSubjectNumber, bool) {
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

func dynamicAmountSubject(tokens []Token, start int, cardName string) (dynamicAmountSubjectMatch, bool) {
	if start >= len(tokens) {
		return dynamicAmountSubjectMatch{}, false
	}
	if subject, ok := dynamicCountAmountSubject(tokens, start); ok {
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
		tokens[start+2].Kind == Apostrophe &&
		equalWord(tokens[start+3], "s") &&
		equalWord(tokens[start+4], "power") &&
		dynamicSubjectBoundary(tokens, start+5):
		return dynamicAmountSubjectMatch{
			amount: CompiledAmount{
				DynamicKind:   DynamicAmountSourcePower,
				ReferenceSpan: spanOf(tokens[start : start+2]),
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
				ReferenceSpan: spanOf(tokens[start : start+2]),
			},
			end: start + 3,
		}, true
	default:
		nameWords := strings.Fields(cardName)
		if possessiveNameAt(tokens, start, nameWords) {
			end := start + len(nameWords)
			if end < len(tokens) &&
				equalWord(tokens[end], "power") &&
				dynamicSubjectBoundary(tokens, end+1) {
				return dynamicAmountSubjectMatch{
					amount: CompiledAmount{
						DynamicKind:   DynamicAmountSourcePower,
						ReferenceSpan: spanOf(tokens[start:end]),
					},
					end: end + 1,
				}, true
			}
		}
		if len(nameWords) > 0 &&
			wordsAt(tokens, start, nameWords...) {
			possessive := start + len(nameWords)
			if possessive+2 < len(tokens) &&
				tokens[possessive].Kind == Apostrophe &&
				equalWord(tokens[possessive+1], "s") &&
				equalWord(tokens[possessive+2], "power") &&
				dynamicSubjectBoundary(tokens, possessive+3) {
				return dynamicAmountSubjectMatch{
					amount: CompiledAmount{
						DynamicKind:   DynamicAmountSourcePower,
						ReferenceSpan: spanOf(tokens[start:possessive]),
					},
					end: possessive + 3,
				}, true
			}

		}
		return dynamicAmountSubjectMatch{}, false
	}
}

func dynamicCountAmountSubject(tokens []Token, start int) (dynamicAmountSubjectMatch, bool) {
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
					Selector: CompiledSelector{
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

func dynamicSubjectBoundary(tokens []Token, end int) bool {
	if end >= len(tokens) {
		return true
	}
	switch tokens[end].Kind {
	case Comma, Period:
		return true
	default:
		return equalWord(tokens[end], "to") || equalWord(tokens[end], "until")
	}
}

func precedingAmountMultiplier(tokens []Token) int {
	multiplier := 0
	for _, token := range tokens {
		value := numberWord(token)
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

func wordsAt(tokens []Token, start int, words ...string) bool {
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

func possessiveNameAt(tokens []Token, start int, nameWords []string) bool {
	if len(nameWords) == 0 || start < 0 || start+len(nameWords) > len(tokens) {
		return false
	}
	last := len(nameWords) - 1
	for i := range last {
		if !equalWord(tokens[start+i], nameWords[i]) {
			return false
		}
	}
	return strings.EqualFold(tokens[start+last].Text, nameWords[last]+"'s")
}

func firstSymbol(tokens []Token) string {
	for _, token := range tokens {
		if token.Kind == Symbol {
			return token.Text
		}
	}
	return ""
}

func effectKindAt(tokens []Token, index int) EffectKind {
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
	if kind == EffectDouble && index+1 < len(tokens) && equalWord(tokens[index+1], "strike") {
		return EffectUnknown
	}
	if kind == EffectGrantKeyword && keywordGrantContinuesPTBuff(tokens, index) {
		return EffectUnknown
	}
	return kind
}

func keywordGrantContinuesPTBuff(tokens []Token, index int) bool {
	for i := range index {
		if !equalWord(tokens[i], "get") && !equalWord(tokens[i], "gets") {
			continue
		}
		power, toughness := compilePTChange(tokens[i+1 : index])
		return power.Known && toughness.Known
	}
	return false
}

func counterIsVerb(tokens []Token, index int) bool {
	if index == 0 {
		return true
	}
	previous := tokens[index-1]
	if previous.Kind == Comma || previous.Kind == Period || previous.Kind == Semicolon {
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

func effectNegated(tokens []Token, verbIndex int) bool {
	start := max(0, verbIndex-3)
	for _, token := range tokens[start:verbIndex] {
		if equalWord(token, "can't") || equalWord(token, "cannot") {
			return true
		}
	}
	return false
}

func abilityBodyTokens(ability Ability) []Token {
	tokens := ability.Tokens
	if ability.AbilityWord != nil {
		if dash := topLevelIndex(tokens, EmDash); dash >= 0 {
			tokens = tokens[dash+1:]
		}
	}
	switch ability.Kind {
	case AbilityActivated, AbilityLoyalty:
		if colon := topLevelIndex(tokens, Colon); colon >= 0 {
			return tokens[colon+1:]
		}
	case AbilityTriggered:
		if comma := triggerBodyComma(tokens); comma >= 0 {
			return tokens[comma+1:]
		}
	default:
	}
	return tokens
}

func effectKind(token Token) EffectKind {
	if token.Kind != Word {
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
	case "lose", "loses":
		return EffectLose
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

func isEffectVerb(token Token) bool {
	return effectKind(token) != EffectUnknown
}

func compileDuration(tokens []Token) DurationKind {
	words := normalizedWords(tokens)
	switch {
	case containsSequence(words, "until", "end", "of", "turn"):
		return DurationUntilEndOfTurn
	case containsSequence(words, "until", "your", "next", "turn"):
		return DurationUntilYourNextTurn
	case containsSequence(words, "this", "combat"):
		return DurationThisCombat
	case containsSequence(words, "this", "turn"):
		return DurationThisTurn
	default:
		return DurationNone
	}
}

var keywordNames = map[string]string{
	"affinity": "Affinity", "annihilator": "Annihilator", "cascade": "Cascade",
	"companion": "Companion", "convoke": "Convoke", "cycling": "Cycling",
	"deathtouch": "Deathtouch", "defender": "Defender", "delve": "Delve", "devoid": "Devoid",
	"disguise": "Disguise", "double strike": "Double strike", "emerge": "Emerge",
	"enchant": "Enchant", "equip": "Equip", "escape": "Escape",
	"eternalize": "Eternalize", "exalted": "Exalted", "first strike": "First strike",
	"flash": "Flash", "flashback": "Flashback", "flying": "Flying",
	"foretell": "Foretell", "haste": "Haste", "hexproof": "Hexproof",
	"improvise": "Improvise", "indestructible": "Indestructible", "infect": "Infect",
	"kicker": "Kicker", "lifelink": "Lifelink", "madness": "Madness",
	"menace": "Menace", "morph": "Morph", "mutate": "Mutate",
	"ninjutsu": "Ninjutsu", "persist": "Persist", "protection": "Protection",
	"prowess": "Prowess", "read ahead": "Read ahead", "reach": "Reach", "shroud": "Shroud",
	"split second": "Split second", "storm": "Storm", "suspend": "Suspend",
	"toxic": "Toxic", "trample": "Trample", "undying": "Undying",
	"vigilance": "Vigilance", "ward": "Ward", "wither": "Wither",
}

func compileKeywords(tokens []Token) []CompiledKeyword {
	var keywords []CompiledKeyword
	for i := 0; i < len(tokens); i++ {
		for width := 2; width >= 1; width-- {
			if i+width > len(tokens) {
				continue
			}
			name := strings.ToLower(joinWords(tokens[i : i+width]))
			canonical, ok := keywordNames[name]
			if !ok {
				continue
			}
			end := i + width
			parameter, end := compileKeywordParameter(tokens, canonical, end)
			phrase := tokens[i:end]
			keywords = append(keywords, CompiledKeyword{
				Name:      canonical,
				Span:      spanOf(phrase),
				Text:      joinedSourceText(phrase),
				Parameter: parameter,
			})
			i = end - 1
			break
		}
	}
	return keywords
}

func compileKeywordParameter(tokens []Token, keyword string, start int) (parameter string, end int) {
	switch keyword {
	case "Protection":
		parameter, end, _ = compileProtectionParameter(tokens, start)
		return parameter, end
	case "Enchant":
		if start < len(tokens) && isEnchantObjectWord(tokens[start]) {
			return strings.ToLower(tokens[start].Text), start + 1
		}
		return "", start
	}
	end = start
	if end < len(tokens) && tokens[end].Kind == Symbol {
		var symbols strings.Builder
		for end < len(tokens) && tokens[end].Kind == Symbol {
			_, _ = symbols.WriteString(tokens[end].Text)
			end++
		}
		return symbols.String(), end
	}
	if end < len(tokens) && tokens[end].Kind == Integer {
		return tokens[end].Text, end + 1
	}
	return "", end
}

func compileProtectionParameter(tokens []Token, start int) (parameter string, end int, ok bool) {
	if start+1 >= len(tokens) ||
		!equalWord(tokens[start], "from") ||
		!isColorWord(tokens[start+1]) {
		return "", start, false
	}
	colors := []string{strings.ToLower(tokens[start+1].Text)}
	end = start + 2
	for end < len(tokens) {
		next := end
		if tokens[next].Kind == Comma {
			next++
		} else if !equalWord(tokens[next], "and") {
			break
		}
		if next < len(tokens) && equalWord(tokens[next], "and") {
			next++
		}
		if next+1 >= len(tokens) ||
			!equalWord(tokens[next], "from") ||
			!isColorWord(tokens[next+1]) {
			break
		}
		colors = append(colors, strings.ToLower(tokens[next+1].Text))
		end = next + 2
	}
	return strings.Join(colors, ","), end, true
}

func isColorWord(token Token) bool {
	if token.Kind != Word {
		return false
	}
	switch strings.ToLower(token.Text) {
	case "black", "blue", "green", "red", "white":
		return true
	default:
		return false
	}
}

func isEnchantObjectWord(token Token) bool {
	if token.Kind != Word {
		return false
	}
	switch strings.ToLower(token.Text) {
	case "artifact", "creature", "enchantment", "land", "permanent", "planeswalker", "player":
		return true
	default:
		return false
	}
}

func compileReferences(tokens []Token, cardName string) []CompiledReference {
	var references []CompiledReference
	if cardName != "" {
		nameWords := strings.Fields(strings.ToLower(cardName))
		for i := 0; i+len(nameWords) <= len(tokens); i++ {
			if possessiveNameAt(tokens, i, nameWords) {
				phrase := tokens[i : i+len(nameWords)]
				references = append(references, CompiledReference{
					Kind: ReferenceSelfName,
					Span: spanOf(phrase),
					Text: joinedSourceText(phrase),
				})
				i += len(nameWords) - 1
				continue
			}
			if tokenWordsEqual(tokens[i:i+len(nameWords)], nameWords) {
				phrase := tokens[i : i+len(nameWords)]
				references = append(references, CompiledReference{
					Kind: ReferenceSelfName,
					Span: spanOf(phrase),
					Text: joinedSourceText(phrase),
				})
				i += len(nameWords) - 1
			}
		}
	}
	for i := 0; i < len(tokens); i++ {
		switch {
		case i+1 < len(tokens) &&
			equalWord(tokens[i], "this") &&
			strings.EqualFold(tokens[i+1].Text, "creature's"):
			phrase := tokens[i : i+2]
			references = append(references, CompiledReference{
				Kind: ReferenceThisObject,
				Span: spanOf(phrase),
				Text: joinedSourceText(phrase),
			})
			i++
		case i+1 < len(tokens) && equalWord(tokens[i], "this") && objectWord(tokens[i+1]):
			phrase := tokens[i : i+2]
			references = append(references, CompiledReference{
				Kind: ReferenceThisObject,
				Span: spanOf(phrase),
				Text: joinedSourceText(phrase),
			})
			i++
		case i+1 < len(tokens) && equalWord(tokens[i], "that") && objectWord(tokens[i+1]):
			phrase := tokens[i : i+2]
			references = append(references, CompiledReference{
				Kind: ReferenceThatObject,
				Span: spanOf(phrase),
				Text: joinedSourceText(phrase),
			})
			i++
		case equalWord(tokens[i], "it") || equalWord(tokens[i], "its") ||
			equalWord(tokens[i], "they") || equalWord(tokens[i], "their") ||
			equalWord(tokens[i], "them") || equalWord(tokens[i], "those"):
			references = append(references, CompiledReference{
				Kind: ReferencePronoun,
				Span: tokens[i].Span,
				Text: tokens[i].Text,
			})
		default:
		}
	}
	return references
}

func semanticTokens(tokens []Token, reminders, quoted []Delimited) []Token {
	excluded := append(append([]Delimited(nil), reminders...), quoted...)
	result := make([]Token, 0, len(tokens))
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

func splitTopLevel(tokens []Token, separator Kind) [][]Token {
	var result [][]Token
	start := 0
	depth := 0
	quoted := false
	for i, token := range tokens {
		switch token.Kind {
		case LeftParen:
			if !quoted {
				depth++
			}
		case RightParen:
			if !quoted && depth > 0 {
				depth--
			}
		case Quote:
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

func allSymbols(tokens []Token) bool {
	if len(tokens) == 0 {
		return false
	}
	for _, token := range tokens {
		if token.Kind != Symbol {
			return false
		}
	}
	return true
}

func relativeSpan(span Span, base int) Span {
	span.Start.Offset -= base
	span.End.Offset -= base
	return span
}

func wordsAfterFirst(tokens []Token) string {
	if len(tokens) < 2 {
		return ""
	}
	return joinedSourceText(tokens[1:])
}

func firstInteger(tokens []Token) string {
	for _, token := range tokens {
		if token.Kind == Integer {
			return token.Text
		}
	}
	return ""
}

func joinedTokenText(tokens []Token) string {
	var builder strings.Builder
	for _, token := range tokens {
		_, _ = builder.WriteString(token.Text)
	}
	return builder.String()
}

func joinedSourceText(tokens []Token) string {
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

func needsSemanticSpace(previous, current Token) bool {
	if current.Kind == Comma || current.Kind == Period || current.Kind == Colon ||
		current.Kind == Semicolon || current.Kind == RightParen ||
		previous.Kind == LeftParen || previous.Kind == Quote || current.Kind == Quote {
		return false
	}
	if previous.Kind == Plus || previous.Kind == Minus || previous.Kind == Slash ||
		current.Kind == Slash {
		return false
	}
	return previous.Kind != Symbol && current.Kind != Symbol
}

func joinWords(tokens []Token) string {
	var words []string
	for _, token := range tokens {
		if token.Kind != Word {
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

func equalWord(token Token, word string) bool {
	return token.Kind == Word && strings.EqualFold(token.Text, word)
}

func numberWord(token Token) int {
	if token.Kind == Integer {
		value, _ := strconv.Atoi(token.Text)
		return value
	}
	switch strings.ToLower(token.Text) {
	case "one":
		return 1
	case "two":
		return 2
	case "three":
		return 3
	case "four":
		return 4
	default:
		return 0
	}
}

func isInterveningIf(tokens []Token, index int) bool {
	comma := triggerBodyComma(tokens)
	return comma >= 0 && index == comma+1
}

func triggerBodyComma(tokens []Token) int {
	comma := topLevelIndex(tokens, Comma)
	for comma > 0 &&
		comma+1 < len(tokens) &&
		strings.EqualFold(tokens[comma-1].Text, "noncreature") &&
		strings.EqualFold(tokens[comma+1].Text, "nonland") {
		next := topLevelIndex(tokens[comma+1:], Comma)
		if next < 0 {
			return -1
		}
		comma += next + 1
	}
	return comma
}

func containsNoun(words []string, singular string) bool {
	return containsWord(words, singular) || containsWord(words, singular+"s")
}

func tokenWordsEqual(tokens []Token, words []string) bool {
	if len(tokens) != len(words) {
		return false
	}
	for i := range words {
		normalized := strings.ToLower(strings.Trim(tokens[i].Text, ",.'\u2019"))
		if tokens[i].Kind != Word || normalized != words[i] {
			return false
		}
	}
	return true
}

func objectWord(token Token) bool {
	switch strings.ToLower(token.Text) {
	case "artifact", "card", "creature", "enchantment", "equipment", "land", "permanent", "spell", "token":
		return token.Kind == Word
	default:
		return false
	}
}

func unsupportedDiagnostic(span Span, text string) Diagnostic {
	return Diagnostic{
		Severity: SeverityWarning,
		Summary:  "unsupported Oracle construct",
		Detail:   "the compiler preserved but did not confidently lower: " + text,
		Span:     span,
	}
}
