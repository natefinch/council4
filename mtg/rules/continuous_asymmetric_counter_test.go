package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// TestAsymmetricCounterModifiesPowerToughnessIndependently covers CR 122.1 and
// 613.4c for asymmetric power/toughness counters: unlike the symmetric +1/+1
// counter, a +1/+0 counter raises only power and a +0/+1 counter raises only
// toughness, so a 2/2 carrying one of each becomes a 3/3 by independent axes.
func TestAsymmetricCounterModifiesPowerToughnessIndependently(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.Counters.Add(counter.PlusOnePlusZero, 1)
	creature.Counters.Add(counter.PlusZeroPlusOne, 1)

	if got := effectivePower(g, creature); got != 3 {
		t.Fatalf("effective power = %d, want 3 (+1/+0 and +0/+1 on 2/2)", got)
	}
	if got, ok := effectiveToughness(g, creature); !ok || got != 3 {
		t.Fatalf("effective toughness = %d ok=%v, want 3 true", got, ok)
	}
}

// TestAsymmetricMinusCounterReducesToughnessOnly covers a negative asymmetric
// counter: a -0/-1 counter lowers only toughness, leaving power untouched
// (CR 122.1, 613.4c). A 2/2 with two -0/-1 counters is a 2/0.
func TestAsymmetricMinusCounterReducesToughnessOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.Counters.Add(counter.MinusZeroMinusOne, 2)

	if got := effectivePower(g, creature); got != 2 {
		t.Fatalf("effective power = %d, want 2 (-0/-1 leaves power)", got)
	}
	if got, ok := effectiveToughness(g, creature); !ok || got != 0 {
		t.Fatalf("effective toughness = %d ok=%v, want 0 true", got, ok)
	}
}

// TestAsymmetricCounterStacksWithSymmetricCounter confirms asymmetric and
// symmetric power/toughness counters sum independently on each axis: a 2/2 with
// a +1/+1 counter and a +1/+2 counter becomes a 4/5 (CR 613.4c).
func TestAsymmetricCounterStacksWithSymmetricCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.Counters.Add(counter.PlusOnePlusOne, 1)
	creature.Counters.Add(counter.PlusOnePlusTwo, 1)

	if got := effectivePower(g, creature); got != 4 {
		t.Fatalf("effective power = %d, want 4 (+1/+1 and +1/+2 on 2/2)", got)
	}
	if got, ok := effectiveToughness(g, creature); !ok || got != 5 {
		t.Fatalf("effective toughness = %d ok=%v, want 5 true", got, ok)
	}
}
