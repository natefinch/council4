// Command council4 runs deterministic test games.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"strings"

	"github.com/natefinch/council4/mtg/agent"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/opt"
)

func main() {
	seed := flag.Uint64("seed", 1, "random seed")
	deckSize := flag.Int("deck-size", 8, "number of Forests in each test deck")
	mode := flag.String("mode", "land", "test game mode: land, spells, or combat")
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
		_, _ = fmt.Fprintln(os.Stderr, "-deck-size must be non-negative")
		os.Exit(2)
	}

	configs, agents, err := gameModeConfig(*mode, *deckSize, deckSizeSet)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
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
			return [game.NumPlayers]game.PlayerConfig{}, [game.NumPlayers]rules.PlayerAgent{}, errors.New("-deck-size is only valid with -mode land")
		}
		return spellConfigs(), agents(agent.SimpleCaster{}), nil
	case "combat":
		if deckSizeSet {
			return [game.NumPlayers]game.PlayerConfig{}, [game.NumPlayers]rules.PlayerAgent{}, errors.New("-deck-size is only valid with -mode land")
		}
		return combatConfigs(), agents(agent.FirstLegal{}), nil
	default:
		return [game.NumPlayers]game.PlayerConfig{}, [game.NumPlayers]rules.PlayerAgent{}, errors.New("-mode must be land, spells, or combat")
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
		Types:    []types.Card{types.Land},
		Subtypes: []types.Sub{types.Forest},
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
			configs[player].Deck = append(configs[player].Deck, divinationLike(), healingSpell(), lavaSpikeLike())
		}
	}
	return configs
}

func combatConfigs() [game.NumPlayers]game.PlayerConfig {
	var configs [game.NumPlayers]game.PlayerConfig
	for player := range configs {
		configs[player].Name = playerName(game.PlayerID(player))
		for range 16 {
			configs[player].Deck = append(configs[player].Deck, forest())
		}
		for range 4 {
			configs[player].Deck = append(configs[player].Deck, trainedArmodon())
		}
		for range 3 {
			configs[player].Deck = append(configs[player].Deck, hastyWolf(), vigilantGuard())
		}
		for range 2 {
			configs[player].Deck = append(configs[player].Deck, wallOfVines())
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
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Bear},
		Power:     opt.Val(power),
		Toughness: opt.Val(toughness),
	}
}

func trainedArmodon() *game.CardDef {
	return creatureCard("Trained Armodon", 3, 3)
}

func hastyWolf() *game.CardDef {
	return creatureCard("Hasty Wolf", 2, 1, game.Haste)
}

func vigilantGuard() *game.CardDef {
	return creatureCard("Vigilant Guard", 2, 2, game.Vigilance)
}

func wallOfVines() *game.CardDef {
	return creatureCard("Wall of Vines", 0, 4, game.Defender)
}

func creatureCard(name string, power, toughness int, keywords ...game.Keyword) *game.CardDef {
	p := game.PT{Value: power}
	t := game.PT{Value: toughness}
	return &game.CardDef{
		Name:      name,
		ManaCost:  greenCost(),
		ManaValue: 1,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(p),
		Toughness: opt.Val(t),
		Abilities: []game.AbilityDef{
			{
				Kind:     game.StaticAbility,
				Keywords: keywords,
			},
		},
	}
}

func divinationLike() *game.CardDef {
	return &game.CardDef{
		Name:      "Simple Divination",
		ManaCost:  greenCost(),
		ManaValue: 1,
		Types:     []types.Card{types.Sorcery},
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
		Types:     []types.Card{types.Sorcery},
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
		Types:     []types.Card{types.Sorcery},
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

func greenCost() opt.V[mana.Cost] {
	cost := mana.Cost{mana.ColoredMana(mana.Green)}
	return opt.Val(cost)
}

func printSummary(g *game.Game, result *rules.GameResult, seed uint64, mode string, deckSize int) {
	fmt.Println("Council4 test game")
	fmt.Printf("Seed: %d\n", seed)
	fmt.Printf("Mode: %s\n", mode)
	switch mode {
	case "land":
		fmt.Printf("Deck size: %d Forests per player\n", deckSize)
	case "combat":
		fmt.Println("Deck: Forests and simple combat creatures")
	default:
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
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Turn log:")
	for i := range result.Turns {
		turn := &result.Turns[i]
		_, _ = fmt.Fprintf(w, "Turn %d (%s)\n", turn.TurnNumber, playerName(turn.ActivePlayer))
		if len(turn.Entries) > 0 {
			printTurnLogEntries(w, g, turn, opts)
			continue
		}
		for _, logged := range turn.Draws {
			_, _ = fmt.Fprintf(w, "  %s: %s\n", playerName(logged.Player), formatDraw(g, logged))
		}
		for i := range turn.Actions {
			logged := &turn.Actions[i]
			if opts.OmitPasses && logged.Action.Kind == action.ActionPass {
				continue
			}
			_, _ = fmt.Fprintf(w, "  %s: %s\n", playerName(logged.Player), formatActionLog(g, logged))
		}
		for _, logged := range turn.Resolves {
			_, _ = fmt.Fprintf(w, "  %s\n", formatResolve(g, logged))
		}
		for _, logged := range turn.CombatDamage {
			_, _ = fmt.Fprintf(w, "  %s\n", formatCombatDamage(g, logged))
		}
		for _, logged := range turn.CreatureDamage {
			_, _ = fmt.Fprintf(w, "  %s\n", formatCreatureDamage(g, logged))
		}
		for _, logged := range turn.Deaths {
			_, _ = fmt.Fprintf(w, "  %s\n", formatPermanentDeath(g, logged))
		}
		for _, logged := range turn.Losses {
			_, _ = fmt.Fprintf(w, "  %s: loses (%s)\n", playerName(logged.Player), logged.Reason)
		}
	}
}

func printTurnLogEntries(w io.Writer, g *game.Game, turn *rules.TurnLog, opts logOptions) {
	for i := range turn.Entries {
		entry := &turn.Entries[i]
		switch entry.Kind {
		case rules.TurnLogEntryDraw:
			_, _ = fmt.Fprintf(w, "  %s: %s\n", playerName(entry.Draw.Player), formatDraw(g, entry.Draw))
		case rules.TurnLogEntryLoss:
			_, _ = fmt.Fprintf(w, "  %s: loses (%s)\n", playerName(entry.Loss.Player), entry.Loss.Reason)
		case rules.TurnLogEntryAction:
			if opts.OmitPasses && entry.Action.Action.Kind == action.ActionPass {
				continue
			}
			_, _ = fmt.Fprintf(w, "  %s: %s\n", playerName(entry.Action.Player), formatActionLog(g, &entry.Action))
		case rules.TurnLogEntryResolve:
			_, _ = fmt.Fprintf(w, "  %s\n", formatResolve(g, entry.Resolve))
		case rules.TurnLogEntryCombatDamage:
			_, _ = fmt.Fprintf(w, "  %s\n", formatCombatDamage(g, entry.CombatDamage))
		case rules.TurnLogEntryCreatureDamage:
			_, _ = fmt.Fprintf(w, "  %s\n", formatCreatureDamage(g, entry.CreatureDamage))
		case rules.TurnLogEntryDeath:
			_, _ = fmt.Fprintf(w, "  %s\n", formatPermanentDeath(g, entry.Death))
		default:
		}
	}
}

func formatDraw(g *game.Game, draw rules.DrawLog) string {
	if draw.Failed {
		return "draw from empty library"
	}
	card, ok := g.GetCardInstance(draw.CardID)
	if !ok {
		return fmt.Sprintf("draw card #%d", draw.CardID)
	}
	return fmt.Sprintf("draw %q", card.Def.Name)
}

func formatAction(g *game.Game, act action.Action) string {
	return formatActionLog(g, &rules.ActionLog{Action: act})
}

func formatActionLog(g *game.Game, logged *rules.ActionLog) string {
	act := logged.Action
	switch act.Kind {
	case action.ActionPass:
		return "pass"
	case action.ActionPlayLand:
		playLand, payloadOK := act.PlayLandPayload()
		card, ok := g.GetCardInstance(playLand.CardID)
		if !ok {
			if !payloadOK {
				return "invalid play land"
			}
			return fmt.Sprintf("play land #%d", playLand.CardID)
		}
		return fmt.Sprintf("play land %q", card.Def.Name)
	case action.ActionCastSpell:
		cast, payloadOK := act.CastSpellPayload()
		card, ok := g.GetCardInstance(cast.CardID)
		if !ok {
			if !payloadOK {
				return "invalid cast spell"
			}
			return fmt.Sprintf("cast spell #%d", cast.CardID)
		}
		return fmt.Sprintf("cast %q", card.Def.Name)
	case action.ActionDeclareAttackers:
		attackers, ok := act.DeclareAttackersPayload()
		if !ok {
			return "invalid declare attackers"
		}
		return formatDeclareAttackers(g, logged, attackers)
	case action.ActionDeclareBlockers:
		blockers, ok := act.DeclareBlockersPayload()
		if !ok {
			return "invalid declare blockers"
		}
		return formatDeclareBlockers(g, logged, blockers)
	default:
		return fmt.Sprintf("action kind %d", act.Kind)
	}
}

func formatDeclareAttackers(g *game.Game, logged *rules.ActionLog, declare action.DeclareAttackersAction) string {
	if len(declare.Attackers) == 0 {
		return "declare no attackers"
	}
	parts := make([]string, 0, len(declare.Attackers))
	for _, declaration := range declare.Attackers {
		parts = append(parts, fmt.Sprintf("%s at %s", formatAttacker(g, logged, declaration), playerName(declaration.Target.Player)))
	}
	return "declare attackers: " + strings.Join(parts, ", ")
}

func formatAttacker(g *game.Game, logged *rules.ActionLog, declaration game.AttackDeclaration) string {
	return formatPermanentForAction(g, logged, declaration.Attacker)
}

func formatDeclareBlockers(g *game.Game, logged *rules.ActionLog, declare action.DeclareBlockersAction) string {
	if len(declare.Blockers) == 0 {
		return "declare no blockers"
	}
	parts := make([]string, 0, len(declare.Blockers))
	for _, declaration := range declare.Blockers {
		parts = append(parts, fmt.Sprintf("%s blocks %s",
			formatPermanentForAction(g, logged, declaration.Blocker),
			formatPermanentForAction(g, logged, declaration.Blocking),
		))
	}
	return "declare blockers: " + strings.Join(parts, ", ")
}

func formatPermanentForAction(g *game.Game, logged *rules.ActionLog, objectID id.ID) string {
	if cardID, ok := logged.PermanentSources[objectID]; ok {
		card, ok := g.GetCardInstance(cardID)
		if ok {
			return fmt.Sprintf("%q", card.Def.Name)
		}
	}
	if tokenName := logged.PermanentTokenNames[objectID]; tokenName != "" {
		return fmt.Sprintf("%q", tokenName)
	}
	return formatPermanent(g, objectID)
}

func formatPermanent(g *game.Game, objectID id.ID) string {
	for _, permanent := range g.Battlefield {
		if permanent == nil || permanent.ObjectID != objectID {
			continue
		}
		card, ok := g.GetCardInstance(permanent.CardInstanceID)
		if ok {
			return fmt.Sprintf("%q", card.Def.Name)
		}
	}
	return fmt.Sprintf("permanent #%d", objectID)
}

func formatResolve(g *game.Game, resolve rules.ResolveLog) string {
	card, ok := g.GetCardInstance(resolve.SourceID)
	if !ok {
		return fmt.Sprintf("resolve %s #%d", stackObjectKindName(resolve.Kind), resolve.SourceID)
	}
	if resolve.Result == "" || resolve.Result == "resolved" {
		return fmt.Sprintf("resolve %s %q", stackObjectKindName(resolve.Kind), card.Def.Name)
	}
	return fmt.Sprintf("resolve %s %q (%s)", stackObjectKindName(resolve.Kind), card.Def.Name, resolve.Result)
}

func formatCombatDamage(g *game.Game, damage rules.CombatDamageLog) string {
	card, ok := g.GetCardInstance(damage.SourceID)
	if !ok {
		return fmt.Sprintf("%s: permanent #%d deals %d combat damage to %s",
			playerName(damage.Controller),
			damage.Attacker,
			damage.Damage,
			playerName(damage.DefendingPlayer),
		)
	}
	return fmt.Sprintf("%s: %q deals %d combat damage to %s",
		playerName(damage.Controller),
		card.Def.Name,
		damage.Damage,
		playerName(damage.DefendingPlayer),
	)
}

func formatCreatureDamage(g *game.Game, damage rules.CreatureDamageLog) string {
	return fmt.Sprintf("%s: %s deals %d combat damage to %s",
		playerName(damage.Controller),
		formatPermanentOrCard(g, damage.SourcePermanent, damage.SourceID),
		damage.Damage,
		formatPermanentOrCard(g, damage.DamagedPermanent, damage.DamagedSourceID),
	)
}

func formatPermanentDeath(g *game.Game, death rules.PermanentDeathLog) string {
	return fmt.Sprintf("%s dies (%s)", formatPermanentOrCard(g, death.Permanent, death.SourceID), death.Reason)
}

func formatPermanentOrCard(g *game.Game, objectID, cardID id.ID) string {
	if formatted := formatPermanent(g, objectID); !strings.HasPrefix(formatted, "permanent #") {
		return formatted
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return fmt.Sprintf("permanent #%d", objectID)
	}
	return fmt.Sprintf("%q", card.Def.Name)
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
		card, ok := g.GetCardInstance(permanent.CardInstanceID)
		if ok && card.Def.HasType(types.Land) {
			count++
		}
	}
	return count
}

func playerName(playerID game.PlayerID) string {
	return fmt.Sprintf("Player %d", int(playerID)+1)
}
