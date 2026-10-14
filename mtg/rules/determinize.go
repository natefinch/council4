package rules

import (
	"math/rand/v2"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

// determinize re-samples the information hidden from observer so a search agent
// searches a plausible world rather than the true one, preserving the fog-of-war
// invariant (docs/adr/0011-search-based-agent-architecture.md, milestone S2). It
// keeps public identities and visible hands, shuffles unknown library order
// while preserving permitted top/search information, and re-deals hidden hands from their
// own hand+library pool — a hand of the same size, drawn from their real deck —
// so the searcher samples unknown information rather than reading the live state, against a
// consistent, legal deck. g must be a clone the caller owns; determinize mutates
// it in place.
func determinize(g *game.Game, observer game.PlayerID, rng *rand.Rand) {
	resampleHiddenFaceDown(g, observer, rng)
	for _, player := range g.Players {
		if player == nil {
			continue
		}
		obs := NewObservation(g, observer)
		// A library search grants full current library information, not merely
		// the matching candidates. Turn control alone grants neither.
		if g.LookingAtLibrary(observer, player.ID) {
			continue
		}
		top, public := obs.LibraryTopRevealed(player.ID)
		if private, visible := obs.LibraryTopLookable(player.ID); visible {
			top, public = private, true
		}
		if public {
			player.Library.Remove(top.CardInstanceID)
		}
		if g.CanSeePlayerInformation(observer, player.ID) {
			player.Library.Shuffle(rng)
		} else {
			resampleHiddenHand(player, rng)
		}
		if public {
			player.Library.Add(top.CardInstanceID)
		}
	}
}

// Full search worlds must not retain an uninvolved face-down card's true
// identity. Sample its printed definition together with that owner's other
// unknown cards, preserving physical IDs, locations and the deck multiset.
// Definitions remain immutable; only cloned CardInstances are reassigned.
func resampleHiddenFaceDown(g *game.Game, observer game.PlayerID, rng *rand.Rand) {
	hidden := make(map[id.ID]bool)
	for _, permanent := range g.Battlefield {
		if g.CanSeePlayerInformation(observer, effectiveController(g, permanent)) {
			continue
		}
		if permanent.FaceDown && !permanent.Token {
			hidden[permanent.CardInstanceID] = true
		}
		for _, component := range permanent.MergedCards {
			if component.FaceDown || permanent.FaceDown {
				hidden[component.CardInstanceID] = true
			}
		}
	}
	for _, obj := range g.Stack.Objects() {
		if obj.FaceDown && !g.CanSeePlayerInformation(observer, obj.Controller) {
			hidden[obj.SourceID] = true
		}
	}
	obs := NewObservation(g, observer)
	for _, player := range g.Players {
		if player == nil {
			continue
		}
		for _, cardID := range player.Exile.All() {
			card, _ := g.GetCardInstance(cardID)
			if player.Exile.IsFaceDown(cardID) && !obs.canLookAtFaceDownExile(&player.Exile, card) {
				hidden[cardID] = true
			}
		}
	}
	ordered := make([]id.ID, 0, len(hidden))
	for cardID := range hidden {
		ordered = append(ordered, cardID)
	}
	slices.Sort(ordered)
	for _, player := range g.Players {
		if player == nil {
			continue
		}
		var pool []id.ID
		for _, cardID := range ordered {
			if card, ok := g.GetCardInstance(cardID); ok && card.Owner == player.ID && !isCommanderCardID(g, cardID) {
				pool = append(pool, cardID)
			}
		}
		if len(pool) == 0 {
			continue
		}
		if !g.CanSeePlayerInformation(observer, player.ID) {
			for _, cardID := range player.Hand.All() {
				if !isCommanderCardID(g, cardID) {
					pool = append(pool, cardID)
				}
			}
		}
		if !g.LookingAtLibrary(observer, player.ID) {
			top, known := obs.LibraryTopRevealed(player.ID)
			if private, visible := obs.LibraryTopLookable(player.ID); visible {
				top, known = private, true
			}
			for _, cardID := range player.Library.All() {
				if (!known || cardID != top.CardInstanceID) && !isCommanderCardID(g, cardID) {
					pool = append(pool, cardID)
				}
			}
		}
		defs := make([]*game.CardDef, len(pool))
		for i, cardID := range pool {
			defs[i] = g.CardInstances[cardID].Def
		}
		rng.Shuffle(len(defs), func(i, j int) { defs[i], defs[j] = defs[j], defs[i] })
		for i, cardID := range pool {
			g.CardInstances[cardID].Def = defs[i]
		}
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
