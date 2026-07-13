package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
)

// commanderColorIdentity returns the union of the color identities of every
// commander the given player owns (CR 903.4), together with an ok flag reporting
// whether that identity could be determined at all. Partner and other
// multiple-commander configurations union their identities per player.
//
// ok is false when the player association is unavailable or the player has no
// modeled commander whose definition can be read, so callers fail closed rather
// than approximate. A legitimately colorless commander yields (empty identity,
// true): the empty identity is a valid answer, distinct from the unavailable
// case.
func commanderColorIdentity(g *game.Game, playerID game.PlayerID) (color.Identity, bool) {
	player, ok := playerByID(g, playerID)
	if !ok {
		return color.Identity{}, false
	}
	seen := make(map[id.ID]bool)
	consider := func(cardID id.ID) bool {
		if cardID == 0 || seen[cardID] {
			return false
		}
		seen[cardID] = true
		card, ok := g.GetCardInstance(cardID)
		if !ok || card.Def == nil || card.Owner != playerID {
			return false
		}
		return true
	}

	var colors []color.Color
	found := false
	appendIdentity := func(cardID id.ID) {
		if !consider(cardID) {
			return
		}
		card, _ := g.GetCardInstance(cardID)
		found = true
		colors = append(colors, card.Def.ColorIdentity.Colors()...)
	}

	appendIdentity(player.CommanderInstanceID)
	for cardID := range g.CommanderIDs {
		appendIdentity(cardID)
	}
	if !found {
		return color.Identity{}, false
	}
	return color.NewIdentity(colors...), true
}

// commanderIdentityComplementColors returns the colors that are NOT in the given
// player's commander color identity (the protected set of "protection from each
// color that's not in your commander's color identity"), together with an ok
// flag. When ok is false the identity is unavailable and the caller must fail
// closed (protect from no color). A colorless commander identity yields all five
// colors; a five-color identity yields none.
func commanderIdentityComplementColors(g *game.Game, playerID game.PlayerID) ([]color.Color, bool) {
	identity, ok := commanderColorIdentity(g, playerID)
	if !ok {
		return nil, false
	}
	var complement []color.Color
	for _, c := range color.AllColors() {
		if !identity.Contains(c) {
			complement = append(complement, c)
		}
	}
	return complement, true
}
