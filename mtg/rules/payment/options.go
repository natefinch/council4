package payment

import (
	"slices"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
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
	bargained       bool
}

// spellCostOptionsForZoneAndKicker returns the available cost options for
// casting a spell from the given zone with the kicker flag. When bargained is
// set the spell's fixed Bargain additional cost (CR 702.166a) is added to every
// option so the cast is legal only when the caster can sacrifice an artifact,
// enchantment, or token.
func spellCostOptionsForZoneAndKicker(s State, playerID game.PlayerID, card *game.CardDef, sourceZone zone.Type, kickerPaid bool, kickerCount int, bargained bool, permissions []SpellCastPermission) []spellCostOption {
	if card == nil {
		return nil
	}
	kicker, kickerOK := spellKicker(card)
	requiredAdditional := card.AdditionalCosts
	if bargained {
		requiredAdditional = append(append([]cost.Additional(nil), requiredAdditional...), BargainSacrificeCost())
	}
	alternatives := card.AlternativeCosts
	hasFlashbackAlternative := slices.ContainsFunc(alternatives, isFlashbackAlternative)
	if flashbackCost, ok := card.FlashbackCost(); ok && !hasFlashbackAlternative {
		alternatives = append(slices.Clone(alternatives), cost.Alternative{
			Label:    flashbackAlternativeLabel,
			ManaCost: opt.Val(slices.Clone(flashbackCost)),
			Mechanic: cost.AlternativeMechanicFlashback,
		})
		hasFlashbackAlternative = true
	}
	if card.HasKeyword(game.JumpStart) && !hasFlashbackAlternative {
		alternatives = append(slices.Clone(alternatives), jumpStartAlternativeCost(card))
		hasFlashbackAlternative = true
	}
	if sourceZone == zone.Graveyard {
		if granted := s.GraveyardCastGrantedAlternatives(playerID, card); len(granted) > 0 {
			alternatives = append(slices.Clone(alternatives), granted...)
		}
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
			bargained:       bargained,
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
		if !alternativeCostConditionSatisfied(s, playerID, alternative) {
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
			bargained:       bargained,
		})
	}
	return options
}

func spellCostOptionsForRequest(s State, req SpellRequest) []spellCostOption {
	options := spellCostOptionsForRequestWithoutModes(s, req)
	options = applyAdditionalCostChoices(options, req.Card)
	addSpreeModeCosts(options, req.Card, req.ChosenModes)
	addEscalateModeCosts(options, req.Card, req.ChosenModes)
	addSpliceCosts(options, req.SpliceManaCosts)
	if req.Offspring {
		addOffspringCost(options, req.Card)
	}
	return options
}

// applyAdditionalCostChoices expands each cost option across the cartesian
// product of the card's printed additional-cost choices ("As an additional cost
// to cast this spell, pay 5 life or pay {2}." — Redirect Lightning). Each chosen
// branch contributes its additive mana to the option's mana cost (it never
// replaces the printed cost, so a later tax or reduction on the printed cost
// still applies to every branch) and its non-mana costs to the option's
// additional costs. The expanded options are re-indexed sequentially so the
// caster's branch selection is a stable option index preserved through payment.
// A card with no additional-cost choices is returned unchanged, so every other
// spell keeps its existing option indices.
func applyAdditionalCostChoices(options []spellCostOption, card *game.CardDef) []spellCostOption {
	if card == nil || len(card.AdditionalCostChoices) == 0 {
		return options
	}
	expanded := options
	for _, choice := range card.AdditionalCostChoices {
		if len(choice.Options) == 0 {
			continue
		}
		next := make([]spellCostOption, 0, len(expanded)*len(choice.Options))
		for _, option := range expanded {
			for _, branch := range choice.Options {
				next = append(next, applyChoiceBranch(option, branch))
			}
		}
		expanded = next
	}
	for i := range expanded {
		expanded[i].index = i
	}
	return expanded
}

// applyChoiceBranch folds one additional-cost choice branch onto a cost option:
// the branch's mana is appended to the option's mana cost and its non-mana costs
// are appended to the option's additional costs. The branch's mana is additional
// to, never a replacement of, the option's existing mana cost.
func applyChoiceBranch(option spellCostOption, branch cost.AdditionalChoiceOption) spellCostOption {
	if len(branch.Mana) > 0 {
		combined := cost.Mana{}
		if option.manaCost != nil {
			combined = append(combined, (*option.manaCost)...)
		}
		combined = append(combined, branch.Mana...)
		option.manaCost = &combined
	}
	if len(branch.Costs) > 0 {
		option.additionalCosts = append(append([]cost.Additional(nil), option.additionalCosts...), branch.Costs...)
	}
	if branch.Label != "" {
		if option.label == "" || option.label == "Normal cost" {
			option.label = branch.Label
		} else {
			option.label += " + " + branch.Label
		}
	}
	return option
}

// addOffspringCost adds the spell's Offspring additional mana cost (CR 702.171a)
// to every payable cost option when the offspring cast branch is chosen, so the
// caster pays the base cost plus the offspring cost. A card without the Offspring
// keyword adds nothing.
func addOffspringCost(options []spellCostOption, card *game.CardDef) {
	offspring, ok := spellOffspring(card)
	if !ok || len(offspring.Cost) == 0 {
		return
	}
	for i := range options {
		combined := cost.Mana{}
		if options[i].manaCost != nil {
			combined = append(combined, (*options[i].manaCost)...)
		}
		combined = append(combined, offspring.Cost...)
		options[i].manaCost = &combined
	}
}

// addSpliceCosts adds the mana splice cost of each card spliced onto an Arcane
// spell (CR 702.47) to every payable cost option, so the controller pays the
// host spell's cost plus each splice cost as an additional cost. A spell with no
// splices adds nothing.
func addSpliceCosts(options []spellCostOption, spliceManaCosts []cost.Mana) {
	var extra cost.Mana
	for _, spliceCost := range spliceManaCosts {
		extra = append(extra, spliceCost...)
	}
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

func spellCostOptionsForRequestWithoutModes(s State, req SpellRequest) []spellCostOption {
	if !req.Alternative.Exists {
		return spellCostOptionsForZoneAndKicker(s, req.PlayerID, req.Card, req.SourceZone, req.KickerPaid, req.KickerCount, req.Bargained, req.CastPermissions)
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
	if !alternativeCostConditionSatisfied(s, req.PlayerID, alternative) {
		return nil
	}
	kicker, kickerOK := spellKicker(req.Card)
	additional := append([]cost.Additional(nil), req.Card.AdditionalCosts...)
	additional = append(additional, alternative.AdditionalCosts...)
	if req.Bargained {
		additional = append(additional, BargainSacrificeCost())
	}
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
		bargained:       req.Bargained,
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

func alternativeCostConditionSatisfied(s State, playerID game.PlayerID, alternative cost.Alternative) bool {
	switch alternative.Condition {
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
	case cost.AlternativeConditionControlsPermanentSubtype:
		for _, permanent := range s.Battlefield() {
			if permanent != nil && !permanent.PhasedOut &&
				s.EffectiveController(permanent) == playerID &&
				s.PermanentHasSubtype(permanent, alternative.ConditionSubtype) {
				return true
			}
		}
		return false
	case cost.AlternativeConditionNotYourTurn:
		return s.ActivePlayer() != playerID
	case cost.AlternativeConditionYourTurn:
		return s.ActivePlayer() == playerID
	case cost.AlternativeConditionOpponentLostLifeThisTurn:
		return s.OpponentLostLifeThisTurn(playerID)
	case cost.AlternativeConditionOpponentGainedLifeThisTurn:
		return s.OpponentGainedLifeThisTurn(playerID)
	case cost.AlternativeConditionCreaturesAttacking:
		count := s.AttackingCreatureCount()
		if alternative.ConditionExactly {
			return count == alternative.ConditionCount
		}
		return count >= alternative.ConditionCount
	case cost.AlternativeConditionPermanentsOnBattlefield:
		selection := game.Selection{RequiredTypesAny: []types.Card{alternative.ConditionPermanentType}}
		count := 0
		for _, permanent := range s.Battlefield() {
			if permanent != nil && !permanent.PhasedOut &&
				s.PermanentMatchesSelection(permanent, selection) {
				count++
			}
		}
		if alternative.ConditionExactly {
			return count == alternative.ConditionCount
		}
		return count >= alternative.ConditionCount
	case cost.AlternativeConditionOpponentCastSpellsThisTurn:
		return s.OpponentCastSpellsThisTurn(playerID, alternative.ConditionCount)
	default:
		return false
	}
}

func isFlashbackAlternative(alternative cost.Alternative) bool {
	return alternative.Mechanic == cost.AlternativeMechanicFlashback
}

// jumpStartAlternativeCost builds the Flashback-style alternative cost a
// Jump-start card (CR 702.134) offers from the graveyard: pay the card's printed
// mana cost and discard a card. It reuses the Flashback label so the graveyard
// cast carries the Flashback permission and is exiled on resolution, and adds
// the discard-a-card additional cost that distinguishes Jump-start.
func jumpStartAlternativeCost(card *game.CardDef) cost.Alternative {
	alternative := cost.Alternative{
		Label:    flashbackAlternativeLabel,
		Mechanic: cost.AlternativeMechanicFlashback,
		AdditionalCosts: []cost.Additional{{
			Kind:   cost.AdditionalDiscard,
			Amount: 1,
			Text:   "Discard a card",
		}},
	}
	if card.ManaCost.Exists {
		alternative.ManaCost = opt.Val(slices.Clone(card.ManaCost.Val))
	}
	return alternative
}

func isEscapeAlternative(alternative cost.Alternative) bool {
	return alternative.Mechanic == cost.AlternativeMechanicEscape
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

func spellOffspring(card *game.CardDef) (game.OffspringKeyword, bool) {
	if card == nil {
		return game.OffspringKeyword{}, false
	}
	return card.OffspringKeyword()
}

// BargainSacrificeCost is the fixed optional additional cost the Bargain keyword
// grants (CR 702.166a): "sacrifice an artifact, enchantment, or token." It is
// added to a spell's costs only on the bargained cast branch, where paying it is
// mandatory; the union of eligible sacrifices is expanded by the payment
// planner's SelectionForAdditionalCost.
func BargainSacrificeCost() cost.Additional {
	return cost.Additional{
		Kind:                            cost.AdditionalSacrifice,
		Amount:                          1,
		MatchArtifactEnchantmentOrToken: true,
		Text:                            "Sacrifice an artifact, enchantment, or token",
	}
}

// payableSpellOptionsFromState returns all spell cost options that can currently be paid.
func payableSpellOptionsFromState(s State, req SpellRequest) []SpellOptionSummary {
	var result []SpellOptionSummary
	for _, option := range spellCostOptionsForRequest(s, req) {
		if _, ok := buildSpellCostPlanForOption(s, req.PlayerID, req.CardID, req.SourceZone, option, req.XValue, req.Targets, req.Bestowed, nil); ok {
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
