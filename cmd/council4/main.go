package main

import (
	"flag"
	"fmt"
	"math/rand/v2"
	"os"

	"github.com/natefinch/council4/mtg/agent"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/rules"
)

func main() {
	seed := flag.Uint64("seed", 1, "random seed")
	deckSize := flag.Int("deck-size", 8, "number of Forests in each test deck")
	verbose := flag.Bool("verbose", false, "print per-turn action log")
	noPass := flag.Bool("nopass", false, "omit pass actions from verbose log output")
	flag.Parse()

	if *deckSize < 0 {
		fmt.Fprintln(os.Stderr, "-deck-size must be non-negative")
		os.Exit(2)
	}

	engine := rules.NewEngine(rand.New(rand.NewPCG(*seed, *seed^0x9e3779b97f4a7c15)))
	gameState := engine.NewGame(landOnlyConfigs(*deckSize))
	agents := [game.NumPlayers]rules.PlayerAgent{
		game.Player1: agent.FirstLegal{},
		game.Player2: agent.FirstLegal{},
		game.Player3: agent.FirstLegal{},
		game.Player4: agent.FirstLegal{},
	}

	result := engine.RunGame(gameState, agents)
	printSummary(gameState, result, *seed, *deckSize)
	if *verbose {
		printTurnLog(gameState, result, logOptions{OmitPasses: *noPass})
	}
}

type logOptions struct {
	OmitPasses bool
}

func landOnlyConfigs(deckSize int) [game.NumPlayers]game.PlayerConfig {
	var configs [game.NumPlayers]game.PlayerConfig
	for player := range configs {
		configs[player].Name = playerName(game.PlayerID(player))
		for range deckSize {
			configs[player].Deck = append(configs[player].Deck, forest())
		}
	}
	return configs
}

func forest() *game.CardDef {
	return &game.CardDef{
		Name:  "Forest",
		Types: []game.CardType{game.TypeLand},
	}
}

func printSummary(g *game.Game, result *rules.GameResult, seed uint64, deckSize int) {
	fmt.Println("Council4 minimal test game")
	fmt.Printf("Seed: %d\n", seed)
	fmt.Printf("Deck size: %d Forests per player\n", deckSize)
	fmt.Printf("Turns: %d\n", result.TurnCount)
	if result.HasWinner {
		fmt.Printf("Winner: %s\n", playerName(result.Winner))
	} else {
		fmt.Println("Winner: none")
	}
	fmt.Printf("Battlefield permanents: %d\n", len(g.Battlefield))
	if len(result.Losses) > 0 {
		fmt.Println("Losses:")
		for _, loss := range result.Losses {
			fmt.Printf("  %s: %s\n", playerName(loss.Player), loss.Reason)
		}
	}
	fmt.Println()
	fmt.Println("Players:")
	for _, player := range g.Players {
		if player == nil {
			continue
		}
		fmt.Printf("  %s: life=%d hand=%d library=%d lands=%d eliminated=%t\n",
			playerName(player.ID),
			player.Life,
			player.Hand.Size(),
			player.Library.Size(),
			countLandsControlled(g, player.ID),
			player.Eliminated,
		)
	}
}

func printTurnLog(g *game.Game, result *rules.GameResult, opts logOptions) {
	fmt.Println()
	fmt.Println("Turn log:")
	for _, turn := range result.Turns {
		fmt.Printf("Turn %d (%s)\n", turn.TurnNumber, playerName(turn.ActivePlayer))
		for _, logged := range turn.Draws {
			fmt.Printf("  %s: %s\n", playerName(logged.Player), formatDraw(g, logged))
		}
		for _, logged := range turn.Losses {
			fmt.Printf("  %s: loses (%s)\n", playerName(logged.Player), logged.Reason)
		}
		for _, logged := range turn.Actions {
			if opts.OmitPasses && logged.Action.Kind == action.ActionPass {
				continue
			}
			fmt.Printf("  %s: %s\n", playerName(logged.Player), formatAction(g, logged.Action))
		}
	}
}

func formatDraw(g *game.Game, draw rules.DrawLog) string {
	if draw.Failed {
		return "draw from empty library"
	}
	card := g.GetCardInstance(draw.CardID)
	if card == nil || card.Def == nil {
		return fmt.Sprintf("draw card #%d", draw.CardID)
	}
	return fmt.Sprintf("draw %q", card.Def.Name)
}

func formatAction(g *game.Game, act action.Action) string {
	switch act.Kind {
	case action.ActionPass:
		return "pass"
	case action.ActionPlayLand:
		card := g.GetCardInstance(act.PlayLand.CardID)
		if card == nil || card.Def == nil {
			return fmt.Sprintf("play land #%d", act.PlayLand.CardID)
		}
		return fmt.Sprintf("play land %q", card.Def.Name)
	default:
		return fmt.Sprintf("action kind %d", act.Kind)
	}
}

func countLandsControlled(g *game.Game, playerID game.PlayerID) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if permanent != nil && permanent.Controller == playerID {
			count++
		}
	}
	return count
}

func playerName(playerID game.PlayerID) string {
	return fmt.Sprintf("Player %d", int(playerID)+1)
}
