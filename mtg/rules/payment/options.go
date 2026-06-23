package payment

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/opt"
)

// flashbackAlternativeLabel is the canonical label for flashback alternative costs.
const flashbackAlternativeLabel = "Flashback"

// escapeAlternativeLabel is the canonical label for escape alternative costs.
const escapeAlternativeLabel = "Escape"

// spellCostOption describes one payable cost option for a spell.
type spellCostOption struct {
	index           int
	label           string
	card            *game.CardDef
	manaCost        *cost.Mana
	additionalCosts []cost.Additional
	castPermission  SpellCastPermission
}

// spellCostOptionsForZoneAndKicker returns the available cost options for
// casting a spell from the given zone with the kicker flag.
func spellCostOptionsForZoneAndKicker(s State, playerID game.PlayerID, card *game.CardDef, sourceZone zone.Type, kickerPaid bool, kickerCount int, permissions []SpellCastPermission) []spellCostOption {
	if card == nil {
		return nil
	}
	kicker, kickerOK := spellKicker(card)
	requiredAdditional := card.AdditionalCosts
	alternatives := card.AlternativeCosts
	hasFlashbackAlternative := slices.ContainsFunc(alternatives, isFlashbackAlternative)
	if flashbackCost, ok := card.FlashbackCost(); ok && !hasFlashbackAlternative {
		alternatives = append(slices.Clone(alternatives), cost.Alternative{
			Label:    flashbackAlternativeLabel,
			ManaCost: opt.Val(slices.Clone(flashbackCost)),
		})
		hasFlashbackAlternative = true
	}
	if len(permissions) == 0 {
		permissions = []SpellCastPermission{SpellCastPermissionDefault}
		if sourceZone == zone.Graveyard && hasFlashbackAlternative {
			permissions[0] = SpellCastPermissionFlashback
		}
	}
	normalPermission, canCastNormally := firstNormalPermission(permissions)
	canCastWithFlashback := sourceZone == zone.Graveyard &&
		hasFlashbackAlternative &&
		slices.Contains(permissions, SpellCastPermissionFlashback)
	canCastWithEscape := sourceZone == zone.Graveyard &&
		slices.Contains(permissions, SpellCastPermissionEscape)
	var options []spellCostOption
	if canCastNormally {
		options = append(options, spellCostOption{
			index:           0,
			label:           "Normal cost",
			card:            card,
			manaCost:        spellManaCostWithKicker(manaCostPtr(card.ManaCost), kicker, kickerOK, kickerPaid, kickerCount),
			additionalCosts: append([]cost.Additional(nil), requiredAdditional...),
			castPermission:  normalPermission,
		})
	}
	for i, alternative := range alternatives {
		permission := normalPermission
		switch {
		case isFlashbackAlternative(alternative):
			if !canCastWithFlashback {
				continue
			}
			permission = SpellCastPermissionFlashback
		case isEscapeAlternative(alternative):
			if !canCastWithEscape {
				continue
			}
			permission = SpellCastPermissionEscape
		default:
			if !canCastNormally {
				continue
			}
		}
		if !alternativeCostConditionSatisfied(s, playerID, alternative.Condition) {
			continue
		}
		additional := append([]cost.Additional(nil), requiredAdditional...)
		additional = append(additional, alternative.AdditionalCosts...)
		label := alternative.Label
		if label == "" {
			label = "Alternative cost"
		}
		options = append(options, spellCostOption{
			index:           i + 1,
			label:           label,
			card:            card,
			manaCost:        spellManaCostWithKicker(manaCostPtr(alternative.ManaCost), kicker, kickerOK, kickerPaid, kickerCount),
			additionalCosts: additional,
			castPermission:  permission,
		})
	}
	return options
}

func spellCostOptionsForRequest(s State, req SpellRequest) []spellCostOption {
	options := spellCostOptionsForRequestWithoutModes(s, req)
	addSpreeModeCosts(options, req.Card, req.ChosenModes)
	addEscalateModeCosts(options, req.Card, req.ChosenModes)
	return options
}

func spellCostOptionsForRequestWithoutModes(s State, req SpellRequest) []spellCostOption {
	if !req.Alternative.Exists {
		return spellCostOptionsForZoneAndKicker(s, req.PlayerID, req.Card, req.SourceZone, req.KickerPaid, req.KickerCount, req.CastPermissions)
	}
	if req.Card == nil {
		return nil
	}
	permissions := req.CastPermissions
	if len(permissions) == 0 {
		permissions = []SpellCastPermission{SpellCastPermissionDefault}
		if req.SourceZone == zone.Graveyard &&
			slices.ContainsFunc(req.Card.AlternativeCosts, isFlashbackAlternative) {
			permissions[0] = SpellCastPermissionFlashback
		}
	}
	castPermission, ok := firstNormalPermission(permissions)
	if !ok {
		return nil
	}
	alternative := req.Alternative.Val
	if !alternativeCostConditionSatisfied(s, req.PlayerID, alternative.Condition) {
		return nil
	}
	kicker, kickerOK := spellKicker(req.Card)
	additional := append([]cost.Additional(nil), req.Card.AdditionalCosts...)
	additional = append(additional, alternative.AdditionalCosts...)
	label := alternative.Label
	if label == "" {
		label = "Alternative cost"
	}
	return []spellCostOption{{
		index:           0,
		label:           label,
		card:            req.Card,
		manaCost:        spellManaCostWithKicker(manaCostPtr(alternative.ManaCost), kicker, kickerOK, req.KickerPaid, req.KickerCount),
		additionalCosts: additional,
		castPermission:  castPermission,
	}}
}

// addSpreeModeCosts adds the additional mana cost of each chosen Spree mode
// (CR 702.171) to every payable cost option. Spree modes carry their own mana
// cost; the controller pays the base cost plus each chosen mode's cost.
func addSpreeModeCosts(options []spellCostOption, card *game.CardDef, chosenModes []int) {
	extra := spreeModeManaCost(card, chosenModes)
	if len(extra) == 0 {
		return
	}
	for i := range options {
		combined := cost.Mana{}
		if options[i].manaCost != nil {
			combined = append(combined, (*options[i].manaCost)...)
		}
		combined = append(combined, extra...)
		options[i].manaCost = &combined
	}
}

// spreeModeManaCost sums the additional mana costs of the chosen modes of a
// Spree spell. Modes without a cost contribute nothing.
func spreeModeManaCost(card *game.CardDef, chosenModes []int) cost.Mana {
	if card == nil || !card.SpellAbility.Exists {
		return nil
	}
	modes := card.SpellAbility.Val.Modes
	var total cost.Mana
	for _, index := range chosenModes {
		if index < 0 || index >= len(modes) {
			continue
		}
		if mode := modes[index]; mode.Cost.Exists {
			total = append(total, mode.Cost.Val...)
		}
	}
	return total
}

// addEscalateModeCosts adds the escalate cost of an Escalate spell (CR 702.121)
// to every payable cost option. The controller pays the spell's base cost plus
// the escalate cost once for each chosen mode beyond the first.
func addEscalateModeCosts(options []spellCostOption, card *game.CardDef, chosenModes []int) {
	extra := escalateModeManaCost(card, chosenModes)
	if len(extra) == 0 {
		return
	}
	for i := range options {
		combined := cost.Mana{}
		if options[i].manaCost != nil {
			combined = append(combined, (*options[i].manaCost)...)
		}
		combined = append(combined, extra...)
		options[i].manaCost = &combined
	}
}

// escalateModeManaCost returns the escalate cost repeated once for each chosen
// mode beyond the first. A spell with no escalate cost, or with one or fewer
// chosen modes, adds nothing.
func escalateModeManaCost(card *game.CardDef, chosenModes []int) cost.Mana {
	if card == nil || !card.SpellAbility.Exists {
		return nil
	}
	escalate := card.SpellAbility.Val.EscalateCost
	if !escalate.Exists || len(chosenModes) <= 1 {
		return nil
	}
	var total cost.Mana
	for range chosenModes[1:] {
		total = append(total, escalate.Val...)
	}
	return total
}

// firstNormalPermission returns the first permission that authorizes paying a
// spell's ordinary (non-graveyard-alternative) cost. Flashback and Escape
// permissions authorize only their graveyard alternative cost, so they are
// skipped here.
func firstNormalPermission(permissions []SpellCastPermission) (SpellCastPermission, bool) {
	for _, permission := range permissions {
		if permission != SpellCastPermissionFlashback && permission != SpellCastPermissionEscape {
			return permission, true
		}
	}
	return SpellCastPermissionDefault, false
}

func alternativeCostConditionSatisfied(s State, playerID game.PlayerID, condition cost.AlternativeCondition) bool {
	switch condition {
	case cost.AlternativeConditionNone:
		return true
	case cost.AlternativeConditionControlsCommander:
		for _, permanent := range s.Battlefield() {
			if permanent != nil && !permanent.PhasedOut &&
				s.EffectiveController(permanent) == playerID &&
				s.IsCommanderPermanent(permanent) {
				return true
			}
		}
		return false
	case cost.AlternativeConditionNotYourTurn:
		return s.ActivePlayer() != playerID
	case cost.AlternativeConditionOpponentLostLifeThisTurn:
		return s.OpponentLostLifeThisTurn(playerID)
	default:
		return false
	}
}

func isFlashbackAlternative(alternative cost.Alternative) bool {
	return strings.EqualFold(strings.TrimSpace(alternative.Label), flashbackAlternativeLabel)
}

func isEscapeAlternative(alternative cost.Alternative) bool {
	return strings.EqualFold(strings.TrimSpace(alternative.Label), escapeAlternativeLabel)
}

func spellManaCostWithKicker(base *cost.Mana, kicker game.KickerKeyword, kickerOK, kickerPaid bool, kickerCount int) *cost.Mana {
	if !kickerPaid || !kickerOK {
		return base
	}
	times := max(kickerCount, 1)
	combined := cost.Mana{}
	if base != nil {
		combined = append(combined, (*base)...)
	}
	for range times {
		combined = append(combined, kicker.Cost...)
	}
	return &combined
}

func spellKicker(card *game.CardDef) (game.KickerKeyword, bool) {
	if card == nil {
		return game.KickerKeyword{}, false
	}
	return card.KickerKeyword()
}

// payableSpellOptionsFromState returns all spell cost options that can currently be paid.
func payableSpellOptionsFromState(s State, req SpellRequest) []SpellOptionSummary {
	var result []SpellOptionSummary
	for _, option := range spellCostOptionsForRequest(s, req) {
		if _, ok := buildSpellCostPlanForOption(s, req.PlayerID, req.CardID, req.SourceZone, option, req.XValue, req.Targets, nil); ok {
			result = append(result, SpellOptionSummary{
				Index:           option.index,
				Label:           option.label,
				ManaCost:        option.manaCost,
				AdditionalCosts: option.additionalCosts,
				CastPermission:  option.castPermission,
			})
		}
	}
	return result
}
