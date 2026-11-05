package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

// handlePutLinkedExiledCardsInLibrary disposes of every card a linked exile
// clause exiled by moving it from exile to its owner's library, to the bottom
// when prim.Bottom is set. The legacy prim.LinkedKey form reads one resolution's
// link and consumes it. The prim.LinkedProducer form reads the original-source
// channel every resolution of the producing body accumulated on this permanent
// occurrence (CR 607.2a), so Trial of a Time Lord's guilty verdict finds the
// creatures chapters I, II and III each exiled; it leaves that channel in place
// because the moved incarnations no longer match it. Each disposed card is
// matched by the linked object identity and exact exiled incarnation, so an
// object that already left exile is skipped. Every card, across
// all owners, moves as one simultaneous batch (CR 603.2c): each card's
// zone-change replacements and its owner's choices are determined in APNAP order
// before any card moves (CR 101.4, 400.6, 616.1), in linked order. Then, in
// APNAP order, each owner privately chooses the order of their cards that
// arrived together (CR 401.4), or with RandomOrder they are randomly ordered
// instead (CR 401.2); either way the replacement choices and zone-change events
// precede the ordering, so they do not reveal the hidden library order.
func handlePutLinkedExiledCardsInLibrary(r *effectResolver, prim game.PutLinkedExiledCardsInLibrary) effectResolved {
	res := effectResolved{accepted: true}
	var key game.LinkedObjectKey
	if prim.LinkedProducer != "" {
		key = originalLinkedProducerKey(r.game, r.obj, prim.LinkedProducer)
	} else {
		key = linkedObjectSourceKey(r.game, r.obj, string(prim.LinkedKey))
	}
	disposed := make(map[id.ID]bool)
	var owners []game.PlayerID
	cardsByOwner := make(map[game.PlayerID][]game.ZoneCardSnapshot)
	for _, ref := range linkedObjects(r.game, key) {
		card, ok := linkedExiledCard(r.game, ref)
		if !ok || disposed[card.ID] {
			continue
		}
		disposed[card.ID] = true
		if _, seen := cardsByOwner[card.Owner]; !seen {
			owners = append(owners, card.Owner)
		}
		cardsByOwner[card.Owner] = append(cardsByOwner[card.Owner], game.ZoneCardSnapshot{
			CardID: card.ID, ZoneVersion: card.ZoneVersion,
		})
	}
	var moves []simultaneousCardMove
	for _, owner := range owners {
		for _, snapshot := range cardsByOwner[owner] {
			moves = append(moves, simultaneousCardMove{card: snapshot, fromZone: zone.Exile})
		}
	}
	arrange := r.ownerOrdersLibraryArrivals(prim.Bottom)
	if prim.RandomOrder {
		arrange = randomLibraryArrivals(r.engine.rng)
	}
	if len(moves) > 0 {
		res.succeeded = moveCardsBetweenZonesSimultaneously(r.game, moves, zone.Library, prim.Bottom, r.game.IDGen.Next(), arrange)
	}
	if prim.LinkedProducer == "" {
		clearLinkedObjects(r.game, key)
	}
	return res
}

func linkedExiledCard(g *game.Game, ref game.LinkedObjectRef) (*game.CardInstance, bool) {
	if ref.CardID == 0 {
		return nil, false
	}
	if ref.ObjectID != 0 {
		snapshot, ok := lastKnownObject(g, ref.ObjectID)
		if !ok || snapshot.CardID != ref.CardID && !linkedZoneCardOf(snapshot, ref) {
			return nil, false
		}
	}
	card, ok := linkedCardInstance(g, ref)
	if !ok {
		return nil, false
	}
	owner, ok := playerByID(g, card.Owner)
	if !ok || !owner.Exile.Contains(card.ID) {
		return nil, false
	}
	return card, true
}

// ownerOrdersLibraryArrivals lets the owner of cards that an effect put into
// the same library position at the same time arrange them in any order
// (CR 401.4). Only that owner is asked, and the default keeps arrival order.
// The options follow the public arrival order, so the decision stays out of
// the public log: its selection would reveal the hidden library order.
func (r *effectResolver) ownerOrdersLibraryArrivals(bottom bool) libraryArrivalArrangement {
	prompt := "Order the cards put on top of your library, top card first"
	if bottom {
		prompt = "Order the cards put on the bottom of your library, top card first"
	}
	return func(owner game.PlayerID, placed []id.ID) []id.ID {
		options := make([]game.ChoiceOption, len(placed))
		defaultOrder := make([]int, len(placed))
		for i, cardID := range placed {
			options[i] = game.ChoiceOption{
				Index: i,
				Label: cardChoiceLabel(r.game, cardID),
				Card:  cardChoiceInfo(r.game, cardID),
			}
			defaultOrder[i] = i
		}
		request := orderChoiceRequest(owner, prompt, options)
		request.DefaultSelection = defaultOrder
		request.Private = true
		selected := r.engine.chooseUnloggedChoice(r.game, r.agents, request, r.log)
		ordered := make([]id.ID, 0, len(selected))
		for _, index := range selected {
			if index < 0 || index >= len(placed) {
				return placed
			}
			ordered = append(ordered, placed[index])
		}
		return ordered
	}
}

// linkedZoneCardOf reports whether ref names one exact card incarnation the
// departed object became, such as a merged permanent's lower component, which
// is exiled with the source along with the rest of that permanent.
func linkedZoneCardOf(snapshot game.ObjectSnapshot, ref game.LinkedObjectRef) bool {
	return ref.CardZoneVersionSet && slices.Contains(snapshot.ZoneCards, game.ZoneCardSnapshot{
		CardID: ref.CardID, ZoneVersion: ref.CardZoneVersion,
	})
}
