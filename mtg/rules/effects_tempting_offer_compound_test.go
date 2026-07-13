package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func rabbitTokenDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Rabbit",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Rabbit},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}
}

// temptingOfferCompoundInstruction models the Tempt with Bunnies idiom: the
// acting player draws a card and creates a 1/1 white Rabbit token as one shared
// multi-primitive body. Both primitives address the acting player through
// GroupOfferMemberReference(), and the body runs once for the controller base,
// once per accepting opponent for that opponent, and once more per accepter for
// the controller reward.
func temptingOfferCompoundInstruction() game.Instruction {
	return game.Instruction{
		Optional:           true,
		OptionalActorGroup: opt.Val(game.OpponentsReference()),
		TemptingOffer:      true,
		TemptingOfferBody: []game.Instruction{
			{Primitive: game.Draw{
				Player: game.GroupOfferMemberReference(),
				Amount: game.Fixed(1),
			}},
			{Primitive: game.CreateToken{
				Amount:    game.Fixed(1),
				Source:    game.TokenDef(rabbitTokenDef()),
				Recipient: opt.Val(game.GroupOfferMemberReference()),
			}},
		},
	}
}

// TestTemptingOfferCompoundDrawAndToken proves the compound-body idiom (Tempt
// with Bunnies) runs the whole draw-and-create-token body for each acting player:
// the controller draws and creates once for the base plus once per accepting
// opponent, each accepting opponent draws and creates once for themselves, and a
// declining opponent does neither.
func TestTemptingOfferCompoundDrawAndToken(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// The controller draws once for the base plus once per accepting opponent
	// (two here); each accepter draws once. Stock enough cards to draw.
	stockLibrary(g, game.Player1, 3)
	stockLibrary(g, game.Player2, 1)
	stockLibrary(g, game.Player4, 1)
	addInstructionSpellToStack(g, []game.Instruction{temptingOfferCompoundInstruction()})

	// Player2 and Player4 accept; Player3 declines.
	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &choiceOnlyAgent{choices: [][]int{{1}}},
		game.Player3: &choiceOnlyAgent{choices: [][]int{{0}}},
		game.Player4: &choiceOnlyAgent{choices: [][]int{{1}}},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	tokens := tokensByController(g)
	// Controller: one base token plus one reward token per accepting opponent (2).
	if tokens[game.Player1] != 3 {
		t.Fatalf("controller controls %d Rabbit tokens, want 3 (1 base + 2 rewards)", tokens[game.Player1])
	}
	if tokens[game.Player2] != 1 || tokens[game.Player4] != 1 {
		t.Fatalf("accepter tokens = %d/%d, want 1/1", tokens[game.Player2], tokens[game.Player4])
	}
	if tokens[game.Player3] != 0 {
		t.Fatalf("decliner controls %d tokens, want 0", tokens[game.Player3])
	}

	// Draws mirror the token counts: the controller drew three, each accepter one,
	// the decliner none.
	if got := g.Players[game.Player1].Hand.Size(); got != 3 {
		t.Fatalf("controller drew to hand size %d, want 3 (1 base + 2 rewards)", got)
	}
	if got := g.Players[game.Player2].Hand.Size(); got != 1 {
		t.Fatalf("Player2 hand size %d, want 1 (accepted)", got)
	}
	if got := g.Players[game.Player4].Hand.Size(); got != 1 {
		t.Fatalf("Player4 hand size %d, want 1 (accepted)", got)
	}
	if got := g.Players[game.Player3].Hand.Size(); got != 0 {
		t.Fatalf("Player3 hand size %d, want 0 (declined)", got)
	}
}

// TestTemptingOfferCompoundAllDecline proves the compound body runs only the
// controller's base draw and token when no opponent accepts.
func TestTemptingOfferCompoundAllDecline(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	stockLibrary(g, game.Player1, 3)
	addInstructionSpellToStack(g, []game.Instruction{temptingOfferCompoundInstruction()})

	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &choiceOnlyAgent{choices: [][]int{{0}}},
		game.Player3: &choiceOnlyAgent{choices: [][]int{{0}}},
		game.Player4: &choiceOnlyAgent{choices: [][]int{{0}}},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	tokens := tokensByController(g)
	if tokens[game.Player1] != 1 {
		t.Fatalf("controller controls %d Rabbit tokens, want 1 (base only)", tokens[game.Player1])
	}
	for _, pid := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if tokens[pid] != 0 {
			t.Fatalf("declining opponent %v controls %d tokens, want 0", pid, tokens[pid])
		}
	}
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("controller hand size %d, want 1 (base only)", got)
	}
}
