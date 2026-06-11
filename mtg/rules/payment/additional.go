package payment

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
)

type additionalCostPlan struct {
	player          game.PlayerID
	sourceCardID    id.ID
	paid            []string
	sacrifices      []*game.Permanent
	permanentsToTap []*game.Permanent
	returnsToHand   []returnToHand
	exilePermanents []*game.Permanent
	discards        []id.ID
	exiles          []cardZoneSelection
	reveals         []cardZoneSelection
	lifePaid        int
	energyPaid      int
	untapSource     *game.Permanent
	counterRemovals []counterRemoval
}

type counterRemoval struct {
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

//nolint:maintidx // Centralized cost dispatch keeps cross-cost reservation checks in one place.
func buildAdditionalCostPlanForCosts(s State, playerID game.PlayerID, costs []cost.Additional, xValue int, prefs *Preferences, source *game.Permanent, sourceCardID id.ID, sourceZone zone.Type, tapReservations ...*game.Permanent) (additionalCostPlan, bool) {
	plan := additionalCostPlan{player: playerID, sourceCardID: sourceCardID}
	reservedTapPermanents := append([]*game.Permanent(nil), tapReservations...)
	if source != nil && hasTapCostOf(costs) {
		reservedTapPermanents = append(reservedTapPermanents, source)
	}
	for _, additional := range costs {
		amount := AdditionalCostAmountFor(additional, xValue)
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
		case cost.AdditionalSacrifice:
			chosen := preferredSacrificePermanents(s, playerID, additional, amount, plannedBattlefieldCosts(plan), prefs)
			if len(chosen) != amount && prefs != nil && len(prefs.SacrificeChoices) > 0 {
				chosen = chooseSacrificePermanents(s, playerID, additional, amount, plannedBattlefieldCosts(plan))
			}
			if len(chosen) != amount {
				return plan, false
			}
			plan.sacrifices = append(plan.sacrifices, chosen...)
			plan.paid = append(plan.paid, AdditionalCostText(additional))
		case cost.AdditionalTapPermanents:
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
			if !ok || player.Life < amount {
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
			chosen := preferredExileCards(s, playerID, additional, amount, plan.exiles, prefs)
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

func plannedBattlefieldCosts(plan additionalCostPlan) []*game.Permanent {
	permanents := make([]*game.Permanent, 0, len(plan.sacrifices)+len(plan.permanentsToTap)+len(plan.returnsToHand)+len(plan.exilePermanents))
	permanents = append(permanents, plan.sacrifices...)
	permanents = append(permanents, plan.permanentsToTap...)
	for _, returned := range plan.returnsToHand {
		permanents = append(permanents, returned.permanent)
	}
	permanents = append(permanents, plan.exilePermanents...)
	return permanents
}

func permanentsInclude(permanents []*game.Permanent, target *game.Permanent) bool {
	return slices.Contains(permanents, target)
}

func chooseExileCards(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []cardZoneSelection) []cardZoneSelection {
	player, ok := s.Player(playerID)
	if !ok {
		return nil
	}
	sourceZone := additionalCostSourceZone(additional.Source)
	chosenIDs := make(map[id.ID]bool)
	for _, chosen := range alreadyChosen {
		chosenIDs[chosen.cardID] = true
	}
	var chosen []cardZoneSelection
	for _, cardID := range cardIDsInZone(player, sourceZone) {
		if chosenIDs[cardID] {
			continue
		}
		card, ok := s.CardInstance(cardID)
		if !ok || !additionalCostMatchesCard(s.CardFace(card, game.FaceFront), additional) {
			continue
		}
		chosen = append(chosen, cardZoneSelection{cardID: cardID, zone: sourceZone})
		if len(chosen) == amount {
			return chosen
		}
	}
	return chosen
}

func preferredExileCards(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []cardZoneSelection, prefs *Preferences) []cardZoneSelection {
	if prefs == nil || len(prefs.ExileChoices) == 0 {
		return chooseExileCards(s, playerID, additional, amount, alreadyChosen)
	}
	sourceZone := additionalCostSourceZone(additional.Source)
	chosenIDs := make(map[id.ID]bool)
	for _, chosen := range alreadyChosen {
		chosenIDs[chosen.cardID] = true
	}
	var chosen []cardZoneSelection
	var consumed int
	for _, cardID := range prefs.ExileChoices {
		card, ok := s.CardInstance(cardID)
		if !ok || !zoneContainsCard(s, playerID, sourceZone, cardID) || chosenIDs[cardID] || !additionalCostMatchesCard(s.CardFace(card, game.FaceFront), additional) {
			return nil
		}
		chosen = append(chosen, cardZoneSelection{cardID: cardID, zone: sourceZone})
		chosenIDs[cardID] = true
		consumed++
		if len(chosen) == amount {
			prefs.ExileChoices = prefs.ExileChoices[consumed:]
			return chosen
		}
	}
	return nil
}

func preferredRevealCards(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []cardZoneSelection, prefs *Preferences) []cardZoneSelection {
	sourceZone := additional.Source
	if sourceZone == zone.None {
		sourceZone = zone.Hand
	}
	if prefs == nil || len(prefs.RevealChoices) == 0 {
		additional.Source = sourceZone
		return chooseExileCards(s, playerID, additional, amount, alreadyChosen)
	}
	chosenIDs := make(map[id.ID]bool, len(alreadyChosen))
	for _, chosen := range alreadyChosen {
		chosenIDs[chosen.cardID] = true
	}
	var chosen []cardZoneSelection
	var consumed int
	for _, cardID := range prefs.RevealChoices {
		card, ok := s.CardInstance(cardID)
		if !ok ||
			!zoneContainsCard(s, playerID, sourceZone, cardID) ||
			chosenIDs[cardID] ||
			!additionalCostMatchesCard(s.CardFace(card, game.FaceFront), additional) {
			return nil
		}
		chosen = append(chosen, cardZoneSelection{cardID: cardID, zone: sourceZone})
		chosenIDs[cardID] = true
		consumed++
		if len(chosen) == amount {
			prefs.RevealChoices = prefs.RevealChoices[consumed:]
			return chosen
		}
	}
	return nil
}

func additionalCostSourceZone(source zone.Type) zone.Type {
	if source == zone.None {
		return zone.Graveyard
	}
	return source
}

func chooseSacrificePermanents(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []*game.Permanent) []*game.Permanent {
	chosenIDs := make(map[id.ID]bool)
	for _, permanent := range alreadyChosen {
		chosenIDs[permanent.ObjectID] = true
	}
	var chosen []*game.Permanent
	for _, permanent := range s.Battlefield() {
		if s.EffectiveController(permanent) != playerID || chosenIDs[permanent.ObjectID] {
			continue
		}
		if additionalCostMatchesPermanent(s, permanent, additional) {
			chosen = append(chosen, permanent)
			if len(chosen) == amount {
				return chosen
			}
		}
	}
	return chosen
}

func preferredSacrificePermanents(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []*game.Permanent, prefs *Preferences) []*game.Permanent {
	if prefs == nil || len(prefs.SacrificeChoices) == 0 {
		return chooseSacrificePermanents(s, playerID, additional, amount, alreadyChosen)
	}
	chosenIDs := make(map[id.ID]bool)
	for _, permanent := range alreadyChosen {
		chosenIDs[permanent.ObjectID] = true
	}
	var chosen []*game.Permanent
	var consumed int
	for _, permanentID := range prefs.SacrificeChoices {
		permanent, ok := s.PermanentByObjectID(permanentID)
		if !ok || s.EffectiveController(permanent) != playerID || chosenIDs[permanentID] || !additionalCostMatchesPermanent(s, permanent, additional) {
			return nil
		}
		chosen = append(chosen, permanent)
		chosenIDs[permanentID] = true
		consumed++
		if len(chosen) == amount {
			prefs.SacrificeChoices = prefs.SacrificeChoices[consumed:]
			return chosen
		}
	}
	return nil
}

func chooseTapPermanents(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []*game.Permanent) []*game.Permanent {
	chosenIDs := make(map[id.ID]bool)
	for _, permanent := range alreadyChosen {
		chosenIDs[permanent.ObjectID] = true
	}
	var chosen []*game.Permanent
	for _, permanent := range s.Battlefield() {
		if permanent.Tapped || s.EffectiveController(permanent) != playerID || chosenIDs[permanent.ObjectID] {
			continue
		}
		if additionalCostMatchesPermanent(s, permanent, additional) {
			chosen = append(chosen, permanent)
			if len(chosen) == amount {
				return chosen
			}
		}
	}
	return chosen
}

func preferredTapPermanents(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []*game.Permanent, prefs *Preferences) []*game.Permanent {
	if prefs == nil || len(prefs.TapChoices) == 0 {
		return chooseTapPermanents(s, playerID, additional, amount, alreadyChosen)
	}
	chosenIDs := make(map[id.ID]bool)
	for _, permanent := range alreadyChosen {
		chosenIDs[permanent.ObjectID] = true
	}
	var chosen []*game.Permanent
	var consumed int
	for _, permanentID := range prefs.TapChoices {
		permanent, ok := s.PermanentByObjectID(permanentID)
		if !ok ||
			permanent.Tapped ||
			s.EffectiveController(permanent) != playerID ||
			chosenIDs[permanentID] ||
			!additionalCostMatchesPermanent(s, permanent, additional) {
			return nil
		}
		chosen = append(chosen, permanent)
		chosenIDs[permanentID] = true
		consumed++
		if len(chosen) == amount {
			prefs.TapChoices = prefs.TapChoices[consumed:]
			return chosen
		}
	}
	return nil
}

func chooseReturnPermanents(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []*game.Permanent) []*game.Permanent {
	chosenIDs := make(map[id.ID]bool)
	for _, permanent := range alreadyChosen {
		chosenIDs[permanent.ObjectID] = true
	}
	var chosen []*game.Permanent
	for _, permanent := range s.Battlefield() {
		if s.EffectiveController(permanent) != playerID || chosenIDs[permanent.ObjectID] {
			continue
		}
		if additionalCostMatchesPermanent(s, permanent, additional) {
			chosen = append(chosen, permanent)
			if len(chosen) == amount {
				return chosen
			}
		}
	}
	return chosen
}

func preferredReturnPermanents(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []*game.Permanent, prefs *Preferences) []*game.Permanent {
	if prefs == nil || len(prefs.ReturnChoices) == 0 {
		return chooseReturnPermanents(s, playerID, additional, amount, alreadyChosen)
	}
	chosenIDs := make(map[id.ID]bool)
	for _, permanent := range alreadyChosen {
		chosenIDs[permanent.ObjectID] = true
	}
	var chosen []*game.Permanent
	var consumed int
	for _, permanentID := range prefs.ReturnChoices {
		permanent, ok := s.PermanentByObjectID(permanentID)
		if !ok ||
			s.EffectiveController(permanent) != playerID ||
			chosenIDs[permanentID] ||
			!additionalCostMatchesPermanent(s, permanent, additional) {
			return nil
		}
		chosen = append(chosen, permanent)
		chosenIDs[permanentID] = true
		consumed++
		if len(chosen) == amount {
			prefs.ReturnChoices = prefs.ReturnChoices[consumed:]
			return chosen
		}
	}
	return nil
}

func chooseDiscardCards(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []id.ID) []id.ID {
	player, ok := s.Player(playerID)
	if !ok {
		return nil
	}
	chosenIDs := make(map[id.ID]bool)
	for _, cardID := range alreadyChosen {
		chosenIDs[cardID] = true
	}
	var chosen []id.ID
	for _, cardID := range player.Hand.All() {
		if chosenIDs[cardID] {
			continue
		}
		card, ok := s.CardInstance(cardID)
		if !ok || !additionalCostMatchesCard(s.CardFace(card, game.FaceFront), additional) {
			continue
		}
		chosen = append(chosen, cardID)
		if len(chosen) == amount {
			return chosen
		}
	}
	return chosen
}

func preferredDiscardCards(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []id.ID, prefs *Preferences) []id.ID {
	if prefs == nil || len(prefs.DiscardChoices) == 0 {
		return chooseDiscardCards(s, playerID, additional, amount, alreadyChosen)
	}
	player, ok := s.Player(playerID)
	if !ok {
		return nil
	}
	chosenIDs := make(map[id.ID]bool)
	for _, cardID := range alreadyChosen {
		chosenIDs[cardID] = true
	}
	var chosen []id.ID
	var consumed int
	for _, cardID := range prefs.DiscardChoices {
		card, ok := s.CardInstance(cardID)
		if !ok || !player.Hand.Contains(cardID) || chosenIDs[cardID] || !additionalCostMatchesCard(s.CardFace(card, game.FaceFront), additional) {
			return nil
		}
		chosen = append(chosen, cardID)
		chosenIDs[cardID] = true
		consumed++
		if len(chosen) == amount {
			prefs.DiscardChoices = prefs.DiscardChoices[consumed:]
			return chosen
		}
	}
	return nil
}

func additionalCostMatchesPermanent(s State, permanent *game.Permanent, additional cost.Additional) bool {
	if additional.RequireTapped && !permanent.Tapped {
		return false
	}
	if additional.RequireSupertype != "" && !s.PermanentHasSupertype(permanent, additional.RequireSupertype) {
		return false
	}
	if additional.MatchPermanentType && !s.PermanentHasType(permanent, additional.PermanentType) {
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
	case cost.AdditionalExile:
		return "Exile a card"
	case cost.AdditionalExileSource:
		return "Exile this card"
	case cost.AdditionalReveal:
		return "Reveal a card"
	case cost.AdditionalTap:
		return "{T}"
	case cost.AdditionalTapPermanents:
		return fmt.Sprintf("Tap %d permanents", AdditionalCostAmount(additional))
	case cost.AdditionalUntap:
		return "{Q}"
	case cost.AdditionalRemoveCounter:
		return "Remove a counter"
	default:
		return "Additional cost"
	}
}

func additionalCostPlanStillValid(s State, player *game.Player, plan additionalCostPlan) bool {
	if plan.untapSource != nil {
		current, ok := s.PermanentByObjectID(plan.untapSource.ObjectID)
		if !ok ||
			current != plan.untapSource ||
			s.EffectiveController(current) != player.ID ||
			!canUntapForAbility(s, current) {
			return false
		}
	}
	plannedCounters := make(map[*game.Permanent]map[counter.Kind]int)
	for _, removal := range plan.counterRemovals {
		current, ok := s.PermanentByObjectID(removal.source.ObjectID)
		if !ok || current != removal.source || s.EffectiveController(current) != player.ID {
			return false
		}
		if plannedCounters[current] == nil {
			plannedCounters[current] = make(map[counter.Kind]int)
		}
		plannedCounters[current][removal.kind] += removal.amount
		if current.Counters.Get(removal.kind) < plannedCounters[current][removal.kind] {
			return false
		}
	}
	for _, sacrifice := range plan.sacrifices {
		permanent, ok := s.PermanentByObjectID(sacrifice.ObjectID)
		if !ok || s.EffectiveController(permanent) != player.ID || permanent != sacrifice {
			return false
		}
	}
	for _, permanentToTap := range plan.permanentsToTap {
		permanent, ok := s.PermanentByObjectID(permanentToTap.ObjectID)
		if !ok ||
			s.EffectiveController(permanent) != player.ID ||
			permanent != permanentToTap ||
			permanent.Tapped {
			return false
		}
	}
	for _, returned := range plan.returnsToHand {
		permanent, ok := s.PermanentByObjectID(returned.permanent.ObjectID)
		if !ok ||
			s.EffectiveController(permanent) != player.ID ||
			permanent != returned.permanent ||
			!additionalCostMatchesPermanent(s, permanent, returned.additional) {
			return false
		}
	}
	for _, permanent := range plan.exilePermanents {
		current, ok := s.PermanentByObjectID(permanent.ObjectID)
		if !ok || s.EffectiveController(current) != player.ID || current != permanent {
			return false
		}
	}
	for _, cardID := range plan.discards {
		if !player.Hand.Contains(cardID) {
			return false
		}
	}
	for _, exile := range plan.exiles {
		if !zoneContainsCard(s, player.ID, exile.zone, exile.cardID) {
			return false
		}
	}
	for _, reveal := range plan.reveals {
		if !zoneContainsCard(s, player.ID, reveal.zone, reveal.cardID) {
			return false
		}
	}
	if plan.lifePaid > 0 && player.Life < plan.lifePaid {
		return false
	}
	if plan.energyPaid > 0 && player.EnergyCounters < plan.energyPaid {
		return false
	}
	return true
}

func applyAdditionalCostPlan(s State, plan additionalCostPlan) bool {
	if plan.untapSource != nil {
		s.SetTapped(plan.untapSource, false)
	}
	for _, removal := range plan.counterRemovals {
		if !s.RemoveCounters(removal.source, removal.kind, removal.amount) {
			return false
		}
	}
	for _, sacrifice := range plan.sacrifices {
		if !s.MovePermanentToZone(sacrifice, zone.Graveyard) {
			return false
		}
	}
	for _, permanentToTap := range plan.permanentsToTap {
		s.SetTapped(permanentToTap, true)
	}
	for _, returned := range plan.returnsToHand {
		if !s.MovePermanentToZone(returned.permanent, zone.Hand) {
			return false
		}
	}
	for _, permanent := range plan.exilePermanents {
		if !s.MovePermanentToZone(permanent, zone.Exile) {
			return false
		}
	}
	for _, cardID := range plan.discards {
		if !s.DiscardFromHand(plan.player, cardID) {
			return false
		}
	}
	for _, exile := range plan.exiles {
		if !s.MoveCard(plan.player, exile.cardID, exile.zone, zone.Exile) {
			return false
		}
	}
	for _, reveal := range plan.reveals {
		s.EmitCardReveal(plan.player, plan.sourceCardID, reveal.cardID, reveal.zone)
	}
	if plan.lifePaid > 0 {
		player, ok := s.Player(plan.player)
		if !ok || player.Life < plan.lifePaid {
			return false
		}
		s.LoseLife(plan.player, plan.lifePaid)
	}
	if plan.energyPaid > 0 {
		player, ok := s.Player(plan.player)
		if !ok || player.EnergyCounters < plan.energyPaid {
			return false
		}
		if !s.SetPlayerEnergyCounters(plan.player, player.EnergyCounters-plan.energyPaid) {
			return false
		}
	}
	return true
}

func zoneContainsCard(s State, playerID game.PlayerID, zoneType zone.Type, cardID id.ID) bool {
	player, ok := s.Player(playerID)
	if !ok {
		return false
	}
	return slices.Contains(cardIDsInZone(player, zoneType), cardID)
}

func cardIDsInZone(player *game.Player, zoneType zone.Type) []id.ID {
	switch zoneType {
	case zone.Library:
		return player.Library.All()
	case zone.Hand:
		return player.Hand.All()
	case zone.Graveyard:
		return player.Graveyard.All()
	case zone.Exile:
		return player.Exile.All()
	case zone.Command:
		return player.CommandZone.All()
	default:
		return nil
	}
}
