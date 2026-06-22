package rules

import "github.com/natefinch/council4/mtg/game"

func (e *Engine) performCommanderMulligan(g *game.Game, playerID game.PlayerID) bool {
	player, ok := playerByID(g, playerID)
	if !ok {
		return false
	}
	for _, cardID := range player.Hand.All() {
		if player.Hand.Remove(cardID) {
			player.Library.AddToBottom(cardID)
		}
	}
	player.Library.Shuffle(e.rng)
	// TODO: real mulligan draws should not emit normal draw events/triggers.
	for range openingHandSize {
		e.drawCard(g, playerID, false)
	}
	player.CommanderMulligansTaken++
	bottomCount := player.CommanderMulligansTaken - 1 // first multiplayer Commander mulligan is free.
	if bottomCount <= 0 {
		return true
	}
	bottomCardsFromHand(player, bottomCount)
	return true
}

func bottomCardsFromHand(player *game.Player, count int) {
	if count <= 0 {
		return
	}
	cards := player.Hand.All()
	for i := len(cards) - 1; i >= 0 && count > 0; i-- {
		cardID := cards[i]
		if player.Hand.Remove(cardID) {
			player.Library.AddToBottom(cardID)
			count--
		}
	}
}
