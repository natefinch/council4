package parser

import "strings"

const shuffleRevealPermanentOracleText = "The owner of target permanent shuffles it into their library, then reveals the top card of their library. If it's a permanent card, they put it onto the battlefield."

// recognizeShuffleRevealPermanentSequence recognizes the closed linked-card
// sequence that shuffles one targeted permanent into its owner's library,
// reveals that owner's top card, and puts the revealed card onto the battlefield
// only when it is a permanent card.
func recognizeShuffleRevealPermanentSequence(sentences []Sentence) {
	if len(sentences) != 2 ||
		len(sentences[0].Effects) != 2 ||
		len(sentences[1].Effects) != 1 ||
		!strings.EqualFold(sentences[0].Text+" "+sentences[1].Text, shuffleRevealPermanentOracleText) {
		return
	}
	shuffle := &sentences[0].Effects[0]
	reveal := &sentences[0].Effects[1]
	put := &sentences[1].Effects[0]
	if shuffle.Kind != EffectShuffle ||
		reveal.Kind != EffectReveal ||
		put.Kind != EffectPut {
		return
	}

	shuffle.Player = EffectPlayerTargetOwner
	shuffle.Exact = true

	reveal.Player = EffectPlayerTargetOwner
	reveal.CardSource = EffectCardSourceTopOfPlayerLibrary
	reveal.Exact = true

	put.Player = EffectPlayerTargetOwner
	put.CardSource = EffectCardSourcePriorInstructionResult
	put.RequirePermanentCard = true
	put.Exact = true
}
