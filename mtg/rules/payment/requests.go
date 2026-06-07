package payment

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SpellRequest bundles all parameters needed to check or pay spell costs.
type SpellRequest struct {
	PlayerID   game.PlayerID
	CardID     id.ID
	SourceZone zone.Type
	Card       *game.CardDef
	XValue     int
	KickerPaid bool
	Prefs      *Preferences
}

// AbilityRequest bundles all parameters needed to check or pay activated
// ability costs.
type AbilityRequest struct {
	PlayerID         game.PlayerID
	Source           *game.Permanent
	SourceCardID     id.ID
	SourceZone       zone.Type
	ManaCost         opt.V[cost.Mana]
	AdditionalCosts  []cost.Additional
	AlternativeCosts []cost.Alternative
	XValue           int
	Prefs            *Preferences
}

// GenericRequest bundles parameters for a generic mana payment — used by
// attack taxes, Cycling, Ward, Madness, Suspend, and resolution-payment effects
// that do not have a full card context.
type GenericRequest struct {
	PlayerID        game.PlayerID
	Cost            *cost.Mana
	XValue          int
	Exclude         map[id.ID]bool
	AdditionalCosts []cost.Additional
	Prefs           *Preferences
}

// Preferences records the player's choices about how to pay optional or
// alternative cost components. It is produced by the Engine's choice layer
// before payment execution and consumed by the planner as a preference hint.
type Preferences struct {
	AlternativeIndex     int
	PhyrexianLifeChoices []bool
	phyrexianIndex       int
	SacrificeChoices     []id.ID
	DiscardChoices       []id.ID
	ExileChoices         []id.ID
}

// SpellOptionSummary is a summary of one payable spell cost option for choice presentation.
type SpellOptionSummary struct {
	Index           int
	Label           string
	ManaCost        *cost.Mana
	AdditionalCosts []cost.Additional
}

// NextPhyrexianLifeChoice returns the next phyrexian payment preference,
// advancing the internal cursor. Returns false (pay mana) when no preference is
// recorded or the list is exhausted.
func (p *Preferences) NextPhyrexianLifeChoice() bool {
	if p == nil || p.phyrexianIndex >= len(p.PhyrexianLifeChoices) {
		return false
	}
	choice := p.PhyrexianLifeChoices[p.phyrexianIndex]
	p.phyrexianIndex++
	return choice
}
