package parser

import (
	"strconv"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/types"
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
	if operation, next, ok := parseStaticBasePowerToughnessOperation(tokens, index, end, subject); ok {
		return operation, next, true
	}
	if operation, next, ok := parseStaticCharacteristicOperation(tokens, index, end, atoms); ok {
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

// parseStaticBasePowerToughnessOperation recognizes the characteristic-setting
// operation "<group> has/have base power and toughness N/N", where N/N are
// non-negative literal integers. Dynamic forms ("base power and toughness X/X,
// where X is ...") carry trailing tokens and fail closed.
func parseStaticBasePowerToughnessOperation(
	tokens []shared.Token,
	index, end int,
	subject StaticDeclarationSubject,
) (StaticDeclarationSyntax, int, bool) {
	if !staticCharacteristicVerb(tokens, index, subject, "has", "have") {
		return StaticDeclarationSyntax{}, 0, false
	}
	if !staticWordsAt(tokens, index+1, "base", "power", "and", "toughness") || index+8 > end {
		return StaticDeclarationSyntax{}, 0, false
	}
	power, powerOK := staticUnsignedInteger(tokens[index+5])
	toughness, toughnessOK := staticUnsignedInteger(tokens[index+7])
	if !powerOK || tokens[index+6].Kind != shared.Slash || !toughnessOK {
		return StaticDeclarationSyntax{}, 0, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationContinuousBasePowerToughness,
		OperationSpan: shared.SpanOf(tokens[index : index+8]),
		BasePower:     power,
		BaseToughness: toughness,
		BasePTSet:     true,
	}, index + 8, true
}

// parseStaticCharacteristicOperation recognizes the characteristic operations
// "<group> is/are <color>" (sets colors) and "<group> is/are [a/an]
// <color>* <type|subtype>* in addition to its/their other (colors|types|colors
// and types)" (adds colors, card types, and subtypes). Card types and subtypes
// always require the explicit "in addition" tail; bare "is/are <color>" sets the
// affected object's colors.
func parseStaticCharacteristicOperation(
	tokens []shared.Token,
	index, end int,
	atoms Atoms,
) (StaticDeclarationSyntax, int, bool) {
	if !staticWordsAt(tokens, index, "is") && !staticWordsAt(tokens, index, "are") {
		return StaticDeclarationSyntax{}, 0, false
	}
	cursor := index + 1
	if staticWordsAt(tokens, cursor, "a") || staticWordsAt(tokens, cursor, "an") {
		cursor++
	}
	if operation, next, ok := parseStaticAllColorsOperation(tokens, index, cursor, end); ok {
		return operation, next, true
	}
	list, next, ok := parseStaticCharacteristicList(tokens, cursor, end, atoms)
	if !ok {
		return StaticDeclarationSyntax{}, 0, false
	}
	operation := StaticDeclarationSyntax{
		Kind:          StaticDeclarationContinuousCharacteristic,
		OperationSpan: shared.SpanOf(tokens[index:next]),
		Colors:        list.colors,
		CardTypes:     list.cardTypes,
		Subtypes:      list.subtypes,
	}
	tail, tailNext, hasTail := parseStaticInAdditionTail(tokens, next, end)
	if !hasTail {
		// Without an explicit "in addition" tail only a bare color set is
		// representable; type and subtype additions fail closed.
		if len(list.cardTypes) != 0 || len(list.subtypes) != 0 || len(list.colors) == 0 {
			return StaticDeclarationSyntax{}, 0, false
		}
		operation.OperationSpan = shared.SpanOf(tokens[index:next])
		return operation, next, true
	}
	if !staticInAdditionTailMatches(tail, list.colors, list.cardTypes, list.subtypes) {
		return StaticDeclarationSyntax{}, 0, false
	}
	operation.ColorsAdd = len(list.colors) != 0
	operation.OperationSpan = shared.SpanOf(tokens[index:tailNext])
	return operation, tailNext, true
}

// staticAllColors lists every Oracle color in canonical WUBRG order; an
// "<group> is/are all colors" declaration SETS the affected object's colors to
// exactly these five.
var staticAllColors = []Color{ColorWhite, ColorBlue, ColorBlack, ColorRed, ColorGreen}

// parseStaticAllColorsOperation recognizes the bare characteristic-set operation
// "<group> is/are all colors" (CR 105.2c), spanning tokens[index] ("is"/"are")
// through "colors". It SETS the affected object's colors to all five colors. A
// trailing "in addition to ..." tail or any other characteristic word fails
// closed: only the exact "all colors" set is representable here.
func parseStaticAllColorsOperation(
	tokens []shared.Token,
	index, cursor, end int,
) (StaticDeclarationSyntax, int, bool) {
	if !staticWordsAt(tokens, cursor, "all", "colors") || cursor+2 > end {
		return StaticDeclarationSyntax{}, 0, false
	}
	next := cursor + 2
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationContinuousCharacteristic,
		OperationSpan: shared.SpanOf(tokens[index:next]),
		Colors:        append([]Color(nil), staticAllColors...),
	}, next, true
}

// staticInAdditionTail records which characteristic categories an "in addition
// to its/their other ..." tail enumerates.
type staticInAdditionTail struct {
	colors bool
	types  bool
}

// parseStaticInAdditionTail consumes "in addition to its/their other
// (colors|types|colors and types)" beginning at start, returning the enumerated
// categories and the index following the tail.
func parseStaticInAdditionTail(tokens []shared.Token, start, end int) (staticInAdditionTail, int, bool) {
	if !staticWordsAt(tokens, start, "in", "addition", "to") {
		return staticInAdditionTail{}, 0, false
	}
	cursor := start + 3
	if !staticWordsAt(tokens, cursor, "its") && !staticWordsAt(tokens, cursor, "their") {
		return staticInAdditionTail{}, 0, false
	}
	cursor++
	if !staticWordsAt(tokens, cursor, "other") {
		return staticInAdditionTail{}, 0, false
	}
	cursor++
	switch {
	case staticWordsAt(tokens, cursor, "colors", "and", "types"):
		return staticInAdditionTail{colors: true, types: true}, cursor + 3, true
	case staticWordsAt(tokens, cursor, "types", "and", "colors"):
		return staticInAdditionTail{colors: true, types: true}, cursor + 3, true
	case staticWordsAt(tokens, cursor, "colors"):
		return staticInAdditionTail{colors: true}, cursor + 1, true
	case staticWordsAt(tokens, cursor, "types"):
		return staticInAdditionTail{types: true}, cursor + 1, true
	default:
		return staticInAdditionTail{}, 0, false
	}
}

// staticInAdditionTailMatches reports whether the enumerated tail categories are
// exactly consistent with the recognized characteristics: colors require a
// "colors" category, card types and subtypes require a "types" category, and the
// tail may not enumerate a category that the operation did not recognize.
func staticInAdditionTailMatches(tail staticInAdditionTail, colors []Color, cardTypes []CardType, subtypes []types.Sub) bool {
	hasColors := len(colors) != 0
	hasTypes := len(cardTypes) != 0 || len(subtypes) != 0
	return tail.colors == hasColors && tail.types == hasTypes && (hasColors || hasTypes)
}

// staticCharacteristicList holds the colors, card types, and subtypes a
// characteristic operation enumerates, in source order.
type staticCharacteristicList struct {
	colors    []Color
	cardTypes []CardType
	subtypes  []types.Sub
}

// parseStaticCharacteristicList consumes a run of color, card-type, and subtype
// atoms beginning at start, returning them in source order with the index
// following the run. Words that are not a recognized characteristic atom stop
// the run.
func parseStaticCharacteristicList(
	tokens []shared.Token,
	start, end int,
	atoms Atoms,
) (staticCharacteristicList, int, bool) {
	var list staticCharacteristicList
	index := start
	for index < end {
		if color, ok := atoms.ColorAt(tokens[index].Span); ok {
			list.colors = append(list.colors, color)
			index++
			continue
		}
		if cardType, ok := atoms.CardTypeAt(tokens[index].Span); ok {
			list.cardTypes = append(list.cardTypes, cardType)
			index++
			continue
		}
		if subtype, width, ok := staticSubtypeAt(tokens, index, end, atoms); ok {
			list.subtypes = append(list.subtypes, subtype)
			index += width
			continue
		}
		break
	}
	if index == start || len(list.colors)+len(list.cardTypes)+len(list.subtypes) == 0 {
		return staticCharacteristicList{}, start, false
	}
	return list, index, true
}

// staticSubtypeAt returns the subtype atom and token width beginning at index, if
// any. Multi-word subtype phrases occupy a single atom spanning several tokens.
func staticSubtypeAt(tokens []shared.Token, index, end int, atoms Atoms) (types.Sub, int, bool) {
	if index >= end {
		return "", 0, false
	}
	for _, atom := range atoms.Subtypes() {
		if atom.Span.Start.Offset != tokens[index].Span.Start.Offset {
			continue
		}
		width := tokensCoveredCount(tokens[index:], atom.Span)
		if width > 0 && index+width <= end {
			return atom.Identity, width, true
		}
	}
	return "", 0, false
}

// staticCharacteristicVerb reports whether the verb beginning at index is the
// group-appropriate singular or plural verb. Source-tied subjects ("this
// creature", "Enchanted creature") use the singular verb; battlefield groups use
// the plural verb.
func staticCharacteristicVerb(tokens []shared.Token, index int, subject StaticDeclarationSubject, singular, plural string) bool {
	if subject.Kind == StaticDeclarationSubjectGroup && subject.Group.Kind != EffectStaticSubjectAttachedObject {
		return staticWordsAt(tokens, index, plural) || staticWordsAt(tokens, index, singular)
	}
	return staticWordsAt(tokens, index, singular)
}

// staticUnsignedInteger returns the value of a non-negative integer token.
func staticUnsignedInteger(token shared.Token) (int, bool) {
	if token.Kind != shared.Integer {
		return 0, false
	}
	value, err := strconv.Atoi(token.Text)
	if err != nil || value < 0 {
		return 0, false
	}
	return value, true
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
	if !staticRuleSubjectKindAllowed(subject) {
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
		next := verb + 1
		var qualifiers []StaticRuleQualifier
		if qualifier, qualifierNext, ok := parseStaticDefenderYouQualifier(tokens, next, end); ok {
			qualifiers = append(qualifiers, qualifier)
			next = qualifierNext
		}
		return staticRuleOperation(tokens, index, next, subject, constraint, StaticRuleOperation{
			Kind:  StaticRuleOperationAttack,
			Voice: StaticRuleVoiceActive,
			Span:  tokens[verb].Span,
		}, qualifiers)
	}
	if staticWordsAt(tokens, verb, "block") {
		return staticRuleOperation(tokens, index, verb+1, subject, constraint, StaticRuleOperation{
			Kind:  StaticRuleOperationBlock,
			Voice: StaticRuleVoiceActive,
			Span:  tokens[verb].Span,
		}, nil)
	}
	if staticWordsAt(tokens, verb, "be", "blocked") {
		next := verb + 2
		var qualifiers []StaticRuleQualifier
		if qualifier, qualifierNext, ok := parseStaticByMoreThanOneQualifier(tokens, next, end); ok {
			qualifiers = append(qualifiers, qualifier)
			next = qualifierNext
		} else if qualifier, qualifierNext, ok := parseStaticBlockerRestrictionQualifier(tokens, next, end); ok {
			qualifiers = append(qualifiers, qualifier)
			next = qualifierNext
		}
		return staticRuleOperation(tokens, index, next, subject, constraint, StaticRuleOperation{
			Kind:  StaticRuleOperationBlock,
			Voice: StaticRuleVoicePassive,
			Span:  shared.SpanOf(tokens[verb : verb+2]),
		}, qualifiers)
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

// parseStaticDefenderYouQualifier consumes the defender restriction "you or
// planeswalkers you control" that scopes an attack prohibition to the source's
// controller. The phrasing is fixed; any deviation fails closed.
func parseStaticDefenderYouQualifier(tokens []shared.Token, start, end int) (StaticRuleQualifier, int, bool) {
	if start+5 > end || !staticWordsAt(tokens, start, "you", "or", "planeswalkers", "you", "control") {
		return StaticRuleQualifier{}, 0, false
	}
	return StaticRuleQualifier{
		Kind: StaticRuleQualifierDefenderYou,
		Span: shared.SpanOf(tokens[start : start+5]),
	}, start + 5, true
}

// parseStaticByMoreThanOneQualifier consumes the bounded block exception "by
// more than one creature". The phrasing is fixed; any deviation fails closed.
func parseStaticByMoreThanOneQualifier(tokens []shared.Token, start, end int) (StaticRuleQualifier, int, bool) {
	if start+5 > end || !staticWordsAt(tokens, start, "by", "more", "than", "one", "creature") {
		return StaticRuleQualifier{}, 0, false
	}
	return StaticRuleQualifier{
		Kind: StaticRuleQualifierByMoreThanOne,
		Span: shared.SpanOf(tokens[start : start+5]),
	}, start + 5, true
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

// staticRuleSubjectKindAllowed reports whether a composed-declaration subject can
// carry a static rule operation: the source object itself (a creature or spell),
// an ambiguous self-name, or the creature an Aura or Equipment is attached to.
func staticRuleSubjectKindAllowed(subject StaticDeclarationSubject) bool {
	switch subject.Kind {
	case StaticDeclarationSubjectSourceCreature,
		StaticDeclarationSubjectSourceSpell,
		StaticDeclarationSubjectSourceNamed:
		return true
	case StaticDeclarationSubjectGroup:
		return subject.Group.Kind == EffectStaticSubjectAttachedObject
	default:
		return false
	}
}

// staticRuleSubjectForDeclaration derives the typed rule subject from the
// declaration subject and the rule operation. A counter operation requires a
// spell subject; block and attack require a creature subject. An ambiguous
// self-name subject adopts whichever the operation implies, while an explicit
// creature, spell, or attached-creature subject must agree with the operation.
func staticRuleSubjectForDeclaration(subject StaticDeclarationSubject, operation StaticRuleOperation) (StaticRuleSubject, bool) {
	if operation.Kind == StaticRuleOperationCounter {
		switch subject.Kind {
		case StaticDeclarationSubjectSourceSpell, StaticDeclarationSubjectSourceNamed:
			return StaticRuleSubject{Kind: StaticRuleSubjectSourceSpell, Span: subject.Span}, true
		default:
			return StaticRuleSubject{}, false
		}
	}
	switch subject.Kind {
	case StaticDeclarationSubjectSourceCreature, StaticDeclarationSubjectSourceNamed:
		return StaticRuleSubject{Kind: StaticRuleSubjectSourceCreature, Span: subject.Span}, true
	case StaticDeclarationSubjectGroup:
		if subject.Group.Kind == EffectStaticSubjectAttachedObject {
			return StaticRuleSubject{Kind: StaticRuleSubjectAttachedObject, Span: subject.Span}, true
		}
	default:
	}
	return StaticRuleSubject{}, false
}

// parseStaticLoseAbilitiesBecomeDeclaration recognizes the "polymorph" static
// shape printed on Auras and a few creatures: "<subject> loses all abilities"
// optionally followed by "and has base power and toughness N/N" or "and is [a]
// <colors>* [<subtype>] [creature] with base power and toughness N/N". The
// colors, card type, and creature subtype are SET (the affected object loses its
// other colors, card types, and creature types). A name-setting tail ("named
// ..."), a "colorless" body, a non-creature card type, or any other trailing
// text fails closed.
func parseStaticLoseAbilitiesBecomeDeclaration(tokens []shared.Token, atoms Atoms) (StaticDeclarationSyntax, bool) {
	if len(tokens) < 5 || tokens[len(tokens)-1].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	subject, index, ok := parseStaticLoseAbilitiesSubject(tokens, atoms)
	if !ok || !staticWordsAt(tokens, index, "loses", "all", "abilities") {
		return StaticDeclarationSyntax{}, false
	}
	index += 3
	end := len(tokens) - 1
	declaration := StaticDeclarationSyntax{
		Kind:             StaticDeclarationLoseAbilitiesBecome,
		Span:             shared.SpanOf(tokens),
		OperationSpan:    shared.SpanOf(tokens[:end]),
		Subject:          subject,
		LoseAllAbilities: true,
	}
	if index == end {
		return declaration, true
	}
	if !staticWordsAt(tokens, index, "and") {
		return StaticDeclarationSyntax{}, false
	}
	next, ok := parseStaticBecomeTail(tokens, index+1, end, &declaration, atoms)
	if !ok || next != end {
		return StaticDeclarationSyntax{}, false
	}
	return declaration, true
}

// parseStaticLoseAbilitiesSubject recognizes the affected object of a polymorph
// declaration: the creature an Aura or Equipment is attached to ("enchanted
// creature", "equipped creature") or the source creature itself ("this
// creature"). It returns the typed subject and the index following it.
func parseStaticLoseAbilitiesSubject(tokens []shared.Token, atoms Atoms) (StaticDeclarationSubject, int, bool) {
	if staticWordsAt(tokens, 0, "this", "creature") {
		return StaticDeclarationSubject{
			Kind: StaticDeclarationSubjectSourceCreature,
			Span: shared.SpanOf(tokens[:2]),
		}, 2, true
	}
	if staticWordsAt(tokens, 0, "enchanted", "creature") || staticWordsAt(tokens, 0, "equipped", "creature") {
		span := shared.SpanOf(tokens[:2])
		return StaticDeclarationSubject{
			Kind:  StaticDeclarationSubjectGroup,
			Span:  span,
			Group: EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAttachedObject, Span: span},
		}, 2, true
	}
	if span, width, ok := staticSourceSubjectAt(tokens, atoms); ok {
		return StaticDeclarationSubject{
			Kind: StaticDeclarationSubjectSourceNamed,
			Span: span,
		}, width, true
	}
	return StaticDeclarationSubject{}, 0, false
}

// parseStaticBecomeTail consumes the optional "and is/has ..." tail of a
// polymorph declaration, recording the set colors, card type, subtype, and base
// power/toughness on the declaration. It returns the index following the tail.
func parseStaticBecomeTail(tokens []shared.Token, index, end int, declaration *StaticDeclarationSyntax, atoms Atoms) (int, bool) {
	if staticWordsAt(tokens, index, "has") {
		basePT, ok := parseStaticBasePowerToughnessAt(tokens, index+1)
		if !ok {
			return 0, false
		}
		declaration.BasePower = basePT.power
		declaration.BaseToughness = basePT.toughness
		declaration.BasePTSet = true
		return basePT.next, true
	}
	if !staticWordsAt(tokens, index, "is") {
		return 0, false
	}
	cursor := index + 1
	if staticWordsAt(tokens, cursor, "a") || staticWordsAt(tokens, cursor, "an") {
		cursor++
	}
	list, next, ok := parseStaticCharacteristicList(tokens, cursor, end, atoms)
	if !ok {
		return 0, false
	}
	for _, cardType := range list.cardTypes {
		if cardType != CardTypeCreature {
			return 0, false
		}
	}
	declaration.Colors = list.colors
	declaration.CardTypes = list.cardTypes
	declaration.Subtypes = list.subtypes
	if !staticWordsAt(tokens, next, "with") {
		return 0, false
	}
	basePT, ok := parseStaticBasePowerToughnessAt(tokens, next+1)
	if !ok {
		return 0, false
	}
	declaration.BasePower = basePT.power
	declaration.BaseToughness = basePT.toughness
	declaration.BasePTSet = true
	return basePT.next, true
}

// staticBasePowerToughness is the result of matching a "base power and toughness
// N/N" phrase: the two literal values and the token index following the pair.
type staticBasePowerToughness struct {
	power     int
	toughness int
	next      int
}

// parseStaticBasePowerToughnessAt matches "base power and toughness N/N"
// beginning at start, where N/N are non-negative literal integers. It returns
// the two values and the index following the slashed pair.
func parseStaticBasePowerToughnessAt(tokens []shared.Token, start int) (staticBasePowerToughness, bool) {
	if !staticWordsAt(tokens, start, "base", "power", "and", "toughness") || start+6 >= len(tokens) {
		return staticBasePowerToughness{}, false
	}
	power, powerOK := staticUnsignedInteger(tokens[start+4])
	toughness, toughnessOK := staticUnsignedInteger(tokens[start+6])
	if !powerOK || tokens[start+5].Kind != shared.Slash || !toughnessOK {
		return staticBasePowerToughness{}, false
	}
	return staticBasePowerToughness{power: power, toughness: toughness, next: start + 7}, true
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
