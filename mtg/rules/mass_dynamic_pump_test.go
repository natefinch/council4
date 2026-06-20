package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// massCountPump models the Craterhoof Behemoth resolving effect: every creature
// you control gets +X/+X until end of turn where X is the number of creatures
// you control, and they gain trample.
func massCountPump() game.ApplyContinuous {
	count := opt.Val(game.DynamicAmount{
		Kind: game.DynamicAmountCountSelector,
		Group: game.BattlefieldGroup(game.Selection{
			Controller:    game.ControllerYou,
			RequiredTypes: []types.Card{types.Creature},
		}),
		Multiplier: 1,
	})
	group := game.BattlefieldGroup(game.Selection{
		Controller:    game.ControllerYou,
		RequiredTypes: []types.Card{types.Creature},
	})
	return game.ApplyContinuous{
		ContinuousEffects: []game.ContinuousEffect{
			{
				Layer:                 game.LayerPowerToughnessModify,
				Group:                 group,
				PowerDeltaDynamic:     count,
				ToughnessDeltaDynamic: count,
			},
			{
				Layer:       game.LayerAbility,
				Group:       group,
				AddKeywords: []game.Keyword{game.Trample},
			},
		},
		Duration: game.DurationUntilEndOfTurn,
	}
}

// TestMassDynamicPumpSnapshotsCountAtResolution verifies that the "+X/+X where X
// is the number of creatures you control" amount is locked at resolution: each
// of two controlled creatures gets +2/+2, a creature that enters afterward is
// unaffected, and the controller's later board changes do not alter the buff.
func TestMassDynamicPumpSnapshotsCountAtResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCombatPermanent(g, game.Player1, creatureWithPT("First", 2, 2))
	second := addCombatPermanent(g, game.Player1, creatureWithPT("Second", 3, 3))
	addCombatPermanent(g, game.Player2, creatureWithPT("Enemy", 4, 4))
	addEffectSpellToStack(g, game.Player1, massCountPump(), nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := effectivePower(g, first); got != 4 {
		t.Fatalf("first creature power = %d, want 4 (2 base + 2 controlled creatures)", got)
	}
	if got := effectivePower(g, second); got != 5 {
		t.Fatalf("second creature power = %d, want 5 (3 base + 2 controlled creatures)", got)
	}
	for _, effect := range g.ContinuousEffects {
		if effect.Layer != game.LayerPowerToughnessModify {
			continue
		}
		if effect.PowerDelta != 2 || effect.ToughnessDelta != 2 ||
			effect.PowerDeltaDynamic.Exists || effect.ToughnessDeltaDynamic.Exists {
			t.Fatalf("runtime effect = %#v, want snapshotted +2/+2", effect)
		}
	}
	if !hasKeyword(g, first, game.Trample) {
		t.Fatal("first creature should have gained trample")
	}

	later := addCombatPermanent(g, game.Player1, creatureWithPT("Later", 5, 5))
	if got := effectivePower(g, later); got != 5 {
		t.Fatalf("later entrant power = %d, want unaffected 5", got)
	}
	if got := effectivePower(g, first); got != 4 {
		t.Fatalf("first creature power after new entrant = %d, want still 4", got)
	}
}
