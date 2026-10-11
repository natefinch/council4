package payment

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

func TestStunCounterPaymentReservations(t *testing.T) {
	for _, tc := range []struct {
		name       string
		stun       int
		reserved   int
		tapped     bool
		manaUntap  bool
		compatible bool
		excluded   bool
	}{
		{"exhausted-by-mana", 1, 1, true, true, false, true},
		{"remaining-counter", 2, 1, true, true, true, false},
		{"all-reserved", 2, 2, true, true, false, true},
		{"tap-mana-preserves-stun", 1, 1, false, false, true, false},
		{"no-mana-untap", 2, 1, true, false, true, false},
		{"stale-counters", 1, 2, true, true, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := &game.Permanent{ObjectID: 1, Tapped: tc.tapped}
			source.Counters.Add(counter.Stun, tc.stun)
			additional := additionalCostPlan{counterRemovals: []counterRemoval{{
				source: source, kind: counter.Stun, amount: tc.reserved,
			}}}
			manaPlan := paymentPlan{manaTaps: []manaTap{{permanent: source, untap: tc.manaUntap}}}
			if got := stunCounterPaymentsCompatible(manaPlan, additional); got != tc.compatible {
				t.Fatalf("compatible = %v, want %v", got, tc.compatible)
			}
			restrictions := additionalManaRestrictions(nil, additional, true)
			if restrictions.excluded[source.ObjectID] != tc.excluded {
				t.Fatalf("source excluded = %v, want %v", restrictions.excluded[source.ObjectID], tc.excluded)
			}
			if source.Tapped != tc.tapped || source.Counters.Get(counter.Stun) != tc.stun {
				t.Fatal("planning/validation mutated the source")
			}
		})
	}
}
