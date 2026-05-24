package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

// paymentTransaction is the internal rules plan for paying a spell, ability,
// or future attack cost. Current payment code still uses paymentPlan; Phase 9B
// migrates behavior into this richer shape one slice at a time.
type paymentTransaction struct {
	poolSpend        []mana.Unit
	manaTaps         []manaTap
	lifePayments     map[game.PlayerID]int
	symbolPayments   []game.SymbolPayment
	additionalCosts  []game.AdditionalCostSelection
	alternativeLabel string
}

type paymentSymbolOption struct {
	symbol  mana.Symbol
	method  game.SymbolPaymentMethod
	color   mana.Color
	generic int
	life    int
	snow    bool
}

type paymentAdditionalCostOption struct {
	cost         game.AdditionalCost
	permanents   []id.ID
	cards        []id.ID
	lifeRequired int
}

// costModificationContext is the future attachment point for cost increases,
// reductions, and taxes produced by static or continuous effects.
type costModificationContext struct {
	player game.PlayerID
	option spellCostOption
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
