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

const reboundReplacementLabel = "Rebound (CR 702.88a)"

// reboundAppliesToResolvingSpell reports whether a resolving spell's rebound
// replacement can apply to its move to the graveyard. "If this spell was cast
// from your hand" (CR 702.88a) refers to the spell's controller, so a spell
// whose control changed after it was cast doesn't rebound for its new
// controller, and copies are never cast.
func reboundAppliesToResolvingSpell(obj *game.StackObject, spellDef *game.CardDef) bool {
	return !obj.Copy && obj.SourceZone == zone.Hand && stackSpellCaster(obj) == obj.Controller && cardHasRebound(spellDef)
}

// reboundReplacement is rebound's "instead of putting it into your graveyard
// as it resolves, exile it" (CR 702.88a). It is a static ability of the spell,
// not a self-replacement, so it competes with other replacements of the same
// stack-to-graveyard move and the spell's controller chooses the order (CR
// 616.1). Its ID is only a local key.
func reboundReplacement(obj *game.StackObject, card *game.CardInstance) game.ReplacementEffect {
	return game.ReplacementEffect{
		ID: obj.ID, Controller: obj.Controller, SourceCardID: card.ID,
		Description: reboundReplacementLabel, MatchEvent: game.EventZoneChanged,
		MatchFromZone: true, FromZone: zone.Stack,
		MatchToZone: true, ToZone: zone.Graveyard, ReplaceToZone: zone.Exile,
	}
}

// recordReboundExile schedules rebound's next-upkeep free recast for the
// spell's controller (CR 702.88a). Callers record it only when the rebound
// replacement was applied and the card actually arrived in its owner's exile;
// a card exiled by another redirect (Dauthi Voidwalker, Rest in Peace) or sent
// elsewhere gets no delayed cast.
func recordReboundExile(g *game.Game, obj *game.StackObject, card *game.CardInstance) {
	if g.ReboundCards == nil {
		g.ReboundCards = make(map[id.ID]game.ReboundCard)
	}
	g.ReboundCards[card.ID] = game.ReboundCard{Owner: card.Owner, Controller: obj.Controller}
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
// exile without paying its mana cost. The cast follows the alternative-cost
// rules (CR 702.88b, 601.2b, 601.2f-h): additional costs, cost increases and
// reductions still apply to the free alternative cost, so an unpayable total
// cost leaves the card in exile. It returns whether the card was cast.
func (e *Engine) offerReboundCast(g *game.Game, playerID game.PlayerID, cardID id.ID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	player, ok := playerByID(g, playerID)
	if !ok || !player.Exile.Contains(cardID) {
		return false
	}
	return e.chooseAndCastFreeSpellFromSourceMay(g, player, playerID, cardID, zone.Exile, game.Selection{}, opt.V[int]{}, false, true, agents, log)
}
