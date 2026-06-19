package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// benchForest is a basic Forest card definition that taps for green mana, like
// the real registry card. Payment activates the mana ability on demand; it is
// not exposed to the agent as a standalone strategic action.
func benchForest() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Forest",
		Supertypes: []types.Super{types.Basic},
		Types:      []types.Card{types.Land},
		Subtypes:   []types.Sub{types.Forest},
		ManaAbilities: []game.ManaAbility{
			game.TapManaAbility(mana.G),
		},
	}}
}

// benchCommander is a minimal legendary creature usable as a commander.
func benchCommander() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Bench Commander",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Power:      opt.Val(game.PT{Value: 3}),
		Toughness:  opt.Val(game.PT{Value: 3}),
	}}
}

// benchHasteBeater is a cheap haste, flying creature so a FirstLegal agent
// attacks with it and games end via combat damage in a realistic number of
// turns, instead of running to deck-out.
func benchHasteBeater() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Bench Beater",
		Types:     []types.Card{types.Creature},
		ManaCost:  opt.Val(cost.Mana{cost.G}),
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
			game.HasteStaticBody,
		},
	}}
}

// benchRealisticConfigs builds four decks of lands plus cheap evasive beaters,
// approximating a real aggressive deck: games end via combat in a modest number
// of turns with a modest board, the common case for playtesting.
func benchRealisticConfigs() [game.NumPlayers]game.PlayerConfig {
	var configs [game.NumPlayers]game.PlayerConfig
	for p := range configs {
		configs[p].Commander = benchCommander()
		deck := make([]*game.CardDef, 0, 99)
		for range 36 {
			deck = append(deck, benchForest())
		}
		for range 63 {
			deck = append(deck, benchHasteBeater())
		}
		configs[p].Deck = deck
	}
	return configs
}

// BenchmarkRealisticGame plays one full game over decks that can actually win,
// the realistic playtest case. Run with -benchtime=Nx.
func BenchmarkRealisticGame(b *testing.B) {
	configs := benchRealisticConfigs()
	b.ResetTimer()
	for range b.N {
		engine := NewEngine(nil)
		g := engine.NewGame(configs)
		engine.RunGame(g, allFirstLegalAgents())
	}
}

func benchForestConfigs() [game.NumPlayers]game.PlayerConfig {
	return benchForestConfigsN(99)
}

func benchForestConfigsN(deckSize int) [game.NumPlayers]game.PlayerConfig {
	var configs [game.NumPlayers]game.PlayerConfig
	for p := range configs {
		configs[p].Commander = benchCommander()
		deck := make([]*game.CardDef, deckSize)
		for i := range deck {
			deck[i] = benchForest()
		}
		configs[p].Deck = deck
	}
	return configs
}

// BenchmarkRealForestGame plays one full four-player game over 99-Forest decks
// with the deterministic FirstLegal agent. It approximates the worst case for
// the continuous-effect recompute: a long deck-out game with a large board.
// Run with -benchtime=1x for a single representative game (and -cpuprofile to
// profile the hot path).
func BenchmarkRealForestGame(b *testing.B) {
	configs := benchForestConfigs()
	b.ResetTimer()
	for range b.N {
		engine := NewEngine(nil)
		g := engine.NewGame(configs)
		engine.RunGame(g, allFirstLegalAgents())
	}
}
