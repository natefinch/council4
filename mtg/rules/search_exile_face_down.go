package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

// handleSearchExileFaceDown resolves the "Search your library for a card, exile
// it face down, then shuffle" search half of a search/exile/conditional-cast
// payoff (Beseech the Mirror). It finds one matching card, exiles it face down in
// the searching player's exile zone, shuffles the library, and publishes the
// exiled card under the primitive's linked key so a following instruction can
// cast it or move it to hand. The card stays hidden in exile until a later
// instruction reveals it. Failing to find leaves nothing exiled and publishes no
// link, so a following move-to-hand fallback finds nothing to move.
func handleSearchExileFaceDown(r *effectResolver, prim game.Search) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	var key game.LinkedObjectKey
	if prim.PublishLinked != "" {
		key = linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked))
		clearLinkedObjects(r.game, key)
	}
	cardID, found := r.engine.searchLibraryExileFaceDown(r.game, r.obj, r.agents, r.log, playerID, prim.Spec, res.amount)
	if !found {
		return res
	}
	res.succeeded = true
	res.amount = 1
	if prim.PublishLinked != "" {
		rememberLinkedObject(r.game, key, game.LinkedObjectRef{CardID: cardID})
	}
	return res
}

// searchLibraryExileFaceDown finds a single card matching spec in playerID's
// library, exiles it face down, and shuffles the library. It returns the exiled
// card's instance ID and true when a card is found, or false when the search
// finds nothing (an empty library, or a legal decline). The library is always
// shuffled because the searching player searched it (CR 701.19a), whether or not
// a card was found.
func (e *Engine) searchLibraryExileFaceDown(g *game.Game, obj *game.StackObject, agents [game.NumPlayers]PlayerAgent, log *TurnLog, playerID game.PlayerID, spec game.SearchSpec, amount int) (id.ID, bool) {
	if amount <= 0 {
		amount = 1
	}
	player, ok := playerByID(g, playerID)
	if !ok {
		return 0, false
	}
	// The player searches their library when this runs regardless of whether a
	// matching card is found (CR 701.19a), so the search event fires once here.
	emitEvent(g, game.Event{
		Kind:       game.EventLibrarySearched,
		Controller: playerID,
		Player:     playerID,
	})
	var candidates []id.ID
	for _, cardID := range player.Library.All() {
		if searchSpecMatches(g, obj, cardID, spec) {
			candidates = append(candidates, cardID)
		}
	}
	minChoices := 0
	if searchMustFindIfAvailable(spec, amount) {
		minChoices = 1
	}
	found := e.chooseSearchMatches(g, agents, log, playerID, candidates, 1, minChoices)
	if len(found) == 0 {
		player.Library.Shuffle(e.rng)
		return 0, false
	}
	cardID := found[0]
	if !player.Library.Remove(cardID) {
		player.Library.Shuffle(e.rng)
		return 0, false
	}
	player.Exile.Add(cardID)
	player.Exile.SetFaceDown(cardID, true)
	emitZoneChangeEvent(g, game.Event{
		SourceID:      stackObjectSourceID(obj),
		StackObjectID: stackObjectID(obj),
		Controller:    stackObjectController(obj),
		Player:        playerID,
		CardID:        cardID,
		FromZone:      zone.Library,
		ToZone:        zone.Exile,
		Amount:        1,
	})
	player.Library.Shuffle(e.rng)
	return cardID, true
}
