package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// creditGoadCreatedTokensRider folds the trailing "The tokens are goaded for the
// rest of the game." rider sentence onto a preceding create-token effect (Life
// of the Party: "each opponent creates a token that's a copy of it. The tokens
// are goaded for the rest of the game."). The rider goads exactly the tokens the
// create produced, permanently. It records the rider span on the create effect
// for lowering and source coverage, clears the rider sentence's effects, and
// marks the sentence so reference and coverage scans credit its "the tokens" back
// reference to the create rather than flagging it as an unrecognized sibling.
//
// It credits only the exact shape: an ability holding exactly one create-token
// effect and exactly one matching rider sentence. Any other shape leaves the
// rider uncredited so the card fails closed. The create must ultimately lower as
// a linked publish so the goad can bind the created tokens; that requirement is
// enforced again at lowering, which fails closed if the create cannot publish.
func creditGoadCreatedTokensRider(ability *Ability) {
	create := loneCreateTokenEffect(ability.Sentences)
	if create == nil {
		return
	}
	riderIdx := -1
	for i := range ability.Sentences {
		sentence := &ability.Sentences[i]
		if len(sentence.Effects) != 0 {
			continue
		}
		if !isGoadCreatedTokensRiderTokens(semanticEffectTokens(sentence.Tokens)) {
			continue
		}
		if riderIdx >= 0 {
			return
		}
		riderIdx = i
	}
	if riderIdx < 0 {
		return
	}
	create.GoadCreatedTokensRiderSpan = ability.Sentences[riderIdx].Span
	create.HasUnrecognizedSibling = false
	create.Exact = exactEffectSyntax(create)
	ability.Sentences[riderIdx].Effects = nil
	ability.Sentences[riderIdx].GoadCreatedTokensRider = true
}

// loneCreateTokenEffect returns the single create-token effect across the
// sentences, or nil when there is not exactly one. Unlike
// loneControllerCreateTokenEffect it accepts any recipient context, because the
// goad-created-tokens rider folds onto a group-recipient create ("each opponent
// creates a token ...", Life of the Party) as well as a controller create. Any
// other effect kind alongside it is tolerated; the caller and lowering apply the
// stricter shape checks.
func loneCreateTokenEffect(sentences []Sentence) *EffectSyntax {
	var found *EffectSyntax
	for i := range sentences {
		for j := range sentences[i].Effects {
			effect := &sentences[i].Effects[j]
			if effect.Kind != EffectCreate {
				continue
			}
			if found != nil {
				return nil
			}
			found = effect
		}
	}
	return found
}

// isGoadCreatedTokensRiderTokens reports whether the sentence tokens are exactly
// "The tokens are goaded for the rest of the game." A trailing period is the only
// content permitted after the ten words; any other token leaves the rider
// unrecognized so it is not mistaken for a standalone effect.
func isGoadCreatedTokensRiderTokens(tokens []shared.Token) bool {
	if !effectWordsAt(tokens, 0, "the", "tokens", "are", "goaded", "for", "the", "rest", "of", "the", "game") {
		return false
	}
	for _, token := range tokens[10:] {
		if token.Kind != shared.Period {
			return false
		}
	}
	return true
}
