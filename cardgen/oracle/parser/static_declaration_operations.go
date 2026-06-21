package parser

import (
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/types"
)

// parseStaticEnchantedTypeChangeDeclaration recognizes the removal-Aura static
// "<attached subject> is [a/an] [colorless] <characteristics> [with
// '<granted mana ability>' | with base power and toughness N/N] [and [it] loses
// all [other] [card types and] abilities[, card types,] [and creature types]]."
// The card types and creature subtypes are SET (replacing the enchanted
// permanent's printed types). A leading "colorless" makes it colorless. The
// optional quoted ability is granted, an optional "with base power and toughness
// N/N" rider sets the affected object's base power and toughness, and the
// optional lose-clause strips the permanent's other abilities.
func parseStaticEnchantedTypeChangeDeclaration(tokens []shared.Token, quoted []Delimited, atoms Atoms) (StaticDeclarationSyntax, bool) {
	if len(tokens) < 4 || tokens[len(tokens)-1].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	subject, index, ok := parseStaticAttachedPermanentSubject(tokens)
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	end := len(tokens) - 1
	if !staticWordsAt(tokens, index, "is") {
		return StaticDeclarationSyntax{}, false
	}
	index++
	if staticWordsAt(tokens, index, "a") || staticWordsAt(tokens, index, "an") {
		index++
	}
	declaration := StaticDeclarationSyntax{
		Kind:          StaticDeclarationEnchantedTypeChange,
		Span:          shared.SpanOf(tokens),
		OperationSpan: shared.SpanOf(tokens[:end]),
		Subject:       subject,
	}
	if staticWordsAt(tokens, index, "colorless") {
		declaration.BecomeColorless = true
		index++
	}
	if list, next, ok := parseStaticCharacteristicList(tokens, index, end, atoms); ok {
		declaration.Colors = list.colors
		declaration.CardTypes = list.cardTypes
		declaration.Subtypes = list.subtypes
		index = next
	}
	if !declaration.BecomeColorless &&
		len(declaration.Colors)+len(declaration.CardTypes)+len(declaration.Subtypes) == 0 {
		return StaticDeclarationSyntax{}, false
	}
	if staticWordsAt(tokens, index, "with") {
		if basePT, ok := parseStaticBasePowerToughnessAt(tokens, index+1); ok {
			declaration.BasePower = basePT.power
			declaration.BaseToughness = basePT.toughness
			declaration.BasePTSet = true
			index = basePT.next
		} else {
			if len(quoted) != 1 {
				return StaticDeclarationSyntax{}, false
			}
			ability, ok := parseStaticGrantedManaAbility(quoted[0])
			if !ok {
				return StaticDeclarationSyntax{}, false
			}
			declaration.GrantedManaAbility = &ability
			index++
		}
	}
	if index < end {
		next, ok := parseStaticEnchantedLoseAbilitiesTail(tokens, index, end)
		if !ok {
			return StaticDeclarationSyntax{}, false
		}
		declaration.LoseAllAbilities = true
		index = next
	}
	if index != end {
		return StaticDeclarationSyntax{}, false
	}
	return declaration, true
}

// parseStaticAttachedPermanentSubject recognizes the affected object of a removal
// Aura: the permanent, creature, land, artifact, enchantment, or planeswalker an
// Aura is attached to ("Enchanted permanent", "Enchanted creature", ...). All
// nouns map to the same attached-object group.
func parseStaticAttachedPermanentSubject(tokens []shared.Token) (StaticDeclarationSubject, int, bool) {
	if len(tokens) < 2 || !staticWordsAt(tokens, 0, "enchanted") {
		return StaticDeclarationSubject{}, 0, false
	}
	switch {
	case staticWordsAt(tokens, 1, "permanent"),
		staticWordsAt(tokens, 1, "creature"),
		staticWordsAt(tokens, 1, "land"),
		staticWordsAt(tokens, 1, "artifact"),
		staticWordsAt(tokens, 1, "enchantment"),
		staticWordsAt(tokens, 1, "planeswalker"):
	default:
		return StaticDeclarationSubject{}, 0, false
	}
	span := shared.SpanOf(tokens[:2])
	return StaticDeclarationSubject{
		Kind:  StaticDeclarationSubjectGroup,
		Span:  span,
		Group: EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAttachedObject, Span: span},
	}, 2, true
}

// parseStaticEnchantedLoseAbilitiesTail consumes the trailing lose-clause of a
// removal Aura: an optional comma, "and", optional "it", then "loses all
// [other]" followed by any combination of "abilities", "card types", and
// "creature types" (the card types and creature types are already SET by the
// body). The clause must include "abilities". It returns the index following the
// clause.
func parseStaticEnchantedLoseAbilitiesTail(tokens []shared.Token, index, end int) (int, bool) {
	cursor := index
	if cursor < end && tokens[cursor].Kind == shared.Comma {
		cursor++
	}
	if !staticWordsAt(tokens, cursor, "and") {
		return 0, false
	}
	cursor++
	if staticWordsAt(tokens, cursor, "it") {
		cursor++
	}
	if !staticWordsAt(tokens, cursor, "loses", "all") {
		return 0, false
	}
	cursor += 2
	if staticWordsAt(tokens, cursor, "other") {
		cursor++
	}
	sawAbilities := false
	for cursor < end {
		switch {
		case tokens[cursor].Kind == shared.Comma:
			cursor++
		case staticWordsAt(tokens, cursor, "and"):
			cursor++
		case staticWordsAt(tokens, cursor, "other"):
			cursor++
		case staticWordsAt(tokens, cursor, "abilities"):
			sawAbilities = true
			cursor++
		case staticWordsAt(tokens, cursor, "card", "types"):
			cursor += 2
		case staticWordsAt(tokens, cursor, "creature", "types"):
			cursor += 2
		default:
			return 0, false
		}
	}
	if !sawAbilities {
		return 0, false
	}
	return cursor, true
}

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

// parseStaticQuotedAbilityGrantDeclarations recognizes the static grant
// "<subject> [gets +X/+Y] [and] [has <keyword>] [and] has '<quoted ability>'."
// in which a permanent (the equipped/enchanted creature, or a controlled group)
// is granted a full quoted triggered or activated ability. Because the quoted
// ability and its terminating period are removed from the semantic token stream
// before static declarations are parsed, the residual body ends in a dangling
// connector ("and", "has", or "have") rather than a period; this recognizer
// detects that residual shape, parses any leading power/toughness or keyword
// operations, and appends the quoted-ability grant. The quoted body is parsed
// once through the same pipeline so downstream layers lower it from typed data.
func parseStaticQuotedAbilityGrantDeclarations(
	tokens []shared.Token,
	quoted []Delimited,
	atoms Atoms,
	conditions []ConditionClause,
) ([]StaticDeclarationSyntax, bool) {
	if len(quoted) != 1 {
		return nil, false
	}
	opTokens, condition, hasCondition := staticOperationTokens(tokens, conditions)
	if len(opTokens) < 3 {
		return nil, false
	}
	subject, verbStart, ok := parseStaticDeclarationSubject(opTokens, atoms)
	if !ok || !staticQuotedGrantSubjectSupported(subject) {
		return nil, false
	}
	leadingEnd, ok := staticQuotedGrantLeadingEnd(opTokens, verbStart)
	if !ok {
		return nil, false
	}
	operations, ok := staticQuotedGrantLeadingOperations(opTokens, verbStart, leadingEnd, subject, atoms)
	if !ok {
		return nil, false
	}
	grant, ok := parseStaticGrantedAbility(quoted[0])
	if !ok {
		return nil, false
	}
	operations = append(operations, StaticDeclarationSyntax{
		Kind:           StaticDeclarationContinuousQuotedAbilityGrant,
		OperationSpan:  quoted[0].Span,
		GrantedAbility: &grant,
	})
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

// staticQuotedGrantSubjectSupported reports whether subject is one the quoted
// ability grant supports: the attached object of an Equipment/Aura ("Equipped
// creature", "Enchanted creature") or a controlled-permanent group.
func staticQuotedGrantSubjectSupported(subject StaticDeclarationSubject) bool {
	if subject.Kind != StaticDeclarationSubjectGroup {
		return false
	}
	switch subject.Group.Kind {
	case EffectStaticSubjectAttachedObject,
		EffectStaticSubjectControlledCreatures,
		EffectStaticSubjectOtherControlledCreatures,
		EffectStaticSubjectControlledPermanents:
		return true
	default:
		return false
	}
}

// staticQuotedGrantLeadingEnd returns the exclusive end index of the leading
// power/toughness and keyword operations that precede the quoted ability grant,
// stripping the dangling connector ("and", "has", or "have") the quoted-text
// removal leaves behind. It fails closed when the residual body does not end in
// one of those connectors.
func staticQuotedGrantLeadingEnd(opTokens []shared.Token, verbStart int) (int, bool) {
	n := len(opTokens)
	if n == 0 {
		return 0, false
	}
	last := opTokens[n-1]
	switch {
	case equalWord(last, "has") || equalWord(last, "have"):
		end := n - 1
		if end > verbStart && equalWord(opTokens[end-1], "and") {
			end--
		}
		return end, true
	case equalWord(last, "and"):
		return n - 1, true
	default:
		return 0, false
	}
}

// staticQuotedGrantLeadingOperations parses the leading operations of a quoted
// ability grant (between verbStart and leadingEnd). When there are no leading
// operations it returns an empty slice; otherwise it synthesizes a sentence
// period and reuses parseStaticOperations so the leading power/toughness and
// keyword grants parse identically to a standalone declaration.
func staticQuotedGrantLeadingOperations(
	opTokens []shared.Token,
	verbStart, leadingEnd int,
	subject StaticDeclarationSubject,
	atoms Atoms,
) ([]StaticDeclarationSyntax, bool) {
	if leadingEnd <= verbStart {
		return nil, true
	}
	leadTokens := make([]shared.Token, 0, leadingEnd+1)
	leadTokens = append(leadTokens, opTokens[:leadingEnd]...)
	// Synthesize a sentence-terminating period strictly past the last operation
	// token so parseStaticOperations sees a terminator. Its offset must lie
	// beyond the final token's span so span-coverage checks (e.g. keyword atom
	// lookups) do not mistake the period for part of the preceding operation.
	endPos := opTokens[leadingEnd-1].Span.End
	endPos.Offset++
	endPos.Column++
	leadTokens = append(leadTokens, shared.Token{
		Kind: shared.Period,
		Text: ".",
		Span: shared.Span{Start: endPos, End: endPos},
	})
	return parseStaticOperations(leadTokens, verbStart, subject, atoms)
}

// parseStaticGrantedAbility parses a quoted full ability body into a typed
// granted-ability syntax, running the quoted text (with its surrounding quotes
// removed) through the same pipeline so downstream layers lower it from typed
// data instead of re-parsing its Oracle wording.
func parseStaticGrantedAbility(quoted Delimited) (StaticGrantedAbilitySyntax, bool) {
	tokens := quoted.Tokens
	if len(tokens) < 3 ||
		tokens[0].Kind != shared.Quote ||
		tokens[len(tokens)-1].Kind != shared.Quote {
		return StaticGrantedAbilitySyntax{}, false
	}
	text := staticGrantedAbilityText(quoted)
	document, diagnostics := Parse(text, Context{})
	if len(document.Abilities) != 1 {
		return StaticGrantedAbilitySyntax{}, false
	}
	return StaticGrantedAbilitySyntax{
		Span:        quoted.Span,
		Text:        text,
		document:    document,
		diagnostics: diagnostics,
	}, true
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
	if span, verbStart, ok := staticAllLandsSubject(tokens); ok {
		return StaticDeclarationSubject{
			Kind:  StaticDeclarationSubjectGroup,
			Span:  span,
			Group: EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAllLands, Span: span},
		}, verbStart, true
	}
	if span, verbStart, ok := staticAttachedObjectSubject(tokens); ok {
		return StaticDeclarationSubject{
			Kind:  StaticDeclarationSubjectGroup,
			Span:  span,
			Group: EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAttachedObject, Span: span},
		}, verbStart, true
	}
	if group, verbStart, ok := staticLinkingVerbGroupSubject(tokens); ok {
		return StaticDeclarationSubject{
			Kind:  StaticDeclarationSubjectGroup,
			Span:  group.Span,
			Group: group,
		}, verbStart, true
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

// staticLinkingVerbGroupSubject recognizes a battlefield-group subject that a
// characteristic-defining static joins to its predicate with the linking verb
// "is"/"are" ("Creatures you control are Slivers ...", "All creatures are ...").
// The shared parseEffectStaticSubject only delimits these groups before an
// action verb (get/have/gain/lose), so the linking-verb forms used by type- and
// color-adding statics are recognized here. It returns the group subject and the
// index of the linking verb that follows the noun phrase.
func staticLinkingVerbGroupSubject(tokens []shared.Token) (EffectStaticSubjectSyntax, int, bool) {
	type groupForm struct {
		words []string
		kind  EffectStaticSubjectKind
	}
	forms := []groupForm{
		{[]string{"other", "creatures", "you", "control"}, EffectStaticSubjectOtherControlledCreatures},
		{[]string{"creatures", "you", "control"}, EffectStaticSubjectControlledCreatures},
		{[]string{"permanents", "you", "control"}, EffectStaticSubjectControlledPermanents},
		{[]string{"all", "other", "creatures"}, EffectStaticSubjectAllOtherCreatures},
		{[]string{"all", "creatures"}, EffectStaticSubjectAllCreatures},
	}
	for _, form := range forms {
		width := len(form.words)
		if !staticWordsAt(tokens, 0, form.words...) || len(tokens) <= width {
			continue
		}
		if !staticLinkingVerb(tokens[width]) {
			continue
		}
		return EffectStaticSubjectSyntax{Kind: form.kind, Span: shared.SpanOf(tokens[:width])}, width, true
	}
	return EffectStaticSubjectSyntax{}, 0, false
}

// staticLinkingVerb reports whether token is the copular verb ("is"/"are") that
// joins a characteristic-defining group subject to its predicate.
func staticLinkingVerb(token shared.Token) bool {
	return equalWord(token, "is") || equalWord(token, "are")
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

// staticAllLandsSubject recognizes the battlefield-wide land subject of a
// continuous land-type-adding static: "Each land is ..." (singular verb) or
// "All lands are ..." (plural verb). It returns the subject span and the index
// of the verb that follows. Any other leading words fail closed so only these
// two exact group phrasings map onto the all-lands group.
func staticAllLandsSubject(tokens []shared.Token) (shared.Span, int, bool) {
	switch {
	case staticWordsAt(tokens, 0, "each", "land", "is"):
		return shared.SpanOf(tokens[:2]), 2, true
	case staticWordsAt(tokens, 0, "all", "lands", "are"):
		return shared.SpanOf(tokens[:2]), 2, true
	default:
		return shared.Span{}, 0, false
	}
}

// staticAttachedObjectSubject recognizes the attached-creature subject an Aura
// or Equipment continuous static applies to ("Equipped creature ...", "Enchanted
// creature ..."). Unlike parseEffectStaticSubject, which only resolves this
// subject when a "has"/"gets" verb follows, this accepts any following operation
// (a prohibition rule, a keyword grant, etc.) so a multi-operation grant such as
// "Equipped creature can't be blocked and has shroud." parses as one declaration
// sequence sharing the attached-object subject. It returns the subject span and
// the index of the verb that follows.
func staticAttachedObjectSubject(tokens []shared.Token) (shared.Span, int, bool) {
	if len(tokens) < 3 {
		return shared.Span{}, 0, false
	}
	if !staticWordsAt(tokens, 0, "equipped", "creature") &&
		!staticWordsAt(tokens, 0, "enchanted", "creature") {
		return shared.Span{}, 0, false
	}
	return shared.SpanOf(tokens[:2]), 2, true
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
	if operation, next, ok := parseStaticEntryChoiceSubtypeOperation(tokens, index, end, subject); ok {
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

func parseStaticEntryChoiceSubtypeOperation(
	tokens []shared.Token,
	index, end int,
	subject StaticDeclarationSubject,
) (StaticDeclarationSyntax, int, bool) {
	const width = 10
	if subject.Kind != StaticDeclarationSubjectSourceCreature ||
		index+width != end ||
		!staticWordsAt(tokens, index,
			"is", "the", "chosen", "type", "in", "addition", "to", "its", "other", "types") {
		return StaticDeclarationSyntax{}, 0, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationContinuousEntryChoiceSubtype,
		OperationSpan: shared.SpanOf(tokens[index:end]),
	}, end, true
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

// parseStaticDynamicPowerToughnessOperation recognizes the characteristic-
// defining operation "<source>'s power and toughness are each equal to
// <count>", where <count> is a supported rules-derived count. The subject must
// be the source object (the card's own name or "this creature"); group subjects
// fail closed. The leading possessive ("'s") follows the source subject the
// caller already consumed.
func parseStaticDynamicPowerToughnessOperation(
	tokens []shared.Token,
	index, end int,
	subject StaticDeclarationSubject,
) (StaticDeclarationSyntax, int, bool) {
	if subject.Kind != StaticDeclarationSubjectSourceCreature &&
		subject.Kind != StaticDeclarationSubjectSourceNamed {
		return StaticDeclarationSyntax{}, 0, false
	}
	if index+1 >= end || tokens[index].Kind != shared.Apostrophe || !staticWordsAt(tokens, index+1, "s") {
		return StaticDeclarationSyntax{}, 0, false
	}
	cursor := index + 2
	if !staticWordsAt(tokens, cursor, "power", "and", "toughness", "are", "each", "equal", "to") {
		return StaticDeclarationSyntax{}, 0, false
	}
	cursor += 7
	value, next, ok := parseStaticDynamicValueCount(tokens, cursor, end)
	if !ok || next != end {
		return StaticDeclarationSyntax{}, 0, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationCharacteristicDefiningPowerToughness,
		OperationSpan: shared.SpanOf(tokens[index:next]),
		DynamicValue:  value,
	}, next, true
}

// parseCharacteristicDefiningPowerToughnessDeclaration recognizes a
// characteristic-defining ability that sets the source object's power and
// toughness equal to a rules-derived count ("<source>'s power and toughness are
// each equal to the number of cards in your hand"). The subject must be the
// source object: the card's own possessive name or "this creature's"/"this
// permanent's". Other subjects (an enchanted or equipped creature) fail closed
// because the runtime models this as the source's printed characteristic only.
func parseCharacteristicDefiningPowerToughnessDeclaration(tokens []shared.Token, atoms Atoms) (StaticDeclarationSyntax, bool) {
	if len(tokens) < 9 || tokens[len(tokens)-1].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	subjectSpan, next, ok := characteristicDefiningSourceSubject(tokens, atoms)
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, next, "power", "and", "toughness", "are", "each", "equal", "to") {
		return StaticDeclarationSyntax{}, false
	}
	value, end, ok := parseStaticDynamicValueCount(tokens, next+7, len(tokens)-1)
	if !ok || end != len(tokens)-1 {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationCharacteristicDefiningPowerToughness,
		Span:          shared.SpanOf(tokens),
		OperationSpan: shared.SpanOf(tokens[next:end]),
		Subject: StaticDeclarationSubject{
			Kind: StaticDeclarationSubjectSourceCreature,
			Span: subjectSpan,
		},
		DynamicValue: value,
	}, true
}

// characteristicDefiningSourceSubject recognizes the possessive source subject
// of a characteristic-defining power/toughness declaration, returning the
// subject span and the index of the first operation token. The card's own name
// (whose self-name span includes the trailing possessive) and the
// "this creature's"/"this permanent's" markers name the source object.
func characteristicDefiningSourceSubject(tokens []shared.Token, atoms Atoms) (shared.Span, int, bool) {
	if len(tokens) >= 2 && equalWord(tokens[0], "this") &&
		(strings.EqualFold(tokens[1].Text, "creature's") || strings.EqualFold(tokens[1].Text, "permanent's")) {
		return shared.SpanOf(tokens[:2]), 2, true
	}
	if span, ok := atoms.SelfNameSpanStartingAt(tokens[0].Span); ok {
		width := tokensCoveredCount(tokens, span)
		if width > 0 && strings.HasSuffix(tokens[width-1].Text, "'s") {
			return span, width, true
		}
	}
	return shared.Span{}, 0, false
}

// parseStaticDynamicValueCount recognizes the supported "the number of <count>"
// phrases a characteristic-defining power/toughness declaration counts. It
// returns the matched count kind and the index past the phrase.
func parseStaticDynamicValueCount(
	tokens []shared.Token,
	start, end int,
) (StaticDeclarationDynamicValueKind, int, bool) {
	if !staticWordsAt(tokens, start, "the", "number", "of") {
		return StaticDeclarationDynamicValueNone, 0, false
	}
	cursor := start + 3
	switch {
	case staticWordsAt(tokens, cursor, "cards", "in", "your", "hand"):
		return StaticDeclarationDynamicValueControllerHandSize, cursor + 4, true
	case staticWordsAt(tokens, cursor, "cards", "in", "your", "graveyard"):
		return StaticDeclarationDynamicValueControllerGraveyardSize, cursor + 4, true
	case staticWordsAt(tokens, cursor, "creatures", "you", "control"):
		return StaticDeclarationDynamicValueControllerCreatureCount, cursor + 3, true
	case staticWordsAt(tokens, cursor, "lands", "you", "control"):
		return StaticDeclarationDynamicValueControllerLandCount, cursor + 3, true
	case staticWordsAt(tokens, cursor, "artifacts", "you", "control"):
		return StaticDeclarationDynamicValueControllerArtifactCount, cursor + 3, true
	case staticWordsAt(tokens, cursor, "creatures", "on", "the", "battlefield"):
		return StaticDeclarationDynamicValueAllBattlefieldCreatureCount, cursor + 4, true
	default:
		return StaticDeclarationDynamicValueNone, 0, false
	}
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
	colors        bool
	types         bool
	creatureTypes bool
	landTypes     bool
}

// parseStaticInAdditionTail consumes "in addition to its/their other
// (colors|types|land types|colors and types)" beginning at start, returning the
// enumerated categories and the index following the tail. The "land types"
// variant is the one printed on continuous land-type-adding statics ("Each land
// is a Forest in addition to its other land types.").
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
	case staticWordsAt(tokens, cursor, "land", "types"):
		return staticInAdditionTail{landTypes: true}, cursor + 2, true
	case staticWordsAt(tokens, cursor, "creature", "types"):
		// "in addition to its/their other creature types" adds creature subtypes
		// without changing card types (Hivestone, Kindred-tribal statics).
		return staticInAdditionTail{creatureTypes: true}, cursor + 2, true
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
// "colors" category, card types and creature subtypes require a "types"
// category, and a "land types" tail requires the operation add only basic land
// subtypes. The tail may not enumerate a category the operation did not
// recognize.
func staticInAdditionTailMatches(tail staticInAdditionTail, colors []Color, cardTypes []CardType, subtypes []types.Sub) bool {
	hasColors := len(colors) != 0
	if tail.landTypes {
		return !hasColors && len(cardTypes) == 0 && len(subtypes) != 0 &&
			allBasicLandSubtypes(subtypes)
	}
	if tail.creatureTypes {
		// A "creature types" tail enumerates only added creature subtypes; it may
		// not accompany a color or card-type addition.
		return !hasColors && len(cardTypes) == 0 && len(subtypes) != 0
	}
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
		return operation, staticDynamicAmountEnd(tokens, next, end), true
	}
	return operation, next, true
}

// staticDynamicAmountEnd returns the index at which a dynamic power/toughness
// tail ends. A conjoined keyword grant ("… for each enchantment you control and
// has first strike") begins a separate "and has/have <keyword>" declaration, so
// the dynamic amount stops at that connector and the surrounding operation loop
// parses the keyword grant next. With no such rider the amount runs to end.
func staticDynamicAmountEnd(tokens []shared.Token, start, end int) int {
	for i := start; i < end; i++ {
		if staticWordsAt(tokens, i, "and") && i+1 < end &&
			(staticWordsAt(tokens, i+1, "has") || staticWordsAt(tokens, i+1, "have")) {
			return i
		}
	}
	return end
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
// [colorless] <colors>* [<subtype>] [creature] with base power and toughness
// N/N". The colors, card type, and creature subtype are SET (the affected object
// loses its other colors, card types, and creature types); a leading "colorless"
// makes it colorless instead. A name-setting tail ("named ..."), a non-creature
// card type, or any other trailing text fails closed.
func parseStaticLoseAbilitiesBecomeDeclaration(tokens []shared.Token, atoms Atoms) (StaticDeclarationSyntax, bool) {
	if len(tokens) < 5 || tokens[len(tokens)-1].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	subject, index, ok := parseStaticLoseAbilitiesSubject(tokens, atoms)
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	end := len(tokens) - 1
	// Order B ("becomes" first): "<subject> is <characteristics> with base
	// power and toughness N/N [and has <keyword>], and it loses all [other]
	// abilities[, card types,] and creature types." (Darksteel Mutation,
	// Lignify). Try it before the "loses all abilities" order so the leading
	// "is" body is consumed as one declaration rather than a bare keyword.
	if staticWordsAt(tokens, index, "is") {
		return parseStaticBecomeThenLoseDeclaration(tokens, index, end, subject, atoms)
	}
	if !staticWordsAt(tokens, index, "loses", "all", "abilities") {
		return StaticDeclarationSyntax{}, false
	}
	index += 3
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

// parseStaticBecomeThenLoseDeclaration recognizes the "becomes-first" polymorph
// order: "<subject> is [a] <colors>* <types>* [<subtype>] with base power and
// toughness N/N [and has <keyword>(s)], and it loses all [other] abilities[,
// card types,] and creature types." The colors, card types, and subtypes are
// SET; the optional "has <keyword>" tail grants keyword abilities. The trailing
// lose-clause is required (this is the vanilla/near-vanilla shape).
func parseStaticBecomeThenLoseDeclaration(tokens []shared.Token, index, end int, subject StaticDeclarationSubject, atoms Atoms) (StaticDeclarationSyntax, bool) {
	declaration := StaticDeclarationSyntax{
		Kind:             StaticDeclarationLoseAbilitiesBecome,
		Span:             shared.SpanOf(tokens),
		OperationSpan:    shared.SpanOf(tokens[:end]),
		Subject:          subject,
		LoseAllAbilities: true,
	}
	cursor := index + 1
	if staticWordsAt(tokens, cursor, "a") || staticWordsAt(tokens, cursor, "an") {
		cursor++
	}
	if staticWordsAt(tokens, cursor, "colorless") {
		declaration.BecomeColorless = true
		cursor++
	}
	list, next, ok := parseStaticCharacteristicList(tokens, cursor, end, atoms)
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	declaration.Colors = list.colors
	declaration.CardTypes = list.cardTypes
	declaration.Subtypes = list.subtypes
	if !staticWordsAt(tokens, next, "with") {
		return StaticDeclarationSyntax{}, false
	}
	basePT, ok := parseStaticBasePowerToughnessAt(tokens, next+1)
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	declaration.BasePower = basePT.power
	declaration.BaseToughness = basePT.toughness
	declaration.BasePTSet = true
	cursor = basePT.next
	if staticWordsAt(tokens, cursor, "and", "has") {
		spans, kwNext, ok := parseStaticKeywordList(tokens, cursor+2, end, atoms)
		if !ok {
			return StaticDeclarationSyntax{}, false
		}
		declaration.KeywordSpans = spans
		cursor = kwNext
	}
	cursor, ok = parseStaticBecomeLoseAbilitiesTail(tokens, cursor, end)
	if !ok || cursor != end {
		return StaticDeclarationSyntax{}, false
	}
	return declaration, true
}

// parseStaticBecomeLoseAbilitiesTail consumes the trailing lose-clause of a
// becomes-first polymorph declaration: an optional comma, "and", optional "it",
// then "loses all [other] abilities", optionally followed by the redundant
// "[, card types,] and creature types" enumeration (all SET by the body). It
// returns the index following the clause.
func parseStaticBecomeLoseAbilitiesTail(tokens []shared.Token, index, end int) (int, bool) {
	cursor := index
	if cursor < end && tokens[cursor].Kind == shared.Comma {
		cursor++
	}
	if !staticWordsAt(tokens, cursor, "and") {
		return 0, false
	}
	cursor++
	if staticWordsAt(tokens, cursor, "it") {
		cursor++
	}
	if !staticWordsAt(tokens, cursor, "loses", "all") {
		return 0, false
	}
	cursor += 2
	if staticWordsAt(tokens, cursor, "other") {
		cursor++
	}
	if !staticWordsAt(tokens, cursor, "abilities") {
		return 0, false
	}
	cursor++
	for cursor < end {
		switch {
		case tokens[cursor].Kind == shared.Comma:
			cursor++
		case staticWordsAt(tokens, cursor, "and"):
			cursor++
		case staticWordsAt(tokens, cursor, "other"):
			cursor++
		case staticWordsAt(tokens, cursor, "card", "types"):
			cursor += 2
		case staticWordsAt(tokens, cursor, "creature", "types"):
			cursor += 2
		default:
			return 0, false
		}
	}
	return cursor, true
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
	if staticWordsAt(tokens, cursor, "colorless") {
		declaration.BecomeColorless = true
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
