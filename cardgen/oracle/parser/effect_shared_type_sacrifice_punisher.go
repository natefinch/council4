package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// sharedTypeSacrificePunisherBody is the fixed token run of the Braids, Arisen
// Nightmare resolving body, matched verbatim after the "At the beginning of your
// end step," trigger. The controller's optional sacrifice, the each-opponent
// shared-card-type sacrifice offer, and the per-declining-opponent "lose 2 life
// and you draw a card" punisher together form one construct the ordinary effect
// vocabulary cannot compose (the "shares a card type with it" back-reference and
// the distributive "for each opponent who doesn't" gate), so the body is captured
// whole and lowered from the kind alone. The trailing "2" is the only integer;
// it is matched literally so any other life amount fails closed.
var sharedTypeSacrificePunisherBody = []struct {
	kind shared.Kind
	text string
}{
	{shared.Word, "you"}, {shared.Word, "may"}, {shared.Word, "sacrifice"},
	{shared.Word, "an"}, {shared.Word, "artifact"}, {shared.Comma, ","},
	{shared.Word, "creature"}, {shared.Comma, ","}, {shared.Word, "enchantment"},
	{shared.Comma, ","}, {shared.Word, "land"}, {shared.Comma, ","},
	{shared.Word, "or"}, {shared.Word, "planeswalker"}, {shared.Period, "."},
	{shared.Word, "If"}, {shared.Word, "you"}, {shared.Word, "do"},
	{shared.Comma, ","}, {shared.Word, "each"}, {shared.Word, "opponent"},
	{shared.Word, "may"}, {shared.Word, "sacrifice"}, {shared.Word, "a"},
	{shared.Word, "permanent"}, {shared.Word, "of"}, {shared.Word, "their"},
	{shared.Word, "choice"}, {shared.Word, "that"}, {shared.Word, "shares"},
	{shared.Word, "a"}, {shared.Word, "card"}, {shared.Word, "type"},
	{shared.Word, "with"}, {shared.Word, "it"}, {shared.Period, "."},
	{shared.Word, "For"}, {shared.Word, "each"}, {shared.Word, "opponent"},
	{shared.Word, "who"}, {shared.Word, "doesn't"}, {shared.Comma, ","},
	{shared.Word, "that"}, {shared.Word, "player"}, {shared.Word, "loses"},
	{shared.Integer, "2"}, {shared.Word, "life"}, {shared.Word, "and"},
	{shared.Word, "you"}, {shared.Word, "draw"}, {shared.Word, "a"},
	{shared.Word, "card"}, {shared.Period, "."},
}

// recognizeSharedTypeSacrificePunisherSequence matches the exact triggered
// resolving body of Braids, Arisen Nightmare and, on success, marks the ability
// with ExactSequenceSharedTypeSacrificePunisher spanning the whole body. The
// lowering separately gates on the controller's own end-step trigger, so this
// recognizer never inspects the trigger words; it matches only a triggered
// ability whose body is exactly the fixed run with no other content, and fails
// closed on any deviation.
func recognizeSharedTypeSacrificePunisherSequence(ability *Ability) bool {
	if ability == nil || ability.Kind != AbilityTriggered {
		return false
	}
	tokens := tokensWithinParserSpan(ability.Tokens, ability.BodySpan)
	cursor := 0
	if !matchTokenRun(tokens, &cursor, sharedTypeSacrificePunisherBody) {
		return false
	}
	if cursor != len(tokens) {
		return false
	}
	ability.ExactSequence = &ExactSequenceSyntax{
		Kind: ExactSequenceSharedTypeSacrificePunisher,
		Span: ability.BodySpan,
	}
	return true
}
