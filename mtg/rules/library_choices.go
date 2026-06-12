package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func millCards(g *game.Game, playerID game.PlayerID, amount int) {
	player, ok := playerByID(g, playerID)
	if !ok || amount <= 0 {
		return
	}
	for range amount {
		cardID, ok := player.Library.Top()
		if !ok {
			return
		}
		player.Library.Remove(cardID)
		destination := commanderReplacementDestination(g, cardID, zone.Graveyard)
		zoneOwner := playerID
		if card, ok := g.GetCardInstance(cardID); destination == zone.Command && ok {
			zoneOwner = card.Owner
		}
		destinationCards, ok := destinationZone(g, zoneOwner, destination)
		if !ok {
			return
		}
		destinationCards.Add(cardID)
		emitZoneChangeEvent(g, game.Event{
			Player:   playerID,
			CardID:   cardID,
			FromZone: zone.Library,
			ToZone:   destination,
			Amount:   1,
		})
	}
}

func (e *Engine) scryCards(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, playerID game.PlayerID, amount int) {
	player, ok := playerByID(g, playerID)
	if !ok || amount <= 0 {
		return
	}
	// TODO: replace sequential prompts with one partition+ordering choice.
	for _, cardID := range peekLibrary(player, amount) {
		selected := e.chooseChoice(g, agents, libraryChoiceRequest(game.ChoiceScry, playerID, "Scry: choose where to put card.", []string{"top", "bottom"}), log)
		if len(selected) == 1 && selected[0] == 1 && player.Library.Remove(cardID) {
			player.Library.AddToBottom(cardID)
		}
	}
	emitEvent(g, game.Event{
		Kind:                       game.EventScry,
		Controller:                 playerID,
		Player:                     playerID,
		Amount:                     amount,
		PlayerEventOrdinalThisTurn: nextPlayerEventOrdinalThisTurn(g, game.EventScry, playerID),
	})
}

func (e *Engine) surveilCards(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, playerID game.PlayerID, amount int) {
	player, ok := playerByID(g, playerID)
	if !ok || amount <= 0 {
		return
	}
	// TODO: replace sequential prompts with one partition+ordering choice.
	for _, cardID := range peekLibrary(player, amount) {
		selected := e.chooseChoice(g, agents, libraryChoiceRequest(game.ChoiceSurveil, playerID, "Surveil: choose where to put card.", []string{"top", "graveyard"}), log)
		if len(selected) == 1 && selected[0] == 1 && player.Library.Remove(cardID) {
			destination := commanderReplacementDestination(g, cardID, zone.Graveyard)
			zoneOwner := playerID
			if card, ok := g.GetCardInstance(cardID); destination == zone.Command && ok {
				zoneOwner = card.Owner
			}
			destinationCards, ok := destinationZone(g, zoneOwner, destination)
			if !ok {
				continue
			}
			destinationCards.Add(cardID)
			emitZoneChangeEvent(g, game.Event{
				Player:   playerID,
				CardID:   cardID,
				FromZone: zone.Library,
				ToZone:   destination,
				Amount:   1,
			})
		}
	}
	emitEvent(g, game.Event{
		Kind:                       game.EventSurveil,
		Controller:                 playerID,
		Player:                     playerID,
		Amount:                     amount,
		PlayerEventOrdinalThisTurn: nextPlayerEventOrdinalThisTurn(g, game.EventSurveil, playerID),
	})
}

func (e *Engine) manifestTopCard(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, playerID game.PlayerID) bool {
	player, ok := playerByID(g, playerID)
	if !ok {
		return false
	}
	cardID, ok := player.Library.Top()
	if !ok {
		return false
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok || !player.Library.Remove(cardID) {
		return false
	}
	_, ok = createCardPermanentFaceDownWithChoices(e, g, card, playerID, zone.Library, game.FaceFront, game.FaceDownManifest, false, agents, log)
	return ok
}

func (e *Engine) manifestDread(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, playerID game.PlayerID) bool {
	player, ok := playerByID(g, playerID)
	if !ok {
		return false
	}
	cards := peekLibrary(player, 2)
	if len(cards) == 0 {
		return false
	}
	chosenIndex := 0
	if len(cards) > 1 {
		selected := e.chooseChoice(g, agents, manifestDreadChoiceRequest(g, playerID, cards), log)
		if len(selected) == 1 && selected[0] >= 0 && selected[0] < len(cards) {
			chosenIndex = selected[0]
		}
	}
	chosenID := cards[chosenIndex]
	chosen, ok := g.GetCardInstance(chosenID)
	if !ok || !player.Library.Remove(chosenID) {
		return false
	}
	if _, ok := createCardPermanentFaceDownWithChoices(e, g, chosen, playerID, zone.Library, game.FaceFront, game.FaceDownManifest, false, agents, log); !ok {
		return false
	}
	for _, cardID := range cards {
		if cardID == chosenID {
			continue
		}
		if !player.Library.Remove(cardID) {
			continue
		}
		destination := commanderReplacementDestination(g, cardID, zone.Graveyard)
		zoneOwner := playerID
		if card, ok := g.GetCardInstance(cardID); destination == zone.Command && ok {
			zoneOwner = card.Owner
		}
		destinationCards, ok := destinationZone(g, zoneOwner, destination)
		if !ok {
			continue
		}
		destinationCards.Add(cardID)
		emitZoneChangeEvent(g, game.Event{
			Player:   playerID,
			CardID:   cardID,
			FromZone: zone.Library,
			ToZone:   destination,
			Amount:   1,
		})
	}
	return true
}

func manifestDreadChoiceRequest(g *game.Game, playerID game.PlayerID, cards []id.ID) game.ChoiceRequest {
	options := make([]game.ChoiceOption, 0, len(cards))
	for i, cardID := range cards {
		label := "unknown card"
		if card, ok := g.GetCardInstance(cardID); ok {
			label = cardFaceOrDefault(card, game.FaceFront).Name
		}
		options = append(options, game.ChoiceOption{Index: i, Label: label})
	}
	return game.ChoiceRequest{
		Kind:             game.ChoiceManifest,
		Player:           playerID,
		Prompt:           "Manifest dread: choose a card to manifest.",
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}
}

func (e *Engine) exploreCreature(
	g *game.Game,
	obj *game.StackObject,
	agents [game.NumPlayers]PlayerAgent,
	log *TurnLog,
	playerID game.PlayerID,
	creature *game.Permanent,
) bool {
	player, ok := playerByID(g, playerID)
	if !ok || creature == nil {
		return false
	}
	cardID, ok := player.Library.Top()
	if !ok {
		return false
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	emitCardRevealEvent(g, obj, playerID, cardID, zone.Library)
	if slices.Contains(cardFaceOrDefault(card, game.FaceFront).Types, types.Land) {
		player.Library.Remove(cardID)
		player.Hand.Add(cardID)
		emitZoneChangeEvent(g, game.Event{
			Player:   playerID,
			CardID:   cardID,
			FromZone: zone.Library,
			ToZone:   zone.Hand,
			Amount:   1,
		})
		return true
	}

	addCountersToPermanentControlledBy(g, playerID, creature, counter.PlusOnePlusOne, 1)
	selected := e.chooseChoice(g, agents, libraryChoiceRequest(game.ChoiceExplore, playerID, "Explore: choose where to put revealed nonland card.", []string{"top", "graveyard"}), log)
	if len(selected) == 1 && selected[0] == 1 && player.Library.Remove(cardID) {
		destination := commanderReplacementDestination(g, cardID, zone.Graveyard)
		zoneOwner := playerID
		if card, ok := g.GetCardInstance(cardID); destination == zone.Command && ok {
			zoneOwner = card.Owner
		}
		destinationCards, ok := destinationZone(g, zoneOwner, destination)
		if !ok {
			return true
		}
		destinationCards.Add(cardID)
		emitZoneChangeEvent(g, game.Event{
			Player:   playerID,
			CardID:   cardID,
			FromZone: zone.Library,
			ToZone:   destination,
			Amount:   1,
		})
	}
	return true
}

func peekLibrary(player *game.Player, amount int) []id.ID {
	if amount <= 0 {
		return nil
	}
	cards := player.Library.All()
	if amount > len(cards) {
		amount = len(cards)
	}
	return append([]id.ID(nil), cards[:amount]...)
}

func reorderLibraryTop(player *game.Player, cards []id.ID) {
	if len(cards) == 0 {
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
