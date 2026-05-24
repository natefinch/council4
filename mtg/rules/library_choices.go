package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

func millCards(g *game.Game, playerID game.PlayerID, amount int) {
	player := playerByID(g, playerID)
	if player == nil || amount <= 0 {
		return
	}
	for range amount {
		cardID, ok := player.Library.Top()
		if !ok {
			return
		}
		player.Library.Remove(cardID)
		player.Graveyard.Add(cardID)
		emitZoneChangeEvent(g, game.GameEvent{
			Player:   playerID,
			CardID:   cardID,
			FromZone: game.ZoneLibrary,
			ToZone:   game.ZoneGraveyard,
			Amount:   1,
		})
	}
}

func (e *Engine) scryCards(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, playerID game.PlayerID, amount int) {
	player := playerByID(g, playerID)
	if player == nil || amount <= 0 {
		return
	}
	// TODO: replace sequential prompts with one partition+ordering choice.
	for _, cardID := range peekLibrary(player, amount) {
		selected := e.chooseChoice(g, agents, libraryChoiceRequest(game.ChoiceScry, playerID, "Scry: choose where to put card.", []string{"top", "bottom"}), log)
		if len(selected) == 1 && selected[0] == 1 && player.Library.Remove(cardID) {
			player.Library.AddToBottom(cardID)
		}
	}
}

func (e *Engine) surveilCards(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, playerID game.PlayerID, amount int) {
	player := playerByID(g, playerID)
	if player == nil || amount <= 0 {
		return
	}
	// TODO: replace sequential prompts with one partition+ordering choice.
	for _, cardID := range peekLibrary(player, amount) {
		selected := e.chooseChoice(g, agents, libraryChoiceRequest(game.ChoiceSurveil, playerID, "Surveil: choose where to put card.", []string{"top", "graveyard"}), log)
		if len(selected) == 1 && selected[0] == 1 && player.Library.Remove(cardID) {
			player.Graveyard.Add(cardID)
			emitZoneChangeEvent(g, game.GameEvent{
				Player:   playerID,
				CardID:   cardID,
				FromZone: game.ZoneLibrary,
				ToZone:   game.ZoneGraveyard,
				Amount:   1,
			})
		}
	}
}

func peekLibrary(player *game.Player, amount int) []id.ID {
	if player == nil || amount <= 0 {
		return nil
	}
	cards := player.Library.All()
	if amount > len(cards) {
		amount = len(cards)
	}
	return append([]id.ID(nil), cards[:amount]...)
}

func reorderLibraryTop(player *game.Player, cards []id.ID) {
	if player == nil || len(cards) == 0 {
		return
	}

	for _, cardID := range cards {
		player.Library.Remove(cardID)
	}
	for i := len(cards) - 1; i >= 0; i-- {
		player.Library.Add(cards[i])
	}
}

func libraryChoiceRequest(kind game.ChoiceKind, playerID game.PlayerID, prompt string, labels []string) game.ChoiceRequest {
	options := make([]game.ChoiceOption, 0, len(labels))
	for i, label := range labels {
		options = append(options, game.ChoiceOption{Index: i, Label: label})
	}
	return game.ChoiceRequest{
		Kind:             kind,
		Player:           playerID,
		Prompt:           prompt,
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}
}
