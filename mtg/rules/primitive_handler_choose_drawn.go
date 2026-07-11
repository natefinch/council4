package rules

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

// handleChooseDrawnPayLifeOrTop resolves the Sylvan Library filter: the player
// chooses up to ChooseCount cards in their hand that were drawn this turn and,
// for each chosen card, pays LifeCost life to keep it or puts it on top of their
// library. A chosen card whose owner cannot pay LifeCost life is put on top.
func handleChooseDrawnPayLifeOrTop(r *effectResolver, prim game.ChooseDrawnPayLifeOrTop) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	candidates := cardsInHandDrawnThisTurn(r.game, playerID)
	if len(candidates) == 0 {
		return res
	}
	chooseCount := min(prim.ChooseCount, len(candidates))
	if chooseCount <= 0 {
		return res
	}
	options := chooseFromZoneOptions(r.game, candidates)
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           playerID,
		Prompt:           "Choose cards drawn this turn to pay life for or put on top of your library",
		Options:          options,
		MinChoices:       chooseCount,
		MaxChoices:       chooseCount,
		DefaultSelection: firstChoiceIndices(chooseCount),
	}, r.log)
	for _, cardID := range chooseFromZoneResolveIndices(candidates, selected) {
		if r.resolveDrawnCardPayLifeOrTop(playerID, cardID, prim.LifeCost) {
			res.succeeded = true
		}
	}
	return res
}

// resolveDrawnCardPayLifeOrTop applies one chosen card's "pay life or put on top"
// choice, reporting whether the card was kept for life or moved to the library.
func (r *effectResolver) resolveDrawnCardPayLifeOrTop(playerID game.PlayerID, cardID id.ID, lifeCost int) bool {
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return false
	}
	canPay := lifeCost <= 0 || player.Life >= lifeCost
	if canPay && r.chooseDrawnCardPayLife(playerID, cardID, lifeCost) {
		loseLife(r.game, playerID, lifeCost)
		return true
	}
	return moveCardBetweenZonesWithPlacement(r.game, playerID, cardID, zone.Hand, zone.Library, false)
}

// chooseDrawnCardPayLife asks playerID whether to pay lifeCost life to keep
// cardID (rather than put it on top of their library).
func (r *effectResolver) chooseDrawnCardPayLife(playerID game.PlayerID, cardID id.ID, lifeCost int) bool {
	cardInfo := cardChoiceInfo(r.game, cardID)
	options := []game.ChoiceOption{
		{Index: 0, Label: fmt.Sprintf("Pay %d life", lifeCost), Card: cardInfo},
		{Index: 1, Label: "Put on top of library", Card: cardInfo},
	}
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           playerID,
		Prompt:           fmt.Sprintf("Pay %d life to keep %s, or put it on top of your library", lifeCost, cardChoiceLabel(r.game, cardID)),
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{1},
		Subject:          cardInfo,
	}, r.log)
	return len(selected) == 1 && selected[0] == 0
}

// cardsInHandDrawnThisTurn returns the cards currently in playerID's hand that
// were drawn this turn, in draw order, deduplicated. It reads the turn's
// EventCardDrawn events, so it tracks card identity rather than a count.
func cardsInHandDrawnThisTurn(g *game.Game, playerID game.PlayerID) []id.ID {
	player, ok := playerByID(g, playerID)
	if !ok {
		return nil
	}
	seen := make(map[id.ID]bool)
	var result []id.ID
	for _, event := range eventsThisTurnWindow(g) {
		if event.Kind != game.EventCardDrawn || event.Player != playerID || event.CardID == 0 {
			continue
		}
		if seen[event.CardID] || !player.Hand.Contains(event.CardID) {
			continue
		}
		seen[event.CardID] = true
		result = append(result, event.CardID)
	}
	return result
}
