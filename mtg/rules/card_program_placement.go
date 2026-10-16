package rules

import (
	"errors"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func (c *programContext) availableFoundCard(handle cardprogram.FoundCard) (*game.CardInstance, zone.Type, bool) {
	card, exists := c.game.GetCardInstance(handle.Identity.CardID)
	if !exists || card.ZoneVersion != handle.Identity.ZoneVersion {
		return nil, zone.None, false
	}
	from := zone.Library
	if recorded, known := c.foundZones[card.ID]; known {
		from = recorded
	}
	source, located := destinationZone(c.game, card.Owner, from)
	if !located || !source.Contains(card.ID) || from == zone.Library &&
		c.game.Players[card.Owner].Library.ReorderVersion() != handle.LibraryEpoch {
		return nil, zone.None, false
	}
	return card, from, true
}

func (c *programContext) putFoundCardsOnBattlefield(placements []cardprogram.CardPlacement) ([]cardprogram.FoundCard, error) {
	var resolved []resolvedBattlefieldCard
	for _, player := range triggerAPNAPPlayers(c.game) {
		for _, placement := range placements {
			for _, handle := range placement.Cards {
				card, from, available := c.availableFoundCard(handle)
				if available && card.Owner == player {
					resolved = append(resolved, resolvedBattlefieldCard{card: card, fromZone: from, controller: player})
				}
			}
		}
	}
	resolver := newEffectResolver(c.engine, c.game, c.obj, c.agents, c.log)
	permanents := resolver.putResolvedCardsOnBattlefieldCollecting(resolved, nil, permanentCreationOptions{})
	arrivals := make([]cardprogram.FoundCard, 0, len(permanents))
	for _, permanent := range permanents {
		card, _ := c.game.GetCardInstance(permanent.CardInstanceID)
		arrivals = append(arrivals, cardprogram.FoundCard{
			Identity: game.ZoneCardSnapshot{CardID: card.ID, ZoneVersion: card.ZoneVersion},
			Info:     cardChoiceInfo(c.game, card.ID).Val,
		})
	}
	return arrivals, nil
}

func (c *programContext) putFoundCardsOnLibraryBottom(placements []cardprogram.CardPlacement) ([]cardprogram.FoundCard, error) {
	byOwner := make(map[game.PlayerID][]cardprogram.FoundCard)
	for _, placement := range placements {
		for _, handle := range placement.Cards {
			if from, recorded := c.foundZones[handle.Identity.CardID]; recorded && from != zone.Library {
				return nil, errors.New("bottom placement requires revealed library handles")
			}
			card, _, available := c.availableFoundCard(handle)
			if available {
				byOwner[card.Owner] = append(byOwner[card.Owner], handle)
			}
		}
	}
	for _, owner := range triggerAPNAPPlayers(c.game) {
		cards := byOwner[owner]
		if len(cards) < 2 {
			continue
		}
		options := make([]game.ChoiceOption, len(cards))
		for index, handle := range cards {
			options[index] = game.ChoiceOption{Index: index, Label: handle.Info.Name, Card: cardChoiceInfo(c.game, handle.Identity.CardID)}
		}
		request := orderChoiceRequest(owner, "Order your revealed cards on the library bottom (first nearest the top).", options)
		request.DefaultSelection = firstChoiceIndices(len(cards))
		selected, err := c.Ask(request)
		if err != nil {
			return nil, err
		}
		ordered := make([]cardprogram.FoundCard, 0, len(cards))
		for _, index := range selected {
			ordered = append(ordered, cards[index])
		}
		byOwner[owner] = ordered
	}
	var moved []cardprogram.FoundCard
	for _, owner := range triggerAPNAPPlayers(c.game) {
		library := &c.game.Players[owner].Library
		for _, handle := range byOwner[owner] {
			if library.Remove(handle.Identity.CardID) {
				library.AddToBottom(handle.Identity.CardID)
				moved = append(moved, handle)
			}
		}
	}
	return moved, nil
}
