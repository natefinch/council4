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
