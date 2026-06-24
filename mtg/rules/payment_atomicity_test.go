package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
)

// TestFailedEnergyPaymentLeavesPriorCostsUnapplied proves a payment that cannot
// satisfy a later resource cost (here energy) leaves every earlier cost's
// objects untouched: the source keeps its counters and stays untapped, and the
// player's energy is unchanged. Before atomic prevalidation the counter cost
// could be applied before the energy shortfall was discovered.
func TestFailedEnergyPaymentLeavesPriorCostsUnapplied(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Reactor"}})
	source.Counters.Add(counter.Charge, 2)
	g.Players[game.Player1].EnergyCounters = 2

	_, ok := paymentOrch.payAbilityCosts(g, payment.AbilityRequest{
		PlayerID: game.Player1,
		Source:   source,
		AdditionalCosts: []cost.Additional{
			{
				Kind:        cost.AdditionalRemoveCounter,
				Text:        "Remove a charge counter from this permanent",
				Amount:      1,
				CounterKind: counter.Charge,
			},
			{
				Kind:   cost.AdditionalEnergy,
				Text:   "Pay {E}{E}{E}{E}{E}",
				Amount: 5,
			},
		},
	})
	if ok {
		t.Fatal("payAbilityCosts paid an unaffordable energy cost")
	}
	if got := source.Counters.Get(counter.Charge); got != 2 {
		t.Fatalf("charge counters = %d, want 2 unchanged after failed payment", got)
	}
	if source.Tapped {
		t.Fatal("source was tapped by a failed payment")
	}
	if got := g.Players[game.Player1].EnergyCounters; got != 2 {
		t.Fatalf("energy counters = %d, want 2 unchanged after failed payment", got)
	}
}

// discardOneCardAbilityPermanent builds a permanent whose activated ability
// costs discarding one card from hand.
func discardOneCardAbilityPermanent() *game.CardDef {
	return activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{{
			Kind:   cost.AdditionalDiscard,
			Text:   "Discard a card",
			Amount: 1,
			Source: zone.Hand,
		}},
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.GainLife{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
			}}},
		}.Ability(),
	})
}

// TestDiscardStalePreferenceFallsBackToDeterministicPlan proves the uniform
// invalid-preference policy now applies to a card cost (discard) that formerly
// rejected outright: a stale discard preference falls back to a deterministic
// legal discard so play continues.
func TestDiscardStalePreferenceFallsBackToDeterministicPlan(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, discardOneCardAbilityPermanent())
	handCard := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Legal Card", Types: []types.Card{types.Instant}}})

	_, ok := paymentOrch.payAbilityCosts(g, payment.AbilityRequest{
		PlayerID:        game.Player1,
		Source:          source,
		AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalDiscard, Text: "Discard a card", Amount: 1, Source: zone.Hand}},
		Prefs:           &payment.Preferences{DiscardChoices: []id.ID{id.ID(99999)}},
	})
	if !ok {
		t.Fatal("stale discard preference did not fall back to a deterministic legal plan")
	}
	if g.Players[game.Player1].Hand.Contains(handCard) || !g.Players[game.Player1].Graveyard.Contains(handCard) {
		t.Fatal("discard fallback did not move the legal hand card to the graveyard")
	}
}

// TestDiscardStalePreferenceRejectedUnderStrictReplayWithoutMutation proves the
// strict-replay policy disables the fallback: the stale discard preference
// rejects the payment and leaves the hand untouched.
func TestDiscardStalePreferenceRejectedUnderStrictReplayWithoutMutation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, discardOneCardAbilityPermanent())
	handCard := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Legal Card", Types: []types.Card{types.Instant}}})

	_, ok := paymentOrch.payAbilityCosts(g, payment.AbilityRequest{
		PlayerID:        game.Player1,
		Source:          source,
		AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalDiscard, Text: "Discard a card", Amount: 1, Source: zone.Hand}},
		Prefs:           &payment.Preferences{DiscardChoices: []id.ID{id.ID(99999)}, StrictReplay: true},
	})
	if ok {
		t.Fatal("strict-replay stale discard preference paid successfully")
	}
	if !g.Players[game.Player1].Hand.Contains(handCard) || g.Players[game.Player1].Graveyard.Contains(handCard) {
		t.Fatal("strict-replay stale discard preference mutated the hand")
	}
}
