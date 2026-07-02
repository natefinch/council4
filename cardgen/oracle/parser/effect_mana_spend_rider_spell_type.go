package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// spellTypeSpendPrefixWords is the shared restriction prefix every spell-type
// mana-spend restriction begins with, before the optional "to".
var spellTypeSpendPrefixWords = []string{"spend", "this", "mana", "only"}

// recognizeSpellTypeRestrictedManaSpendRider reports whether the sentence tokens
// are a spell-type-restricted mana-spend restriction "Spend this mana only to
// cast <spell type> spell[s]." and, if so, returns its typed syntax. It
// recognizes a closed set of spell-type selectors, each mapping to one closed
// restriction-only condition:
//
//   - "a creature spell" / "creature spells" → ManaSpendCastCreatureSpell
//     (Somberwald Sage, Metamorphosis; the bare singular form is also owned by
//     recognizeCreatureSpellRestrictedManaSpendRider).
//   - "a noncreature spell" / "noncreature spells" → ManaSpendCastNoncreatureSpell
//     (Nardole, Resourceful Cyborg).
//   - "an instant or sorcery spell" / "instant and/or sorcery spells" /
//     "instant and sorcery spells" → ManaSpendCastInstantOrSorcerySpell (Vodalian
//     Arcanist, Cormela, Glamour Thief, Abstract Paintmage).
//   - "a multicolored spell" / "multicolored spells" → ManaSpendCastMulticoloredSpell
//     (Pillar of the Paruns).
//   - "a planeswalker spell" / "planeswalker spells" → ManaSpendCastPlaneswalkerSpell
//     (Interplanar Beacon).
//
// Any other selector, qualifier, or trailing content fails closed, so wordings
// such as "cast spells", "cast a spell of the chosen type", subtype filters, and
// mana-value or ownership qualifiers stay rejected.
func recognizeSpellTypeRestrictedManaSpendRider(tokens []shared.Token) (ManaSpendRiderSyntax, bool) {
	prefix := len(spellTypeSpendPrefixWords)
	if len(tokens) <= prefix || !effectWordsAt(tokens, 0, spellTypeSpendPrefixWords...) {
		return ManaSpendRiderSyntax{}, false
	}
	index := prefix
	if index < len(tokens) && equalWord(tokens[index], "to") {
		index++
	}
	if index >= len(tokens) || !equalWord(tokens[index], "cast") {
		return ManaSpendRiderSyntax{}, false
	}
	index++
	condition, next, ok := spellTypeSpendSelector(tokens, index)
	if !ok {
		return ManaSpendRiderSyntax{}, false
	}
	for i := next; i < len(tokens); i++ {
		if tokens[i].Kind != shared.Period {
			return ManaSpendRiderSyntax{}, false
		}
	}
	return ManaSpendRiderSyntax{
		Span:          shared.SpanOf(tokens),
		ConditionSpan: shared.SpanOf(tokens[:next]),
		Condition:     condition,
		Effect:        ManaSpendRiderEffectUnknown,
		Restricted:    true,
	}, true
}

// spellTypeSpendSelector parses one spell-type selector "<a|an>? <type> spell[s]"
// (or the "instant <connector> sorcery" union) beginning at start. It returns the
// mapped restriction condition and the index just past "spell"/"spells". Unknown
// selectors fail closed.
func spellTypeSpendSelector(tokens []shared.Token, start int) (ManaSpendConditionKind, int, bool) {
	i := start
	if i < len(tokens) && (equalWord(tokens[i], "a") || equalWord(tokens[i], "an")) {
		i++
	}
	if i >= len(tokens) {
		return ManaSpendConditionUnknown, 0, false
	}
	// The instant-or-sorcery union reads "instant <or|and|and/or> sorcery".
	if equalWord(tokens[i], "instant") {
		j := i + 1
		var connected bool
		if next, ok := consumeDynamicUnionConnector(tokens, j); ok {
			j = next
			connected = true
		} else if j < len(tokens) && equalWord(tokens[j], "and") {
			j++
			connected = true
		}
		if !connected || j >= len(tokens) || !equalWord(tokens[j], "sorcery") {
			return ManaSpendConditionUnknown, 0, false
		}
		end, ok := spellNounEnd(tokens, j+1)
		if !ok {
			return ManaSpendConditionUnknown, 0, false
		}
		return ManaSpendCastInstantOrSorcerySpell, end, true
	}
	condition, ok := spellTypeSpendCondition(tokens[i])
	if !ok {
		return ManaSpendConditionUnknown, 0, false
	}
	end, ok := spellNounEnd(tokens, i+1)
	if !ok {
		return ManaSpendConditionUnknown, 0, false
	}
	return condition, end, true
}

// spellTypeSpendCondition maps a single spell-type word to its restriction
// condition, failing closed on any unmodeled spell type.
func spellTypeSpendCondition(token shared.Token) (ManaSpendConditionKind, bool) {
	switch {
	case equalWord(token, "creature"):
		return ManaSpendCastCreatureSpell, true
	case equalWord(token, "noncreature"):
		return ManaSpendCastNoncreatureSpell, true
	case equalWord(token, "multicolored"):
		return ManaSpendCastMulticoloredSpell, true
	case equalWord(token, "planeswalker"):
		return ManaSpendCastPlaneswalkerSpell, true
	default:
		return ManaSpendConditionUnknown, false
	}
}

// spellNounEnd requires the singular or plural spell noun at index and returns
// the index just past it, failing closed on any other word.
func spellNounEnd(tokens []shared.Token, index int) (int, bool) {
	if index >= len(tokens) ||
		(!equalWord(tokens[index], "spell") && !equalWord(tokens[index], "spells")) {
		return 0, false
	}
	return index + 1, true
}
