package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/types"
)

func recognizeDrawPutLandSubtypeLifeSequence(ability *Ability) bool {
	if ability == nil || ability.Kind != AbilityTriggered || ability.Trigger == nil {
		return false
	}
	tokens := bodyTokensExcludingReminders(ability)
	cursor := 0
	takeWord := func(word string) bool {
		if cursor >= len(tokens) || !equalWord(tokens[cursor], word) {
			return false
		}
		cursor++
		return true
	}
	takeKind := func(kind shared.Kind) bool {
		if cursor >= len(tokens) || tokens[cursor].Kind != kind {
			return false
		}
		cursor++
		return true
	}
	for _, word := range []string{"draw", "a", "card"} {
		if !takeWord(word) {
			return false
		}
	}
	if !takeKind(shared.Comma) {
		return false
	}
	for _, word := range []string{"then", "you", "may", "put", "a", "land", "card", "from", "your", "hand", "onto", "the", "battlefield"} {
		if !takeWord(word) {
			return false
		}
	}
	if !takeKind(shared.Period) {
		return false
	}
	for _, word := range []string{"if", "you", "put", "a"} {
		if !takeWord(word) {
			return false
		}
	}
	if cursor >= len(tokens) || tokens[cursor].Kind != shared.Word {
		return false
	}
	subtype := types.Sub(tokens[cursor].Text)
	cursor++
	for _, word := range []string{"onto", "the", "battlefield", "this", "way"} {
		if !takeWord(word) {
			return false
		}
	}
	if !takeKind(shared.Comma) {
		return false
	}
	for _, word := range []string{"you", "gain"} {
		if !takeWord(word) {
			return false
		}
	}
	if cursor >= len(tokens) {
		return false
	}
	life, ok := effectNumber(tokens[cursor], ability.Atoms)
	if !ok || life < 1 {
		return false
	}
	cursor++
	if !takeWord("life") || !takeKind(shared.Period) || cursor != len(tokens) {
		return false
	}
	ability.ExactSequence = &ExactSequenceSyntax{
		Kind:       ExactSequenceDrawPutLandSubtypeLife,
		Span:       ability.BodySpan,
		PutSubtype: subtype,
		LifeAmount: life,
	}
	return true
}
