package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

func conditionCardTypeQualifierContinues(tokens []shared.Token, comma int) bool {
	if comma <= 0 || comma+1 >= len(tokens) {
		return false
	}
	if _, before := recognizeExcludedCardTypeWord(tokens[comma-1].Text); !before {
		return false
	}
	if _, after := recognizeExcludedCardTypeWord(tokens[comma+1].Text); !after {
		return false
	}
	for i := comma + 1; i < len(tokens); i++ {
		if equalWord(tokens[i], "card") || equalWord(tokens[i], "cards") {
			return i+1 == len(tokens) || tokens[i+1].Kind == shared.Comma || tokens[i+1].Kind == shared.Period
		}
		if tokens[i].Kind == shared.Comma {
			if i+1 >= len(tokens) {
				return false
			}
			_, before := recognizeExcludedCardTypeWord(tokens[i-1].Text)
			_, after := recognizeExcludedCardTypeWord(tokens[i+1].Text)
			if !before || !after {
				return false
			}
			continue
		}
		if _, excluded := recognizeExcludedCardTypeWord(tokens[i].Text); excluded {
			continue
		}
		if _, cardType := recognizeCardTypeWord(tokens[i].Text); !cardType {
			return false
		}
	}
	return false
}
