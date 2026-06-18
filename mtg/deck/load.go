package deck

import (
	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/rules"
)

// PlayerInput pairs a player's display name with their parsed decklist.
type PlayerInput struct {
	Name     string
	Decklist *Decklist
}

// UnresolvedCard records a decklist entry whose name was not found in the card
// registry.
type UnresolvedCard struct {
	Player game.PlayerID
	Name   string
}

// LoadResult is the outcome of loading four decklists into game configs.
type LoadResult struct {
	// Configs are the assembled player configs, ready for rules.Engine.NewGame.
	Configs [game.NumPlayers]game.PlayerConfig

	// UnderTest identifies the player whose deck is being tested.
	UnderTest game.PlayerID

	// Unresolved lists decklist entries whose names were not in the registry.
	Unresolved []UnresolvedCard

	// Legality lists conservative Commander deck-legality violations, as
	// reported by rules.ValidateCommanderConfigs.
	Legality []rules.CommanderLegalityError
}

// OK reports whether every card resolved and every deck is Commander-legal.
func (r *LoadResult) OK() bool {
	return len(r.Unresolved) == 0 && len(r.Legality) == 0
}

// Load resolves four parsed decklists into validated PlayerConfigs using reg,
// designating underTest as the deck being tested.
//
// Load never panics on malformed input: card names missing from reg are
// collected in Unresolved, and conservative Commander legality violations are
// collected in Legality. Because an unresolved card is omitted from its deck, a
// player with unresolved names may also surface a deck-size legality error.
func Load(inputs [game.NumPlayers]PlayerInput, underTest game.PlayerID, reg *cards.Registry) *LoadResult {
	result := &LoadResult{UnderTest: underTest}
	for i := range inputs {
		config, unresolved := buildConfig(game.PlayerID(i), &inputs[i], reg)
		result.Configs[i] = config
		result.Unresolved = append(result.Unresolved, unresolved...)
	}
	result.Legality = rules.ValidateCommanderConfigs(result.Configs)
	return result
}

// buildConfig resolves one player's decklist into a PlayerConfig, returning any
// card names that were not found in reg. The first commander entry becomes the
// PlayerConfig commander; additional commander entries (partners) are not yet
// represented and are ignored.
func buildConfig(playerID game.PlayerID, input *PlayerInput, reg *cards.Registry) (game.PlayerConfig, []UnresolvedCard) {
	config := game.PlayerConfig{Name: input.Name}
	if input.Decklist == nil {
		return config, nil
	}

	var unresolved []UnresolvedCard
	if len(input.Decklist.Commander) > 0 {
		name := input.Decklist.Commander[0].Name
		if def := reg.Lookup(name); def != nil {
			config.Commander = def
		} else {
			unresolved = append(unresolved, UnresolvedCard{Player: playerID, Name: name})
		}
	}

	for i := range input.Decklist.Cards {
		entry := &input.Decklist.Cards[i]
		def := reg.Lookup(entry.Name)
		if def == nil {
			unresolved = append(unresolved, UnresolvedCard{Player: playerID, Name: entry.Name})
			continue
		}
		for range entry.Quantity {
			config.Deck = append(config.Deck, def)
		}
	}
	return config, unresolved
}
