package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

// costModificationContext is the future attachment point for cost increases,
// reductions, and taxes produced by static or continuous effects.
type costModificationContext struct {
	player     game.PlayerID
	card       *game.CardDef
	cardID     id.ID
	sourceZone game.ZoneType
	option     spellCostOption
}

// spellPaymentRequest bundles all parameters needed to check or pay spell costs,
// replacing the old WithKickerFromZoneAndPreferences overload chain.
type spellPaymentRequest struct {
	playerID   game.PlayerID
	cardID     id.ID
	sourceZone game.ZoneType
	card       *game.CardDef
	xValue     int
	kickerPaid bool
	prefs      *paymentPreferences
}

// abilityPaymentRequest bundles all parameters needed to check or pay ability costs,
// replacing the old WithPreferences overload chain.
type abilityPaymentRequest struct {
	playerID game.PlayerID
	source   *game.Permanent
	ability  *game.AbilityDef
	xValue   int
	prefs    *paymentPreferences
}

type paymentPreferences struct {
	alternativeIndex     int
	phyrexianLifeChoices []bool
	phyrexianIndex       int
	sacrificeChoices     []id.ID
	discardChoices       []id.ID
}

func (p *paymentPreferences) nextPhyrexianLifeChoice() bool {
	if p == nil || p.phyrexianIndex >= len(p.phyrexianLifeChoices) {
		return false
	}
	choice := p.phyrexianLifeChoices[p.phyrexianIndex]
	p.phyrexianIndex++
	return choice
}
