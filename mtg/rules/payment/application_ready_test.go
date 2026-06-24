package payment

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestPaymentApplicationReadyRejectsCombinedLifeOverflow proves the combined
// prevalidation gate refuses a plan whose mana life cost and additional life
// cost together exceed the player's life, even though each plan's own validity
// check (which only sees its own share) would pass. Catching this before any
// mutation is what makes a failed payment atomic: applyPaymentPlan would
// otherwise spend the mana-side life and leave the additional side to fail.
func TestPaymentApplicationReadyRejectsCombinedLifeOverflow(t *testing.T) {
	state := fakePaymentState{}
	player := &game.Player{ID: game.Player1, Life: 3}
	manaPlan := paymentPlan{lifePayment: 2}
	additionalPlan := additionalCostPlan{player: game.Player1, lifePaid: 2}

	if paymentApplicationReady(state, player, manaPlan, additionalPlan) {
		t.Fatal("paymentApplicationReady = true for combined life 4 > life 3, want false")
	}

	// Each plan in isolation is affordable at the same life total.
	if !paymentApplicationReady(state, player, manaPlan, additionalCostPlan{player: game.Player1}) {
		t.Fatal("mana-only life payment 2 <= life 3 should be ready")
	}
	if !paymentApplicationReady(state, player, paymentPlan{}, additionalCostPlan{player: game.Player1, lifePaid: 3}) {
		t.Fatal("additional-only life payment 3 <= life 3 should be ready")
	}

	// The combined cost becomes payable exactly when life covers the sum.
	player.Life = 4
	if !paymentApplicationReady(state, player, manaPlan, additionalPlan) {
		t.Fatal("combined life payment 4 == life 4 should be ready")
	}
}
