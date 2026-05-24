package main

import (
	"bytes"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/rules"
)

func TestSpellModeRunsDeterministicallyWithCastsAndResolves(t *testing.T) {
	first := runSpellMode(1)
	second := runSpellMode(1)

	if first.TurnCount == 0 {
		t.Fatal("spell mode produced zero turns")
	}
	if first.TurnCount != second.TurnCount {
		t.Fatalf("turn count differs: %d != %d", first.TurnCount, second.TurnCount)
	}
	if first.HasWinner != second.HasWinner || first.Winner != second.Winner {
		t.Fatalf("winner differs: (%v,%v) != (%v,%v)", first.HasWinner, first.Winner, second.HasWinner, second.Winner)
	}
	casts, resolves := countCastsAndResolves(first)
	if casts == 0 {
		t.Fatal("spell mode produced no casts")
	}
	if resolves == 0 {
		t.Fatal("spell mode produced no resolves")
	}
}

func TestPrintTurnLogIncludesCastAndResolve(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addTestCard(g, game.Player1, &game.CardDef{Name: "Test Spell"})
	result := &rules.GameResult{
		Turns: []rules.TurnLog{
			{
				TurnNumber:   1,
				ActivePlayer: game.Player1,
				Actions: []rules.ActionLog{
					{Player: game.Player1, Action: action.CastSpell(cardID, nil, 0, nil)},
				},
				Resolves: []rules.ResolveLog{
					{SourceID: cardID, Controller: game.Player1, Kind: game.StackSpell, Result: "resolved"},
				},
			},
		},
	}
	var out bytes.Buffer

	printTurnLog(&out, g, result, logOptions{})

	got := out.String()
	for _, want := range []string{`Player 1: cast "Test Spell"`, `resolve spell "Test Spell"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q:\n%s", want, got)
		}
	}
}

func TestPrintTurnLogNoPassKeepsOtherEvents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addTestCard(g, game.Player1, &game.CardDef{Name: "Test Spell"})
	drawID := addTestCard(g, game.Player1, &game.CardDef{Name: "Drawn Card"})
	result := &rules.GameResult{
		Turns: []rules.TurnLog{
			{
				TurnNumber:   1,
				ActivePlayer: game.Player1,
				Draws: []rules.DrawLog{
					{Player: game.Player1, CardID: drawID},
				},
				Losses: []rules.LossLog{
					{Player: game.Player2, Reason: rules.LossReasonZeroLife},
				},
				Actions: []rules.ActionLog{
					{Player: game.Player1, Action: action.CastSpell(cardID, nil, 0, nil)},
					{Player: game.Player2, Action: action.Pass()},
				},
				Resolves: []rules.ResolveLog{
					{SourceID: cardID, Controller: game.Player1, Kind: game.StackSpell, Result: "graveyard"},
				},
			},
		},
	}
	var out bytes.Buffer

	printTurnLog(&out, g, result, logOptions{OmitPasses: true})

	got := out.String()
	if strings.Contains(got, "Player 2: pass") {
		t.Fatalf("output included pass action:\n%s", got)
	}
	for _, want := range []string{`draw "Drawn Card"`, `loses (0 life)`, `cast "Test Spell"`, `resolve spell "Test Spell" (graveyard)`} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q:\n%s", want, got)
		}
	}
}

func addTestCard(g *game.Game, owner game.PlayerID, def *game.CardDef) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Def:   def,
		Owner: owner,
	}
	return cardID
}

func runSpellMode(seed uint64) *rules.GameResult {
	configs, agents, err := gameModeConfig("spells", 8, false)
	if err != nil {
		panic(err)
	}
	engine := rules.NewEngine(rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15)))
	return engine.RunGame(engine.NewGame(configs), agents)
}

func countCastsAndResolves(result *rules.GameResult) (int, int) {
	casts := 0
	resolves := 0
	for _, turn := range result.Turns {
		for _, logged := range turn.Actions {
			if logged.Action.Kind == action.ActionCastSpell {
				casts++
			}
		}
		resolves += len(turn.Resolves)
	}
	return casts, resolves
}
