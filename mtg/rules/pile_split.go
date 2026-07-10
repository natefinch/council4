package rules

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

// handlePileSplit reveals the top cards of the controller's library, has the
// separating player divide them into two piles, has the choosing player pick
// one pile to keep, and routes the kept pile to prim.Kept and the other pile to
// prim.Other.
func handlePileSplit(r *effectResolver, prim game.PileSplit) effectResolved {
	look := r.quantity(prim.Amount)
	res := effectResolved{accepted: true, amount: look}
	controllerID, ok := r.resolvePlayer(prim.Player)
	if ok {
		res.succeeded = r.engine.pileSplitCards(r.game, r.agents, r.log, controllerID, look, prim)
	}
	return res
}

// pileSplitCards implements the Fact-or-Fiction reveal-and-split process. It
// reveals the top look cards of controllerID's library, asks the separating
// player to divide them into two piles, asks the choosing player which pile is
// kept, then moves the kept pile to prim.Kept and the other pile to prim.Other
// (both belonging to controllerID).
func (e *Engine) pileSplitCards(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, controllerID game.PlayerID, look int, prim game.PileSplit) bool {
	player, ok := playerByID(g, controllerID)
	if !ok || look <= 0 {
		return false
	}
	seen := peekLibrary(player, look)
	if len(seen) == 0 {
		return false
	}
	separatorID := pileSplitActor(g, controllerID, prim.SeparatorOpponent)
	chooserID := pileSplitActor(g, controllerID, prim.ChooserOpponent)

	firstPile := e.choosePileSeparation(g, agents, log, separatorID, seen)
	keepFirstPile := e.choosePileKept(g, agents, log, chooserID, seen, firstPile)

	var keptCards, otherCards []id.ID
	for _, cardID := range seen {
		inFirstPile := slices.Contains(firstPile, cardID)
		if inFirstPile == keepFirstPile {
			keptCards = append(keptCards, cardID)
		} else {
			otherCards = append(otherCards, cardID)
		}
	}
	movePileSplitPile(g, controllerID, player, keptCards, prim.Kept)
	movePileSplitPile(g, controllerID, player, otherCards, prim.Other)
	return true
}

// pileSplitActor returns the player who performs a separate-or-choose role: the
// controller when opponent is false, otherwise the next non-eliminated opponent
// in turn order (the controller again only when no opponent remains).
func pileSplitActor(g *game.Game, controllerID game.PlayerID, opponent bool) game.PlayerID {
	if !opponent {
		return controllerID
	}
	return g.TurnOrder.NextActivePlayer(controllerID)
}

// choosePileSeparation asks the separating player which of the seen cards form
// the first pile; the remaining seen cards form the second pile. The default
// splits the revealed cards as evenly as possible.
func (e *Engine) choosePileSeparation(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, separatorID game.PlayerID, seen []id.ID) []id.ID {
	options := make([]game.ChoiceOption, len(seen))
	for i, cardID := range seen {
		options[i] = game.ChoiceOption{Index: i, Label: cardChoiceLabel(g, cardID), Card: cardChoiceInfo(g, cardID)}
	}
	defaults := make([]int, 0, len(seen)/2)
	for i := range len(seen) / 2 {
		defaults = append(defaults, i)
	}
	selected := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind:             game.ChoicePileSeparate,
		Player:           separatorID,
		Prompt:           "Pile split: separate the revealed cards into two piles (selected cards form the first pile).",
		Options:          options,
		MinChoices:       0,
		MaxChoices:       len(seen),
		DefaultSelection: defaults,
	}, log)
	firstPile := make([]id.ID, 0, len(selected))
	for _, index := range selected {
		if index >= 0 && index < len(seen) {
			firstPile = append(firstPile, seen[index])
		}
	}
	return firstPile
}

// choosePileKept asks the choosing player which pile is kept, returning true
// when the first pile is kept. The default keeps the larger pile.
func (e *Engine) choosePileKept(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, chooserID game.PlayerID, seen, firstPile []id.ID) bool {
	firstCount := len(firstPile)
	secondCount := len(seen) - firstCount
	defaultIndex := 0
	if secondCount > firstCount {
		defaultIndex = 1
	}
	options := []game.ChoiceOption{
		{Index: 0, Label: fmt.Sprintf("First pile (%d cards)", firstCount)},
		{Index: 1, Label: fmt.Sprintf("Second pile (%d cards)", secondCount)},
	}
	selected := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind:             game.ChoicePileChoose,
		Player:           chooserID,
		Prompt:           "Pile split: choose which pile is put into the revealing player's hand.",
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{defaultIndex},
	}, log)
	return len(selected) == 1 && selected[0] == 0
}

// movePileSplitPile moves a pile of cards from controllerID's library to a
// destination zone. The hand placement is direct; a library placement puts the
// cards on the bottom; a graveyard placement routes through
// putLibraryCardIntoGraveyard so it honors both the commander replacement
// (CR 903.9a) and graveyard-redirect replacements (CR 614; e.g. an opponent's
// Dauthi Voidwalker exiles the losing pile with void counters). The whole
// graveyard-bound pile shares one simultaneous ID, since its cards move at once.
func movePileSplitPile(g *game.Game, controllerID game.PlayerID, player *game.Player, cards []id.ID, destination zone.Type) {
	var graveyardBatchID id.ID
	for _, cardID := range cards {
		if !player.Library.Remove(cardID) {
			continue
		}
		switch destination {
		case zone.Hand:
			player.Hand.Add(cardID)
			emitZoneChangeEvent(g, game.Event{
				Player:   controllerID,
				CardID:   cardID,
				FromZone: zone.Library,
				ToZone:   zone.Hand,
				Amount:   1,
			})
		case zone.Library:
			player.Library.AddToBottom(cardID)
			emitZoneChangeEvent(g, game.Event{
				Player:   controllerID,
				CardID:   cardID,
				FromZone: zone.Library,
				ToZone:   zone.Library,
				Amount:   1,
			})
		default:
			if graveyardBatchID == 0 {
				graveyardBatchID = g.IDGen.Next()
			}
			putLibraryCardIntoGraveyard(g, controllerID, cardID, graveyardBatchID)
		}
	}
}
