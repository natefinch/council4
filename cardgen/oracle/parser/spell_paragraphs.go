package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// RulesFreeAbilityWord reports whether a recognized label has no rules meaning.
func RulesFreeAbilityWord(label string) bool {
	switch label {
	case "",
		"Adamant", "Addendum", "Bloodrush", "Celebration", "Channel", "Cohort",
		"Converge", "Corrupted", "Coven", "Crown of Madness", "Delirium", "Descend",
		"Domain", "Enrage", "Fateful hour", "Ferocious", "Flurry", "Formidable",
		"Hellbent", "History Teacher", "Inspired", "Kinship", "Lieutenant",
		"Magecraft", "Metalcraft", "Morbid", "Opus", "Parley", "Raid", "Revolt",
		"Spell mastery", "Survival", "Teleport", "Threshold", "Those Who Came Before",
		"Undergrowth", "Unlock Ability", "Vivid", "Void", "Will of the council":
		return true
	default:
		return false
	}
}

func joinResolvingAlternativeParagraphs(source string, context Context, abilities []Ability) ([]Ability, []shared.Diagnostic) {
	if !context.InstantOrSorcery {
		return abilities, nil
	}
	var result []Ability
	var diagnostics []shared.Diagnostic
	for _, ability := range abilities {
		if len(result) == 0 || !plainResolvingParagraph(result[len(result)-1]) ||
			!plainResolvingParagraph(ability) ||
			!conditionalInsteadParagraph(ability) &&
				!(result[len(result)-1].paragraphSourceTokens != nil && leadingConditionalParagraph(ability)) {
			result = append(result, ability)
			continue
		}
		previous := &result[len(result)-1]
		tokens := append(cloneTokens(previous.Tokens), TokensInSpan(ability.Tokens, ability.BodySpan)...)
		joined, parsedDiagnostics := parseAbility(source, tokens, context)
		diagnostics = append(diagnostics, parsedDiagnostics...)
		joined.paragraphSourceTokens = append(cloneTokens(previous.originalParagraphTokens()), ability.originalParagraphTokens()...)
		joined.paragraphAbilityWords = append([]AbilityWordClause(nil), previous.paragraphAbilityWords...)
		if ability.AbilityWord != nil {
			joined.paragraphAbilityWords = append(joined.paragraphAbilityWords, *ability.AbilityWord)
		}
		*previous = joined
	}

	return result, diagnostics
}

func (a *Ability) originalParagraphTokens() []shared.Token {
	if a.paragraphSourceTokens != nil {
		return a.paragraphSourceTokens
	}
	return a.Tokens
}

// ParagraphLabelSpans reports the rules-free labels consumed by paragraph joining.
func (a *Ability) ParagraphLabelSpans() []shared.Span {
	var spans []shared.Span
	for _, word := range a.paragraphAbilityWords {
		spans = append(spans, word.Span, word.SeparatorSpan)
	}
	return spans
}

func plainResolvingParagraph(ability Ability) bool {
	return ability.Kind == AbilitySpell && ability.Modal == nil && ability.DiceTable == nil &&
		ability.costPhrase == nil && len(ability.Sentences) > 0 &&
		(ability.AbilityWord == nil || RulesFreeAbilityWord(ability.AbilityWord.Label))
}

func conditionalInsteadParagraph(ability Ability) bool {
	if len(ability.Sentences) == 0 {
		return false
	}
	body := semanticEffectTokens(TokensInSpan(ability.Tokens, ability.Sentences[0].Span))
	hasIf := false
	for i, token := range body {
		hasIf = hasIf || equalWord(token, "if")
		if i > 0 && equalWord(token, "if") && equalWord(body[i-1], "instead") {
			return true
		}
	}
	if !hasIf {
		return false
	}
	if replacement, ok := leadingInsteadReplacement(body); ok {
		return replacement.Kind == EffectReplacementInstead
	}
	if replacement, ok := trailingInsteadBeforeConditionReplacement(body); ok {
		return replacement.Kind == EffectReplacementInstead
	}
	return len(body) >= 2 && equalWord(body[len(body)-2], "instead") &&
		body[len(body)-1].Kind == shared.Period
}

func leadingConditionalParagraph(ability Ability) bool {
	body := semanticEffectTokens(TokensInSpan(ability.Tokens, ability.BodySpan))
	return len(body) > 0 && (equalWord(body[0], "if") || equalWord(body[0], "unless"))
}

// ReplacedClauseID names the parser-owned resolving clause this alternative replaces.
func (r EffectReplacementSyntax) ReplacedClauseID() int {
	return r.replacedClauseID
}
