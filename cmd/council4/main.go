package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand/v2"
	"os"

	"github.com/natefinch/council4/mtg/agent"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/rules"
)

func main() {
	seed := flag.Uint64("seed", 1, "random seed")
	deckSize := flag.Int("deck-size", 8, "number of Forests in each test deck")
	mode := flag.String("mode", "land", "test game mode: land or spells")
	verbose := flag.Bool("verbose", false, "print per-turn action log")
	noPass := flag.Bool("nopass", false, "omit pass actions from verbose log output")
	flag.Parse()

	deckSizeSet := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "deck-size" {
			deckSizeSet = true
		}
	})

	if *deckSize < 0 {
		fmt.Fprintln(os.Stderr, "-deck-size must be non-negative")
		os.Exit(2)
	}

	configs, agents, err := gameModeConfig(*mode, *deckSize, deckSizeSet)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	engine := rules.NewEngine(rand.New(rand.NewPCG(*seed, *seed^0x9e3779b97f4a7c15)))
	gameState := engine.NewGame(configs)

	result := engine.RunGame(gameState, agents)
	printSummary(gameState, result, *seed, *mode, *deckSize)
	if *verbose {
		printTurnLog(os.Stdout, gameState, result, logOptions{OmitPasses: *noPass})
	}
}

type logOptions struct {
	OmitPasses bool
}

func gameModeConfig(mode string, deckSize int, deckSizeSet bool) ([game.NumPlayers]game.PlayerConfig, [game.NumPlayers]rules.PlayerAgent, error) {
	switch mode {
	case "land":
		return landOnlyConfigs(deckSize), agents(agent.FirstLegal{}), nil
	case "spells":
		if deckSizeSet {
			return [game.NumPlayers]game.PlayerConfig{}, [game.NumPlayers]rules.PlayerAgent{}, fmt.Errorf("-deck-size is only valid with -mode land")
		}
		return spellConfigs(), agents(agent.SimpleCaster{}), nil
	default:
		return [game.NumPlayers]game.PlayerConfig{}, [game.NumPlayers]rules.PlayerAgent{}, fmt.Errorf("-mode must be land or spells")
	}
}

func agents(playerAgent rules.PlayerAgent) [game.NumPlayers]rules.PlayerAgent {
	return [game.NumPlayers]rules.PlayerAgent{
		game.Player1: playerAgent,
		game.Player2: playerAgent,
		game.Player3: playerAgent,
		game.Player4: playerAgent,
	}
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
		Name:     "Forest",
		Types:    []game.CardType{game.TypeLand},
		Subtypes: []string{"Forest"},
	}
}

func spellConfigs() [game.NumPlayers]game.PlayerConfig {
	var configs [game.NumPlayers]game.PlayerConfig
	for player := range configs {
		configs[player].Name = playerName(game.PlayerID(player))
		for range 12 {
			configs[player].Deck = append(configs[player].Deck, forest())
		}
		for range 4 {
			configs[player].Deck = append(configs[player].Deck, grizzlyBears())
		}
		for range 2 {
			configs[player].Deck = append(configs[player].Deck, divinationLike())
			configs[player].Deck = append(configs[player].Deck, healingSpell())
			configs[player].Deck = append(configs[player].Deck, lavaSpikeLike())
		}
	}
	return configs
}

func grizzlyBears() *game.CardDef {
	power := game.PT{Value: 2}
	toughness := game.PT{Value: 2}
	return &game.CardDef{
		Name:      "Grizzly Bears",
		ManaCost:  greenCost(),
		ManaValue: 1,
		Types:     []game.CardType{game.TypeCreature},
		Subtypes:  []string{"Bear"},
		Power:     &power,
		Toughness: &toughness,
	}
}

func divinationLike() *game.CardDef {
	return &game.CardDef{
		Name:      "Simple Divination",
		ManaCost:  greenCost(),
		ManaValue: 1,
		Types:     []game.CardType{game.TypeSorcery},
		Abilities: []game.AbilityDef{
			{
				Kind: game.SpellAbility,
				Effects: []game.Effect{
					{Type: game.EffectDraw, Amount: 1, TargetIndex: -1},
				},
			},
		},
	}
}

func healingSpell() *game.CardDef {
	return &game.CardDef{
		Name:      "Simple Healing",
		ManaCost:  greenCost(),
		ManaValue: 1,
		Types:     []game.CardType{game.TypeSorcery},
		Abilities: []game.AbilityDef{
			{
				Kind: game.SpellAbility,
				Effects: []game.Effect{
					{Type: game.EffectGainLife, Amount: 3, TargetIndex: -1},
				},
			},
		},
	}
}

func lavaSpikeLike() *game.CardDef {
	return &game.CardDef{
		Name:      "Simple Lava Spike",
		ManaCost:  greenCost(),
		ManaValue: 1,
		Types:     []game.CardType{game.TypeSorcery},
		Abilities: []game.AbilityDef{
			{
				Kind: game.SpellAbility,
				Targets: []game.TargetSpec{
					{MinTargets: 1, MaxTargets: 1, Constraint: "player"},
				},
				Effects: []game.Effect{
					{Type: game.EffectDamage, Amount: 3, TargetIndex: 0},
				},
			},
		},
	}
}

func greenCost() *mana.Cost {
	cost := mana.Cost{mana.ColoredMana(mana.Green)}
	return &cost
}

func printSummary(g *game.Game, result *rules.GameResult, seed uint64, mode string, deckSize int) {
	fmt.Println("Council4 test game")
	fmt.Printf("Seed: %d\n", seed)
	fmt.Printf("Mode: %s\n", mode)
	if mode == "land" {
		fmt.Printf("Deck size: %d Forests per player\n", deckSize)
	} else {
		fmt.Println("Deck: Forests, simple creatures, and simple spells")
	}
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

func printTurnLog(w io.Writer, g *game.Game, result *rules.GameResult, opts logOptions) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Turn log:")
	for _, turn := range result.Turns {
		fmt.Fprintf(w, "Turn %d (%s)\n", turn.TurnNumber, playerName(turn.ActivePlayer))
		for _, logged := range turn.Draws {
			fmt.Fprintf(w, "  %s: %s\n", playerName(logged.Player), formatDraw(g, logged))
		}
		for _, logged := range turn.Losses {
			fmt.Fprintf(w, "  %s: loses (%s)\n", playerName(logged.Player), logged.Reason)
		}
		for _, logged := range turn.Actions {
			if opts.OmitPasses && logged.Action.Kind == action.ActionPass {
				continue
			}
			fmt.Fprintf(w, "  %s: %s\n", playerName(logged.Player), formatAction(g, logged.Action))
		}
		for _, logged := range turn.Resolves {
			fmt.Fprintf(w, "  %s\n", formatResolve(g, logged))
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
	case action.ActionCastSpell:
		card := g.GetCardInstance(act.CastSpell.CardID)
		if card == nil || card.Def == nil {
			return fmt.Sprintf("cast spell #%d", act.CastSpell.CardID)
		}
		return fmt.Sprintf("cast %q", card.Def.Name)
	default:
		return fmt.Sprintf("action kind %d", act.Kind)
	}
}

func formatResolve(g *game.Game, resolve rules.ResolveLog) string {
	card := g.GetCardInstance(resolve.SourceID)
	if card == nil || card.Def == nil {
		return fmt.Sprintf("resolve %s #%d", stackObjectKindName(resolve.Kind), resolve.SourceID)
	}
	if resolve.Result == "" || resolve.Result == "resolved" {
		return fmt.Sprintf("resolve %s %q", stackObjectKindName(resolve.Kind), card.Def.Name)
	}
	return fmt.Sprintf("resolve %s %q (%s)", stackObjectKindName(resolve.Kind), card.Def.Name, resolve.Result)
}

func stackObjectKindName(kind game.StackObjectKind) string {
	switch kind {
	case game.StackSpell:
		return "spell"
	case game.StackActivatedAbility:
		return "activated ability"
	case game.StackTriggeredAbility:
		return "triggered ability"
	default:
		return "stack object"
	}
}

func countLandsControlled(g *game.Game, playerID game.PlayerID) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if permanent == nil || permanent.Controller != playerID {
			continue
		}
		card := g.GetCardInstance(permanent.CardInstanceID)
		if card != nil && card.Def != nil && card.Def.HasType(game.TypeLand) {
			count++
		}
	}
	return count
}

func playerName(playerID game.PlayerID) string {
	return fmt.Sprintf("Player %d", int(playerID)+1)
}
