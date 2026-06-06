package rules

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

const commanderDeckCardCount = 99

// CommanderLegalityError describes one conservative Commander deck legality
// violation.
type CommanderLegalityError struct {
	Player game.PlayerID
	Reason string
}

func (e CommanderLegalityError) Error() string {
	return fmt.Sprintf("player %d: %s", e.Player, e.Reason)
}

// ValidateCommanderConfigs returns commander legality errors for each configured player.
//
//nolint:gocritic // Keep the exported API accepting fixed-size player configs by value.
func ValidateCommanderConfigs(configs [game.NumPlayers]game.PlayerConfig) []CommanderLegalityError {
	var errs []CommanderLegalityError
	for i, config := range configs {
		errs = append(errs, validateCommanderConfig(game.PlayerID(i), config)...)
	}
	return errs
}

func validateCommanderConfig(playerID game.PlayerID, config game.PlayerConfig) []CommanderLegalityError {
	var errs []CommanderLegalityError
	add := func(reason string) {
		errs = append(errs, CommanderLegalityError{Player: playerID, Reason: reason})
	}

	if config.Commander == nil {
		add("missing commander")
		return errs
	}
	if len(config.Deck) != commanderDeckCardCount {
		add(fmt.Sprintf("deck has %d cards, want %d", len(config.Deck), commanderDeckCardCount))
	}
	if !config.Commander.HasSupertype(types.Legendary) || !config.Commander.HasType(types.Creature) {
		add("commander must be a legendary creature")
	}
	seen := make(map[string]bool)
	for _, card := range config.Deck {
		if card == nil {
			add("deck contains nil card")
			continue
		}
		if !config.Commander.ColorIdentity.ContainsAll(card.ColorIdentity) {
			add(fmt.Sprintf("%q has color identity outside commander's color identity", card.Name))
		}
		if card.Name == config.Commander.Name {
			add(fmt.Sprintf("commander %q is also present in deck", card.Name))
		}
		if card.HasSupertype(types.Basic) {
			continue
		}
		if seen[card.Name] {
			add(fmt.Sprintf("duplicate nonbasic card %q", card.Name))
			continue
		}
		seen[card.Name] = true
	}
	return errs
}

func isCommanderCardID(g *game.Game, cardID id.ID) bool {
	if cardID == 0 {
		return false
	}
	if g.CommanderIDs[cardID] {
		return true
	}
	for _, player := range g.Players {
		if player.CommanderInstanceID == cardID {
			return true
		}
	}
	return false
}

func commanderReplacementDestination(g *game.Game, cardID id.ID, destination zone.Type) zone.Type {
	if !isCommanderCardID(g, cardID) {
		return destination
	}
	switch destination {
	case zone.Graveyard, zone.Exile, zone.Hand, zone.Library:
		return zone.Command
	default:
		return destination
	}
}
