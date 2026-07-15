package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// jetmirThresholdAnthem builds one of Jetmir, Nexus of Revels' stacked anthem
// static abilities: while the source's controller controls at least minCount
// creatures, every creature that controller controls (including Jetmir itself)
// gets +1/+0 and the given keyword.
func jetmirThresholdAnthem(minCount int, keyword game.Keyword) game.StaticAbility {
	return game.StaticAbility{
		Condition: opt.Val(game.Condition{
			ControlsMatching: opt.Val(game.SelectionCount{
				Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
				MinCount:  minCount,
			}),
		}),
		ContinuousEffects: []game.ContinuousEffect{
			{
				Layer: game.LayerPowerToughnessModify,
				Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{
					RequiredTypes: []types.Card{types.Creature},
				}),
				PowerDelta: 1,
			},
			{
				Layer: game.LayerAbility,
				Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{
					RequiredTypes: []types.Card{types.Creature},
				}),
				AddKeywords: []game.Keyword{keyword},
			},
		},
	}
}

// jetmirPermanent mirrors the lowered shape of Jetmir, Nexus of Revels: a 3/3
// creature carrying three independently threshold-gated anthems at three, six,
// and nine controlled creatures.
func jetmirPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "Jetmir, Nexus of Revels",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{
			jetmirThresholdAnthem(3, game.Vigilance),
			jetmirThresholdAnthem(6, game.Trample),
			jetmirThresholdAnthem(9, game.DoubleStrike),
		},
	}})
}

// TestJetmirThresholdAnthemsStackAndTrack verifies Jetmir's three stacked
// anthems each switch on independently as the controller's creature count
// crosses three, six, and nine, that the bonuses accumulate (+1/+0 per active
// threshold) on every creature the controller controls including Jetmir, and
// that the bonuses disappear again as the count falls back below each threshold.
func TestJetmirThresholdAnthemsStackAndTrack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	jetmir := jetmirPermanent(g, game.Player1)
	witness := addCombatCreaturePermanentWithPower(g, game.Player1, 2)

	// creatureCount returns the number of creatures Player1 controls: Jetmir,
	// the witness, and any extra vanilla creatures added so far.
	extras := []*game.Permanent{}
	setCount := func(target int) {
		for len(extras)+2 < target {
			extras = append(extras, addCombatCreaturePermanent(g, game.Player1))
		}
		for len(extras)+2 > target {
			last := extras[len(extras)-1]
			g.Battlefield = removePermanent(g.Battlefield, last)
			extras = extras[:len(extras)-1]
		}
	}

	assert := func(count, wantWitnessPower, wantJetmirPower int, vig, tramp, dbl bool) {
		t.Helper()
		if got := effectivePower(g, witness); got != wantWitnessPower {
			t.Fatalf("%d creatures: witness power = %d, want %d", count, got, wantWitnessPower)
		}
		if got := effectivePower(g, jetmir); got != wantJetmirPower {
			t.Fatalf("%d creatures: Jetmir power = %d, want %d", count, got, wantJetmirPower)
		}
		for _, tc := range []struct {
			kw   game.Keyword
			want bool
		}{{game.Vigilance, vig}, {game.Trample, tramp}, {game.DoubleStrike, dbl}} {
			if got := hasKeyword(g, witness, tc.kw); got != tc.want {
				t.Fatalf("%d creatures: witness %v = %v, want %v", count, tc.kw, got, tc.want)
			}
			if got := hasKeyword(g, jetmir, tc.kw); got != tc.want {
				t.Fatalf("%d creatures: Jetmir %v = %v, want %v", count, tc.kw, got, tc.want)
			}
		}
	}

	// Below three creatures: no anthem active.
	setCount(2)
	assert(2, 2, 3, false, false, false)

	// Three creatures: +1/+0 and vigilance.
	setCount(3)
	assert(3, 3, 4, true, false, false)

	// Five creatures: still only the first threshold.
	setCount(5)
	assert(5, 3, 4, true, false, false)

	// Six creatures: first two thresholds stack (+2/+0, vigilance and trample).
	setCount(6)
	assert(6, 4, 5, true, true, false)

	// Nine creatures: all three thresholds stack (+3/+0 and all keywords).
	setCount(9)
	assert(9, 5, 6, true, true, true)

	// Dropping back below nine removes double strike and the third +1/+0.
	setCount(8)
	assert(8, 4, 5, true, true, false)

	// Dropping below three removes every bonus again.
	setCount(2)
	assert(2, 2, 3, false, false, false)
}

// TestJetmirAnthemFollowsControllerChange verifies that when Jetmir changes
// controller, both the creature-count condition and the affected group move to
// the new controller: the old controller's creatures lose the bonus and the new
// controller's creatures (including Jetmir) gain it based on the new
// controller's creature count.
func TestJetmirAnthemFollowsControllerChange(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	jetmir := jetmirPermanent(g, game.Player1)
	p1Creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	addCombatCreaturePermanent(g, game.Player1) // Player1: Jetmir + 2 others = 3 creatures.

	p2CreatureA := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	addCombatCreaturePermanent(g, game.Player2) // Player2: 2 creatures before Jetmir moves.

	// While Player1 controls Jetmir (3 creatures) the first anthem is active for
	// Player1's creatures; Player2 (2 creatures) gets nothing.
	if got := effectivePower(g, p1Creature); got != 3 {
		t.Fatalf("Player1 creature power under Player1 Jetmir = %d, want 3", got)
	}
	if !hasKeyword(g, p1Creature, game.Vigilance) {
		t.Fatal("Player1 creature missing vigilance under Player1 Jetmir")
	}
	if got := effectivePower(g, p2CreatureA); got != 2 {
		t.Fatalf("Player2 creature power under Player1 Jetmir = %d, want 2", got)
	}
	if hasKeyword(g, p2CreatureA, game.Vigilance) {
		t.Fatal("Player2 creature has vigilance while Player1 controls Jetmir")
	}

	// Move Jetmir to Player2. Now Player1 controls 2 creatures (below three) and
	// Player2 controls Jetmir + 2 = 3 creatures (at the threshold).
	jetmir.Controller = game.Player2

	if got := effectivePower(g, p1Creature); got != 2 {
		t.Fatalf("Player1 creature power after control change = %d, want 2 (anthem left)", got)
	}
	if hasKeyword(g, p1Creature, game.Vigilance) {
		t.Fatal("Player1 creature kept vigilance after Jetmir changed controller")
	}
	if got := effectivePower(g, p2CreatureA); got != 3 {
		t.Fatalf("Player2 creature power after control change = %d, want 3 (anthem arrived)", got)
	}
	if !hasKeyword(g, p2CreatureA, game.Vigilance) {
		t.Fatal("Player2 creature missing vigilance after gaining Jetmir")
	}
	if got := effectivePower(g, jetmir); got != 4 {
		t.Fatalf("Jetmir power under Player2 = %d, want 4 (counts itself in group)", got)
	}
}
