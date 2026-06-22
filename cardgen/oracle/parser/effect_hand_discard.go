package parser

import (
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// recognizeRevealChooseHandDiscardSequence recognizes the closed targeted hand-
// disruption family "Target player reveals their hand. You choose a [filter]
// card from it. That player discards that card." (Coercion, Duress,
// Thoughtseize, Inquisition of Kozilek). The resolving controller looks at the
// target player's revealed hand and chooses one card matching the filter, which
// that player then discards. A trailing "You lose N life." rider (Thoughtseize)
// is left as its own effect and lowered separately.
//
// The middle "You choose a [filter] card from it." sentence carries no
// resolving effect, so the parser owns its wording here: it confirms the
// reveal-hand / choose-card / discard-that-card shape across three consecutive
// sentences, extracts the noncreature / nonland / mana-value filter from the
// choose sentence, and folds it onto the EffectDiscard half via
// HandChoiceDiscard (with the choose sentence's coverage span) plus the shared
// RevealChooseDiscard marker on both the reveal and the discard. The text-blind
// lowering reads only those typed fields. Any filter or shape it does not model
// fails closed, leaving the generic sequence path untouched.
func recognizeRevealChooseHandDiscardSequence(sentences []Sentence) {
	for i := 0; i+2 < len(sentences); i++ {
		reveal := &sentences[i]
		choose := &sentences[i+1]
		discard := &sentences[i+2]
		if !isRevealHandSentence(reveal) ||
			len(choose.Effects) != 0 ||
			!isDiscardThatCardSentence(discard) {
			continue
		}
		filter, ok := parseHandChoiceDiscardFilter(choose.Text, choose.Span)
		if !ok {
			continue
		}
		reveal.Effects[0].Exact = true
		reveal.Effects[0].RevealChooseDiscard = true
		discard.Effects[0].Exact = true
		discard.Effects[0].RevealChooseDiscard = true
		discard.Effects[0].HandChoiceDiscard = filter
		return
	}
}

// isRevealHandSentence reports whether the sentence is a lone "Target player
// reveals their hand." reveal of a target player's hand.
func isRevealHandSentence(sentence *Sentence) bool {
	if len(sentence.Effects) != 1 || len(sentence.Targets) == 0 {
		return false
	}
	effect := sentence.Effects[0]
	return effect.Kind == EffectReveal &&
		effect.Context == EffectContextTarget &&
		strings.Contains(strings.ToLower(effect.Selection.Text), "hand")
}

// isDiscardThatCardSentence reports whether the sentence is a lone "That player
// discards that card." discard performed by the referenced (revealing) player.
func isDiscardThatCardSentence(sentence *Sentence) bool {
	if len(sentence.Effects) != 1 {
		return false
	}
	effect := sentence.Effects[0]
	return effect.Kind == EffectDiscard &&
		effect.Context == EffectContextReferencedPlayer
}

// parseHandChoiceDiscardFilter parses the middle "You choose a [filter] card
// from it." sentence into a typed filter. It accepts the any-card, noncreature,
// and nonland descriptors, the "from it" / "from among them" / "from those
// cards" source phrasings, and an optional "with mana value N or less" bound;
// every other wording fails closed.
func parseHandChoiceDiscardFilter(text string, span shared.Span) (HandChoiceDiscardSyntax, bool) {
	normalized := strings.TrimSuffix(strings.TrimSpace(strings.ToLower(text)), ".")
	const prefix = "you choose a "
	rest, ok := strings.CutPrefix(normalized, prefix)
	if !ok {
		return HandChoiceDiscardSyntax{}, false
	}
	descriptor, after, ok := strings.Cut(rest, "card")
	if !ok {
		return HandChoiceDiscardSyntax{}, false
	}
	filter := HandChoiceDiscardSyntax{Present: true, ChooseSpan: span}
	if descriptor = strings.TrimSpace(descriptor); descriptor != "" {
		for token := range strings.FieldsSeq(strings.ReplaceAll(descriptor, ",", " ")) {
			switch token {
			case "noncreature":
				filter.ExcludeCreature = true
			case "nonland":
				filter.ExcludeLand = true
			default:
				return HandChoiceDiscardSyntax{}, false
			}
		}
	}
	bound, ok := cutHandChoiceSource(strings.TrimSpace(after))
	if !ok {
		return HandChoiceDiscardSyntax{}, false
	}
	if bound != "" {
		value, ok := parseManaValueOrLess(bound)
		if !ok {
			return HandChoiceDiscardSyntax{}, false
		}
		filter.HasMaxManaValue = true
		filter.MaxManaValue = value
	}
	return filter, true
}

// cutHandChoiceSource strips the "from it" / "from among them" / "from those
// cards" source phrasing from the choose sentence's remainder and returns any
// trailing bound clause. It fails closed when no recognized source is present.
func cutHandChoiceSource(after string) (string, bool) {
	for _, source := range []string{"from among them", "from those cards", "from it"} {
		if rest, ok := strings.CutPrefix(after, source); ok {
			return strings.TrimSpace(rest), true
		}
	}
	return "", false
}

// parseManaValueOrLess parses a "with mana value N or less" bound into N.
func parseManaValueOrLess(text string) (int, bool) {
	const prefix = "with mana value "
	rest, ok := strings.CutPrefix(text, prefix)
	if !ok {
		return 0, false
	}
	number, ok := strings.CutSuffix(rest, " or less")
	if !ok {
		return 0, false
	}
	value, err := strconv.Atoi(strings.TrimSpace(number))
	if err != nil || value < 0 {
		return 0, false
	}
	return value, true
}
