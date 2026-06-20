package payment

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
)

// flashbackAlternativeLabel is the canonical label for flashback alternative costs.
const flashbackAlternativeLabel = "Flashback"

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
func spellCostOptionsForZoneAndKicker(s State, playerID game.PlayerID, card *game.CardDef, sourceZone zone.Type, kickerPaid bool, permissions []SpellCastPermission) []spellCostOption {
	if card == nil {
		return nil
	}
	kicker, kickerOK := spellKicker(card)
	requiredAdditional := card.AdditionalCosts
	hasFlashbackAlternative := slices.ContainsFunc(card.AlternativeCosts, isFlashbackAlternative)
	if len(permissions) == 0 {
		permissions = []SpellCastPermission{SpellCastPermissionDefault}
		if sourceZone == zone.Graveyard && hasFlashbackAlternative {
			permissions[0] = SpellCastPermissionFlashback
		}
	}
	nonFlashbackPermission, canCastWithoutFlashback := firstNonFlashbackPermission(permissions)
	canCastWithFlashback := sourceZone == zone.Graveyard &&
		hasFlashbackAlternative &&
		slices.Contains(permissions, SpellCastPermissionFlashback)
	var options []spellCostOption
	if canCastWithoutFlashback {
		options = append(options, spellCostOption{
			index:           0,
			label:           "Normal cost",
			card:            card,
			manaCost:        spellManaCostWithKicker(manaCostPtr(card.ManaCost), kicker, kickerOK, kickerPaid),
			additionalCosts: append([]cost.Additional(nil), requiredAdditional...),
			castPermission:  nonFlashbackPermission,
		})
	}
	for i, alternative := range card.AlternativeCosts {
		flashback := isFlashbackAlternative(alternative)
		if flashback && !canCastWithFlashback {
			continue
		}
		if !flashback && !canCastWithoutFlashback {
			continue
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
			manaCost:        spellManaCostWithKicker(manaCostPtr(alternative.ManaCost), kicker, kickerOK, kickerPaid),
			additionalCosts: additional,
			castPermission:  nonFlashbackPermission,
		})
		if flashback {
			options[len(options)-1].castPermission = SpellCastPermissionFlashback
		}
	}
	return options
}

func spellCostOptionsForRequest(s State, req SpellRequest) []spellCostOption {
	if !req.Alternative.Exists {
		return spellCostOptionsForZoneAndKicker(s, req.PlayerID, req.Card, req.SourceZone, req.KickerPaid, req.CastPermissions)
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
	castPermission, ok := firstNonFlashbackPermission(permissions)
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
		manaCost:        spellManaCostWithKicker(manaCostPtr(alternative.ManaCost), kicker, kickerOK, req.KickerPaid),
		additionalCosts: additional,
		castPermission:  castPermission,
	}}
}

func firstNonFlashbackPermission(permissions []SpellCastPermission) (SpellCastPermission, bool) {
	for _, permission := range permissions {
		if permission != SpellCastPermissionFlashback {
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
	default:
		return false
	}
}

func isFlashbackAlternative(alternative cost.Alternative) bool {
	return strings.EqualFold(strings.TrimSpace(alternative.Label), flashbackAlternativeLabel)
}

func spellManaCostWithKicker(base *cost.Mana, kicker game.KickerKeyword, kickerOK, kickerPaid bool) *cost.Mana {
	if !kickerPaid || !kickerOK {
		return base
	}
	combined := cost.Mana{}
	if base != nil {
		combined = append(combined, (*base)...)
	}
	combined = append(combined, kicker.Cost...)
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
		if _, ok := buildSpellCostPlanForOption(s, req.PlayerID, req.CardID, req.SourceZone, option, req.XValue, nil); ok {
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
