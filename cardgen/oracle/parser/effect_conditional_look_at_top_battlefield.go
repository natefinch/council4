package parser

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// conditionalLookAtTopBattlefieldPutRun is the fixed token run that follows the
// variable card-type disjunction and the trailing "card" word: ", you may put
// it onto the battlefield". The optional "tapped" entry rider and the closing
// period are matched separately so both the plain and tapped entries are
// recognized.
var conditionalLookAtTopBattlefieldPutRun = []struct {
	kind shared.Kind
	text string
}{
	{shared.Comma, ","}, {shared.Word, "you"}, {shared.Word, "may"},
	{shared.Word, "put"}, {shared.Word, "it"}, {shared.Word, "onto"},
	{shared.Word, "the"}, {shared.Word, "battlefield"},
}

// conditionalLookAtTopBattlefieldHandRuns are the recognized trailing fallback
// clauses that send the card into the controller's hand when it is not put onto
// the battlefield. Each begins after the put clause's closing period and must
// consume the remainder of the body. The three wordings ("the card" / "it" in
// the conditional, and the "Otherwise," form) are equivalent: a mandatory move
// into the hand.
var conditionalLookAtTopBattlefieldHandRuns = [][]struct {
	kind shared.Kind
	text string
}{
	{
		{shared.Word, "If"}, {shared.Word, "you"}, {shared.Word, "don't"},
		{shared.Word, "put"}, {shared.Word, "the"}, {shared.Word, "card"},
		{shared.Word, "onto"}, {shared.Word, "the"}, {shared.Word, "battlefield"},
		{shared.Comma, ","}, {shared.Word, "put"}, {shared.Word, "it"},
		{shared.Word, "into"}, {shared.Word, "your"}, {shared.Word, "hand"},
		{shared.Period, "."},
	},
	{
		{shared.Word, "If"}, {shared.Word, "you"}, {shared.Word, "don't"},
		{shared.Word, "put"}, {shared.Word, "it"}, {shared.Word, "onto"},
		{shared.Word, "the"}, {shared.Word, "battlefield"}, {shared.Comma, ","},
		{shared.Word, "put"}, {shared.Word, "it"}, {shared.Word, "into"},
		{shared.Word, "your"}, {shared.Word, "hand"}, {shared.Period, "."},
	},
	{
		{shared.Word, "Otherwise"}, {shared.Comma, ","}, {shared.Word, "put"},
		{shared.Word, "it"}, {shared.Word, "into"}, {shared.Word, "your"},
		{shared.Word, "hand"}, {shared.Period, "."},
	},
}

// conditionalLookAtTopBattlefieldBottomRuns are the recognized trailing fallback
// clauses that optionally send the card to the bottom of the controller's
// library when it is not put onto the battlefield. The two wordings ("If you
// don't put the card onto the battlefield, you may put it on the bottom of your
// library." and the "Otherwise," form) are equivalent: an optional move to the
// bottom of the library.
var conditionalLookAtTopBattlefieldBottomRuns = [][]struct {
	kind shared.Kind
	text string
}{
	{
		{shared.Word, "If"}, {shared.Word, "you"}, {shared.Word, "don't"},
		{shared.Word, "put"}, {shared.Word, "the"}, {shared.Word, "card"},
		{shared.Word, "onto"}, {shared.Word, "the"}, {shared.Word, "battlefield"},
		{shared.Comma, ","}, {shared.Word, "you"}, {shared.Word, "may"},
		{shared.Word, "put"}, {shared.Word, "it"}, {shared.Word, "on"},
		{shared.Word, "the"}, {shared.Word, "bottom"}, {shared.Word, "of"},
		{shared.Word, "your"}, {shared.Word, "library"}, {shared.Period, "."},
	},
	{
		{shared.Word, "Otherwise"}, {shared.Comma, ","}, {shared.Word, "you"},
		{shared.Word, "may"}, {shared.Word, "put"}, {shared.Word, "it"},
		{shared.Word, "on"}, {shared.Word, "the"}, {shared.Word, "bottom"},
		{shared.Word, "of"}, {shared.Word, "your"}, {shared.Word, "library"},
		{shared.Period, "."},
	},
}

// permanentCardTypes is the set of card types that can be put onto the
// battlefield, i.e. the expansion of a "permanent card" condition.
var permanentCardTypes = []CardType{
	CardTypeArtifact, CardTypeBattle, CardTypeCreature,
	CardTypeEnchantment, CardTypeLand, CardTypePlaneswalker,
}

// scanConditionalBattlefieldCardTypes scans the card-type condition for the
// look-at-top battlefield sequence. A leading "permanent" word expands to every
// permanent card type; otherwise it falls back to the plain disjunction scan so
// "<type[ or type...]>" still matches. It fails closed on any unknown word.
func scanConditionalBattlefieldCardTypes(tokens []shared.Token, start int) ([]CardType, int, bool) {
	if start < len(tokens) && tokens[start].Kind == shared.Word &&
		strings.EqualFold(tokens[start].Text, "permanent") {
		return permanentCardTypes, start + 1, true
	}
	return scanCardTypeDisjunction(tokens, start)
}

// recognizeConditionalLookAtTopBattlefieldSequence matches the exact resolving
// body "look at the top card of your library. If it's a <type[ or type...]>
// card, you may put it onto the battlefield[ tapped]." optionally followed by a
// mandatory "if you don't put it onto the battlefield, put it into your hand."
// fallback or an optional "if you don't put it onto the battlefield, you may put
// it on the bottom of your library." fallback. The card-type condition is a
// disjunction of named types or the single word "permanent", which expands to
// every permanent card type. It records the disjunctive card types, the tapped
// entry rider, and the fallback disposition text-blind so the
// compiler and lowering stay free of Oracle wording. The whole body must match
// exactly, so any extra text fails closed. Both triggered abilities ("Whenever
// ..., look at ...") and activated abilities ("{cost}: Look at ...") carry this
// body, so both kinds are recognized.
func recognizeConditionalLookAtTopBattlefieldSequence(ability *Ability) bool {
	if ability == nil || (ability.Kind != AbilityTriggered && ability.Kind != AbilityActivated) {
		return false
	}
	tokens := tokensWithinParserSpan(ability.Tokens, ability.BodySpan)
	cursor := 0
	// An activated body begins after the cost colon, so its first word "Look" is
	// capitalized; a triggered body begins mid-sentence with lowercase "look".
	// Fold the leading word so both kinds match the same fixed prefix.
	if len(tokens) == 0 || tokens[0].Kind != shared.Word ||
		!strings.EqualFold(tokens[0].Text, "look") {
		return false
	}
	cursor = 1
	if !matchTokenRun(tokens, &cursor, conditionalLookAtTopPrefix[1:]) {
		return false
	}
	cardTypes, next, ok := scanConditionalBattlefieldCardTypes(tokens, cursor)
	if !ok {
		return false
	}
	cursor = next
	if cursor >= len(tokens) || tokens[cursor].Kind != shared.Word || tokens[cursor].Text != "card" {
		return false
	}
	cursor++
	if !matchTokenRun(tokens, &cursor, conditionalLookAtTopBattlefieldPutRun) {
		return false
	}
	entersTapped := false
	if cursor < len(tokens) && tokens[cursor].Kind == shared.Word && tokens[cursor].Text == "tapped" {
		entersTapped = true
		cursor++
	}
	if cursor >= len(tokens) || tokens[cursor].Kind != shared.Period || tokens[cursor].Text != "." {
		return false
	}
	cursor++

	elseDisposition := LookAtTopBattlefieldElseNone
	if cursor != len(tokens) {
		matched := false
		for _, run := range conditionalLookAtTopBattlefieldHandRuns {
			runCursor := cursor
			if matchTokenRun(tokens, &runCursor, run) && runCursor == len(tokens) {
				elseDisposition = LookAtTopBattlefieldElseHand
				matched = true
				break
			}
		}
		for _, run := range conditionalLookAtTopBattlefieldBottomRuns {
			if matched {
				break
			}
			runCursor := cursor
			if matchTokenRun(tokens, &runCursor, run) && runCursor == len(tokens) {
				elseDisposition = LookAtTopBattlefieldElseBottom
				matched = true
			}
		}
		if !matched {
			return false
		}
	}

	ability.ExactSequence = &ExactSequenceSyntax{
		Kind:                  ExactSequenceConditionalLookAtTopBattlefield,
		Span:                  ability.BodySpan,
		LookAtTopCardTypes:    cardTypes,
		LookAtTopEntersTapped: entersTapped,
		LookAtTopBattlefield:  elseDisposition,
	}
	return true
}
