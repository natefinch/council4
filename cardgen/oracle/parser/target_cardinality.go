package parser

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// variableTargetCardinality owns the count prefix of one target production.
// Its span remains part of the target in the owning effect's ClauseID.
func variableTargetCardinality(tokens []shared.Token, targetIndex int) (int, TargetCardinalitySyntax, bool) {
	countIndex := targetIndex - 1
	if countIndex >= 0 && (equalWord(tokens[countIndex], "other") || equalWord(tokens[countIndex], "another")) {
		countIndex--
	}
	if countIndex < 0 || !equalWord(tokens[countIndex], "X") {
		return 0, TargetCardinalitySyntax{}, false
	}
	count := TargetCardinalitySyntax{Min: 0, Max: 99, CountEqualsX: true}
	if countIndex >= 2 && effectWordsAt(tokens, countIndex-2, "up", "to") {
		count.CountEqualsX = false
		count.MaxX = true
		countIndex -= 2
	}
	return countIndex, count, true
}

// Exactness reuses the single-target selection grammar after singularizing only
// its typed noun vocabulary; every other word must still round-trip.
func exactCountedPermanentTargetSyntax(text, prefix string, selection SelectionSyntax) bool {
	if !strings.HasPrefix(strings.ToLower(text), strings.ToLower(prefix)) {
		return false
	}
	words := strings.Fields(text[len(prefix):])
	if selection.Controller == SelectionControllerOpponent && len(words) >= 3 &&
		strings.EqualFold(strings.Join(words[len(words)-3:], " "), "your opponents control") {
		words[len(words)-3], words[len(words)-2], words[len(words)-1] = "an", "opponent", "controls"
	}
	nouns := []string{"artifact", "creature", "enchantment", "land", "permanent", "planeswalker", "battle", "spell", "player", "opponent"}
	for _, subtype := range selection.SubtypesAny {
		nouns = append(nouns, string(subtype))
	}
	for i, word := range words {
		if word == "and/or" && len(selection.RequiredTypesAny) > 1 && !selection.ConjunctiveTypes {
			words[i] = "or"
			continue
		}
		bare := strings.TrimSuffix(word, ",")
		for _, noun := range nouns {
			if strings.EqualFold(bare, noun+"s") {
				words[i] = noun + strings.TrimPrefix(word, bare)
				break
			}
		}
	}
	return exactSinglePermanentTargetSyntax(strings.Join(words, " "), selection)
}
