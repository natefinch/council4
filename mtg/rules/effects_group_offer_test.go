package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// groupOfferInstructions models "Any player may have <source> deal 5 damage to
// them. If no one does, <controller loses 3 life>. If a player does, <controller
// gains 3 life>." — a multiplayer "may have" offer whose consequence branches on
// whether at least one player accepted.
func groupOfferInstructions(group game.PlayerGroupReference) []game.Instruction {
	return []game.Instruction{
		{
			Primitive: game.Damage{
				Amount:    game.Fixed(5),
				Recipient: game.PlayerDamageRecipient(game.GroupOfferMemberReference()),
			},
			Optional:           true,
			OptionalActorGroup: opt.Val(group),
			PublishResult:      "offer",
		},
		{
			Primitive:  game.LoseLife{Player: game.ControllerReference(), Amount: game.Fixed(3)},
			ResultGate: opt.Val(game.InstructionResultGate{Key: "offer", Accepted: game.TriFalse}),
		},
		{
			Primitive:  game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(3)},
			ResultGate: opt.Val(game.InstructionResultGate{Key: "offer", Accepted: game.TriTrue}),
		},
	}
}

func TestGroupOfferDealsDamageToEachAcceptingPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addInstructionSpellToStack(g, groupOfferInstructions(game.AllPlayersReference()))

	// Player1 and Player3 accept; Player2 and Player4 decline.
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
		game.Player2: &choiceOnlyAgent{choices: [][]int{{0}}},
		game.Player3: &choiceOnlyAgent{choices: [][]int{{1}}},
		game.Player4: &choiceOnlyAgent{choices: [][]int{{0}}},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != 38 {
		t.Fatalf("Player1 life = %d, want 38 (accepter took 5, controller gained 3)", got)
	}
	if got := g.Players[game.Player3].Life; got != 35 {
		t.Fatalf("Player3 life = %d, want 35 (accepted, took 5)", got)
	}
	if got := g.Players[game.Player2].Life; got != 40 {
		t.Fatalf("Player2 life = %d, want 40 (declined)", got)
	}
	if got := g.Players[game.Player4].Life; got != 40 {
		t.Fatalf("Player4 life = %d, want 40 (declined)", got)
	}
}

func TestGroupOfferNoneAcceptFiresIfNoOneDoesBranch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// Offer opponents only so the controller is not asked; every opponent
	// declines.
	addInstructionSpellToStack(g, groupOfferInstructions(game.OpponentsReference()))

	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &choiceOnlyAgent{choices: [][]int{{0}}},
		game.Player3: &choiceOnlyAgent{choices: [][]int{{0}}},
		game.Player4: &choiceOnlyAgent{choices: [][]int{{0}}},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	for _, pid := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if got := g.Players[pid].Life; got != 40 {
			t.Fatalf("Player %d life = %d, want 40 (no damage, none accepted)", pid, got)
		}
	}
	// No opponent accepted, so the controller loses 3 (37) and does not gain.
	if got := g.Players[game.Player1].Life; got != 37 {
		t.Fatalf("controller life = %d, want 37 (lose-3 if-no-one-does branch)", got)
	}
}

func TestGroupOfferSomeAcceptFiresIfAPlayerDoesBranch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addInstructionSpellToStack(g, groupOfferInstructions(game.OpponentsReference()))

	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &choiceOnlyAgent{choices: [][]int{{0}}},
		game.Player3: &choiceOnlyAgent{choices: [][]int{{1}}},
		game.Player4: &choiceOnlyAgent{choices: [][]int{{0}}},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Players[game.Player3].Life; got != 35 {
		t.Fatalf("Player3 life = %d, want 35 (accepted, took 5)", got)
	}
	// A player accepted, so the controller gains 3 (43) and does not lose.
	if got := g.Players[game.Player1].Life; got != 43 {
		t.Fatalf("controller life = %d, want 43 (gain-3 if-a-player-does branch)", got)
	}
}
