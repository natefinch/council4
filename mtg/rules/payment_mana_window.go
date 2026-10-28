package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
)

func (e *Engine) choosePreparedSpellCost(g *game.Game, request payment.SpellRequest, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (payment.PreparedSpellCost, bool) {
	costs := paymentOrch.planner(g).PrepareSpellCosts(request)
	manual := e.hasManualPaymentManaAbility(g, request.PlayerID)
	payable := make([]payment.PreparedSpellCost, 0, len(costs))
	fallback := 0
	foundPayable := false
	for _, prepared := range costs {
		canPay, feasible := preparedSpellCostFeasibility(g, request, prepared, manual)
		if feasible {
			if canPay && !foundPayable {
				fallback, foundPayable = len(payable), true
			}
			payable = append(payable, prepared)
		}
	}
	costs = payable
	if len(costs) == 0 {
		return payment.PreparedSpellCost{}, false
	}
	options := make([]game.ChoiceOption, len(costs))
	for index, prepared := range costs {
		options[index] = game.ChoiceOption{Index: index, Label: prepared.Summary.Label}
	}
	selected := []int{fallback}
	if len(options) > 1 {
		selected = e.chooseChoice(g, agents, game.ChoiceRequest{
			Kind: game.ChoicePayment, Player: request.PlayerID, Prompt: "Choose spell cost before activating mana abilities.",
			Options: options, MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{fallback},
		}, log)
	}
	return costs[selected[0]], true
}

// spellCostFeasibility reports whether request's total cost can be paid from
// current resources (payableNow), and whether it can be paid now or after the
// caster activates payment-time mana abilities (feasible). manual reports
// whether the caster has such an ability.
func spellCostFeasibility(g *game.Game, request payment.SpellRequest, manual bool) (payableNow, feasible bool) {
	if paymentOrch.canPaySpellCosts(g, request) {
		return true, true
	}
	if !manual {
		return false, false
	}
	for _, prepared := range paymentOrch.planner(g).PrepareSpellCosts(request) {
		if _, feasible := preparedSpellCostFeasibility(g, request, prepared, true); feasible {
			return false, true
		}
	}
	return false, false
}

// preparedSpellCostFeasibility classifies one locked total cost (CR 601.2f).
// Mana abilities may be activated while paying only when the total includes a
// mana payment (CR 601.2g), and they add only mana, so a total that is not
// payable now is feasible only if it includes mana and its non-mana costs are
// payable now.
func preparedSpellCostFeasibility(g *game.Game, request payment.SpellRequest, prepared payment.PreparedSpellCost, manual bool) (payableNow, feasible bool) {
	proposal := request
	proposal.PreparedCost = &prepared
	if paymentOrch.canPaySpellCosts(g, proposal) {
		return true, true
	}
	return false, manual && prepared.IncludesManaPayment() &&
		paymentOrch.planner(g).CanPayPreparedNonManaCosts(proposal)
}

func (e *Engine) hasManualPaymentManaAbility(g *game.Game, player game.PlayerID) bool {
	g.ManaPaymentScopes = append(g.ManaPaymentScopes, player)
	defer func() { g.ManaPaymentScopes = g.ManaPaymentScopes[:len(g.ManaPaymentScopes)-1] }()
	return len(e.legalManaAbilityActions(g, player)) != 0
}

func (e *Engine) activateManualPaymentManaAbilities(g *game.Game, player game.PlayerID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) func() {
	g.ManaPaymentScopes = append(g.ManaPaymentScopes, player)
	defer func() { g.ManaPaymentScopes = g.ManaPaymentScopes[:len(g.ManaPaymentScopes)-1] }()
	var undo func()
	finish := func() {
		if undo != nil {
			undo()
		}
	}
	for {
		actions := e.legalManaAbilityActions(g, player)
		if len(actions) == 0 {
			return finish
		}
		options := []game.ChoiceOption{{Index: 0, Label: "Finish activating mana abilities"}}
		for _, candidate := range actions {
			activation, _ := candidate.ActivateAbilityPayload()
			label := activatedAbilityText(g, player, activation)
			options = append(options, game.ChoiceOption{Index: len(options), Label: label, Card: cardChoiceInfo(g, activation.SourceID)})
		}
		selected := e.chooseChoice(g, agents, game.ChoiceRequest{
			Kind: game.ChoicePayment, Player: player, Prompt: "Activate a mana ability while paying?",
			Options: options, MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{0},
		}, log)
		if selected[0] == 0 {
			return finish
		}
		if undo == nil {
			undo = reversibleManaCheckpoint(g, log)
		}
		var libraries [game.NumPlayers]zone.Zone
		for owner := range libraries {
			libraries[owner] = g.Players[owner].Library.Clone()
		}
		eventStart := len(g.Events)
		activation, _ := actions[selected[0]-1].ActivateAbilityPayload()
		if !e.applyActivateAbilityWithChoices(g, player, activation, agents, log) {
			panic("validated payment-time mana ability failed")
		}
		if !manaActivationIsReversible(g, libraries, eventStart) {
			// CR 733.1 forbids reversing a draw, library move, shuffle or reveal.
			// Keep that activation and any mana it consumed; only later
			// reversible activations may be undone on an unpaid proposal.
			undo = nil
		}
	}
}
