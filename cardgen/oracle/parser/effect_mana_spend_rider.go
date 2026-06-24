package parser

import (
	"strconv"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// stripCreatureSpellHasteRiderTokens removes the trailing creature-spell haste
// mana-spend rider sentence ("If that mana is spent on a creature spell, it
// gains haste until end of turn.") from tokens so the rider's "haste" is not
// scanned as a keyword the source permanent itself gains. The rider grants haste
// to the creature spell paid with the tagged mana, modeled through the typed
// rider effect rather than as a static keyword on the source.
func stripCreatureSpellHasteRiderTokens(tokens []shared.Token) []shared.Token {
	for i := range tokens {
		if !effectWordsAt(tokens, i, creatureSpellHasteConditionWords...) {
			continue
		}
		end := i
		for end < len(tokens) && tokens[end].Kind != shared.Period {
			end++
		}
		for end < len(tokens) && tokens[end].Kind == shared.Period {
			end++
		}
		result := append([]shared.Token(nil), tokens[:i]...)
		return append(result, tokens[end:]...)
	}
	return tokens
}

// manaSpendRiderWords is the exact leading token sequence of the
// commander-creature-type spend condition.
var manaSpendRiderWords = []string{
	"when", "that", "mana", "is", "spent", "to", "cast",
	"a", "creature", "spell", "that", "shares", "a", "creature",
	"type", "with", "your", "commander",
}

var chosenTypeManaSpendConditionWords = []string{
	"spend", "this", "mana", "only", "to", "cast", "a", "creature", "spell",
	"of", "the", "chosen", "type",
}

var legendaryManaSpendConditionWords = []string{
	"spend", "this", "mana", "only", "to", "cast", "a", "legendary", "spell",
}

// creatureSpellRestrictedConditionWords is the bare restricted spend condition
// "Spend this mana only to cast a creature spell" (Beastcaller Savant, Dwynen's
// Elite). It restricts the tagged mana to creature spells with no further
// qualifier or rider effect.
var creatureSpellRestrictedConditionWords = []string{
	"spend", "this", "mana", "only", "to", "cast", "a", "creature", "spell",
}

// creature-spell haste bonus rider (Arena of Glory, Generator Servant).
var creatureSpellHasteConditionWords = []string{
	"if", "that", "mana", "is", "spent", "on", "a", "creature", "spell",
}

// creatureSpellHasteEffectWords is the bonus granted when the tagged mana pays
// for a creature spell.
var creatureSpellHasteEffectWords = []string{
	"it", "gains", "haste", "until", "end", "of", "turn",
}

var cantBeCounteredSpendEffectWords = []string{
	"and", "that", "spell", "can't", "be", "countered",
}

// chosenTypeOrActivateWords is the continuation that extends the chosen-type
// spend restriction to also permit activating an ability of a creature source of
// the chosen type (Secluded Courtyard).
var chosenTypeOrActivateWords = []string{
	"or", "activate", "an", "ability", "of", "a", "creature", "source",
	"of", "the", "chosen", "type",
}

// recognizeManaSpendRider reports whether the sentence tokens are exactly the
// Path of Ancestry mana-spend rider "When that mana is spent to cast a creature
// spell that shares a creature type with your commander, scry N." and, if so,
// returns its typed syntax. It matches the entire token stream so that an
// unmodeled rider effect or trailing qualifier fails closed.
func recognizeManaSpendRider(tokens []shared.Token) (ManaSpendRiderSyntax, bool) {
	n := len(manaSpendRiderWords)
	// Layout: <n condition words> , scry <integer> [periods...]
	if len(tokens) < n+3 {
		return ManaSpendRiderSyntax{}, false
	}
	if !effectWordsAt(tokens, 0, manaSpendRiderWords...) {
		return ManaSpendRiderSyntax{}, false
	}
	if tokens[n].Kind != shared.Comma {
		return ManaSpendRiderSyntax{}, false
	}
	if !equalWord(tokens[n+1], "scry") {
		return ManaSpendRiderSyntax{}, false
	}
	amountToken := tokens[n+2]
	if amountToken.Kind != shared.Integer {
		return ManaSpendRiderSyntax{}, false
	}
	amount, err := strconv.Atoi(amountToken.Text)
	if err != nil || amount < 1 {
		return ManaSpendRiderSyntax{}, false
	}
	// Only trailing periods may follow the scry amount; any further word or
	// punctuation means extra unmodeled content, so fail closed.
	for i := n + 3; i < len(tokens); i++ {
		if tokens[i].Kind != shared.Period {
			return ManaSpendRiderSyntax{}, false
		}
	}
	return ManaSpendRiderSyntax{
		Span:          shared.SpanOf(tokens),
		ConditionSpan: shared.SpanOf(tokens[:n]),
		EffectSpan:    shared.SpanOf(tokens[n+1 : n+3]),
		Condition:     ManaSpendCastCommanderCreatureType,
		Effect:        ManaSpendRiderEffectScry,
		ScryAmount:    amount,
	}, true
}

// recognizeChosenTypeManaSpendRider reports whether the sentence tokens are the
// chosen-creature-type mana-spend restriction and, if so, returns its typed
// syntax. It recognizes three forms after the shared restriction prefix "Spend
// this mana only to cast a creature spell of the chosen type":
//   - ", and that spell can't be countered." (Cavern of Souls): the spell is
//     additionally made uncounterable.
//   - " or activate an ability of a creature source of the chosen type."
//     (Secluded Courtyard): the spend is also permitted on activated abilities of
//     creature sources of the chosen type.
//   - bare, ending the sentence (Unclaimed Territory, Pillar of Origins): the
//     restriction stands alone.
//
// Any other trailing content fails closed.
func recognizeChosenTypeManaSpendRider(tokens []shared.Token) (ManaSpendRiderSyntax, bool) {
	conditionEnd := len(chosenTypeManaSpendConditionWords)
	if len(tokens) < conditionEnd ||
		!effectWordsAt(tokens, 0, chosenTypeManaSpendConditionWords...) {
		return ManaSpendRiderSyntax{}, false
	}
	condition := ManaSpendCastChosenCreatureType
	effect := ManaSpendRiderEffectUnknown
	effectSpan := shared.Span{}
	tailStart := conditionEnd
	switch {
	case conditionEnd < len(tokens) && tokens[conditionEnd].Kind == shared.Comma:
		effectStart := conditionEnd + 1
		effectEnd := effectStart + len(cantBeCounteredSpendEffectWords)
		if len(tokens) < effectEnd ||
			!effectWordsAt(tokens, effectStart, cantBeCounteredSpendEffectWords...) {
			return ManaSpendRiderSyntax{}, false
		}
		effect = ManaSpendRiderEffectCantBeCountered
		effectSpan = shared.SpanOf(tokens[effectStart:effectEnd])
		tailStart = effectEnd
	case conditionEnd < len(tokens) && equalWord(tokens[conditionEnd], "or"):
		clauseEnd := conditionEnd + len(chosenTypeOrActivateWords)
		if len(tokens) < clauseEnd ||
			!effectWordsAt(tokens, conditionEnd, chosenTypeOrActivateWords...) {
			return ManaSpendRiderSyntax{}, false
		}
		condition = ManaSpendCastOrActivateChosenCreatureType
		tailStart = clauseEnd
	default:
	}
	for i := tailStart; i < len(tokens); i++ {
		if tokens[i].Kind != shared.Period {
			return ManaSpendRiderSyntax{}, false
		}
	}
	return ManaSpendRiderSyntax{
		Span:          shared.SpanOf(tokens),
		ConditionSpan: shared.SpanOf(tokens[:conditionEnd]),
		EffectSpan:    effectSpan,
		Condition:     condition,
		Effect:        effect,
		Restricted:    true,
	}, true
}

// recognizeLegendaryManaSpendRider reports whether the sentence tokens are
// exactly "Spend this mana only to cast a legendary spell" optionally followed by
// ", and that spell can't be countered." (Delighted Halfling) and, if so, returns
// its typed syntax. The trailing can't-be-countered clause is optional so the
// bare restriction is also recognized; any other trailing content fails closed.
func recognizeLegendaryManaSpendRider(tokens []shared.Token) (ManaSpendRiderSyntax, bool) {
	conditionEnd := len(legendaryManaSpendConditionWords)
	if len(tokens) < conditionEnd ||
		!effectWordsAt(tokens, 0, legendaryManaSpendConditionWords...) {
		return ManaSpendRiderSyntax{}, false
	}
	effect := ManaSpendRiderEffectUnknown
	effectSpan := shared.Span{}
	tailStart := conditionEnd
	if conditionEnd < len(tokens) && tokens[conditionEnd].Kind == shared.Comma {
		effectStart := conditionEnd + 1
		effectEnd := effectStart + len(cantBeCounteredSpendEffectWords)
		if len(tokens) < effectEnd ||
			!effectWordsAt(tokens, effectStart, cantBeCounteredSpendEffectWords...) {
			return ManaSpendRiderSyntax{}, false
		}
		effect = ManaSpendRiderEffectCantBeCountered
		effectSpan = shared.SpanOf(tokens[effectStart:effectEnd])
		tailStart = effectEnd
	}
	for i := tailStart; i < len(tokens); i++ {
		if tokens[i].Kind != shared.Period {
			return ManaSpendRiderSyntax{}, false
		}
	}
	return ManaSpendRiderSyntax{
		Span:          shared.SpanOf(tokens),
		ConditionSpan: shared.SpanOf(tokens[:conditionEnd]),
		EffectSpan:    effectSpan,
		Condition:     ManaSpendCastLegendarySpell,
		Effect:        effect,
		Restricted:    true,
	}, true
}

// recognizeCreatureSpellHasteManaSpendRider reports whether the sentence tokens
// are exactly "If that mana is spent on a creature spell, it gains haste until
// end of turn." (Arena of Glory, Generator Servant) and, if so, returns its
// typed syntax. It is an unrestricted bonus rider: the tagged mana may be spent
// on anything, but a creature spell paid for with it gains haste until end of
// turn. Any other trailing content fails closed.
func recognizeCreatureSpellHasteManaSpendRider(tokens []shared.Token) (ManaSpendRiderSyntax, bool) {
	conditionEnd := len(creatureSpellHasteConditionWords)
	effectStart := conditionEnd + 1
	effectEnd := effectStart + len(creatureSpellHasteEffectWords)
	if len(tokens) < effectEnd ||
		!effectWordsAt(tokens, 0, creatureSpellHasteConditionWords...) ||
		tokens[conditionEnd].Kind != shared.Comma ||
		!effectWordsAt(tokens, effectStart, creatureSpellHasteEffectWords...) {
		return ManaSpendRiderSyntax{}, false
	}
	for i := effectEnd; i < len(tokens); i++ {
		if tokens[i].Kind != shared.Period {
			return ManaSpendRiderSyntax{}, false
		}
	}
	return ManaSpendRiderSyntax{
		Span:          shared.SpanOf(tokens),
		ConditionSpan: shared.SpanOf(tokens[:conditionEnd]),
		EffectSpan:    shared.SpanOf(tokens[effectStart:effectEnd]),
		Condition:     ManaSpendCastCreatureSpell,
		Effect:        ManaSpendRiderEffectGainsHasteUntilEndOfTurn,
	}, true
}

// recognizeCreatureSpellRestrictedManaSpendRider reports whether the sentence
// tokens are exactly "Spend this mana only to cast a creature spell." and, if
// so, returns its typed syntax. It is the bare restricted creature-spell spend
// rider (Beastcaller Savant, Dwynen's Elite): the tagged mana may be spent only
// on creature spells. Any trailing qualifier ("of the chosen type", "or activate
// …") is handled by other recognizers, so any extra content here fails closed.
func recognizeCreatureSpellRestrictedManaSpendRider(tokens []shared.Token) (ManaSpendRiderSyntax, bool) {
	conditionEnd := len(creatureSpellRestrictedConditionWords)
	if len(tokens) < conditionEnd ||
		!effectWordsAt(tokens, 0, creatureSpellRestrictedConditionWords...) {
		return ManaSpendRiderSyntax{}, false
	}
	for i := conditionEnd; i < len(tokens); i++ {
		if tokens[i].Kind != shared.Period {
			return ManaSpendRiderSyntax{}, false
		}
	}
	return ManaSpendRiderSyntax{
		Span:          shared.SpanOf(tokens),
		ConditionSpan: shared.SpanOf(tokens[:conditionEnd]),
		Condition:     ManaSpendCastCreatureSpell,
		Effect:        ManaSpendRiderEffectUnknown,
		Restricted:    true,
	}, true
}

// collapseManaSpendRiderSentence replaces a recognized mana-spend rider
// sentence's generic effects with a single typed EffectManaSpendRider effect
// that spans the whole sentence, so the rider rides on the preceding add-mana
// effect rather than splitting into uncoordinated cast/scry effects. It returns
// true when it collapsed the sentence. The synthesized effect credits the full
// sentence span for coverage and round-trips exactly.
func collapseManaSpendRiderSentence(sentence *Sentence, tokens []shared.Token) bool {
	rider, ok := recognizeManaSpendRider(tokens)
	if !ok {
		rider, ok = recognizeChosenTypeManaSpendRider(tokens)
	}
	if !ok {
		rider, ok = recognizeLegendaryManaSpendRider(tokens)
	}
	if !ok {
		rider, ok = recognizeCreatureSpellHasteManaSpendRider(tokens)
	}
	if !ok {
		rider, ok = recognizeCreatureSpellRestrictedManaSpendRider(tokens)
	}
	if !ok {
		return false
	}
	span := shared.SpanOf(tokens)
	riderCopy := rider
	sentence.Effects = []EffectSyntax{{
		Kind:           EffectManaSpendRider,
		VerbSpan:       tokens[0].Span,
		ClauseSpan:     span,
		Span:           span,
		Text:           sentence.Text,
		Tokens:         tokens,
		Exact:          true,
		ManaSpendRider: &riderCopy,
	}}
	return true
}
