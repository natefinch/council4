package parser

func createHasLeadingIterator(effect *EffectSyntax) bool {
	for i, token := range effect.Tokens {
		if token.Span.Start.Offset >= effect.VerbSpan.Start.Offset {
			break
		}
		if effectWordsAt(effect.Tokens, i, "for", "each") {
			return true
		}
	}
	return false
}
