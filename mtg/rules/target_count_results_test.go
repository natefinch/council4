package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestCompiledLeadingTokenCountPreservesTypedSelection(t *testing.T) {
	def := compileTargetCountSpell(t, "For each token you control, create a 1/1 white Rabbit creature token.")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for range 2 {
		addCreaturePermanent(g, game.Player1)
		token := addCreaturePermanent(g, game.Player1)
		token.Token, token.TokenDef = true, g.CardInstances[token.CardInstanceID].Def
	}
	opponentToken := addCreaturePermanent(g, game.Player2)
	opponentToken.Token, opponentToken.TokenDef = true, g.CardInstances[opponentToken.CardInstanceID].Def
	spell := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].Hand.Remove(spell)
	g.Stack.Push(&game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: spell, Controller: game.Player1,
	})
	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
	tokens := 0
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.Controller == game.Player1 {
			tokens++
		}
	}
	if tokens != 4 {
		t.Fatalf("controlled tokens=%d, want two originals and two created; non-tokens and opponents must not count", tokens)
	}
}
