package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func punisherStackObject(g *game.Game) *game.StackObject {
	source := addCreaturePermanent(g, game.Player1)
	return &game.StackObject{
		ID:         g.IDGen.Next(),
		Controller: game.Player1,
		SourceID:   source.ObjectID,
	}
}

// TestPunisherEachLoseLifeNoAlternativeLosesLife proves that when the affected
// opponent can neither sacrifice nor discard, the punisher choice falls through
// to the life loss, as Hag of Ceaseless Torment relies on.
func TestPunisherEachLoseLifeNoAlternativeLosesLife(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	obj := punisherStackObject(g)

	resolveInstruction(engine, g, obj, game.PunisherEachLoseLife{
		PlayerGroup:        game.OpponentsReference(),
		Amount:             game.Fixed(3),
		AllowSacrifice:     true,
		SacrificeSelection: game.Selection{ExcludedTypes: []types.Card{types.Land}},
		AllowDiscard:       true,
	}, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != 37 {
		t.Fatalf("Player2 life = %d, want 37 (lost 3)", got)
	}
}

// TestPunisherEachLoseLifeSacrificeAvoidsLifeLoss proves that an opponent who
// chooses to sacrifice a permanent avoids the life loss, and the chosen
// permanent leaves the battlefield.
func TestPunisherEachLoseLifeSacrificeAvoidsLifeLoss(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	obj := punisherStackObject(g)
	victim := addCreaturePermanent(g, game.Player2)

	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{
		Primitive: game.PunisherEachLoseLife{
			PlayerGroup:        game.OpponentsReference(),
			Amount:             game.Fixed(3),
			AllowSacrifice:     true,
			SacrificeSelection: game.Selection{ExcludedTypes: []types.Card{types.Land}},
			AllowDiscard:       true,
		},
	}, agents, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != 40 {
		t.Fatalf("Player2 life = %d, want 40 (sacrificed instead)", got)
	}
	for _, permanent := range g.Battlefield {
		if permanent.ObjectID == victim.ObjectID {
			t.Fatal("victim permanent still on battlefield, want sacrificed")
		}
	}
}
