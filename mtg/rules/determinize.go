package rules

import (
	"math/rand/v2"

	"github.com/natefinch/council4/mtg/game"
)

// determinize re-samples the information hidden from observer so a search agent
// searches a plausible world rather than the true one, preserving the fog-of-war
// invariant (docs/adr/0011-search-based-agent-architecture.md, milestone S2). It
// keeps every public zone and the observer's own hand, shuffles every library
// (its order is hidden), and re-deals each opponent's hand from that opponent's
// own hand+library pool — a hand of the same size, drawn from their real deck —
// so the searcher never sees an opponent's true hand yet still searches against a
// consistent, legal deck. g must be a clone the caller owns; determinize mutates
// it in place.
func determinize(g *game.Game, observer game.PlayerID, rng *rand.Rand) {
	for _, player := range g.Players {
		if player == nil {
			continue
		}
		if player.ID == observer {
			// The observer knows its own hand but not its library's order.
			player.Library.Shuffle(rng)
			continue
		}
		resampleHiddenHand(player, rng)
	}
}

// resampleHiddenHand pools an opponent's hand and library, shuffles the pool, and
// deals a fresh hand of the original size with the rest forming the library, so
// the opponent holds a random, deck-consistent hand the searcher could not know.
func resampleHiddenHand(player *game.Player, rng *rand.Rand) {
	handSize := player.Hand.Size()
	pool := append(player.Hand.All(), player.Library.All()...)
	if len(pool) == 0 {
		return
	}
	rng.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })

	for _, cardID := range player.Hand.All() {
		player.Hand.Remove(cardID)
	}
	for _, cardID := range player.Library.All() {
		player.Library.Remove(cardID)
	}
	for i, cardID := range pool {
		if i < handSize {
			player.Hand.Add(cardID)
		} else {
			player.Library.Add(cardID)
		}
	}
}
