package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestPopulateCopiesChosenControlledCreatureToken(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	tokenDef := &game.CardDef{CardFace: game.CardFace{
		Name:      "Saproling",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Saproling},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}
	if _, ok := createTokenPermanent(g, game.Player1, tokenDef); !ok {
		t.Fatal("failed to create source token")
	}
	addCombatCreaturePermanent(g, game.Player1)
	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1}, game.CreateToken{
		Amount: game.Fixed(1),
		Source: game.TokenCopyOf(game.TokenCopySpec{
			Source: game.TokenCopySourceChosenControlledCreatureToken,
		}),
	}, &TurnLog{})
	if got := countTokenPermanentsNamed(g, "Saproling"); got != 2 {
		t.Fatalf("Saproling tokens = %d, want 2 after populate", got)
	}
}
