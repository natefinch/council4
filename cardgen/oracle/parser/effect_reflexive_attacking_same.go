package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// creditEachOpponentAttackingSameRider folds the trailing "Each opponent
// attacking that player does the same." rider sentence onto an enchanted-player
// combat trigger's lone controller create-token effect (Curse of Opulence, Curse
// of Disturbance). The rider widens the token creation so that, in addition to
// the controller, each opponent attacking the enchanted player also creates that
// token. It records the rider span on the create effect for lowering and source
// coverage, clears the rider sentence's effects, and marks the sentence so
// reference and coverage scans credit its "that player" back reference and tokens
// to the create rather than flagging them as an unrecognized sibling.
//
// It credits only the exact shape: a triggered ability whose event is the passive
// "enchanted player is attacked" clause, holding exactly one exact controller
// create-token effect that is not a copy, choice, or multi-token creation, and
// exactly one matching rider sentence. Any other shape leaves the rider
// uncredited so the card fails closed. The strict token compatibility check (a
// fixed-count creature or predefined artifact token) is enforced again at
// lowering, which fails closed if the folded create cannot widen to a group.
func creditEachOpponentAttackingSameRider(ability *Ability) {
	if ability.Trigger == nil || ability.Trigger.TriggerEvent == nil ||
		!ability.Trigger.TriggerEvent.EnchantedPlayerIsAttacked {
		return
	}
	create := loneControllerCreateTokenEffect(ability.Sentences)
	if create == nil {
		return
	}
	if create.TokenChoice || len(create.AdditionalTokens) > 0 ||
		create.TokenCopyOfTarget || create.TokenCopyOfReference || create.TokenCopyOfAttached ||
		create.TokenCopyOfTriggeringSet || create.TokenCopyOfForEach {
		return
	}
	riderIdx := -1
	for i := range ability.Sentences {
		sentence := &ability.Sentences[i]
		if len(sentence.Effects) != 0 {
			continue
		}
		if !isEachOpponentAttackingSameRiderTokens(semanticEffectTokens(sentence.Tokens)) {
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
	create.EachOpponentAttackingSameRiderSpan = ability.Sentences[riderIdx].Span
	create.HasUnrecognizedSibling = false
	create.Exact = exactEffectSyntax(create)
	ability.Sentences[riderIdx].Effects = nil
	ability.Sentences[riderIdx].EachOpponentAttackingSameRider = true
}

// loneControllerCreateTokenEffect returns the single controller-recipient
// create-token effect across the sentences, or nil when there is not exactly one
// create effect or the sole create is not controller-scoped. Any other effect
// kind alongside it is tolerated; the caller applies the stricter shape checks.
func loneControllerCreateTokenEffect(sentences []Sentence) *EffectSyntax {
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
	if found == nil || found.Context != EffectContextController {
		return nil
	}
	return found
}

// isEachOpponentAttackingSameRiderTokens reports whether the sentence tokens are
// exactly "Each opponent attacking that player does the same." A trailing period
// is the only content permitted after the eight words; any other token leaves the
// rider unrecognized so it is not mistaken for a standalone effect.
func isEachOpponentAttackingSameRiderTokens(tokens []shared.Token) bool {
	if !effectWordsAt(tokens, 0, "each", "opponent", "attacking", "that", "player", "does", "the", "same") {
		return false
	}
	for _, token := range tokens[8:] {
		if token.Kind != shared.Period {
			return false
		}
	}
	return true
}
