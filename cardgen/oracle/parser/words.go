package parser

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// normalizedWords returns the lower-cased text of word tokens. Oracle vocabulary
// normalization belongs to the parser, which owns spelling and grammar; the
// compiler and lowering never normalize Oracle wording.
func normalizedWords(tokens []shared.Token) []string {
	words := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if token.Kind == shared.Word {
			words = append(words, strings.ToLower(token.Text))
		}
	}
	return words
}
