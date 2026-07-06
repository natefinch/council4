package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

// handlePartitionExiledCostCards disposes of the cards exiled to pay the
// resolving ability's activation cost. One player (the next opponent of the
// controller when prim.ChooserOpponent is set, otherwise the controller) chooses
// a single exiled card; that card goes to the bottom (or top) of its owner's
// library, and every other exiled card returns to the battlefield under the
// controller's control, tapped when prim.OtherEntersTapped is set. It backs "An
// opponent chooses one of the exiled cards. You put that card on the bottom of
// your library and return the other to the battlefield tapped." (Coin of Fate).
// Only cards still in exile are considered, so a card that already moved is
// skipped.
func handlePartitionExiledCostCards(r *effectResolver, prim game.PartitionExiledCostCards) effectResolved {
	res := effectResolved{accepted: true}
	exiled := exiledCostCardsStillInExile(r.game, r.obj)
	if len(exiled) == 0 {
		return res
	}
	controllerID := r.obj.Controller
	chooserID := pileSplitActor(r.game, controllerID, prim.ChooserOpponent)
	chosenID := r.chooseExiledCostCard(chooserID, exiled)
	if r.putExiledCardInLibrary(chosenID, prim.ChosenToLibraryBottom) {
		res.succeeded = true
	}
	options := permanentCreationOptions{ForceTapped: prim.OtherEntersTapped}
	for _, cardID := range exiled {
		if cardID == chosenID {
			continue
		}
		if r.returnExiledCostCardToBattlefield(cardID, controllerID, options) {
			res.succeeded = true
		}
	}
	return res
}

// exiledCostCardsStillInExile returns the card-instance IDs the resolving object
// recorded as exiled to pay its cost that remain in their owner's exile zone, in
// cost order. A card that has since left exile is skipped so a disposal never
// acts on a stale identity.
func exiledCostCardsStillInExile(g *game.Game, obj *game.StackObject) []id.ID {
	if len(obj.ExiledAsCostIDs) == 0 {
		return nil
	}
	ids := make([]id.ID, 0, len(obj.ExiledAsCostIDs))
	for _, cardID := range obj.ExiledAsCostIDs {
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		owner, ok := playerByID(g, card.Owner)
		if !ok || !owner.Exile.Contains(cardID) {
			continue
		}
		ids = append(ids, cardID)
	}
	return ids
}

// chooseExiledCostCard asks chooserID to select one of the exiled cards and
// returns the chosen card's ID. The default selection keeps the disposal
// deterministic when no agent answers.
func (r *effectResolver) chooseExiledCostCard(chooserID game.PlayerID, exiled []id.ID) id.ID {
	options := make([]game.ChoiceOption, len(exiled))
	for i, cardID := range exiled {
		options[i] = game.ChoiceOption{Index: i, Label: cardChoiceLabel(r.game, cardID), Card: cardChoiceInfo(r.game, cardID)}
	}
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           chooserID,
		Prompt:           "Choose one of the exiled cards to put on the bottom of its owner's library.",
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}, r.log)
	if len(selected) == 1 && selected[0] >= 0 && selected[0] < len(exiled) {
		return exiled[selected[0]]
	}
	return exiled[0]
}

// putExiledCardInLibrary moves a cost-exiled card from its owner's exile zone to
// their library, to the bottom when bottom is set and the top otherwise.
func (r *effectResolver) putExiledCardInLibrary(cardID id.ID, bottom bool) bool {
	card, ok := r.game.GetCardInstance(cardID)
	if !ok {
		return false
	}
	owner, ok := playerByID(r.game, card.Owner)
	if !ok || !owner.Exile.Remove(cardID) {
		return false
	}
	if bottom {
		owner.Library.AddToBottom(cardID)
	} else {
		owner.Library.Add(cardID)
	}
	emitZoneChangeEvent(r.game, game.Event{
		Player:   card.Owner,
		CardID:   cardID,
		FromZone: zone.Exile,
		ToZone:   zone.Library,
		Amount:   1,
	})
	return true
}

// returnExiledCostCardToBattlefield moves a cost-exiled card from its owner's
// exile zone onto the battlefield under controllerID's control, applying the
// entry options (e.g. entering tapped).
func (r *effectResolver) returnExiledCostCardToBattlefield(cardID id.ID, controllerID game.PlayerID, options permanentCreationOptions) bool {
	card, ok := r.game.GetCardInstance(cardID)
	if !ok {
		return false
	}
	owner, ok := playerByID(r.game, card.Owner)
	if !ok || !owner.Exile.Remove(cardID) {
		return false
	}
	if _, ok := createCardPermanentFaceWithOptions(r.engine, r.game, card, controllerID, zone.Exile, game.FaceFront, nil, options, r.agents, r.log); ok {
		return true
	}
	owner.Exile.Add(cardID)
	return false
}
