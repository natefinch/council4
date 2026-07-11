package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestCullingRitualDestroysAndAddsCombinationMana proves the Culling Ritual
// runtime shape: the mass destroy removes every nonland permanent with mana value
// two or less (leaving lands and higher-cost permanents), publishes how many it
// destroyed, and the add-mana payoff hands the controller exactly that many mana
// split freely between {B} and {G} as chosen.
func TestCullingRitualDestroysAndAddsCombinationMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	// Four nonland permanents with mana value <= 2 across both players are
	// destroyed; the land and the mana-value-five permanent survive.
	destroyed := []*game.Permanent{
		addPermanentWithManaCost(g, game.Player1, 0),
		addPermanentWithManaCost(g, game.Player1, 2),
		addPermanentWithManaCost(g, game.Player2, 1),
		addPermanentWithManaCost(g, game.Player2, 2),
	}
	survivingLand := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Island",
		Types: []types.Card{types.Land},
	}})
	survivingBig := addPermanentWithManaCost(g, game.Player2, 5)

	addInstructionSpellToStack(g, []game.Instruction{
		{
			Primitive: game.Destroy{Group: game.BattlefieldGroup(game.Selection{
				ExcludedTypes: []types.Card{types.Land},
				ManaValue:     opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2}),
			})},
			PublishResult: game.ResultKey("destroyed-this-way"),
		},
		{
			Primitive: game.AddMana{
				Amount: game.Dynamic(game.DynamicAmount{
					Kind:       game.DynamicAmountPreviousEffectResult,
					Multiplier: 1,
					ResultKey:  game.ResultKey("destroyed-this-way"),
				}),
				CombinationColors: []mana.Color{mana.B, mana.G},
			},
		},
	})

	// The controller splits the four produced mana as two black and two green
	// (color index 0 is {B}, index 1 is {G}).
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0, 0, 1, 1}}},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	for _, permanent := range destroyed {
		if _, ok := permanentByObjectID(g, permanent.ObjectID); ok {
			t.Fatal("nonland permanent with mana value <= 2 survived the mass destroy")
		}
	}
	if _, ok := permanentByObjectID(g, survivingLand.ObjectID); !ok {
		t.Fatal("land was destroyed but should have been excluded")
	}
	if _, ok := permanentByObjectID(g, survivingBig.ObjectID); !ok {
		t.Fatal("mana-value-five permanent was destroyed but should have survived")
	}

	pool := g.Players[game.Player1].ManaPool
	if got := pool.Total(); got != len(destroyed) {
		t.Fatalf("total mana = %d, want %d (one per destroyed permanent)", got, len(destroyed))
	}
	if got := pool.Amount(mana.B); got != 2 {
		t.Fatalf("black mana = %d, want 2", got)
	}
	if got := pool.Amount(mana.G); got != 2 {
		t.Fatalf("green mana = %d, want 2", got)
	}
	if got := pool.Amount(mana.W) + pool.Amount(mana.U) + pool.Amount(mana.R) + pool.Amount(mana.C); got != 0 {
		t.Fatalf("off-color mana = %d, want 0 (only {B}/{G} may be produced)", got)
	}
}
