package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// cardHasRebound reports whether the spell's resolving face carries the Rebound
// keyword (CR 702.88).
func cardHasRebound(spellDef *game.CardDef) bool {
	return spellDef.HasKeyword(game.Rebound)
}

// reboundExileResolvingSpell exiles a resolving Rebound spell that was cast from
// its owner's hand instead of putting it into the graveyard, and records it for
// the caster's next-upkeep free recast (CR 702.88a). It mirrors the stack-to-zone
// move used for ordinary resolution but forces the exile destination.
func (*Engine) reboundExileResolvingSpell(g *game.Game, obj *game.StackObject, card *game.CardInstance) bool {
	player, ok := playerByID(g, card.Owner)
	if !ok {
		return false
	}
	player.Exile.Add(card.ID)
	if g.ReboundCards == nil {
		g.ReboundCards = make(map[id.ID]game.ReboundCard)
	}
	g.ReboundCards[card.ID] = game.ReboundCard{Owner: card.Owner, Controller: obj.Controller}
	emitZoneChangeEvent(g, game.Event{
		SourceID:      card.ID,
		StackObjectID: obj.ID,
		Controller:    obj.Controller,
		Player:        card.Owner,
		CardID:        card.ID,
		FromZone:      zone.Stack,
		ToZone:        zone.Exile,
	})
	return true
}

// reboundCardIDsInOrder lists the tracked Rebound card IDs in a stable order.
func reboundCardIDsInOrder(g *game.Game) []id.ID {
	ids := make([]id.ID, 0, len(g.ReboundCards))
	for cardID := range g.ReboundCards {
		ids = append(ids, cardID)
	}
	slices.Sort(ids)
	return ids
}

// processReboundUpkeep resolves the rebound delayed trigger at the beginning of
// each player's upkeep (CR 702.88a): for every card this player rebounded, they
// may cast it from exile without paying its mana cost. The permission lasts only
// this upkeep, so the tracking entry is consumed whether or not the card is
// recast; a card left uncast simply remains in exile.
func (e *Engine) processReboundUpkeep(g *game.Game, playerID game.PlayerID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	for _, cardID := range reboundCardIDsInOrder(g) {
		rebound := g.ReboundCards[cardID]
		if rebound.Controller != playerID {
			continue
		}
		delete(g.ReboundCards, cardID)
		e.offerReboundCast(g, playerID, cardID, agents, log)
	}
}

// offerReboundCast lets the controller optionally cast a rebounded card from
// exile without paying its mana cost. It returns whether the card was cast.
func (e *Engine) offerReboundCast(g *game.Game, playerID game.PlayerID, cardID id.ID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	player, ok := playerByID(g, playerID)
	if !ok || !player.Exile.Contains(cardID) {
		return false
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	spellDef := cardFaceOrDefault(card, game.FaceFront)
	modes, targets, ok := firstLegalSpellCastChoice(g, playerID, spellDef)
	if !ok {
		return false
	}
	if !e.chooseMay(g, agents, playerID, "Cast "+spellDef.Name+" from exile with rebound?", log) {
		return false
	}
	targetCounts, ok := spellTargetCounts(g, playerID, spellDef, modes, targets)
	if !ok {
		panic("validated rebounded spell targets could not be segmented")
	}
	if !player.Exile.Remove(cardID) {
		return false
	}
	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackSpell,
		SourceID:     cardID,
		Face:         game.FaceFront,
		Controller:   playerID,
		Targets:      append([]game.Target(nil), targets...),
		TargetCounts: targetCounts,
		ChosenModes:  append([]int(nil), modes...),
		SourceZone:   zone.Exile,
	}
	pushSpellToStack(g, obj, game.Event{
		SourceID:        cardID,
		StackObjectID:   obj.ID,
		Controller:      playerID,
		CardID:          cardID,
		CardTypes:       cardTypes(spellDef),
		CardSupertypes:  cardSupertypes(spellDef),
		CardSubtypes:    cardSubtypes(spellDef),
		Colors:          spellColors(spellDef),
		ManaValue:       opt.Val(stackManaValue(spellDef, 0)),
		ManaSpentToCast: opt.Val(0),
		FromZone:        zone.Exile,
		ToZone:          zone.Stack,
	})
	return true
}
