package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// hideawayLinkID is the fixed source-scoped link key under which a Hideaway
// permanent records the card it exiled face down (CR 702.75a), so its later
// activated ability can find and play that exact card (CR 702.75c). Both the
// enters-the-battlefield trigger and the activated ability resolve from the same
// land, so they share this source-scoped key.
const hideawayLinkID = "hideaway"

// handleHideawayExile resolves the Hideaway N enters action: the controller
// looks at the top Amount cards of their library, exiles one of them face down
// linked to the source permanent, and puts the rest on the bottom of their
// library in a random order.
func handleHideawayExile(r *effectResolver, prim game.HideawayExile) effectResolved {
	res := effectResolved{accepted: true}
	playerID := stackObjectController(r.obj)
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return res
	}
	amount := min(r.quantity(prim.Amount), player.Library.Size())
	res.amount = amount
	if amount <= 0 {
		return res
	}
	looked := peekLibrary(player, amount)
	if len(looked) == 0 {
		return res
	}
	chosen := r.chooseHideawayCard(playerID, looked)
	rest := make([]id.ID, 0, len(looked)-1)
	for _, cardID := range looked {
		if cardID != chosen {
			rest = append(rest, cardID)
		}
	}
	key := linkedObjectSourceKey(r.game, r.obj, hideawayLinkID)
	clearLinkedObjects(r.game, key)
	destination, moved := moveCardBetweenZonesWithEvent(r.game, game.Event{
		SourceID: stackObjectSourceID(r.obj), StackObjectID: stackObjectID(r.obj),
		Controller: playerID, Player: playerID, CardID: chosen,
		FromZone: zone.Library, ToZone: zone.Exile,
	}, zoneMoveOptions{faceDownExile: true})
	if moved && destination == zone.Exile {
		card, _ := r.game.GetCardInstance(chosen)
		rememberLinkedObject(r.game, key, game.LinkedObjectRef{
			CardID: chosen, CardZoneVersion: card.ZoneVersion, CardZoneVersionSet: true,
		})
		res.succeeded = true
	}
	if r.engine.rng != nil && len(rest) > 1 {
		r.engine.rng.Shuffle(len(rest), func(i, j int) { rest[i], rest[j] = rest[j], rest[i] })
	}
	for _, cardID := range rest {
		if player.Library.Remove(cardID) {
			player.Library.AddToBottom(cardID)
		}
	}
	return res
}

// chooseHideawayCard has playerID choose which of the looked-at cards to exile
// face down. With a single card there is no decision.
func (r *effectResolver) chooseHideawayCard(playerID game.PlayerID, cards []id.ID) id.ID {
	if len(cards) == 1 {
		return cards[0]
	}
	options := make([]game.ChoiceOption, len(cards))
	for i, cardID := range cards {
		options[i] = game.ChoiceOption{
			Index: i,
			Label: cardChoiceLabel(r.game, cardID),
			Card:  cardChoiceInfo(r.game, cardID),
		}
	}
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           playerID,
		Private:          true,
		Prompt:           "Choose a card to exile face down",
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: firstChoiceIndices(1),
	}, r.log)
	for _, idx := range selected {
		if idx >= 0 && idx < len(cards) {
			return cards[idx]
		}
	}
	return cards[0]
}

// handlePlayHideawayCard resolves the "you may play the exiled card without
// paying its mana cost" half of the Hideaway mechanic. The enclosing
// instruction's Condition gate and Optional "may" are evaluated by the
// resolution envelope, so by the time this runs the controller has chosen to
// play the card and the printed condition holds. A land is put onto the
// battlefield; any other card is cast as a free spell from exile.
func handlePlayHideawayCard(r *effectResolver, _ game.PlayHideawayCard) effectResolved {
	res := effectResolved{accepted: true}
	playerID := stackObjectController(r.obj)
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return res
	}
	key := linkedObjectSourceKey(r.game, r.obj, hideawayLinkID)
	refs := linkedObjects(r.game, key)
	if len(refs) == 0 {
		return res
	}
	cardID := refs[0].CardID
	if cardID == 0 || !player.Exile.Contains(cardID) {
		return res
	}
	card, ok := r.game.GetCardInstance(cardID)
	if !ok {
		return res
	}
	player.Exile.SetFaceDown(cardID, false)
	def := cardFaceOrDefault(card, game.FaceFront)
	if def.HasType(types.Land) {
		if r.playHideawayLand(playerID, cardID) {
			clearLinkedObjects(r.game, key)
			res.succeeded = true
		}
		return res
	}
	if r.engine.castFreeSpellFromExile(r.game, playerID, cardID, r.agents, r.log) {
		clearLinkedObjects(r.game, key)
		res.succeeded = true
	}
	return res
}

// playHideawayLand puts the exiled hideaway land onto the battlefield under
// playerID's control from exile.
func (r *effectResolver) playHideawayLand(playerID game.PlayerID, cardID id.ID) bool {
	card, ok := r.game.GetCardInstance(cardID)
	if !ok {
		return false
	}
	_, placed := r.putResolvedCardOnBattlefieldValue(card, zone.Exile, playerID, nil, permanentCreationOptions{})
	return placed
}
