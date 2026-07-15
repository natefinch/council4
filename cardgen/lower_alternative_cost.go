package cardgen

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func lowerSpellAlternativeCost(cardName string, ability compiler.CompiledAbility) (abilityLowering, *shared.Diagnostic) {
	if ability.AlternativeCost != nil && ability.AlternativeCost.Kind == compiler.AlternativeCostFlashback {
		return lowerFlashbackAlternativeCost(cardName, ability)
	}
	if ability.AlternativeCost != nil && ability.AlternativeCost.Kind == compiler.AlternativeCostEscape {
		return lowerEscapeAlternativeCost(cardName, ability)
	}
	if ability.AlternativeCost != nil &&
		ability.AlternativeCost.Kind == compiler.AlternativeCostOverload &&
		ability.AlternativeCost.ReplaceTargetWithEach &&
		len(ability.AlternativeCost.ManaCost) > 0 &&
		overloadManaCostSupported(ability.AlternativeCost.ManaCost) &&
		ability.Cost == nil &&
		len(ability.Content.Effects) == 0 &&
		len(ability.Content.Targets) == 0 &&
		len(ability.Content.Conditions) == 0 &&
		len(ability.Content.Keywords) == 0 &&
		len(ability.Content.Modes) == 0 {
		return abilityLowering{
			overloadCost: opt.Val(slices.Clone(ability.AlternativeCost.ManaCost)),
			consumed: semanticConsumption{
				alternativeCost: true,
				references:      len(ability.Content.References),
			},
			sourceSpans: []shared.Span{ability.Span},
		}, nil
	}
	if ability.AlternativeCost != nil && ability.AlternativeCost.Kind == compiler.AlternativeCostPitch {
		return lowerPitchAlternativeCost(cardName, ability)
	}
	if ability.AlternativeCost != nil && ability.AlternativeCost.Kind == compiler.AlternativeCostDiscard {
		return lowerDiscardAlternativeCost(cardName, ability)
	}
	if ability.AlternativeCost != nil && ability.AlternativeCost.Kind == compiler.AlternativeCostBorderpost {
		return lowerBorderpostAlternativeCost(cardName, ability)
	}
	if ability.AlternativeCost != nil && ability.AlternativeCost.Kind == compiler.AlternativeCostFree {
		return lowerFreeAlternativeCost(cardName, ability)
	}
	if ability.AlternativeCost != nil && ability.AlternativeCost.Kind == compiler.AlternativeCostMana {
		return lowerManaAlternativeCost(cardName, ability)
	}
	if ability.AlternativeCost == nil ||
		(ability.AlternativeCost.Kind != compiler.AlternativeCostUnknown &&
			ability.AlternativeCost.Kind != compiler.AlternativeCostCommander) ||
		ability.AlternativeCost.Condition != compiler.AlternativeCostConditionControlsCommander ||
		!ability.AlternativeCost.WithoutPayingManaCost ||
		ability.Cost != nil ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported alternative spell cost",
			"the executable source backend could not recognize the spell's alternative cost",
		)
	}

	return abilityLowering{
		alternativeCosts: []cost.Alternative{{
			Label:     "Cast without paying mana cost",
			Condition: cost.AlternativeConditionControlsCommander,
		}},
		consumed: semanticConsumption{
			alternativeCost: true,
			references:      len(ability.Content.References),
		},
		sourceSpans: []shared.Span{ability.Span},
	}, nil
}

func lowerBorderpostAlternativeCost(cardName string, ability compiler.CompiledAbility) (abilityLowering, *shared.Diagnostic) {
	if ability.AlternativeCost == nil ||
		len(ability.AlternativeCost.ManaCost) == 0 ||
		ability.Cost == nil ||
		len(ability.Cost.Components) != 1 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported alternative spell cost",
			"the executable source backend could not recognize the Borderpost alternative cost",
		)
	}
	additional, ok := lowerActivatedAdditionalCost(cardName, ability.Cost.Components[0])
	if !ok || additional.Kind != cost.AdditionalReturnToHand {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported alternative spell cost",
			"the executable source backend could not lower the Borderpost return cost",
		)
	}
	return abilityLowering{
		alternativeCosts: []cost.Alternative{{
			Label:           "Pay {1} and return a basic land",
			ManaCost:        opt.Val(slices.Clone(ability.AlternativeCost.ManaCost)),
			AdditionalCosts: []cost.Additional{additional},
		}},
		consumed: semanticConsumption{
			alternativeCost: true,
			cost:            true,
			references:      len(ability.Content.References),
		},
		sourceSpans: []shared.Span{ability.Span},
	}, nil
}

func overloadManaCostSupported(manaCost cost.Mana) bool {
	for _, symbol := range manaCost {
		if symbol.Kind == cost.VariableSymbol {
			return false
		}
	}
	return true
}

// lowerFlashbackAlternativeCost lowers the em-dash Flashback form
// "Flashback—<cost>" into a SimpleKeyword(Flashback) grant plus a Flashback
// alternative cost carrying the non-mana (or compound) cost typed by the shared
// cost machinery. The runtime gates graveyard flashback casting on the keyword
// grant and pays the alternative's mana and additional costs, then exiles the
// spell. It fails closed when the cost is unrecognized.
func lowerFlashbackAlternativeCost(cardName string, ability compiler.CompiledAbility) (abilityLowering, *shared.Diagnostic) {
	if ability.Cost == nil || len(ability.Cost.Components) == 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported alternative spell cost",
			"the executable source backend could not recognize the flashback cost",
		)
	}
	manaCost, additionalCosts, ok := lowerActivationCostComponents(cardName, ability.Cost)
	if !ok {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported alternative spell cost",
			"the executable source backend does not yet lower this flashback cost",
		)
	}
	alternative := cost.Alternative{
		Label:           "Flashback",
		Mechanic:        cost.AlternativeMechanicFlashback,
		AdditionalCosts: additionalCosts,
	}
	if len(manaCost) > 0 {
		alternative.ManaCost = opt.Val(manaCost)
	}
	return abilityLowering{
		staticAbilities: []loweredStaticAbility{{
			Body: game.StaticAbility{
				KeywordAbilities: []game.KeywordAbility{game.SimpleKeyword{Kind: game.Flashback}},
			},
		}},
		alternativeCosts: []cost.Alternative{alternative},
		consumed: semanticConsumption{
			cost:            true,
			alternativeCost: true,
			references:      len(ability.Content.References),
		},
		sourceSpans: []shared.Span{ability.Span},
	}, nil
}

// lowerEscapeAlternativeCost lowers the em-dash Escape form
// "Escape—<cost>, Exile N cards from your graveyard." into a
// SimpleKeyword(Escape) grant plus an Escape alternative cost carrying the
// compound escape cost typed by the shared cost machinery (its mana cost plus
// the graveyard-exile additional cost). The runtime gates graveyard escape
// casting on the keyword grant and pays the alternative's mana and additional
// costs. Unlike Flashback the spell is not exiled, so it can be escaped again.
// It fails closed when the cost is unrecognized.
func lowerEscapeAlternativeCost(cardName string, ability compiler.CompiledAbility) (abilityLowering, *shared.Diagnostic) {
	if ability.Cost == nil || len(ability.Cost.Components) == 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported alternative spell cost",
			"the executable source backend could not recognize the escape cost",
		)
	}
	manaCost, additionalCosts, ok := lowerActivationCostComponents(cardName, ability.Cost)
	if !ok {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported alternative spell cost",
			"the executable source backend does not yet lower this escape cost",
		)
	}
	alternative := cost.Alternative{
		Label:           "Escape",
		Mechanic:        cost.AlternativeMechanicEscape,
		AdditionalCosts: additionalCosts,
	}
	if len(manaCost) > 0 {
		alternative.ManaCost = opt.Val(manaCost)
	}
	return abilityLowering{
		staticAbilities: []loweredStaticAbility{{
			Body: game.StaticAbility{
				KeywordAbilities: []game.KeywordAbility{game.SimpleKeyword{Kind: game.Escape}},
			},
		}},
		alternativeCosts: []cost.Alternative{alternative},
		consumed: semanticConsumption{
			cost:            true,
			alternativeCost: true,
			references:      len(ability.Content.References),
		},
		sourceSpans: []shared.Span{ability.Span},
	}, nil
}

// lowerPitchAlternativeCost lowers the Force of Will pitch family into a free
// (no-mana) alternative whose non-mana cost components — an optional pay-life
// and an exile-a-colored-card-from-hand — are lowered through the shared cost
// machinery used for activated, additional, and resolution costs. The cost
// rides on the ability's compiled Cost; the not-your-turn condition gates the
// option. It fails closed when the cost is unrecognized.
func lowerPitchAlternativeCost(cardName string, ability compiler.CompiledAbility) (abilityLowering, *shared.Diagnostic) {
	alternative := ability.AlternativeCost
	unsupported := alternative == nil ||
		alternative.WithoutPayingManaCost ||
		len(alternative.ManaCost) != 0 ||
		ability.Cost == nil ||
		len(ability.Cost.Components) == 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0
	condition, _, conditionOK := lowerAlternativeCostCondition(alternative)
	if unsupported || !conditionOK {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported alternative spell cost",
			"the executable source backend could not recognize the spell's alternative cost",
		)
	}
	manaCost, additionalCosts, ok := lowerActivationCostComponents(cardName, ability.Cost)
	if !ok || len(manaCost) != 0 || len(additionalCosts) == 0 {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported alternative spell cost",
			"the executable source backend does not yet lower this pitch cost",
		)
	}
	return abilityLowering{
		alternativeCosts: []cost.Alternative{{
			Label:           pitchAlternativeLabel(additionalCosts),
			AdditionalCosts: additionalCosts,
			Condition:       condition,
		}},
		consumed: semanticConsumption{
			cost:            true,
			alternativeCost: true,
			references:      len(ability.Content.References),
		},
		sourceSpans: []shared.Span{ability.Span},
	}, nil
}

// lowerDiscardAlternativeCost lowers the Foil/Outbreak family: a free (no-mana)
// alternative whose additional costs discard one or more cards from hand,
// optionally constrained by subtype, rather than paying the printed mana cost.
func lowerDiscardAlternativeCost(cardName string, ability compiler.CompiledAbility) (abilityLowering, *shared.Diagnostic) {
	alternative := ability.AlternativeCost
	unsupported := alternative == nil ||
		alternative.WithoutPayingManaCost ||
		len(alternative.ManaCost) != 0 ||
		ability.Cost == nil ||
		len(ability.Cost.Components) == 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0
	condition, _, conditionOK := lowerAlternativeCostCondition(alternative)
	if unsupported || !conditionOK {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported alternative spell cost",
			"the executable source backend could not recognize the spell's alternative cost",
		)
	}
	manaCost, additionalCosts, ok := lowerActivationCostComponents(cardName, ability.Cost)
	if !ok || len(manaCost) != 0 || len(additionalCosts) == 0 {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported alternative spell cost",
			"the executable source backend does not yet lower this discard cost",
		)
	}
	return abilityLowering{
		alternativeCosts: []cost.Alternative{{
			Label:           discardAlternativeLabel(additionalCosts),
			AdditionalCosts: additionalCosts,
			Condition:       condition,
		}},
		consumed: semanticConsumption{
			cost:            true,
			alternativeCost: true,
			references:      len(ability.Content.References),
		},
		sourceSpans: []shared.Span{ability.Span},
	}, nil
}

// lowerFreeAlternativeCost lowers the "free spell" family: a no-mana alternative
// whose single non-mana additional cost (pay life, sacrifice, tap, return, ...)
// is lowered through the shared cost machinery, optionally gated by a condition
// ("If you control a Swamp,", "If it's your turn,"). It fails closed when the
// cost carries mana, is empty, or is not recognized, so a mana-bearing or
// compound alternative is never mistaken for a free spell.
func lowerFreeAlternativeCost(cardName string, ability compiler.CompiledAbility) (abilityLowering, *shared.Diagnostic) {
	alternative := ability.AlternativeCost
	unsupported := alternative == nil ||
		alternative.WithoutPayingManaCost ||
		len(alternative.ManaCost) != 0 ||
		ability.Cost == nil ||
		len(ability.Cost.Components) == 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0
	condition, conditionSubtype, conditionOK := lowerAlternativeCostCondition(alternative)
	if unsupported || !conditionOK {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported alternative spell cost",
			"the executable source backend could not recognize the spell's alternative cost",
		)
	}
	manaCost, additionalCosts, ok := lowerActivationCostComponents(cardName, ability.Cost)
	if !ok || len(manaCost) != 0 || len(additionalCosts) != 1 {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported alternative spell cost",
			"the executable source backend does not yet lower this free alternative cost",
		)
	}
	return abilityLowering{
		alternativeCosts: []cost.Alternative{{
			Label:            freeAlternativeLabel(additionalCosts),
			AdditionalCosts:  additionalCosts,
			Condition:        condition,
			ConditionSubtype: conditionSubtype,
		}},
		consumed: semanticConsumption{
			cost:            true,
			alternativeCost: true,
			references:      len(ability.Content.References),
		},
		sourceSpans: []shared.Span{ability.Span},
	}, nil
}

// lowerManaAlternativeCost lowers the "conditional mana-only" family: "[If
// <condition>,] you may pay {MANA} rather than pay this spell's mana cost." into
// an optional cost.Alternative whose ManaCost replaces the spell's printed mana
// cost. The replacement carries a real mana cost even for {0} (an explicit
// cost.O(0) symbol), so it is a distinct payable option that additional costs
// and cost modifiers still apply to, never a cast-for-free absence.
//
// It fails closed on a payment carrying {X} (whose alternative-cost X semantics
// this backend does not model), on any residual ability content, and on any
// condition lowerManaAlternativeCostCondition does not recognize, so an
// unmodeled Trap condition is never approximated.
func lowerManaAlternativeCost(cardName string, ability compiler.CompiledAbility) (abilityLowering, *shared.Diagnostic) {
	alternative := ability.AlternativeCost
	unsupported := alternative == nil ||
		alternative.WithoutPayingManaCost ||
		len(alternative.ManaCost) == 0 ||
		!overloadManaCostSupported(alternative.ManaCost) ||
		ability.Cost != nil ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0
	lowered, conditionOK := lowerManaAlternativeCostCondition(alternative)
	if unsupported || !conditionOK {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported alternative spell cost",
			"the executable source backend could not recognize the mana-only alternative cost",
		)
	}
	return abilityLowering{
		alternativeCosts: []cost.Alternative{{
			Label:                  "Pay " + alternative.ManaCost.String(),
			ManaCost:               opt.Val(slices.Clone(alternative.ManaCost)),
			Condition:              lowered.Condition,
			ConditionCount:         lowered.Count,
			ConditionExactly:       lowered.Exactly,
			ConditionPermanentType: lowered.PermanentType,
		}},
		consumed: semanticConsumption{
			alternativeCost: true,
			references:      len(ability.Content.References),
		},
		sourceSpans: []shared.Span{ability.Span},
	}, nil
}

// loweredManaAlternativeCondition is a compiled mana-only alternative-cost
// condition lowered onto its runtime condition, count threshold, exact-comparison
// flag, and (for a board-state gate) the counted permanent type.
type loweredManaAlternativeCondition struct {
	Condition     cost.AlternativeCondition
	Count         int
	Exactly       bool
	PermanentType types.Card
}

// lowerManaAlternativeCostCondition maps a compiled mana-only alternative-cost
// condition onto its runtime condition, count threshold, exact-comparison flag,
// and counted permanent type. It fails closed on any condition the mana-only
// family does not model.
func lowerManaAlternativeCostCondition(alternative *compiler.CompiledAlternativeCost) (loweredManaAlternativeCondition, bool) {
	switch alternative.Condition {
	case compiler.AlternativeCostConditionUnknown:
		return loweredManaAlternativeCondition{Condition: cost.AlternativeConditionNone}, true
	case compiler.AlternativeCostConditionOpponentGainedLifeThisTurn:
		return loweredManaAlternativeCondition{Condition: cost.AlternativeConditionOpponentGainedLifeThisTurn}, true
	case compiler.AlternativeCostConditionCreaturesAttacking:
		if alternative.ConditionCount < 1 {
			return loweredManaAlternativeCondition{}, false
		}
		return loweredManaAlternativeCondition{
			Condition: cost.AlternativeConditionCreaturesAttacking,
			Count:     alternative.ConditionCount,
			Exactly:   alternative.ConditionExactly,
		}, true
	case compiler.AlternativeCostConditionPermanentsOnBattlefield:
		if alternative.ConditionCount < 1 || alternative.ConditionPermanentType == "" {
			return loweredManaAlternativeCondition{}, false
		}
		return loweredManaAlternativeCondition{
			Condition:     cost.AlternativeConditionPermanentsOnBattlefield,
			Count:         alternative.ConditionCount,
			PermanentType: alternative.ConditionPermanentType,
		}, true
	case compiler.AlternativeCostConditionOpponentCastSpellsThisTurn:
		if alternative.ConditionCount < 1 {
			return loweredManaAlternativeCondition{}, false
		}
		return loweredManaAlternativeCondition{
			Condition: cost.AlternativeConditionOpponentCastSpellsThisTurn,
			Count:     alternative.ConditionCount,
		}, true
	default:
		return loweredManaAlternativeCondition{}, false
	}
}

// capitalizing the payment's own printed cost text (e.g. "pay 4 life" becomes
// "Pay 4 life", "sacrifice a creature" becomes "Sacrifice a creature").
func freeAlternativeLabel(additionalCosts []cost.Additional) string {
	if len(additionalCosts) != 1 || len(additionalCosts[0].Text) == 0 {
		return "Alternative cost"
	}
	text := additionalCosts[0].Text
	return strings.ToUpper(text[:1]) + text[1:]
}

// discardAlternativeLabel builds the display label for a discard alternative
// from its lowered discard costs, naming each discarded card's subtype filter
// when present.
func discardAlternativeLabel(additionalCosts []cost.Additional) string {
	parts := make([]string, 0, len(additionalCosts))
	for _, additional := range additionalCosts {
		if additional.Kind != cost.AdditionalDiscard {
			continue
		}
		switch {
		case additional.SubtypesAny[0] != "":
			sub := string(additional.SubtypesAny[0])
			parts = append(parts, indefiniteArticle(sub)+" "+sub+" card")
		case len(parts) > 0:
			parts = append(parts, "another card")
		default:
			parts = append(parts, "a card")
		}
	}
	return "Discard " + strings.Join(parts, " and ")
}

func indefiniteArticle(word string) string {
	if word == "" {
		return "a"
	}
	switch word[0] {
	case 'A', 'E', 'I', 'O', 'U', 'a', 'e', 'i', 'o', 'u':
		return "an"
	default:
		return "a"
	}
}

func lowerAlternativeCostCondition(alternative *compiler.CompiledAlternativeCost) (cost.AlternativeCondition, types.Sub, bool) {
	switch alternative.Condition {
	case compiler.AlternativeCostConditionUnknown:
		return cost.AlternativeConditionNone, "", true
	case compiler.AlternativeCostConditionNotYourTurn:
		return cost.AlternativeConditionNotYourTurn, "", true
	case compiler.AlternativeCostConditionYourTurn:
		return cost.AlternativeConditionYourTurn, "", true
	case compiler.AlternativeCostConditionControlsSubtype:
		if alternative.ConditionSubtype == "" {
			return cost.AlternativeConditionNone, "", false
		}
		return cost.AlternativeConditionControlsPermanentSubtype, alternative.ConditionSubtype, true
	default:
		return cost.AlternativeConditionNone, "", false
	}
}

// pitchAlternativeLabel builds the display label for a pitch alternative from
// its lowered exile cost, naming the exiled card's color when known.
func pitchAlternativeLabel(additionalCosts []cost.Additional) string {
	for _, additional := range additionalCosts {
		if additional.Kind != cost.AdditionalExile {
			continue
		}
		if additional.MatchCardColor {
			if name, ok := colorDisplayName(additional.CardColor); ok {
				return "Exile a " + name + " card"
			}
		}
		return "Exile a card"
	}
	return "Alternative cost"
}

func colorDisplayName(c color.Color) (string, bool) {
	switch c {
	case color.White:
		return "white", true
	case color.Blue:
		return "blue", true
	case color.Black:
		return "black", true
	case color.Red:
		return "red", true
	case color.Green:
		return "green", true
	default:
		return "", false
	}
}
