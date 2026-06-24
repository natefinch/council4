package rules

import (
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/opt"
)

const flashbackAlternativeLabel = "Flashback"

// escapeAlternativeLabel is the label carried by the Escape alternative cost so
// the cast flow can recognize a graveyard escape cast (CR 702.139).
const escapeAlternativeLabel = "Escape"

// evokeAlternativeLabel is the label carried by the Evoke alternative cost so
// the cast flow can recognize that a spell was cast for its Evoke cost and mark
// the resulting permanent to be sacrificed when it enters (CR 702.74).
const evokeAlternativeLabel = "Evoke"

// evokeAlternativeChosen reports whether the spell cost option selected by the
// payment preferences is the spell's Evoke alternative cost. The normal cost is
// option index 0 and each alternative cost is option index i+1 into the face's
// AlternativeCosts, so index-1 selects the chosen alternative.
func evokeAlternativeChosen(card *game.CardDef, alternativeIndex int) bool {
	index := alternativeIndex - 1
	if card == nil || index < 0 || index >= len(card.AlternativeCosts) {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(card.AlternativeCosts[index].Label), evokeAlternativeLabel)
}

func manaCostPtr(manaCost opt.V[cost.Mana]) *cost.Mana {
	if !manaCost.Exists {
		return nil
	}
	return &manaCost.Val
}

func isFlashbackAlternative(alternative cost.Alternative) bool {
	return strings.EqualFold(strings.TrimSpace(alternative.Label), flashbackAlternativeLabel)
}

func isEscapeAlternative(alternative cost.Alternative) bool {
	return strings.EqualFold(strings.TrimSpace(alternative.Label), escapeAlternativeLabel)
}

func spellAdditionalCosts(card *game.CardDef) []cost.Additional {
	if card == nil {
		return nil
	}
	return append([]cost.Additional(nil), card.AdditionalCosts...)
}

func abilityAdditionalCosts(additionalCosts []cost.Additional) []cost.Additional {
	return append([]cost.Additional(nil), additionalCosts...)
}
