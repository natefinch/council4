package payment

import (
	"strings"

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
	additionalCosts []game.AdditionalCost
}

// spellCostOptionsForZoneAndKicker returns the available cost options for
// casting a spell from the given zone with the kicker flag.
func spellCostOptionsForZoneAndKicker(card *game.CardDef, sourceZone game.ZoneType, kickerPaid bool) []spellCostOption {
	if card == nil {
		return nil
	}
	ability, ok := firstSpellAbility(card)
	if !ok {
		return []spellCostOption{{index: 0, label: "Normal cost", card: card, manaCost: manaCostPtr(card.ManaCost)}}
	}
	requiredAdditional := abilityAdditionalCosts(ability)
	options := []spellCostOption{
		{
			index:           0,
			label:           "Normal cost",
			card:            card,
			manaCost:        spellManaCostWithKicker(manaCostPtr(card.ManaCost), ability, kickerPaid),
			additionalCosts: append([]game.AdditionalCost(nil), requiredAdditional...),
		},
	}
	for i, alternative := range ability.AlternativeCosts {
		if isFlashbackAlternative(alternative) && sourceZone != game.ZoneGraveyard {
			continue
		}
		if sourceZone == game.ZoneGraveyard && !isFlashbackAlternative(alternative) {
			continue
		}
		additional := append([]game.AdditionalCost(nil), requiredAdditional...)
		additional = append(additional, alternative.AdditionalCosts...)
		label := alternative.Label
		if label == "" {
			label = "Alternative cost"
		}
		options = append(options, spellCostOption{
			index:           i + 1,
			label:           label,
			card:            card,
			manaCost:        spellManaCostWithKicker(manaCostPtr(alternative.ManaCost), ability, kickerPaid),
			additionalCosts: additional,
		})
	}
	if sourceZone == game.ZoneGraveyard {
		return options[1:]
	}
	return options
}

func isFlashbackAlternative(alternative game.AlternativeCost) bool {
	return strings.EqualFold(strings.TrimSpace(alternative.Label), flashbackAlternativeLabel)
}

func spellManaCostWithKicker(base *cost.Mana, ability *game.AbilityDef, kickerPaid bool) *cost.Mana {
	if !kickerPaid || ability == nil {
		return base
	}
	kickerCost, ok := ability.KickerCost()
	if !ok {
		return base
	}
	combined := cost.Mana{}
	if base != nil {
		combined = append(combined, (*base)...)
	}
	combined = append(combined, kickerCost...)
	return &combined
}

// firstSpellAbility returns the first spell ability from a card, if any.
func firstSpellAbility(card *game.CardDef) (*game.AbilityDef, bool) {
	abilities := card.AbilityDefs()
	for i := range abilities {
		if abilities[i].IsSpell() {
			return &abilities[i], true
		}
	}
	return nil, false
}

// payableSpellOptionsFromState returns all spell cost options that can currently be paid.
func payableSpellOptionsFromState(s State, req SpellRequest) []SpellOptionSummary {
	var result []SpellOptionSummary
	for _, option := range spellCostOptionsForZoneAndKicker(req.Card, req.SourceZone, req.KickerPaid) {
		if _, ok := buildSpellCostPlanForOption(s, req.PlayerID, req.CardID, req.SourceZone, option, req.XValue, nil); ok {
			result = append(result, SpellOptionSummary{
				Index:           option.index,
				Label:           option.label,
				ManaCost:        option.manaCost,
				AdditionalCosts: option.additionalCosts,
			})
		}
	}
	return result
}
