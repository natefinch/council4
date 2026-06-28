package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func spiritTokenDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Spirit",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Sub("Spirit")},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}
}

// A CreateToken whose Recipient references a targeted player creates the token
// under that player's control and ownership, not the resolving controller's
// ("Target opponent creates a 1/1 colorless Spirit creature token.", Forbidden
// Orchard).
func TestCreateTokenTargetedPlayerRecipientControlsToken(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	obj := &game.StackObject{
		Controller: game.Player1,
		Targets:    []game.Target{{Kind: game.TargetPlayer, PlayerID: game.Player2}},
	}
	resolveInstruction(engine, g, obj, game.CreateToken{
		Amount:    game.Fixed(1),
		Source:    game.TokenDef(spiritTokenDef()),
		Recipient: opt.Val(game.TargetPlayerReference(0)),
	}, &TurnLog{})

	token := newlyCreatedToken(g)
	if token == nil {
		t.Fatal("targeted-recipient token did not enter the battlefield")
	}
	if token.Controller != game.Player2 {
		t.Fatalf("token controller = %v, want the targeted opponent Player2", token.Controller)
	}
	if token.Owner != game.Player2 {
		t.Fatalf("token owner = %v, want the targeted opponent Player2", token.Owner)
	}
}

// tokensByController counts the tokens each player controls on the battlefield.
func tokensByController(g *game.Game) map[game.PlayerID]int {
	counts := make(map[game.PlayerID]int)
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			counts[permanent.Controller]++
		}
	}
	return counts
}

// A CreateToken whose RecipientGroup is all players creates one token under each
// player's control ("Each player creates a 1/1 ... creature token.", Grismold,
// the Dreadsower).
func TestCreateTokenEachPlayerRecipientGroup(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	obj := &game.StackObject{Controller: game.Player1}
	resolveInstruction(engine, g, obj, game.CreateToken{
		Amount:         game.Fixed(1),
		Source:         game.TokenDef(spiritTokenDef()),
		RecipientGroup: game.AllPlayersReference(),
	}, &TurnLog{})

	counts := tokensByController(g)
	if len(counts) != game.NumPlayers {
		t.Fatalf("tokens created for %d players, want all %d", len(counts), game.NumPlayers)
	}
	for player, n := range counts {
		if n != 1 {
			t.Fatalf("player %v controls %d tokens, want 1", player, n)
		}
	}
}

// A CreateToken whose RecipientGroup is the controller's opponents creates one
// token under each opponent's control and none for the controller ("Each
// opponent creates a 1/1 white Human creature token.", Slaughter Specialist).
func TestCreateTokenEachOpponentRecipientGroup(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	obj := &game.StackObject{Controller: game.Player1}
	resolveInstruction(engine, g, obj, game.CreateToken{
		Amount:         game.Fixed(1),
		Source:         game.TokenDef(spiritTokenDef()),
		RecipientGroup: game.OpponentsReference(),
	}, &TurnLog{})

	counts := tokensByController(g)
	if counts[game.Player1] != 0 {
		t.Fatalf("controller Player1 controls %d tokens, want 0", counts[game.Player1])
	}
	if len(counts) != game.NumPlayers-1 {
		t.Fatalf("tokens created for %d players, want %d opponents", len(counts), game.NumPlayers-1)
	}
	for player, n := range counts {
		if n != 1 {
			t.Fatalf("opponent %v controls %d tokens, want 1", player, n)
		}
	}
}
