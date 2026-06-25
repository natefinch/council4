package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// handleExileEntireHand exiles every card in the resolving player's hand at once
// and remembers each under the source-keyed linked set prim.LinkedKey names, so
// a paired handleReturnExiledCardsToHand on the same source returns exactly that
// set ("exile all cards from your hand" — Wormfang Behemoth). The link is keyed
// by the source permanent's card identity (like the O-Ring exile-until-leaves
// key), so the set survives the source leaving the battlefield, when the return
// trigger resolves from last-known information. An empty hand exiles nothing.
func handleExileEntireHand(r *effectResolver, prim game.ExileEntireHand) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return res
	}
	key := linkedObjectSourceKey(r.game, r.obj, string(prim.LinkedKey))
	clearLinkedObjects(r.game, key)
	for _, cardID := range slices.Clone(player.Hand.All()) {
		card, cardOK := r.game.GetCardInstance(cardID)
		if !cardOK {
			continue
		}
		if !moveCardBetweenZones(r.game, card.Owner, cardID, zone.Hand, zone.Exile) {
			continue
		}
		rememberLinkedObject(r.game, key, game.LinkedObjectRef{CardID: cardID})
		res.succeeded = true
		res.amount++
	}
	return res
}

// handleReturnExiledCardsToHand returns the cards an earlier handleExileEntireHand
// exiled under prim.LinkedKey to their owners' hands ("return the exiled cards to
// their owner's hand" — Wormfang Behemoth). Each remembered card moves from exile
// to its owner's hand; a card no longer in exile is skipped. The linked set is
// cleared afterward so a later exile/return cycle starts fresh.
func handleReturnExiledCardsToHand(r *effectResolver, prim game.ReturnExiledCardsToHand) effectResolved {
	res := effectResolved{accepted: true}
	key := linkedObjectSourceKey(r.game, r.obj, string(prim.LinkedKey))
	for _, ref := range linkedObjects(r.game, key) {
		card, ok := r.game.GetCardInstance(ref.CardID)
		if !ok {
			continue
		}
		if !moveCardBetweenZones(r.game, card.Owner, ref.CardID, zone.Exile, zone.Hand) {
			continue
		}
		res.succeeded = true
		res.amount++
	}
	clearLinkedObjects(r.game, key)
	return res
}
