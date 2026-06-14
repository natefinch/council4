package compiler

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

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

func unsupportedDiagnostic(span shared.Span, text string) shared.Diagnostic {
	return shared.Diagnostic{
		Severity: shared.SeverityWarning,
		Summary:  "unsupported Oracle construct",
		Detail:   "the compiler preserved but did not confidently lower: " + text,
		Span:     span,
	}
}
