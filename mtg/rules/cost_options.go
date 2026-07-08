package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/opt"
)

// evokeAlternativeChosen reports whether the spell cost option selected by the
// payment preferences is the spell's Evoke alternative cost. The normal cost is
// option index 0 and each alternative cost is option index i+1 into the face's
// AlternativeCosts, so index-1 selects the chosen alternative.
func evokeAlternativeChosen(card *game.CardDef, alternativeIndex int) bool {
	index := alternativeIndex - 1
	if card == nil || index < 0 || index >= len(card.AlternativeCosts) {
		return false
	}
	return card.AlternativeCosts[index].Mechanic == cost.AlternativeMechanicEvoke
}

// convertedAlternativeChosen reports whether the spell cost option selected by
// the payment preferences is the spell's "More Than Meets the Eye" alternative
// cost. The normal cost is option index 0 and each alternative cost is option
// index i+1 into the face's AlternativeCosts, so index-1 selects the chosen
// alternative.
func convertedAlternativeChosen(card *game.CardDef, alternativeIndex int) bool {
	index := alternativeIndex - 1
	if card == nil || index < 0 || index >= len(card.AlternativeCosts) {
		return false
	}
	return card.AlternativeCosts[index].Mechanic == cost.AlternativeMechanicMoreThanMeetsTheEye
}

func manaCostPtr(manaCost opt.V[cost.Mana]) *cost.Mana {
	if !manaCost.Exists {
		return nil
	}
	return &manaCost.Val
}

func isFlashbackAlternative(alternative cost.Alternative) bool {
	return alternative.Mechanic == cost.AlternativeMechanicFlashback
}

func isEscapeAlternative(alternative cost.Alternative) bool {
	return alternative.Mechanic == cost.AlternativeMechanicEscape
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
