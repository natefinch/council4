package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// beastmasterAscensionPermanent mirrors the lowered shape of Beastmaster
// Ascension's "As long as Beastmaster Ascension has seven or more quest counters
// on it, creatures you control get +5/+5.": a StaticAbility whose Condition
// inspects the source's own quest-counter count and whose continuous effect
// pumps the controller's creatures only while that threshold is met.
func beastmasterAscensionPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Beastmaster Ascension",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			Condition: opt.Val(game.Condition{
				Object: opt.Val(game.SourcePermanentReference()),
				ObjectMatches: opt.Val(game.Selection{
					RequiredCounter:      counter.Quest,
					RequiredCounterCount: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 7}),
				}),
			}),
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerPowerToughnessModify,
				Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{
					RequiredTypes: []types.Card{types.Creature},
				}),
				PowerDelta:     5,
				ToughnessDelta: 5,
			}},
		}},
	}})
}

// TestCounterThresholdStaticAppliesOnlyAtThreshold proves the Beastmaster
// Ascension anthem pumps the controller's creatures only once the enchantment
// accumulates the required number of quest counters.
func TestCounterThresholdStaticAppliesOnlyAtThreshold(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	ascension := beastmasterAscensionPermanent(g, game.Player1)

	powerOf := func() int {
		for _, view := range observe(g, game.Player1).Battlefield() {
			if view.Name == "Powered Combat Creature" {
				return view.Power
			}
		}
		t.Fatal("creature not found on battlefield")
		return 0
	}

	if got := powerOf(); got != 2 {
		t.Fatalf("power without quest counters = %d, want 2", got)
	}

	ascension.Counters.Add(counter.Quest, 6)
	if got := powerOf(); got != 2 {
		t.Fatalf("power with 6 quest counters = %d, want 2 (below threshold)", got)
	}

	ascension.Counters.Add(counter.Quest, 1)
	if got := powerOf(); got != 7 {
		t.Fatalf("power with 7 quest counters = %d, want 7 (+5/+5 applies)", got)
	}
}
