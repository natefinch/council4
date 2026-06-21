package cardgen

import (
	"slices"
	"strings"
)

// CorpusExclusionReason identifies why a Scryfall record is outside the card
// generation corpus.
type CorpusExclusionReason string

// Corpus exclusion reasons are stable report values.
const (
	ExcludeAlchemy                 CorpusExclusionReason = "alchemy"
	ExcludeDigitalOnly             CorpusExclusionReason = "digital-only"
	ExcludeMemorabilia             CorpusExclusionReason = "memorabilia"
	ExcludeNoSanctionedPaperFormat CorpusExclusionReason = "no-sanctioned-paper-legality"
	ExcludeSpecialFormat           CorpusExclusionReason = "special-format"
)

// disownedCardNames lists cards Wizards of the Coast has officially disowned and
// removed from supported play for containing racist or culturally offensive
// content (announced 2020). These cards are never generated and never appear in
// any supported, unsupported, or excluded list — they are treated as if they do
// not exist, regardless of any other corpus rule.
var disownedCardNames = map[string]struct{}{
	"invoke prejudice":      {},
	"cleanse":               {},
	"stone-throwing devils": {},
	"pradesh gypsies":       {},
	"jihad":                 {},
	"imprison":              {},
	"crusade":               {},
}

// DisownedCard reports whether the card is one Wizards of the Coast has officially
// disowned. Disowned cards must be omitted entirely: never generated and never
// listed as supported, unsupported, or excluded. The match is on the card's name,
// case-insensitively, and applies before every other corpus rule.
func DisownedCard(card ScryfallCard) bool {
	_, disowned := disownedCardNames[strings.ToLower(strings.TrimSpace(card.Name))]
	return disowned
}

// CorpusPolicy selects the Scryfall records considered for card generation.
type CorpusPolicy struct{}

// Exclusion reports whether card is outside the supported corpus policy.
func (CorpusPolicy) Exclusion(card ScryfallCard) (CorpusExclusionReason, bool) {
	switch card.SetType {
	case "alchemy":
		return ExcludeAlchemy, true
	case "memorabilia":
		return ExcludeMemorabilia, true
	case "minigame":
		return ExcludeSpecialFormat, true
	case "funny":
		if !hasSanctionedPaperLegality(card.Legalities) {
			return ExcludeNoSanctionedPaperFormat, true
		}
	default:
	}
	switch card.Layout {
	case "art_series", "emblem", "planar", "scheme", "vanguard":
		return ExcludeSpecialFormat, true
	case "token", "double_faced_token":
		if !slices.Contains(card.Games, "paper") {
			return ExcludeDigitalOnly, true
		}
		if !hasPermanentFace(card) {
			return ExcludeSpecialFormat, true
		}
		return "", false
	}
	if hasSanctionedPaperLegality(card.Legalities) {
		return "", false
	}
	if card.Digital || !slices.Contains(card.Games, "paper") {
		return ExcludeDigitalOnly, true
	}
	return ExcludeNoSanctionedPaperFormat, true
}

func hasPermanentFace(card ScryfallCard) bool {
	if typeLineHasPermanentType(card.TypeLine) {
		return true
	}
	return slices.ContainsFunc(card.CardFaces, func(face ScryfallCardFace) bool {
		return typeLineHasPermanentType(face.TypeLine)
	})
}

func typeLineHasPermanentType(typeLine string) bool {
	parsed := ParseTypeLine(typeLine)
	return slices.ContainsFunc(parsed.Types, func(cardType string) bool {
		switch cardType {
		case "Artifact", "Battle", "Creature", "Enchantment", "Land", "Planeswalker":
			return true
		default:
			return false
		}
	})
}

func hasSanctionedPaperLegality(legalities map[string]string) bool {
	for _, format := range []string{
		"standard",
		"pioneer",
		"modern",
		"legacy",
		"pauper",
		"vintage",
		"commander",
	} {
		switch legalities[format] {
		case "legal", "restricted", "banned":
			return true
		}
	}
	return false
}
