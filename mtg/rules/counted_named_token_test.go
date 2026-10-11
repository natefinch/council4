package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompiledNamedTokenCountDoesNotCountNontokenSubtype(t *testing.T) {
	for _, body := range []string{
		"For each Blood token you control, create a 1/1 white Rabbit creature token.",
		"Create X 1/1 white Rabbit creature tokens, where X is the number of Blood tokens you control.",
	} {
		t.Run(body, func(t *testing.T) {
			testNamedTokenCount(t, body)
		})
	}
}

func testNamedTokenCount(t *testing.T, body string) {
	t.Helper()
	def := compileTargetCountSpell(t, body)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for _, fixture := range []struct {
		controller game.PlayerID
		token      bool
		subtype    types.Sub
	}{
		{game.Player1, true, "Blood"}, {game.Player1, true, "Blood"},
		{game.Player1, false, "Blood"}, {game.Player1, true, "Clue"},
		{game.Player2, true, "Blood"},
	} {
		permanent := addCombatPermanent(g, fixture.controller, &game.CardDef{CardFace: game.CardFace{
			Name: "Count Subject", Types: []types.Card{types.Artifact}, Subtypes: []types.Sub{fixture.subtype},
		}})
		permanent.Token = fixture.token
		if fixture.token {
			permanent.TokenDef = g.CardInstances[permanent.CardInstanceID].Def
		}
	}
	spell := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].Hand.Remove(spell)
	g.Stack.Push(&game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: spell, Controller: game.Player1})
	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
	created := 0
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.TokenDef.Name != "Count Subject" && permanent.Controller == game.Player1 {
			created++
		}
	}
	if created != 2 {
		t.Fatalf("created tokens=%d, want only the two controlled Blood tokens", created)
	}
}
