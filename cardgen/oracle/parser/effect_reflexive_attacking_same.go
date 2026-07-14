package parser

import (
	"reflect"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// creditEachOpponentAttackingSameRider folds the trailing anaphoric "Each
// opponent attacking that player does the same." rider sentence onto an
// enchanted-player combat trigger's lone controller effect. The rider widens the
// controller's action so that, in addition to the controller, each opponent
// attacking the enchanted player does the same: creating that token (Curse of
// Opulence, Curse of Disturbance), gaining that life (Curse of Vitality), or
// drawing those cards (Curse of Verbosity). It records the rider span on the
// widened effect for lowering and source coverage, clears the rider sentence's
// effects, and marks the sentence so reference and coverage scans credit its
// "that player" back reference to the widened effect rather than flagging it as
// an unrecognized sibling.
//
// It credits only the exact shape: a triggered ability whose event is the passive
// "enchanted player is attacked" clause, holding exactly one exact controller
// create-token, gain-life, or draw effect that is not a copy, choice, or
// multi-token creation, and exactly one matching rider sentence. Any other shape
// leaves the rider uncredited so the card fails closed. The strict effect
// compatibility check is enforced again at lowering, which fails closed if the
// widened effect cannot be mirrored to the attacking-opponent group.
func creditEachOpponentAttackingSameRider(ability *Ability) {
	if !enchantedPlayerAttackedTrigger(ability) {
		return
	}
	effect := loneControllerReflexiveSameEffect(ability.Sentences)
	if effect == nil {
		return
	}
	if effect.Kind == EffectCreate &&
		(effect.TokenChoice || len(effect.AdditionalTokens) > 0 ||
			effect.TokenCopyOfTarget || effect.TokenCopyOfReference || effect.TokenCopyOfAttached ||
			effect.TokenCopyOfTriggeringSet || effect.TokenCopyOfForEach) {
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
	effect.EachOpponentAttackingSameRiderSpan = ability.Sentences[riderIdx].Span
	effect.HasUnrecognizedSibling = false
	effect.Exact = exactEffectSyntax(effect)
	ability.Sentences[riderIdx].Effects = nil
	ability.Sentences[riderIdx].EachOpponentAttackingSameRider = true
}

// creditEachOpponentAttackingUntapRider folds the trailing explicit "Each
// opponent attacking that player untaps all nonland permanents they control."
// rider sentence onto an enchanted-player combat trigger's lone controller untap
// effect (Curse of Bounty). Unlike the anaphoric "does the same." family, this
// rider spells its action out, so it parses to a real "each opponent" untap of
// the same nonland group the controller untaps. It records the rider span on the
// controller untap for lowering and source coverage, clears the rider sentence's
// untap effect, and marks the sentence so its "that player"/"they" back
// references are credited rather than flagged as an unrecognized sibling.
//
// It credits only the exact shape: a triggered ability whose event is the passive
// "enchanted player is attacked" clause, holding exactly one exact controller
// untap of "all nonland permanents you control", and exactly one rider sentence
// whose leading words are "each opponent attacking that player" and whose sole
// effect is an each-opponent untap of the same permanent group the controller
// untaps. Any other shape leaves the rider uncredited so the card fails closed.
func creditEachOpponentAttackingUntapRider(ability *Ability) {
	if !enchantedPlayerAttackedTrigger(ability) {
		return
	}
	controller := loneControllerUntapEffect(ability.Sentences)
	if controller == nil || !controller.Exact {
		return
	}
	riderIdx := -1
	for i := range ability.Sentences {
		sentence := &ability.Sentences[i]
		if len(sentence.Effects) != 1 {
			continue
		}
		if !isEachOpponentAttackingRiderPrefix(semanticEffectTokens(sentence.Tokens)) {
			continue
		}
		rider := &sentence.Effects[0]
		if rider.Kind != EffectUntap ||
			rider.Context != EffectContextEachOpponent ||
			!reflexiveUntapSelectionsMatch(&controller.Selection, &rider.Selection) {
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
	// The controller untap and the rider each-opponent untap must be the ability's
	// only two effects. A third sibling effect would survive the fold and route the
	// ability through multi-effect lowering, which never mirrors the rider, so
	// reject it here and fail closed.
	if totalEffectCount(ability.Sentences) != 2 {
		return
	}
	controller.EachOpponentAttackingSameRiderSpan = ability.Sentences[riderIdx].Span
	controller.HasUnrecognizedSibling = false
	// The controller untap and the now-folded rider untap were the ability's two
	// legacy effects, so emitSentenceResolvingSyntax flagged both as requiring
	// ordered lowering. Folding the rider leaves the controller untap as the
	// ability's sole real effect — a single continuous untap plus its group
	// mirror — so clear the flag it no longer needs, mirroring how the other
	// multi-sentence rider folds (attacker mana, discard-then-draw) clear it.
	controller.RequiresOrderedLowering = false
	ability.Sentences[riderIdx].Effects = nil
	ability.Sentences[riderIdx].EachOpponentAttackingSameRider = true
}

// enchantedPlayerAttackedTrigger reports whether the ability is triggered by the
// passive "enchanted player is attacked" combat event, the shared gate of the
// reflexive attacking-opponent rider family.
func enchantedPlayerAttackedTrigger(ability *Ability) bool {
	return ability.Trigger != nil && ability.Trigger.TriggerEvent != nil &&
		ability.Trigger.TriggerEvent.EnchantedPlayerIsAttacked
}

// loneControllerReflexiveSameEffect returns the ability's sole effect when it is a
// controller-recipient create-token, gain-life, or draw effect — the kinds the
// anaphoric "does the same." rider can widen to the attacking-opponent group. It
// returns nil unless that candidate is the ability's only effect, so a trigger
// that pairs the candidate with any sibling effect fails closed instead of folding
// the rider and then silently dropping the group mirror in multi-effect lowering.
// The "does the same." rider sentence itself carries no effects, so it never
// counts against this lone-effect requirement.
func loneControllerReflexiveSameEffect(sentences []Sentence) *EffectSyntax {
	var found *EffectSyntax
	for i := range sentences {
		for j := range sentences[i].Effects {
			if found != nil {
				return nil
			}
			found = &sentences[i].Effects[j]
		}
	}
	if found == nil || !reflexiveSameCandidate(found) || found.Context != EffectContextController {
		return nil
	}
	return found
}

// reflexiveSameCandidate reports whether effect is one the anaphoric "does the
// same." rider can widen: a token creation, a life gain (a "gain N life" whose
// grammatical object is life, not a keyword grant), or a card draw.
func reflexiveSameCandidate(effect *EffectSyntax) bool {
	switch effect.Kind {
	case EffectCreate, EffectDraw:
		return true
	case EffectGain:
		return effect.LifeObject
	default:
		return false
	}
}

// loneControllerUntapEffect returns the single controller-recipient untap effect
// across the sentences, or nil when there is not exactly one controller untap.
// The explicit Curse of Bounty rider parses to its own each-opponent untap
// sentence, so this counts only controller-context untaps; the caller pairs it
// with the matching rider sentence and requires those two to be the ability's only
// effects, so any additional sibling effect fails closed.
func loneControllerUntapEffect(sentences []Sentence) *EffectSyntax {
	var found *EffectSyntax
	for i := range sentences {
		for j := range sentences[i].Effects {
			effect := &sentences[i].Effects[j]
			if effect.Kind != EffectUntap || effect.Context != EffectContextController {
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

// totalEffectCount returns the number of parsed effects across every sentence of
// an ability. The reflexive untap fold uses it to require that the controller
// untap and the rider each-opponent untap are the ability's only effects, so a
// third sibling effect fails closed instead of surviving the fold unmirrored.
func totalEffectCount(sentences []Sentence) int {
	n := 0
	for i := range sentences {
		n += len(sentences[i].Effects)
	}
	return n
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

// isEachOpponentAttackingRiderPrefix reports whether the sentence tokens begin
// with the reflexive framing "Each opponent attacking that player", the shared
// five-word prefix that ties an explicit rider action to the opponents attacking
// the enchanted player. The remaining tokens spell the rider's own action and are
// validated through the parsed effect shape by the caller.
func isEachOpponentAttackingRiderPrefix(tokens []shared.Token) bool {
	return effectWordsAt(tokens, 0, "each", "opponent", "attacking", "that", "player")
}

// reflexiveUntapSelectionsMatch reports whether the controller untap's permanent
// selection and the rider untap's permanent selection describe the same group of
// permanents up to who controls them: the controller untaps "nonland permanents
// you control" and the rider untaps "nonland permanents they control", so the two
// selections agree on every dimension except the controller relation and the
// verbatim text and pronoun-reconstruction flags. Requiring this match keeps the
// rider strictly the reflexive mirror of the controller untap and fails closed if
// the two untaps name different permanent groups.
func reflexiveUntapSelectionsMatch(controller, rider *SelectionSyntax) bool {
	a := *controller
	b := *rider
	a.Span, b.Span = shared.Span{}, shared.Span{}
	a.Text, b.Text = "", ""
	a.Controller, b.Controller = SelectionControllerAny, SelectionControllerAny
	a.OpponentEach, b.OpponentEach = false, false
	a.OpponentThey, b.OpponentThey = false, false
	return reflect.DeepEqual(a, b)
}
