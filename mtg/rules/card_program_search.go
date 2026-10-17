package rules

import (
	"errors"
	"slices"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func (c *programContext) mintFoundCard(card *game.CardInstance, from zone.Type) cardprogram.FoundCard {
	if c.found == nil {
		c.found = make(map[game.ObjectID]game.ZoneCardSnapshot)
	}
	if c.foundZones == nil {
		c.foundZones = make(map[game.ObjectID]zone.Type)
	}
	identity := game.ZoneCardSnapshot{CardID: card.ID, ZoneVersion: card.ZoneVersion}
	c.found[card.ID], c.foundZones[card.ID] = identity, from
	handle := cardprogram.FoundCard{Identity: identity, Info: cardChoiceInfo(c.game, card.ID).Val}
	if from == zone.Library {
		if c.libraryEpoch == nil {
			c.libraryEpoch = make(map[game.ObjectID]uint64)
		}
		handle.LibraryEpoch = c.game.Players[card.Owner].Library.ReorderVersion()
		c.libraryEpoch[card.ID] = handle.LibraryEpoch
	} else {
		delete(c.libraryEpoch, card.ID)
	}
	return handle
}

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
		targets = append(targets, c.mintFoundCard(card, from))
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
	for _, cardID := range found {
		card, _ := c.game.GetCardInstance(cardID)
		outcome.Cards = append(outcome.Cards, c.mintFoundCard(card, zone.Library))
	}
	return outcome, nil
}

func (c *programContext) MoveCards(placements ...cardprogram.CardPlacement) ([]cardprogram.FoundCard, error) {
	if err := c.valid(); err != nil {
		return nil, err
	}
	seen := make(map[game.ObjectID]bool)
	entry, bottom, ordinary := false, false, false
	for _, placement := range placements {
		modifiedEntry := placement.Recipient.Exists || len(placement.ContinuousEffects) > 0 ||
			len(placement.EntryCounters) > 0 || placement.EntryTapped
		if modifiedEntry {
			if placement.Destination != zone.Battlefield || len(placements) != 1 {
				return nil, errors.New("modified battlefield entry requires one battlefield placement")
			}
			if err := game.ValidatePrimitive(game.PutOnBattlefield{
				Source:    game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource}),
				Recipient: placement.Recipient, ContinuousEffects: placement.ContinuousEffects,
				EntryCounters: placement.EntryCounters, EntryTapped: placement.EntryTapped,
			}, c.specs); err != nil {
				return nil, err
			}
			for _, counter := range placement.EntryCounters {
				if !counter.Kind.Valid() || counter.Kind.PlayerOnly() {
					return nil, errors.New("battlefield entry requires permanent counter kinds")
				}
			}
			validation := &game.CardDef{CardFace: game.CardFace{
				Name:            "authored battlefield entry",
				StaticAbilities: []game.StaticAbility{{ContinuousEffects: placement.ContinuousEffects}},
			}}
			if issues := game.ValidateCardDef(validation); len(issues) > 0 {
				return nil, errors.New("invalid battlefield entry effect: " + issues[0].Message)
			}
			if err := c.engine.programs.ValidateCard(validation); err != nil {
				return nil, err
			}
		}
		if placement.LibraryBottom && placement.Destination != zone.Library ||
			placement.Destination == zone.Library && !placement.LibraryBottom {
			return nil, errors.New("authored library placement requires an explicit bottom position")
		}
		switch placement.Destination {
		case zone.Battlefield:
			entry = true
		case zone.Library:
			bottom = true
		case zone.Graveyard, zone.Hand, zone.Exile:
			ordinary = true
		default:
			return nil, errors.New("authored found-card placement has an unsupported destination")
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
	if entry && (bottom || ordinary) || bottom && ordinary {
		return nil, errors.New("battlefield entry and library reordering require their own simultaneous placement")
	}
	previous, _ := c.game.ChoiceContext()
	c.engine.setReplacementChoiceContext(c.game, c.agents, c.log)
	defer c.game.SetChoiceContext(previous)
	if entry {
		return c.putFoundCardsOnBattlefield(placements)
	}
	if bottom {
		return c.putFoundCardsOnLibraryBottom(placements)
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
			card, fromZone, available := c.availableFoundCard(member)
			if !available {
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
			move.fromZone, replacement, move.event, zoneMoveOptions{}, batch) {
			if move.replacement.shuffleIntoLibrary && replacement.destination == zone.Library {
				shuffle[move.owner] = true
			}
			card, _ := c.game.GetCardInstance(move.member.Identity.CardID)
			if card.ZoneVersion == move.member.Identity.ZoneVersion {
				continue
			}
			identity := game.ZoneCardSnapshot{CardID: card.ID, ZoneVersion: card.ZoneVersion}
			if replacement.destination == zone.Graveyard || replacement.destination == zone.Exile &&
				!c.game.Players[card.Owner].Exile.IsFaceDown(card.ID) {
				moved = append(moved, c.mintFoundCard(card, replacement.destination))
			} else {
				moved = append(moved, cardprogram.FoundCard{Identity: identity, Info: move.member.Info})
			}
		}
	}
	for player := range game.NumPlayers {
		if shuffle[game.PlayerID(player)] {
			c.game.Players[player].Library.Shuffle(c.game.RNG)
		}
	}
	return moved, nil
}
