package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// parseStaticSpellKeywordGrantDeclaration recognizes the spell keyword-grant
// family "[<filter>] spells you cast have <keyword>." (Inspiring Statuary's
// "Nonartifact spells you cast have improvise."; Ironheart, Clever Champion's
// "Noncreature spells you cast have improvise."). The declaration grants a
// parameterless keyword to the matching spells its controller casts. An optional
// leading "nonartifact"/"noncreature" word narrows the affected spells by card
// type; the bare "Spells you cast have <keyword>." form grants the keyword to
// every spell the controller casts. The granted keyword atom is carried in
// KeywordSpans and resolved downstream from the ability's recognized keyword
// content.
func parseStaticSpellKeywordGrantDeclaration(tokens []shared.Token, atoms Atoms) (StaticDeclarationSyntax, bool) {
	if len(tokens) < 5 || tokens[len(tokens)-1].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	end := len(tokens) - 1
	index := 0
	filter := StaticDeclarationCardFilterNone
	if width, kind, ok := staticSpellKeywordGrantFilter(tokens, index); ok {
		filter = kind
		index += width
	}
	if !staticWordsAt(tokens, index, "spells", "you", "cast", "have") {
		return StaticDeclarationSyntax{}, false
	}
	index += 4
	keyword, keywordWidth, ok := staticKeywordAt(tokens, index, end, atoms)
	if !ok || keyword.Parameter.Kind != KeywordParameterNone || index+keywordWidth != end {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationSpellKeywordGrant,
		Span:          shared.SpanOf(tokens),
		OperationSpan: keyword.Span,
		Subject: StaticDeclarationSubject{
			Kind:       StaticDeclarationSubjectControllerSpells,
			Span:       shared.SpanOf(tokens[:index]),
			CardFilter: filter,
		},
		KeywordSpans: []shared.Span{keyword.Span},
	}, true
}

// staticSpellKeywordGrantFilter matches an optional leading "non<type>" word that
// narrows the affected spells of a "[<filter>] spells you cast have <keyword>."
// declaration, returning the recognized filter and the number of tokens it
// consumed. Only the closed nonartifact and noncreature forms the corpus needs
// are recognized; any other prefix fails so the caller falls back to the
// unfiltered form or rejects the sentence.
func staticSpellKeywordGrantFilter(tokens []shared.Token, index int) (int, StaticDeclarationCardFilterKind, bool) {
	if index >= len(tokens) || tokens[index].Kind != shared.Word {
		return 0, StaticDeclarationCardFilterNone, false
	}
	excluded, ok := recognizeExcludedCardTypeWord(tokens[index].Text)
	if !ok {
		return 0, StaticDeclarationCardFilterNone, false
	}
	switch excluded {
	case CardTypeArtifact:
		return 1, StaticDeclarationCardFilterNonartifact, true
	case CardTypeCreature:
		return 1, StaticDeclarationCardFilterNoncreature, true
	default:
		return 0, StaticDeclarationCardFilterNone, false
	}
}
