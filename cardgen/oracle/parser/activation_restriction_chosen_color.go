package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// spendChosenColorManaActivationWords is the exact trailing restriction "Spend
// only mana of the chosen color to activate this ability." (Throne of Eldraine),
// which limits the ability's cost payment to the source's entry-time chosen
// color. The wording is fixed and card-specific, so it is matched literally and
// fails closed on any other selector.
var spendChosenColorManaActivationWords = []string{
	"spend", "only", "mana", "of", "the", "chosen", "color",
	"to", "activate", "this", "ability",
}

// parseSpendChosenColorManaActivationRestriction recognizes the exact trailing
// sentence "Spend only mana of the chosen color to activate this ability." and
// returns an ActivationRestrictionSpendChosenColorMana clause spanning the whole
// sentence. Any deviation returns ok=false so the "Activate only …" path is
// tried instead.
func parseSpendChosenColorManaActivationRestriction(
	tokens []shared.Token,
	fullSpan shared.Span,
) (ActivationRestriction, bool) {
	if len(tokens) != len(spendChosenColorManaActivationWords) ||
		!effectWordsAt(tokens, 0, spendChosenColorManaActivationWords...) {
		return ActivationRestriction{}, false
	}
	return ActivationRestriction{
		Kind: ActivationRestrictionSpendChosenColorMana,
		Span: fullSpan,
	}, true
}
