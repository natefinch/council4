package rules

import (
	"errors"
	"slices"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func (c *programContext) TargetCards() ([]cardprogram.FoundCard, error) {
	if err := c.valid(); err != nil {
		return nil, err
	}
	var targets []cardprogram.FoundCard
	seen := make(map[game.ObjectID]bool)
	for _, target := range c.obj.Targets {
		if target.Kind != game.TargetCard || seen[target.CardID] {
			continue
		}
		card, exists := c.game.GetCardInstance(target.CardID)
		from, located := cardZone(c.game, target.CardID)
		if !exists || !located || !target.CardZoneVersionSet || card.ZoneVersion != target.CardZoneVersion {
			continue
		}
		if from != zone.Graveyard && from != zone.Exile ||
			from == zone.Exile && c.game.Players[card.Owner].Exile.IsFaceDown(card.ID) {
			return nil, errors.New("authored target handles require public graveyard or face-up exile cards")
		}
		seen[card.ID] = true
		if c.found == nil {
			c.found = make(map[game.ObjectID]game.ZoneCardSnapshot)
		}
		if c.foundZones == nil {
			c.foundZones = make(map[game.ObjectID]zone.Type)
		}
		identity := game.ZoneCardSnapshot{CardID: card.ID, ZoneVersion: card.ZoneVersion}
		c.found[card.ID], c.foundZones[card.ID] = identity, from
		delete(c.libraryEpoch, card.ID)
		targets = append(targets, cardprogram.FoundCard{Identity: identity, Info: cardChoiceInfo(c.game, card.ID).Val})
	}
	return targets, nil
}

func (c *programContext) Search(operation game.Search) (cardprogram.SearchResult, error) {
	if err := c.valid(); err != nil {
		return cardprogram.SearchResult{}, err
	}
	if !operation.Spec.RevealOnly || operation.PublishLinked != "" {
		return cardprogram.SearchResult{}, errors.New("authored search returns revealed local handles, not linked or placed cards")
	}
	if err := game.ValidatePrimitive(operation, c.specs); err != nil {
		return cardprogram.SearchResult{}, err
	}
	player, available := resolvePlayerReference(c.game, c.obj, operation.Player)
	outcome := cardprogram.SearchResult{Available: available}
	if !available {
		return outcome, nil
	}
	found := c.engine.searchLibraryRevealOnlyCards(c.game, c.obj, c.agents, c.log, player, operation.Spec, operation.Amount.Value())
	if c.libraryEpoch == nil {
		c.libraryEpoch = make(map[game.ObjectID]uint64)
	}
	for _, cardID := range found {
		card, _ := c.game.GetCardInstance(cardID)
		identity := game.ZoneCardSnapshot{CardID: cardID, ZoneVersion: card.ZoneVersion}
		if c.found == nil {
			c.found = make(map[game.ObjectID]game.ZoneCardSnapshot)
		}
		c.found[cardID] = identity
		delete(c.foundZones, cardID)
		c.libraryEpoch[cardID] = c.game.Players[player].Library.ReorderVersion()
		outcome.Cards = append(outcome.Cards, cardprogram.FoundCard{
			Identity: identity, Info: cardChoiceInfo(c.game, cardID).Val, LibraryEpoch: c.libraryEpoch[cardID],
		})
	}
	return outcome, nil
}

func (c *programContext) MoveCards(placements ...cardprogram.CardPlacement) ([]cardprogram.FoundCard, error) {
	if err := c.valid(); err != nil {
		return nil, err
	}
	seen := make(map[game.ObjectID]bool)
	for _, placement := range placements {
		if placement.Destination != zone.Graveyard && placement.Destination != zone.Hand && placement.Destination != zone.Exile {
			return nil, errors.New("authored found-card placement requires hand, graveyard or exile")
		}
		for _, member := range placement.Cards {
			minted, available := c.found[member.Identity.CardID]
			if !available || minted != member.Identity ||
				c.libraryEpoch[member.Identity.CardID] != member.LibraryEpoch || seen[member.Identity.CardID] {
				return nil, errors.New("found-card placement requires unique handles from this invocation")
			}
			seen[member.Identity.CardID] = true
		}
	}
	type preparedMove struct {
		member      cardprogram.FoundCard
		owner       game.PlayerID
		fromZone    zone.Type
		event       game.Event
		replacement zoneChangeReplacementResult
	}
	var prepared []preparedMove
	var moved []cardprogram.FoundCard
	var batch game.ObjectID
	if len(seen) > 1 {
		batch = c.game.IDGen.Next()
	}
	for _, placement := range placements {
		for _, member := range placement.Cards {
			card, exists := c.game.GetCardInstance(member.Identity.CardID)
			fromZone := zone.Library
			if from, recorded := c.foundZones[member.Identity.CardID]; recorded {
				fromZone = from
			}
			if !exists || card.ZoneVersion != member.Identity.ZoneVersion {
				continue
			}
			source, located := destinationZone(c.game, card.Owner, fromZone)
			if !located || !source.Contains(card.ID) ||
				fromZone == zone.Library && c.game.Players[card.Owner].Library.ReorderVersion() != c.libraryEpoch[card.ID] {
				continue
			}
			event := game.Event{
				Kind: game.EventZoneChanged, Controller: c.obj.Controller, Player: card.Owner,
				SourceID: c.source.CardID, SourceObjectID: c.source.ObjectID, StackObjectID: c.source.StackID,
				CardID: card.ID, FromZone: fromZone, ToZone: placement.Destination, SimultaneousID: batch,
			}
			replacement := replacementZoneChange(c.game, event)
			replacement.destination = commanderReplacementDestination(c.game, card.ID, replacement.destination)
			prepared = append(prepared, preparedMove{member: member, owner: card.Owner, fromZone: fromZone, event: event, replacement: replacement})
		}
	}
	// CR 404.3: order the actual graveyard arrivals after destination replacements.
	for _, player := range triggerAPNAPPlayers(c.game) {
		var indices []int
		var options []game.ChoiceOption
		for index, move := range prepared {
			if move.owner == player && move.replacement.destination == zone.Graveyard {
				indices = append(indices, index)
				options = append(options, game.ChoiceOption{
					Index: len(options), Label: move.member.Info.Name, Card: cardChoiceInfo(c.game, move.member.Identity.CardID),
				})
			}
		}
		if len(indices) < 2 {
			continue
		}
		request := orderChoiceRequest(player, "Order the cards entering your graveyard (first nearest the top).", options)
		for index := len(options) - 1; index >= 0; index-- {
			request.DefaultSelection = append(request.DefaultSelection, index)
		}
		selected, err := c.Ask(request)
		if err != nil {
			return nil, err
		}
		original := slices.Clone(prepared)
		// Zone.Add prepends, so commit the chosen order from bottom to top.
		for index, position := range indices {
			prepared[position] = original[indices[selected[len(selected)-1-index]]]
		}
	}
	// Replacement shuffles occur only after the entire simultaneous placement.
	// They must not hide or invalidate other members of this same event.
	shuffle := make(map[game.PlayerID]bool)
	for _, move := range prepared {
		replacement := move.replacement
		replacement.shuffleIntoLibrary = false
		if moveCardBetweenZonesAfterReplacement(c.game, move.owner, move.member.Identity.CardID,
			move.fromZone, replacement, move.event, false, batch) {
			if move.replacement.shuffleIntoLibrary && replacement.destination == zone.Library {
				shuffle[move.owner] = true
			}
			card, _ := c.game.GetCardInstance(move.member.Identity.CardID)
			if card.ZoneVersion == move.member.Identity.ZoneVersion {
				continue
			}
			identity := game.ZoneCardSnapshot{CardID: card.ID, ZoneVersion: card.ZoneVersion}
			moved = append(moved, cardprogram.FoundCard{
				Identity: identity, Info: game.ChoiceCardInfo{
					CardID: move.member.Info.CardID, Name: move.member.Info.Name, ManaValue: move.member.Info.ManaValue,
					Types: slices.Clone(move.member.Info.Types), Colors: slices.Clone(move.member.Info.Colors),
				},
			})
		}
	}
	for player := range game.NumPlayers {
		if shuffle[game.PlayerID(player)] {
			c.game.Players[player].Library.Shuffle(c.game.RNG)
		}
	}
	return moved, nil
}
