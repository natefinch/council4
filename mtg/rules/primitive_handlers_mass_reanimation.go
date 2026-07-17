package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

// handleChooseCardFromEachGraveyard walks every player in prim.Players in APNAP
// order and, for each, has prim.Chooser pick a card in that player's graveyard
// matching prim.Selection — exactly one when able, or up to one when
// prim.Optional — remembering each chosen card, card-scoped, under
// prim.LinkedKey. Each graveyard is an independent pool, so at most one card per
// player is chosen and a player with no eligible card contributes none. The
// chosen cards stay in their graveyards; a paired ReanimateLinkedCards puts
// exactly those cards onto the battlefield.
func handleChooseCardFromEachGraveyard(r *effectResolver, prim game.ChooseCardFromEachGraveyard) effectResolved {
	res := effectResolved{accepted: true}
	chooser, ok := r.resolvePlayer(prim.Chooser)
	if !ok {
		return res
	}
	key := linkedObjectSourceKey(r.game, r.obj, string(prim.LinkedKey))
	for _, ownerID := range playersInAPNAPOrder(r.game, r.playerGroupMembers(prim.Players)) {
		graveyard, ok := destinationZone(r.game, ownerID, zone.Graveyard)
		if !ok {
			continue
		}
		var pool []id.ID
		for _, cardID := range graveyard.All() {
			card, ok := r.game.GetCardInstance(cardID)
			if !ok {
				continue
			}
			if handCardMatchesSelection(r.game, card, prim.Selection, ownerID) {
				pool = append(pool, cardID)
			}
		}
		if len(pool) == 0 {
			continue
		}
		cardID, chosen := r.chooseCardFromGraveyardPool(chooser, pool, prim.Optional)
		if !chosen {
			continue
		}
		rememberLinkedObject(r.game, key, game.LinkedObjectRef{CardID: cardID})
		res.succeeded = true
	}
	return res
}

// chooseCardFromGraveyardPool has chooser pick one card from pool, or none when
// optional, returning the chosen card and whether one was chosen. A mandatory
// pick from a non-empty pool always yields a card; an optional pick may yield
// none.
func (r *effectResolver) chooseCardFromGraveyardPool(chooser game.PlayerID, pool []id.ID, optional bool) (id.ID, bool) {
	minChoices := 1
	if optional {
		minChoices = 0
	}
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           chooser,
		Prompt:           "Choose a card in that player's graveyard to put onto the battlefield",
		Options:          chooseFromZoneOptions(r.game, pool),
		MinChoices:       minChoices,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}, r.log)
	if len(selected) == 1 && selected[0] >= 0 && selected[0] < len(pool) {
		return pool[selected[0]], true
	}
	return 0, false
}

// handleReanimateLinkedCards puts every card a sibling ChooseCardFromEachGraveyard
// remembered under prim.LinkedKey from its owner's graveyard onto the battlefield
// at once under prim.Controller's control. Routing every card through
// putResolvedCardsOnBattlefieldCollecting makes them enter as one simultaneous
// batch, so commander-zone replacements, enter-the-battlefield replacements and
// triggers, ownership, control, and legend/state-based checks all resolve the way
// a single mass reanimation should. It clears the consumed link once resolved.
func handleReanimateLinkedCards(r *effectResolver, prim game.ReanimateLinkedCards) effectResolved {
	res := effectResolved{accepted: true}
	controller, ok := r.recipientController(prim.Controller)
	if !ok {
		return res
	}
	key := linkedObjectSourceKey(r.game, r.obj, string(prim.LinkedKey))
	var resolved []resolvedBattlefieldCard
	for _, ref := range linkedObjects(r.game, key) {
		if ref.CardID == 0 {
			continue
		}
		card, ok := r.game.GetCardInstance(ref.CardID)
		if !ok {
			continue
		}
		graveyard, ok := destinationZone(r.game, card.Owner, zone.Graveyard)
		if !ok || !graveyard.Contains(card.ID) {
			continue
		}
		resolved = append(resolved, resolvedBattlefieldCard{
			card:       card,
			fromZone:   zone.Graveyard,
			controller: controller,
		})
	}
	clearLinkedObjects(r.game, key)
	if len(resolved) == 0 {
		return res
	}
	entered := r.putResolvedCardsOnBattlefieldCollecting(resolved, nil, permanentCreationOptions{})
	if len(entered) > 0 {
		res.succeeded = true
	}
	return res
}
