package parser

import (
	"strconv"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

// PayRepeatedlyAnimateSyntax holds the structured fields of an
// EffectPayRepeatedlyAnimate effect (Primal Adversary's enters trigger). Cost is
// the repeatable mana cost offered any number of times; CounterPower and
// CounterToughness are the dimensions of the per-payment counters placed on the
// source ("+1/+1 counters"); LandPower and LandToughness are the base
// power/toughness the chosen lands gain; LandSubtypes are the creature subtype(s)
// added to each chosen land; and LandKeywords are the keyword(s) granted to each
// chosen land. The number of times paid sizes both the counters and the maximum
// number of lands animated.
type PayRepeatedlyAnimateSyntax struct {
	Cost             cost.Mana
	CounterPower     int
	CounterToughness int
	LandPower        int
	LandToughness    int
	LandSubtypes     []types.Sub
	LandKeywords     []KeywordKind
}

// recognizePayRepeatedlyAnimateSequence folds the exact two-sentence Primal
// Adversary enters trigger onto a single typed EffectPayRepeatedlyAnimate effect.
// The first sentence is the repeatable payment offer ("you may pay {1}{G} any
// number of times.") and the second is the reflexive consequence keyed on the
// number of times paid ("When you pay this cost one or more times, put that many
// +1/+1 counters on this creature, then up to that many target lands you control
// become 3/3 Wolf creatures with haste that are still lands."). It runs after
// the per-sentence effect scan, replacing the mangled per-sentence effects with
// one effect whose Span covers the whole body; stripPayRepeatedlyAnimateSemantics
// then clears the residual optional, reference, keyword, and condition semantics
// so the collapsed effect is the ability's sole resolving content. Any other
// shape leaves the ability untouched so it stays unsupported and fails closed.
func recognizePayRepeatedlyAnimateSequence(ability *Ability) {
	if ability.Kind != AbilityTriggered {
		return
	}
	payIndex, reflexIndex, ok := payRepeatedlyAnimateSentenceIndices(ability.Sentences)
	if !ok {
		return
	}
	paymentSentence := &ability.Sentences[payIndex]
	reflexSentence := &ability.Sentences[reflexIndex]
	manaCost, ok := parsePayRepeatedlyOffer(semanticEffectTokens(paymentSentence.Tokens))
	if !ok {
		return
	}
	payload, ok := parsePayRepeatedlyAnimateReflex(semanticEffectTokens(reflexSentence.Tokens), manaCost)
	if !ok {
		return
	}

	body := append(cloneTokens(paymentSentence.Tokens), cloneTokens(reflexSentence.Tokens)...)
	span := shared.Span{Start: paymentSentence.Span.Start, End: reflexSentence.Span.End}
	effect := EffectSyntax{
		Kind:                 EffectPayRepeatedlyAnimate,
		Span:                 span,
		ClauseSpan:           span,
		Tokens:               body,
		PayRepeatedlyAnimate: payload,
	}
	paymentSentence.Effects = []EffectSyntax{effect}
	paymentSentence.Targets = nil
	reflexSentence.Effects = nil
	reflexSentence.Targets = nil
}

// payRepeatedlyAnimateSentenceIndices returns the indices of the two semantic
// sentences (the payment offer and the reflexive consequence) when the ability
// has exactly two sentences carrying effect tokens, failing otherwise.
func payRepeatedlyAnimateSentenceIndices(sentences []Sentence) (payIndex, reflexIndex int, ok bool) {
	var semantic []int
	for i := range sentences {
		if len(semanticEffectTokens(sentences[i].Tokens)) != 0 {
			semantic = append(semantic, i)
		}
	}
	if len(semantic) != 2 {
		return 0, 0, false
	}
	return semantic[0], semantic[1], true
}

// parsePayRepeatedlyOffer matches the payment-offer sentence "you may pay <mana>
// any number of times." and returns the repeatable mana cost.
func parsePayRepeatedlyOffer(tokens []shared.Token) (cost.Mana, bool) {
	if !effectWordsAt(tokens, 0, "you", "may", "pay") {
		return nil, false
	}
	manaCost, end, ok := parseKeywordManaCost(tokens, 3)
	if !ok {
		return nil, false
	}
	if !effectWordsAt(tokens, end, "any", "number", "of", "times") {
		return nil, false
	}
	end += 4
	if end != len(tokens)-1 || tokens[end].Kind != shared.Period {
		return nil, false
	}
	return manaCost, true
}

// parsePayRepeatedlyAnimateReflex matches the reflexive consequence sentence
// "When you pay this cost one or more times, put that many +N/+N counters on
// this creature, then up to that many target lands you control become P/T
// <subtype...> creatures with <keyword...> that are still lands." and returns the
// typed payload sized by the supplied repeatable mana cost.
func parsePayRepeatedlyAnimateReflex(tokens []shared.Token, manaCost cost.Mana) (*PayRepeatedlyAnimateSyntax, bool) {
	if !effectWordsAt(tokens, 0, "when", "you", "pay", "this", "cost", "one", "or", "more", "times") {
		return nil, false
	}
	index := 9
	if index >= len(tokens) || tokens[index].Kind != shared.Comma {
		return nil, false
	}
	index++
	if !effectWordsAt(tokens, index, "put", "that", "many") {
		return nil, false
	}
	index += 3
	counters, ok := parseSignedPowerToughness(tokens, index)
	if !ok {
		return nil, false
	}
	index = counters.Next
	if !effectWordsAt(tokens, index, "counters", "on", "this", "creature") {
		return nil, false
	}
	index += 4
	if index >= len(tokens) || tokens[index].Kind != shared.Comma {
		return nil, false
	}
	index++
	if !effectWordsAt(tokens, index, "then", "up", "to", "that", "many", "target", "lands", "you", "control", "become") {
		return nil, false
	}
	index += 10
	land, ok := parsePowerToughness(tokens, index)
	if !ok {
		return nil, false
	}
	index = land.Next
	subtypes, index, ok := parseAnimatedSubtypeRun(tokens, index)
	if !ok {
		return nil, false
	}
	if !effectWordsAt(tokens, index, "creatures", "with") {
		return nil, false
	}
	index += 2
	keywords, index, ok := parseAnimatedKeywordRun(tokens, index)
	if !ok {
		return nil, false
	}
	if !effectWordsAt(tokens, index, "that", "are", "still", "lands") {
		return nil, false
	}
	index += 4
	if index != len(tokens)-1 || tokens[index].Kind != shared.Period {
		return nil, false
	}
	return &PayRepeatedlyAnimateSyntax{
		Cost:             manaCost,
		CounterPower:     counters.Power,
		CounterToughness: counters.Toughness,
		LandPower:        land.Power,
		LandToughness:    land.Toughness,
		LandSubtypes:     subtypes,
		LandKeywords:     keywords,
	}, true
}

// powerToughness is a parsed "N/N" power/toughness pair together with the token
// index immediately past the matched sequence.
type powerToughness struct {
	Power     int
	Toughness int
	Next      int
}

// parseSignedPowerToughness matches a "+N/+N" counter dimension at index.
func parseSignedPowerToughness(tokens []shared.Token, index int) (powerToughness, bool) {
	if index+4 >= len(tokens) ||
		tokens[index].Kind != shared.Plus ||
		tokens[index+1].Kind != shared.Integer ||
		tokens[index+2].Kind != shared.Slash ||
		tokens[index+3].Kind != shared.Plus ||
		tokens[index+4].Kind != shared.Integer {
		return powerToughness{}, false
	}
	power, err := strconv.Atoi(tokens[index+1].Text)
	if err != nil {
		return powerToughness{}, false
	}
	toughness, err := strconv.Atoi(tokens[index+4].Text)
	if err != nil {
		return powerToughness{}, false
	}
	return powerToughness{Power: power, Toughness: toughness, Next: index + 5}, true
}

// parsePowerToughness matches an "N/N" base power/toughness at index.
func parsePowerToughness(tokens []shared.Token, index int) (powerToughness, bool) {
	if index+2 >= len(tokens) ||
		tokens[index].Kind != shared.Integer ||
		tokens[index+1].Kind != shared.Slash ||
		tokens[index+2].Kind != shared.Integer {
		return powerToughness{}, false
	}
	power, err := strconv.Atoi(tokens[index].Text)
	if err != nil {
		return powerToughness{}, false
	}
	toughness, err := strconv.Atoi(tokens[index+2].Text)
	if err != nil {
		return powerToughness{}, false
	}
	return powerToughness{Power: power, Toughness: toughness, Next: index + 3}, true
}

// parseAnimatedSubtypeRun consumes one or more consecutive creature-subtype words
// (the animated lands' added types, "Wolf") up to the following "creatures" noun.
func parseAnimatedSubtypeRun(tokens []shared.Token, index int) ([]types.Sub, int, bool) {
	var subtypes []types.Sub
	for index < len(tokens) {
		subtype, ok := recognizeSubtypePhrase(tokens[index].Text)
		if !ok {
			break
		}
		subtypes = append(subtypes, subtype)
		index++
	}
	if len(subtypes) == 0 {
		return nil, index, false
	}
	return subtypes, index, true
}

// parseAnimatedKeywordRun consumes one or more keyword names (the animated lands'
// granted keywords, "haste"), separated by commas or "and", up to the closing
// "that are still lands" clause introduced by "that".
func parseAnimatedKeywordRun(tokens []shared.Token, index int) ([]KeywordKind, int, bool) {
	var keywords []KeywordKind
	for index < len(tokens) && !equalWord(tokens[index], "that") {
		if tokens[index].Kind == shared.Comma || equalWord(tokens[index], "and") {
			index++
			continue
		}
		kind, length, ok := recognizeKeywordNameAt(tokens, index)
		if !ok {
			return nil, index, false
		}
		keywords = append(keywords, kind)
		index += length
	}
	if len(keywords) == 0 {
		return nil, index, false
	}
	return keywords, index, true
}

// abilityHasPayRepeatedlyAnimate reports whether the ability carries a recognized
// EffectPayRepeatedlyAnimate effect.
func abilityHasPayRepeatedlyAnimate(ability *Ability) bool {
	for i := range ability.Sentences {
		for j := range ability.Sentences[i].Effects {
			if ability.Sentences[i].Effects[j].Kind == EffectPayRepeatedlyAnimate {
				return true
			}
		}
	}
	return false
}

// stripPayRepeatedlyAnimateSemantics clears the residual optional, reference,
// keyword, condition, and static-declaration semantics of an ability whose body
// collapsed onto a single EffectPayRepeatedlyAnimate effect. The collapsed effect
// owns the whole body, so the "you may" optional, the "one or more times"
// condition, and the per-sentence references/keywords/declarations the general
// scans re-derived would otherwise leave the ability over-counted and fail the
// lowering coverage gate. It mirrors stripImpulseExileSemantics and runs after
// emitSemanticAccessors re-derives those fields.
func stripPayRepeatedlyAnimateSemantics(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		if !abilityHasPayRepeatedlyAnimate(ability) {
			continue
		}
		ability.Optional = false
		ability.OptionalSpan = shared.Span{}
		ability.SemanticReferences = nil
		ability.SemanticKeywords = nil
		ability.ConditionBoundaries = nil
		ability.EventHistoryConditions = nil
		ability.ConditionClauses = nil
		ability.ConditionSegments = nil
		ability.TriggerConditionSegments = nil
		ability.StaticDeclarations = nil
	}
}
