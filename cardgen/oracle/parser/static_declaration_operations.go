package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func parseStaticSubjectDeclarations(
	tokens []shared.Token,
	atoms Atoms,
	conditions []ConditionClause,
) ([]StaticDeclarationSyntax, bool) {
	if len(tokens) < 3 || tokens[len(tokens)-1].Kind != shared.Period {
		return nil, false
	}
	opTokens, condition, hasCondition := staticOperationTokens(tokens, conditions)
	if len(opTokens) < 3 || opTokens[len(opTokens)-1].Kind != shared.Period {
		return nil, false
	}
	subject, verbStart, ok := parseStaticDeclarationSubject(opTokens, atoms)
	if !ok {
		return nil, false
	}
	operations, ok := parseStaticOperations(opTokens, verbStart, subject, atoms)
	if !ok {
		return nil, false
	}
	span := shared.SpanOf(tokens)
	for i := range operations {
		operations[i].Span = span
		operations[i].Subject = subject
		if hasCondition {
			operations[i].HasCondition = true
			operations[i].ConditionSpan = condition.Span
		}
	}
	return operations, true
}

func parseStaticDeclarationSubject(tokens []shared.Token, atoms Atoms) (StaticDeclarationSubject, int, bool) {
	if staticWordsAt(tokens, 0, "this", "creature") {
		return StaticDeclarationSubject{
			Kind: StaticDeclarationSubjectSourceCreature,
			Span: shared.SpanOf(tokens[:2]),
		}, 2, true
	}
	if staticWordsAt(tokens, 0, "this", "spell") {
		return StaticDeclarationSubject{
			Kind: StaticDeclarationSubjectSourceSpell,
			Span: shared.SpanOf(tokens[:2]),
		}, 2, true
	}
	if span, width, ok := staticSourceSubjectAt(tokens, atoms); ok {
		return StaticDeclarationSubject{
			Kind: StaticDeclarationSubjectSourceNamed,
			Span: span,
		}, width, true
	}
	group := parseEffectStaticSubject(tokens, atoms)
	if group.Kind == EffectStaticSubjectNone {
		return StaticDeclarationSubject{}, 0, false
	}
	verbStart := tokensCoveredCount(tokens, group.Span)
	if verbStart == 0 {
		return StaticDeclarationSubject{}, 0, false
	}
	return StaticDeclarationSubject{
		Kind:  StaticDeclarationSubjectGroup,
		Span:  group.Span,
		Group: group,
	}, verbStart, true
}

// staticSourceSubjectAt returns the span and token width of a source-marker
// ("this <marker>") or self-name subject phrase beginning at tokens[0].
func staticSourceSubjectAt(tokens []shared.Token, atoms Atoms) (shared.Span, int, bool) {
	if len(tokens) == 0 {
		return shared.Span{}, 0, false
	}
	spans := append(append([]shared.Span(nil), atoms.SourceMarkerSpans()...), atoms.SourceNameSpans()...)
	for _, span := range spans {
		if span.Start.Offset != tokens[0].Span.Start.Offset {
			continue
		}
		width := tokensCoveredCount(tokens, span)
		if width > 0 {
			return span, width, true
		}
	}
	return shared.Span{}, 0, false
}

func parseStaticOperations(
	tokens []shared.Token,
	start int,
	subject StaticDeclarationSubject,
	atoms Atoms,
) ([]StaticDeclarationSyntax, bool) {
	end := len(tokens) - 1
	var operations []StaticDeclarationSyntax
	index := start
	lastConnectorHadAnd := false
	for index < end {
		if len(operations) > 0 {
			next, hadAnd, ok := consumeStaticConnector(tokens, index, end)
			if !ok {
				return nil, false
			}
			lastConnectorHadAnd = hadAnd
			index = next
		}
		operation, next, ok := parseStaticOperation(tokens, index, end, subject, atoms)
		if !ok {
			return nil, false
		}
		operations = append(operations, operation)
		index = next
	}
	if len(operations) == 0 {
		return nil, false
	}
	// A multi-operation sequence must join its final operation with "and"
	// ("A and B", "A, B, and C"); a bare comma alone is not a sentence-level
	// conjunction and fails closed.
	if len(operations) >= 2 && !lastConnectorHadAnd {
		return nil, false
	}
	return operations, true
}

func consumeStaticConnector(tokens []shared.Token, index, end int) (next int, hadAnd, ok bool) {
	consumed := false
	if index < end && tokens[index].Kind == shared.Comma {
		index++
		consumed = true
	}
	if index < end && staticWordsAt(tokens, index, "and") {
		index++
		consumed = true
		hadAnd = true
	}
	if !consumed || index >= end {
		return 0, false, false
	}
	return index, hadAnd, true
}

func parseStaticOperation(
	tokens []shared.Token,
	index, end int,
	subject StaticDeclarationSubject,
	atoms Atoms,
) (StaticDeclarationSyntax, int, bool) {
	if operation, next, ok := parseStaticPowerToughnessOperation(tokens, index, end, subject); ok {
		return operation, next, true
	}
	if operation, next, ok := parseStaticKeywordGrantOperation(tokens, index, end, atoms); ok {
		return operation, next, true
	}
	if operation, next, ok := parseStaticRuleOperation(tokens, index, end, subject); ok {
		return operation, next, true
	}
	return StaticDeclarationSyntax{}, 0, false
}

func parseStaticPowerToughnessOperation(
	tokens []shared.Token,
	index, end int,
	subject StaticDeclarationSubject,
) (StaticDeclarationSyntax, int, bool) {
	if !staticPowerToughnessVerb(tokens, index, subject) || index+6 > end {
		return StaticDeclarationSyntax{}, 0, false
	}
	power, powerOK := parseSignedAmount(tokens[index+1], tokens[index+2])
	toughness, toughnessOK := parseSignedAmount(tokens[index+4], tokens[index+5])
	if !powerOK || tokens[index+3].Kind != shared.Slash || !toughnessOK {
		return StaticDeclarationSyntax{}, 0, false
	}
	operation := StaticDeclarationSyntax{
		Kind:           StaticDeclarationContinuousPowerToughness,
		OperationSpan:  tokens[index].Span,
		PowerDelta:     power,
		ToughnessDelta: toughness,
	}
	next := index + 6
	if next < end {
		if _, _, ok := consumeStaticConnector(tokens, next, end); ok {
			return operation, next, true
		}
		if !staticDynamicAmountTail(tokens, next) {
			return StaticDeclarationSyntax{}, 0, false
		}
		operation.Dynamic = true
		return operation, end, true
	}
	return operation, next, true
}

// staticDynamicAmountTail reports whether the tokens beginning at start open a
// recognized dynamic-amount tail ("for each ..." or "equal to ...") that scales
// a power/toughness change. Any other trailing tokens fail closed.
func staticDynamicAmountTail(tokens []shared.Token, start int) bool {
	return staticWordsAt(tokens, start, "for", "each") ||
		staticWordsAt(tokens, start, "equal", "to")
}

func staticPowerToughnessVerb(tokens []shared.Token, index int, subject StaticDeclarationSubject) bool {
	if subject.Kind == StaticDeclarationSubjectGroup {
		return staticWordsAt(tokens, index, "get") || staticWordsAt(tokens, index, "gets")
	}
	return staticWordsAt(tokens, index, "gets")
}

func parseStaticKeywordGrantOperation(
	tokens []shared.Token,
	index, end int,
	atoms Atoms,
) (StaticDeclarationSyntax, int, bool) {
	if !staticWordsAt(tokens, index, "has") && !staticWordsAt(tokens, index, "have") {
		return StaticDeclarationSyntax{}, 0, false
	}
	spans, next, ok := parseStaticKeywordList(tokens, index+1, end, atoms)
	if !ok {
		return StaticDeclarationSyntax{}, 0, false
	}
	operationSpan := spans[0]
	operationSpan.End = spans[len(spans)-1].End
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationKeywordGrant,
		OperationSpan: operationSpan,
		KeywordSpans:  spans,
	}, next, true
}

func parseStaticKeywordList(tokens []shared.Token, start, end int, atoms Atoms) ([]shared.Span, int, bool) {
	var spans []shared.Span
	index := start
	for index < end {
		keyword, width, ok := staticKeywordAt(tokens, index, end, atoms)
		if !ok {
			break
		}
		spans = append(spans, keyword.Span)
		next := index + width
		separator := next
		if separator < end && tokens[separator].Kind == shared.Comma {
			separator++
		}
		if separator < end && staticWordsAt(tokens, separator, "and") {
			separator++
		}
		if separator > next {
			if _, _, ok := staticKeywordAt(tokens, separator, end, atoms); ok {
				index = separator
				continue
			}
		}
		index = next
		break
	}
	if len(spans) == 0 {
		return nil, start, false
	}
	return spans, index, true
}

func staticKeywordAt(tokens []shared.Token, index, end int, atoms Atoms) (Keyword, int, bool) {
	if index >= end {
		return Keyword{}, 0, false
	}
	for _, keyword := range atoms.Keywords() {
		if keyword.NameSpan.Start.Offset != tokens[index].Span.Start.Offset {
			continue
		}
		width := tokensCoveredCount(tokens[index:], keyword.Span)
		if width > 0 && index+width <= end {
			return keyword, width, true
		}
	}
	return Keyword{}, 0, false
}

func parseStaticRuleOperation(
	tokens []shared.Token,
	index, end int,
	subject StaticDeclarationSubject,
) (StaticDeclarationSyntax, int, bool) {
	if subject.Kind != StaticDeclarationSubjectSourceCreature &&
		subject.Kind != StaticDeclarationSubjectSourceSpell &&
		subject.Kind != StaticDeclarationSubjectSourceNamed {
		return StaticDeclarationSyntax{}, 0, false
	}
	if staticWordsAt(tokens, index, "can't") || staticWordsAt(tokens, index, "cannot") {
		return parseStaticProhibitionRuleOperation(tokens, index, end, subject)
	}
	if rule, next, ok := parseStaticAttackRuleOperation(tokens, index, end, subject); ok {
		return rule, next, true
	}
	if rule, next, ok := parseStaticRequiredBlockRuleOperation(tokens, index, end, subject); ok {
		return rule, next, true
	}
	return StaticDeclarationSyntax{}, 0, false
}

func parseStaticProhibitionRuleOperation(
	tokens []shared.Token,
	index, end int,
	subject StaticDeclarationSubject,
) (StaticDeclarationSyntax, int, bool) {
	constraint := StaticRuleConstraint{Kind: StaticRuleConstraintProhibition, Span: tokens[index].Span}
	verb := index + 1
	if staticWordsAt(tokens, verb, "attack") {
		return staticRuleOperation(tokens, index, verb+1, subject, constraint, StaticRuleOperation{
			Kind:  StaticRuleOperationAttack,
			Voice: StaticRuleVoiceActive,
			Span:  tokens[verb].Span,
		}, nil)
	}
	if staticWordsAt(tokens, verb, "block") {
		return staticRuleOperation(tokens, index, verb+1, subject, constraint, StaticRuleOperation{
			Kind:  StaticRuleOperationBlock,
			Voice: StaticRuleVoiceActive,
			Span:  tokens[verb].Span,
		}, nil)
	}
	if staticWordsAt(tokens, verb, "be", "blocked") {
		return staticRuleOperation(tokens, index, verb+2, subject, constraint, StaticRuleOperation{
			Kind:  StaticRuleOperationBlock,
			Voice: StaticRuleVoicePassive,
			Span:  shared.SpanOf(tokens[verb : verb+2]),
		}, nil)
	}
	if staticWordsAt(tokens, verb, "be", "countered") {
		return staticRuleOperation(tokens, index, verb+2, subject, constraint, StaticRuleOperation{
			Kind:  StaticRuleOperationCounter,
			Voice: StaticRuleVoicePassive,
			Span:  shared.SpanOf(tokens[verb : verb+2]),
		}, nil)
	}
	return StaticDeclarationSyntax{}, 0, false
}

func parseStaticAttackRuleOperation(
	tokens []shared.Token,
	index, end int,
	subject StaticDeclarationSubject,
) (StaticDeclarationSyntax, int, bool) {
	constraintStart := index
	operationStart := index
	if staticWordsAt(tokens, index, "must") {
		operationStart++
	}
	explicit := operationStart != constraintStart
	if (explicit && !staticWordsAt(tokens, operationStart, "attack")) ||
		(!explicit && !staticWordsAt(tokens, operationStart, "attacks")) {
		return StaticDeclarationSyntax{}, 0, false
	}
	qualifierStart := operationStart + 1
	if !staticWordsAt(tokens, qualifierStart, "each", "combat", "if", "able") ||
		qualifierStart+4 > end {
		return StaticDeclarationSyntax{}, 0, false
	}
	constraintSpan := shared.SpanOf(tokens[constraintStart : qualifierStart+4])
	if explicit {
		constraintSpan = tokens[constraintStart].Span
	}
	qualifiers := []StaticRuleQualifier{
		{Kind: StaticRuleQualifierEachCombat, Span: shared.SpanOf(tokens[qualifierStart : qualifierStart+2])},
		{Kind: StaticRuleQualifierIfAble, Span: shared.SpanOf(tokens[qualifierStart+2 : qualifierStart+4])},
	}
	return staticRuleOperation(tokens, index, qualifierStart+4, subject,
		StaticRuleConstraint{Kind: StaticRuleConstraintRequirement, Span: constraintSpan},
		StaticRuleOperation{Kind: StaticRuleOperationAttack, Voice: StaticRuleVoiceActive, Span: tokens[operationStart].Span},
		qualifiers,
	)
}

func parseStaticRequiredBlockRuleOperation(
	tokens []shared.Token,
	index, end int,
	subject StaticDeclarationSubject,
) (StaticDeclarationSyntax, int, bool) {
	if !staticWordsAt(tokens, index, "must", "be", "blocked", "if", "able") ||
		index+5 > end {
		return StaticDeclarationSyntax{}, 0, false
	}
	qualifiers := []StaticRuleQualifier{
		{Kind: StaticRuleQualifierIfAble, Span: shared.SpanOf(tokens[index+3 : index+5])},
	}
	return staticRuleOperation(tokens, index, index+5, subject,
		StaticRuleConstraint{Kind: StaticRuleConstraintRequirement, Span: tokens[index].Span},
		StaticRuleOperation{Kind: StaticRuleOperationBlock, Voice: StaticRuleVoicePassive, Span: shared.SpanOf(tokens[index+1 : index+3])},
		qualifiers,
	)
}

func staticRuleOperation(
	tokens []shared.Token,
	start, next int,
	subject StaticDeclarationSubject,
	constraint StaticRuleConstraint,
	operation StaticRuleOperation,
	qualifiers []StaticRuleQualifier,
) (StaticDeclarationSyntax, int, bool) {
	ruleSubject, ok := staticRuleSubjectForDeclaration(subject, operation)
	if !ok {
		return StaticDeclarationSyntax{}, 0, false
	}
	rule := StaticRuleSyntax{
		Span:       shared.SpanOf(tokens[start:next]),
		Subject:    ruleSubject,
		Constraint: constraint,
		Operation:  operation,
		Qualifiers: qualifiers,
	}
	if !validStaticRuleSyntax(rule) {
		return StaticDeclarationSyntax{}, 0, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationRule,
		OperationSpan: operation.Span,
		Rule:          rule,
	}, next, true
}

// staticRuleSubjectForDeclaration derives the typed rule subject from the
// declaration subject and the rule operation. A counter operation requires a
// spell subject; block and attack require a creature subject. An ambiguous
// self-name subject adopts whichever the operation implies, while an explicit
// creature or spell subject must agree with the operation.
func staticRuleSubjectForDeclaration(subject StaticDeclarationSubject, operation StaticRuleOperation) (StaticRuleSubject, bool) {
	kind := StaticRuleSubjectSourceCreature
	if operation.Kind == StaticRuleOperationCounter {
		kind = StaticRuleSubjectSourceSpell
	}
	switch subject.Kind {
	case StaticDeclarationSubjectSourceCreature:
		if kind != StaticRuleSubjectSourceCreature {
			return StaticRuleSubject{}, false
		}
	case StaticDeclarationSubjectSourceSpell:
		if kind != StaticRuleSubjectSourceSpell {
			return StaticRuleSubject{}, false
		}
	case StaticDeclarationSubjectSourceNamed:
	default:
		return StaticRuleSubject{}, false
	}
	return StaticRuleSubject{Kind: kind, Span: subject.Span}, true
}

func tokensCoveredCount(tokens []shared.Token, span shared.Span) int {
	count := 0
	for count < len(tokens) && spanCovers(span, tokens[count].Span) {
		count++
	}
	return count
}

func staticWordsAt(tokens []shared.Token, start int, words ...string) bool {
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
