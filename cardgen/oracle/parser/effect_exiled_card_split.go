package parser

import (
	"slices"

	"github.com/natefinch/council4/mtg/game/zone"
)

// isExiledCardOpponentChoiceWords reports whether the normalized words are the
// recognized zero-effect antecedent "An opponent chooses one of the exiled
// cards." that names an opponent as the chooser of one of the cost-exiled cards.
func isExiledCardOpponentChoiceWords(words []string) bool {
	return slices.Equal(words,
		[]string{"an", "opponent", "chooses", "one", "of", "the", "exiled", "cards"})
}

// recognizeExiledCardOpponentChoiceSplit recognizes the opponent-choice disposal
// antecedent "An opponent chooses one of the exiled cards. You put that card on
// the bottom of your library and return the other to the battlefield tapped."
// (Coin of Fate). The zero-effect antecedent names an opponent as the chooser of
// one of the cost-exiled cards; the immediately following put/return sentence
// sends the chosen card ("that card") to the bottom of the controller's library
// and the other to the battlefield tapped. It marks the antecedent as a credited
// rider and records the opponent-chooser role and antecedent span on the put
// effect so text-blind lowering can synthesize the opponent's choice. Any other
// wording fails closed.
func recognizeExiledCardOpponentChoiceSplit(sentences []Sentence) bool {
	for i := range sentences {
		if len(sentences[i].Effects) != 0 ||
			!isExiledCardOpponentChoiceWords(normalizedWords(semanticEffectTokens(sentences[i].Tokens))) {
			continue
		}
		put, ok := exiledCardSplitPutReturn(sentences, i+1)
		if !ok {
			continue
		}
		put.ExiledCardSplitOpponentChooses = true
		put.ExiledCardChoiceRiderSpan = sentences[i].Span
		sentences[i].ExiledCardChoiceRider = true
		return true
	}
	return false
}

// exiledCardSplitPutReturn validates that the sentence at index holds the
// put/return disposal "You put that card on the bottom of your library and
// return the other to the battlefield tapped." and returns the put effect for
// annotation. Any other shape fails closed.
func exiledCardSplitPutReturn(sentences []Sentence, index int) (*EffectSyntax, bool) {
	if index < 0 || index >= len(sentences) {
		return nil, false
	}
	effects := sentences[index].Effects
	if len(effects) != 2 {
		return nil, false
	}
	put := &effects[0]
	other := &effects[1]
	if put.Kind != EffectPut ||
		put.Context != EffectContextController ||
		put.ToZone != zone.Library ||
		put.Destination != EffectDestinationBottom {
		return nil, false
	}
	if other.Kind != EffectReturn ||
		other.Context != EffectContextController ||
		other.ToZone != zone.Battlefield ||
		!other.EntersTapped ||
		!other.Selection.Other {
		return nil, false
	}
	return put, true
}
