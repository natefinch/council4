package cardgen

import "slices"

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
