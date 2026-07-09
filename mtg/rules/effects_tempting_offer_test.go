package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// temptingOfferTokenInstruction models the Tempt with Vengeance idiom: the
// controller creates a token, each opponent may create the same token for
// themselves, and for each opponent who does the controller creates another
// token. The acting player is addressed through GroupOfferMemberReference(),
// which resolveTemptingOffer binds to the controller for the base and reward
// creations and to each accepting opponent for their own creation.
func temptingOfferTokenInstruction() game.Instruction {
	return game.Instruction{
		Primitive: game.CreateToken{
			Amount:    game.Fixed(1),
			Source:    game.TokenDef(spiritTokenDef()),
			Recipient: opt.Val(game.GroupOfferMemberReference()),
		},
		Optional:           true,
		OptionalActorGroup: opt.Val(game.OpponentsReference()),
		TemptingOffer:      true,
	}
}

// When some opponents accept the tempting offer, the controller creates the base
// token plus one additional token for each accepting opponent, and each accepting
// opponent creates one token for themselves.
func TestTemptingOfferControllerRepeatsPerAcceptingOpponent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addInstructionSpellToStack(g, []game.Instruction{temptingOfferTokenInstruction()})

	// Player2 and Player3 accept; Player4 declines.
	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &choiceOnlyAgent{choices: [][]int{{1}}},
		game.Player3: &choiceOnlyAgent{choices: [][]int{{1}}},
		game.Player4: &choiceOnlyAgent{choices: [][]int{{0}}},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	counts := tokensByController(g)
	// Controller: one base token plus one reward per accepting opponent (2) = 3.
	if counts[game.Player1] != 3 {
		t.Fatalf("controller Player1 controls %d tokens, want 3 (1 base + 2 rewards)", counts[game.Player1])
	}
	if counts[game.Player2] != 1 {
		t.Fatalf("Player2 controls %d tokens, want 1 (accepted)", counts[game.Player2])
	}
	if counts[game.Player3] != 1 {
		t.Fatalf("Player3 controls %d tokens, want 1 (accepted)", counts[game.Player3])
	}
	if counts[game.Player4] != 0 {
		t.Fatalf("Player4 controls %d tokens, want 0 (declined)", counts[game.Player4])
	}
}

// When no opponent accepts the tempting offer, the controller creates only the
// base token and no rewards, and no opponent creates anything.
func TestTemptingOfferNoAcceptorsCreatesOnlyBase(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addInstructionSpellToStack(g, []game.Instruction{temptingOfferTokenInstruction()})

	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &choiceOnlyAgent{choices: [][]int{{0}}},
		game.Player3: &choiceOnlyAgent{choices: [][]int{{0}}},
		game.Player4: &choiceOnlyAgent{choices: [][]int{{0}}},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	counts := tokensByController(g)
	if counts[game.Player1] != 1 {
		t.Fatalf("controller Player1 controls %d tokens, want 1 (base only, no acceptors)", counts[game.Player1])
	}
	for _, pid := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if counts[pid] != 0 {
			t.Fatalf("opponent %v controls %d tokens, want 0 (declined)", pid, counts[pid])
		}
	}
}

// When every opponent accepts, the controller creates the base plus one reward
// per opponent and each opponent creates one for themselves.
func TestTemptingOfferAllAcceptorsRewardEach(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addInstructionSpellToStack(g, []game.Instruction{temptingOfferTokenInstruction()})

	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &choiceOnlyAgent{choices: [][]int{{1}}},
		game.Player3: &choiceOnlyAgent{choices: [][]int{{1}}},
		game.Player4: &choiceOnlyAgent{choices: [][]int{{1}}},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	counts := tokensByController(g)
	// One base plus three rewards for the controller; one each for the opponents.
	if counts[game.Player1] != 4 {
		t.Fatalf("controller Player1 controls %d tokens, want 4 (1 base + 3 rewards)", counts[game.Player1])
	}
	for _, pid := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if counts[pid] != 1 {
			t.Fatalf("opponent %v controls %d tokens, want 1 (accepted)", pid, counts[pid])
		}
	}
}
