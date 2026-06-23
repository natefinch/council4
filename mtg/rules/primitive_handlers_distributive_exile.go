package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// handleExileForEachPlayer resolves the distributive Saga chapter "For each
// player, exile up to one [other] target <permanent> that player controls until
// this Saga leaves the battlefield." (Vault 13: Dweller's Journey, Battle at the
// Helvault). Walking every player in APNAP order, prim.Chooser picks up to one
// permanent that player controls matching prim.Selection and the runtime exiles
// it, remembering each exiled permanent under prim.LinkedKey keyed by the source
// so the paired return brings back exactly this set. The chosen permanents
// accumulate across chapters under the same key; nothing is cleared here.
func handleExileForEachPlayer(r *effectResolver, prim game.ExileForEachPlayer) effectResolved {
	res := effectResolved{accepted: true}
	chooser, ok := r.resolvePlayer(prim.Chooser)
	if !ok {
		return res
	}
	source, _ := sourcePermanent(r.game, r.obj)
	resolver := newReferenceResolver(r.game, r.obj)
	key := linkedObjectSourceKey(r.game, r.obj, string(prim.LinkedKey))
	for _, playerID := range playersInAPNAPOrder(r.game, r.playerGroupMembers(game.AllPlayersReference())) {
		candidates := playerControlledSelectionCandidates(r.game, resolver, source, playerID, prim.Selection)
		permanent, chosen := r.engine.chooseUpToOnePermanent(r.game, candidates, chooser, "Choose a permanent to exile", r.agents, r.log)
		if !chosen {
			continue
		}
		linkedRef := permanentLinkedObjectRef(permanent)
		if movePermanentToZone(r.game, permanent, zone.Exile) {
			rememberLinkedObject(r.game, key, linkedRef)
			res.succeeded = true
		}
	}
	return res
}

// handleReturnLinkedExiledCardsToBattlefield resolves the partial Saga payoff
// "Return N cards exiled with this Saga to the battlefield under their owners'
// control and put the rest on the bottom of their owners' libraries." (Vault 13:
// Dweller's Journey). prim.Chooser picks up to prim.Amount of the still-exiled
// cards a sibling ExileForEachPlayer recorded under prim.LinkedKey to put onto
// the battlefield under their owners' control; when prim.RestToLibraryBottom is
// set every remaining linked card goes to the bottom of its owner's library. It
// clears the link so the synthesized leaves-the-battlefield return finds nothing
// left.
func handleReturnLinkedExiledCardsToBattlefield(r *effectResolver, prim game.ReturnLinkedExiledCardsToBattlefield) effectResolved {
	res := effectResolved{accepted: true}
	amount := r.quantity(prim.Amount)
	key := linkedObjectSourceKey(r.game, r.obj, string(prim.LinkedKey))
	live := liveLinkedExiledCards(r.game, key)
	returned := r.engine.chooseLinkedExiledCardsToReturn(r.game, live, amount, r.obj.Controller, r.agents, r.log)
	returnedSet := make(map[*game.CardInstance]bool, len(returned))
	for _, card := range returned {
		owner, ok := playerByID(r.game, card.Owner)
		if !ok || !owner.Exile.Remove(card.ID) {
			continue
		}
		if _, ok := createCardPermanentFaceWithOptions(r.engine, r.game, card, card.Owner, zone.Exile, game.FaceFront, nil, permanentCreationOptions{}, r.agents, r.log); ok {
			returnedSet[card] = true
			res.succeeded = true
		}
	}
	if prim.RestToLibraryBottom {
		for _, entry := range live {
			if returnedSet[entry.card] {
				continue
			}
			owner, ok := playerByID(r.game, entry.card.Owner)
			if !ok || !owner.Exile.Remove(entry.card.ID) {
				continue
			}
			owner.Library.AddToBottom(entry.card.ID)
			emitZoneChangeEvent(r.game, game.Event{
				Player:   entry.card.Owner,
				CardID:   entry.card.ID,
				FromZone: zone.Exile,
				ToZone:   zone.Library,
			})
			res.succeeded = true
		}
	}
	clearLinkedObjects(r.game, key)
	return res
}

// liveLinkedExiledCard pairs a still-exiled linked card with its owning player.
type liveLinkedExiledCard struct {
	card  *game.CardInstance
	owner *game.Player
}

// liveLinkedExiledCards resolves the cards recorded under key that are still in
// their owners' exile zones, in record order. A linked object whose last-known
// identity no longer matches, whose card instance is gone, or that has left
// exile is skipped, mirroring the existing linked-disposal handlers.
func liveLinkedExiledCards(g *game.Game, key game.LinkedObjectKey) []liveLinkedExiledCard {
	var live []liveLinkedExiledCard
	for _, ref := range linkedObjects(g, key) {
		if snapshot, ok := lastKnownObject(g, ref.ObjectID); !ok || snapshot.CardID != ref.CardID {
			continue
		}
		card, ok := g.GetCardInstance(ref.CardID)
		if !ok {
			continue
		}
		owner, ok := playerByID(g, card.Owner)
		if !ok || !owner.Exile.Contains(card.ID) {
			continue
		}
		live = append(live, liveLinkedExiledCard{card: card, owner: owner})
	}
	return live
}

// chooseLinkedExiledCardsToReturn has chooser pick up to amount of the live
// linked exiled cards to return; fewer live cards than amount returns all of
// them, matching "Return N cards ..." returning every exiled card when fewer
// than N remain.
func (e *Engine) chooseLinkedExiledCardsToReturn(g *game.Game, live []liveLinkedExiledCard, amount int, chooser game.PlayerID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) []*game.CardInstance {
	if amount <= 0 || len(live) == 0 {
		return nil
	}
	if len(live) <= amount {
		cards := make([]*game.CardInstance, len(live))
		for i, entry := range live {
			cards[i] = entry.card
		}
		return cards
	}
	options := make([]game.ChoiceOption, len(live))
	for i, entry := range live {
		options[i] = game.ChoiceOption{Index: i, Label: cardChoiceLabel(g, entry.card.ID), Card: cardChoiceInfo(g, entry.card.ID)}
	}
	request := game.ChoiceRequest{
		Kind:             game.ChoicePayment,
		Player:           chooser,
		Prompt:           "Choose cards to return to the battlefield",
		Options:          options,
		MinChoices:       amount,
		MaxChoices:       amount,
		DefaultSelection: firstChoiceIndices(amount),
	}
	selected := e.chooseChoice(g, agents, request, log)
	cards := make([]*game.CardInstance, 0, len(selected))
	for _, idx := range selected {
		if idx >= 0 && idx < len(live) {
			cards = append(cards, live[idx].card)
		}
	}
	return cards
}
