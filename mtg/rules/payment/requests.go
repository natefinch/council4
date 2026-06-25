package payment

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SpellCastPermission identifies the permission selected to cast a spell.
type SpellCastPermission uint8

const (
	// SpellCastPermissionDefault is the ordinary permission for the card's zone.
	SpellCastPermissionDefault SpellCastPermission = iota
	// SpellCastPermissionRuleEffect is an independent permission granted by a rule effect.
	SpellCastPermissionRuleEffect
	// SpellCastPermissionFlashback is the permission supplied by flashback.
	SpellCastPermissionFlashback
	// SpellCastPermissionEscape is the permission supplied by escape. Like
	// flashback it authorizes casting only from the graveyard via the escape
	// alternative cost, but the spell is not exiled afterward so it can be
	// escaped again.
	SpellCastPermissionEscape
)

// SpellRequest bundles all parameters needed to check or pay spell costs.
type SpellRequest struct {
	PlayerID        game.PlayerID
	CardID          id.ID
	SourceZone      zone.Type
	Card            *game.CardDef
	XValue          int
	KickerPaid      bool
	KickerCount     int
	ChosenModes     []int
	Alternative     opt.V[cost.Alternative]
	CastPermissions []SpellCastPermission
	Prefs           *Preferences
	// Targets are the spell's chosen targets, supplied so target-dependent cost
	// modifiers ("Spells your opponents cast that target this creature cost {N}
	// more to cast.") can match. It is empty when the spell has no targets or is
	// being cost-checked before targets are announced.
	Targets []game.Target
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
	// ForMana marks the {T} cost of a mana ability so the source's tap records
	// tapped-for-mana provenance.
	ForMana bool
}

// GenericRequest bundles parameters for a generic mana payment. Spell is set
// only when the payment is part of casting a spell outside the normal spell
// planner, such as a madness cost.
type GenericRequest struct {
	PlayerID        game.PlayerID
	SourceCardID    id.ID
	Spell           *game.CardDef
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
	TapChoices           []id.ID
	ReturnChoices        []id.ID
	DiscardChoices       []id.ID
	ExileChoices         []id.ID
	RevealChoices        []id.ID
	EvidenceChoices      []id.ID
	// RemoveCounterChoices lists the permanents chosen to lose counters for an
	// AdditionalRemoveCounterAmong cost, one entry per counter removed. The same
	// permanent may appear multiple times when several of its counters are
	// removed.
	RemoveCounterChoices []id.ID
	// StrictReplay selects the strict invalid-preference policy. By default an
	// additional cost whose recorded preference is stale or now illegal falls
	// back to a deterministic legal selection so play continues; under strict
	// replay that fallback is disabled and an unsatisfiable preference rejects
	// the whole payment, so a recorded game replays exactly or not at all. The
	// policy applies uniformly to sacrifice, tap, return, discard, exile,
	// reveal, evidence, and counter-removal preferences.
	StrictReplay bool
}

// SpellOptionSummary is a summary of one payable spell cost option for choice presentation.
type SpellOptionSummary struct {
	Index           int
	Label           string
	ManaCost        *cost.Mana
	AdditionalCosts []cost.Additional
	CastPermission  SpellCastPermission
}

// SpellPaymentResult records the paid costs and casting permission selected by
// the payment plan.
type SpellPaymentResult struct {
	AdditionalCostsPaid []string
	PoolSpend           map[mana.Unit]int
	CastPermission      SpellCastPermission
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

func clonePreferences(prefs *Preferences) *Preferences {
	if prefs == nil {
		return nil
	}
	return &Preferences{
		AlternativeIndex:     prefs.AlternativeIndex,
		PhyrexianLifeChoices: append([]bool(nil), prefs.PhyrexianLifeChoices...),
		phyrexianIndex:       prefs.phyrexianIndex,
		SacrificeChoices:     append([]id.ID(nil), prefs.SacrificeChoices...),
		TapChoices:           append([]id.ID(nil), prefs.TapChoices...),
		ReturnChoices:        append([]id.ID(nil), prefs.ReturnChoices...),
		DiscardChoices:       append([]id.ID(nil), prefs.DiscardChoices...),
		ExileChoices:         append([]id.ID(nil), prefs.ExileChoices...),
		RevealChoices:        append([]id.ID(nil), prefs.RevealChoices...),
		EvidenceChoices:      append([]id.ID(nil), prefs.EvidenceChoices...),
		RemoveCounterChoices: append([]id.ID(nil), prefs.RemoveCounterChoices...),
		StrictReplay:         prefs.StrictReplay,
	}
}
