package payment

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

type additionalCostPlan struct {
	player          game.PlayerID
	sourceCardID    id.ID
	paid            []string
	sacrifices      []*game.Permanent
	permanentsToTap []*game.Permanent
	returnsToHand   []returnToHand
	exilePermanents []*game.Permanent
	exertSource     *game.Permanent
	millAmount      int
	discards        []id.ID
	exiles          []cardZoneSelection
	reveals         []cardZoneSelection
	evidence        []evidencePayment
	lifePaid        int
	energyPaid      int
	untapSource     *game.Permanent
	counterRemovals []counterRemoval
	counterAdds     []counterPlacement
}

type counterRemoval struct {
	source *game.Permanent
	kind   counter.Kind
	amount int
}

type counterPlacement struct {
	source *game.Permanent
	kind   counter.Kind
	amount int
}

type returnToHand struct {
	permanent  *game.Permanent
	additional cost.Additional
}

type cardZoneSelection struct {
	cardID id.ID
	zone   zone.Type
}

type evidencePayment struct {
	cards     []cardZoneSelection
	threshold int
}

//nolint:maintidx // Centralized cost dispatch keeps cross-cost reservation checks in one place.
func buildAdditionalCostPlanForCosts(s State, playerID game.PlayerID, costs []cost.Additional, xValue int, prefs *Preferences, source *game.Permanent, sourceCardID id.ID, sourceZone zone.Type, tapReservations ...*game.Permanent) (additionalCostPlan, bool) {
	if costsHaveChoiceGroup(costs) {
		concrete, ok := resolveAdditionalCostChoices(s, playerID, costs, xValue, prefs, source, sourceCardID, sourceZone, tapReservations...)
		if !ok {
			return additionalCostPlan{player: playerID, sourceCardID: sourceCardID}, false
		}
		return buildAdditionalCostPlanForCosts(s, playerID, concrete, xValue, prefs, source, sourceCardID, sourceZone, tapReservations...)
	}
	plan := additionalCostPlan{player: playerID, sourceCardID: sourceCardID}
	reservedTapPermanents := append([]*game.Permanent(nil), tapReservations...)
	if source != nil && hasTapCostOf(costs) {
		reservedTapPermanents = append(reservedTapPermanents, source)
	}
	for i, additional := range costs {
		amount := AdditionalCostAmountFor(additional, xValue)
		if additional.AmountDynamic != cost.AdditionalDynamicAmountNone {
			amount = s.AdditionalDynamicAmountValue(playerID, additional.AmountDynamic)
		}
		if amount < 0 {
			return plan, false
		}
		switch additional.Kind {
		case cost.AdditionalUnknown:
			if additional.Text == "" {
				continue
			}
			return plan, false
		case cost.AdditionalTap:
			continue
		case cost.AdditionalUntap:
			if amount != 1 ||
				source == nil ||
				s.EffectiveController(source) != playerID ||
				!canUntapForAbility(s, source) ||
				plan.untapSource != nil {
				return plan, false
			}
			plan.untapSource = source
			plan.paid = append(plan.paid, AdditionalCostText(additional))
		case cost.AdditionalRemoveCounter:
			if source == nil || s.EffectiveController(source) != playerID {
				return plan, false
			}
			planned := 0
			for _, removal := range plan.counterRemovals {
				if removal.source == source && removal.kind == additional.CounterKind {
					planned += removal.amount
				}
			}
			if source.Counters.Get(additional.CounterKind) < planned+amount {
				return plan, false
			}
			plan.counterRemovals = append(plan.counterRemovals, counterRemoval{
				source: source,
				kind:   additional.CounterKind,
				amount: amount,
			})
			plan.paid = append(plan.paid, AdditionalCostText(additional))
		case cost.AdditionalRemoveCounterAmong:
			if amount <= 0 {
				return plan, false
			}
			removals, ok := planRemoveCounterAmong(s, playerID, additional, amount, plan.counterRemovals, prefs)
			if !ok {
				return plan, false
			}
			plan.counterRemovals = append(plan.counterRemovals, removals...)
			plan.paid = append(plan.paid, AdditionalCostText(additional))
		case cost.AdditionalExert:
			if amount != 1 ||
				source == nil ||
				s.EffectiveController(source) != playerID ||
				plan.exertSource != nil {
				return plan, false
			}
			plan.exertSource = source
			plan.paid = append(plan.paid, AdditionalCostText(additional))
		case cost.AdditionalMill:
			plan.millAmount += amount
			plan.paid = append(plan.paid, AdditionalCostText(additional))
		case cost.AdditionalPutCounter:
			if source == nil ||
				s.EffectiveController(source) != playerID ||
				amount <= 0 ||
				!additional.CounterKind.Valid() ||
				additional.CounterKind.PlayerOnly() {
				return plan, false
			}
			plan.counterAdds = append(plan.counterAdds, counterPlacement{
				source: source,
				kind:   additional.CounterKind,
				amount: amount,
			})
			plan.paid = append(plan.paid, AdditionalCostText(additional))
		case cost.AdditionalCollectEvidence:
			if amount <= 0 {
				return plan, false
			}
			chosen := preferredEvidenceCards(s, playerID, amount, plannedEvidenceCards(plan), costs[i+1:], xValue, sourceCardID, sourceZone, prefs)
			if len(chosen) == 0 {
				return plan, false
			}
			plan.evidence = append(plan.evidence, evidencePayment{cards: chosen, threshold: amount})
			plan.paid = append(plan.paid, AdditionalCostText(additional))
		case cost.AdditionalSacrifice:
			chosen := preferredSacrificePermanents(s, playerID, additional, amount, plannedBattlefieldCosts(plan), prefs, source)
			if len(chosen) != amount && prefs != nil && len(prefs.SacrificeChoices) > 0 {
				chosen = chooseSacrificePermanents(s, playerID, additional, amount, plannedBattlefieldCosts(plan), source)
			}
			if len(chosen) != amount {
				return plan, false
			}
			plan.sacrifices = append(plan.sacrifices, chosen...)
			plan.paid = append(plan.paid, AdditionalCostText(additional))
		case cost.AdditionalTapPermanents:
			if additional.TotalPowerAtLeast > 0 {
				chosen := preferredTapPermanentsTotalPower(s, playerID, additional, append(plannedBattlefieldCosts(plan), reservedTapPermanents...), source, prefs)
				if len(chosen) == 0 {
					return plan, false
				}
				plan.permanentsToTap = append(plan.permanentsToTap, chosen...)
				reservedTapPermanents = append(reservedTapPermanents, chosen...)
				plan.paid = append(plan.paid, AdditionalCostText(additional))
				continue
			}
			chosen := preferredTapPermanents(s, playerID, additional, amount, append(plannedBattlefieldCosts(plan), reservedTapPermanents...), prefs)
			if len(chosen) != amount && prefs != nil && len(prefs.TapChoices) > 0 {
				chosen = chooseTapPermanents(s, playerID, additional, amount, append(plannedBattlefieldCosts(plan), reservedTapPermanents...))
			}
			if len(chosen) != amount {
				return plan, false
			}
			plan.permanentsToTap = append(plan.permanentsToTap, chosen...)
			reservedTapPermanents = append(reservedTapPermanents, chosen...)
			plan.paid = append(plan.paid, AdditionalCostText(additional))
		case cost.AdditionalReturnToHand:
			chosen := preferredReturnPermanents(s, playerID, additional, amount, plannedBattlefieldCosts(plan), prefs)
			if len(chosen) != amount {
				return plan, false
			}
			for _, permanent := range chosen {
				plan.returnsToHand = append(plan.returnsToHand, returnToHand{
					permanent:  permanent,
					additional: additional,
				})
			}
			plan.paid = append(plan.paid, AdditionalCostText(additional))
		case cost.AdditionalSacrificeSource:
			if amount != 1 ||
				source == nil ||
				permanentsInclude(plan.permanentsToTap, source) ||
				s.EffectiveController(source) != playerID ||
				!additionalCostMatchesPermanent(s, source, additional) {
				return plan, false
			}
			plan.sacrifices = append(plan.sacrifices, source)
			plan.paid = append(plan.paid, AdditionalCostText(additional))
		case cost.AdditionalDiscard:
			chosen := preferredDiscardCards(s, playerID, additional, amount, plan.discards, prefs)
			if len(chosen) != amount {
				return plan, false
			}
			plan.discards = append(plan.discards, chosen...)
			plan.paid = append(plan.paid, AdditionalCostText(additional))
		case cost.AdditionalPayLife:
			player, ok := s.Player(playerID)
			if !ok || !s.CanPayLife(playerID) || player.Life < plan.lifePaid+amount {
				return plan, false
			}
			plan.lifePaid += amount
			plan.paid = append(plan.paid, AdditionalCostText(additional))
		case cost.AdditionalEnergy:
			player, ok := s.Player(playerID)
			if !ok || player.EnergyCounters < plan.energyPaid+amount {
				return plan, false
			}
			plan.energyPaid += amount
			plan.paid = append(plan.paid, AdditionalCostText(additional))
		case cost.AdditionalExile:
			if additional.TotalManaValueAtLeast > 0 {
				chosen := preferredThresholdExileCards(s, playerID, additional, plannedEvidenceCards(plan), costs[i+1:], xValue, sourceCardID, sourceZone, prefs)
				if len(chosen) == 0 {
					return plan, false
				}
				plan.exiles = append(plan.exiles, chosen...)
				plan.paid = append(plan.paid, AdditionalCostText(additional))
				continue
			}
			chosen := preferredExileCards(s, playerID, additional, amount, plannedEvidenceCards(plan), costs[i+1:], xValue, sourceCardID, sourceZone, prefs)
			if len(chosen) != amount {
				return plan, false
			}
			plan.exiles = append(plan.exiles, chosen...)
			plan.paid = append(plan.paid, AdditionalCostText(additional))
		case cost.AdditionalReveal:
			if amount == 0 {
				plan.paid = append(plan.paid, AdditionalCostText(additional))
				continue
			}
			chosen := preferredRevealCards(s, playerID, additional, amount, plan.reveals, prefs)
			if len(chosen) != amount {
				return plan, false
			}
			plan.reveals = append(plan.reveals, chosen...)
			plan.paid = append(plan.paid, AdditionalCostText(additional))
		case cost.AdditionalExileSource:
			if sourceZone == zone.Battlefield {
				if amount != 1 ||
					source == nil ||
					permanentsInclude(plan.permanentsToTap, source) ||
					s.EffectiveController(source) != playerID ||
					!additionalCostMatchesPermanent(s, source, additional) {
					return plan, false
				}
				plan.exilePermanents = append(plan.exilePermanents, source)
				plan.paid = append(plan.paid, AdditionalCostText(additional))
				continue
			}
			if amount != 1 || sourceCardID == 0 || sourceZone == zone.None || !zoneContainsCard(s, playerID, sourceZone, sourceCardID) {
				return plan, false
			}
			if cardZoneSelectionsInclude(plannedEvidenceCards(plan), sourceCardID) {
				return plan, false
			}
			card, ok := s.CardInstance(sourceCardID)
			if !ok || !additionalCostMatchesCard(s.CardFace(card, game.FaceFront), additional) {
				return plan, false
			}
			plan.exiles = append(plan.exiles, cardZoneSelection{cardID: sourceCardID, zone: sourceZone})
			plan.paid = append(plan.paid, AdditionalCostText(additional))
		default:
			return plan, false
		}
	}

	return plan, true
}

// costsHaveChoiceGroup reports whether any cost belongs to a printed "or" choice
// group whose alternatives must be resolved to a single concrete cost before
// planning.
func costsHaveChoiceGroup(costs []cost.Additional) bool {
	for _, additional := range costs {
		if additional.ChoiceGroup != 0 {
			return true
		}
	}
	return false
}

// resolveAdditionalCostChoices collapses each printed "or" choice group into one
// payable alternative, returning the concrete cost list to plan. Mandatory costs
// (ChoiceGroup zero) are kept; for each choice group the first alternative that
// is payable in context is selected. It fails closed when any group has no
// payable alternative.
func resolveAdditionalCostChoices(s State, playerID game.PlayerID, costs []cost.Additional, xValue int, prefs *Preferences, source *game.Permanent, sourceCardID id.ID, sourceZone zone.Type, tapReservations ...*game.Permanent) ([]cost.Additional, bool) {
	concrete := make([]cost.Additional, 0, len(costs))
	var groups []uint8
	for _, additional := range costs {
		if additional.ChoiceGroup == 0 {
			concrete = append(concrete, additional)
			continue
		}
		if !slices.Contains(groups, additional.ChoiceGroup) {
			groups = append(groups, additional.ChoiceGroup)
		}
	}
	for _, group := range groups {
		picked := false
		for _, additional := range costs {
			if additional.ChoiceGroup != group {
				continue
			}
			member := additional
			member.ChoiceGroup = 0
			trial := append(append([]cost.Additional(nil), concrete...), member)
			if _, ok := buildAdditionalCostPlanForCosts(s, playerID, trial, xValue, prefs, source, sourceCardID, sourceZone, tapReservations...); ok {
				concrete = append(concrete, member)
				picked = true
				break
			}
		}
		if !picked {
			return nil, false
		}
	}
	return concrete, true
}

func plannedBattlefieldCosts(plan additionalCostPlan) []*game.Permanent {
	permanents := make([]*game.Permanent, 0, len(plan.sacrifices)+len(plan.permanentsToTap)+len(plan.returnsToHand)+len(plan.exilePermanents)+1)
	permanents = append(permanents, plan.sacrifices...)
	permanents = append(permanents, plan.permanentsToTap...)
	for _, returned := range plan.returnsToHand {
		permanents = append(permanents, returned.permanent)
	}
	permanents = append(permanents, plan.exilePermanents...)
	if plan.exertSource != nil {
		permanents = append(permanents, plan.exertSource)
	}
	return permanents
}

func plannedEvidenceCards(plan additionalCostPlan) []cardZoneSelection {
	var cards []cardZoneSelection
	cards = append(cards, plan.exiles...)
	for _, evidence := range plan.evidence {
		cards = append(cards, evidence.cards...)
	}
	return cards
}

func permanentsInclude(permanents []*game.Permanent, target *game.Permanent) bool {
	return slices.Contains(permanents, target)
}

func additionalCostMatchesPermanent(s State, permanent *game.Permanent, additional cost.Additional) bool {
	if additional.RequireTapped && !permanent.Tapped {
		return false
	}
	if additional.RequireToken && !permanent.Token {
		return false
	}
	if additional.RequireSupertype != "" && !s.PermanentHasSupertype(permanent, additional.RequireSupertype) {
		return false
	}
	if additional.MatchPermanentType && !s.PermanentHasType(permanent, additional.PermanentType) &&
		(additional.PermanentTypeAlt == "" || !s.PermanentHasType(permanent, additional.PermanentTypeAlt)) {
		return false
	}
	if additional.ExcludePermanentType != "" && s.PermanentHasType(permanent, additional.ExcludePermanentType) {
		return false
	}
	if additional.MatchHistoric &&
		!s.PermanentHasType(permanent, types.Artifact) &&
		!s.PermanentHasSupertype(permanent, types.Legendary) &&
		!s.PermanentHasSubtype(permanent, types.Saga) {
		return false
	}
	if additional.MatchCardColor && !slices.Contains(s.PermanentEffectiveColors(permanent), additional.CardColor) {
		return false
	}
	if additional.SubtypesAny != (cost.SubtypeSet{}) {
		for _, subtype := range additional.SubtypesAny {
			if subtype != "" && s.PermanentHasSubtype(permanent, subtype) {
				return true
			}
		}
		return false
	}
	return true
}

func additionalCostMatchesCard(card *game.CardDef, additional cost.Additional) bool {
	if card == nil {
		return false
	}
	if additional.MatchCardType && !card.HasType(additional.CardType) {
		return false
	}
	if additional.MatchHistoric && !cardDefIsHistoric(card) {
		return false
	}
	if additional.MatchCardColor && !slices.Contains(card.Colors, additional.CardColor) {
		return false
	}
	if additional.SubtypesAny != (cost.SubtypeSet{}) {
		for _, subtype := range additional.SubtypesAny {
			if subtype != "" && card.HasSubtype(subtype) {
				return true
			}
		}
		return false
	}
	return true
}

// cardDefIsHistoric reports whether a card is historic: an artifact, a
// legendary, or a Saga (CR 702.61b).
func cardDefIsHistoric(card *game.CardDef) bool {
	return card.HasType(types.Artifact) ||
		card.HasSupertype(types.Legendary) ||
		card.HasSubtype(types.Saga)
}

func evidenceCardManaValue(s State, cardID id.ID) (int, bool) {
	card, ok := s.CardInstance(cardID)
	if !ok {
		return 0, false
	}
	face := s.CardFace(card, game.FaceFront)
	if face == nil {
		return 0, false
	}
	if !face.ManaCost.Exists {
		return 0, true
	}
	for _, symbol := range face.ManaCost.Val {
		if symbol.Kind == cost.VariableSymbol {
			return 0, false
		}
	}
	return face.ManaValue(), true
}

func evidenceCardsMeetThreshold(s State, playerID game.PlayerID, cards []cardZoneSelection, threshold int) bool {
	total := 0
	for _, selection := range cards {
		if selection.zone != zone.Graveyard || !zoneContainsCard(s, playerID, zone.Graveyard, selection.cardID) {
			return false
		}
		manaValue, ok := evidenceCardManaValue(s, selection.cardID)
		if !ok {
			return false
		}
		total += manaValue
	}
	return total >= threshold
}

// AdditionalCostAmount returns the effective amount for an additional cost.
func AdditionalCostAmount(additional cost.Additional) int {
	return AdditionalCostAmountFor(additional, 0)
}

// AdditionalCostAmountFor returns the effective amount for an additional cost
// using the announced X value when the cost is variable.
func AdditionalCostAmountFor(additional cost.Additional, xValue int) int {
	if additional.AmountFromX {
		return xValue
	}
	if additional.Amount > 0 {
		return additional.Amount
	}
	return 1
}

// AdditionalCostText returns display text for an additional cost.
func AdditionalCostText(additional cost.Additional) string {
	if additional.Text != "" {
		return additional.Text
	}
	switch additional.Kind {
	case cost.AdditionalSacrifice:
		return "Sacrifice a permanent"
	case cost.AdditionalSacrificeSource:
		return "Sacrifice this permanent"
	case cost.AdditionalDiscard:
		return "Discard a card"
	case cost.AdditionalPayLife:
		return "Pay life"
	case cost.AdditionalEnergy:
		return fmt.Sprintf("Pay {E}x%d", AdditionalCostAmount(additional))
	case cost.AdditionalReturnToHand:
		return fmt.Sprintf("Return %d permanents to hand", AdditionalCostAmount(additional))
	case cost.AdditionalExert:
		return "Exert this permanent"
	case cost.AdditionalMill:
		return fmt.Sprintf("Mill %d cards", AdditionalCostAmount(additional))
	case cost.AdditionalPutCounter:
		return fmt.Sprintf("Put %d %s counters on source", AdditionalCostAmount(additional), additional.CounterKind)
	case cost.AdditionalCollectEvidence:
		return fmt.Sprintf("Collect evidence %d", AdditionalCostAmount(additional))
	case cost.AdditionalExile:
		return "Exile a card"
	case cost.AdditionalExileSource:
		return "Exile this card"
	case cost.AdditionalReveal:
		return "Reveal a card"
	case cost.AdditionalTap:
		return "{T}"
	case cost.AdditionalTapPermanents:
		if additional.TotalPowerAtLeast > 0 {
			if additional.Text != "" {
				return additional.Text
			}
			return fmt.Sprintf("Tap creatures with total power %d", additional.TotalPowerAtLeast)
		}
		return fmt.Sprintf("Tap %d permanents", AdditionalCostAmount(additional))
	case cost.AdditionalUntap:
		return "{Q}"
	case cost.AdditionalRemoveCounter:
		return "Remove a counter"
	case cost.AdditionalRemoveCounterAmong:
		if additional.AnyCounterKind {
			return fmt.Sprintf("Remove %d counters from among permanents you control", AdditionalCostAmount(additional))
		}
		return fmt.Sprintf("Remove %d %s counters from among permanents you control", AdditionalCostAmount(additional), additional.CounterKind)
	default:
		return "Additional cost"
	}
}
