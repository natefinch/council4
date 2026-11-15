package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

// castPaidTargetedSpell ignores timing during resolution, but retains printed
// restrictions, prohibitions, cast limits and the card's mana/additional costs.
func (e *Engine) castPaidTargetedSpell(g *game.Game, controllerID game.PlayerID, cardID id.ID, fromZone zone.Type, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	source, ok := playerHoldingCastSource(g, cardID, fromZone)
	if !ok {
		return false
	}
	return e.castPaidSpellFromSource(g, source, controllerID, cardID, fromZone, agents, log)
}

func (e *Engine) castPaidSpellFromSource(g *game.Game, sourcePlayer *game.Player, controllerID game.PlayerID, cardID id.ID, fromZone zone.Type, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if sourcePlayer == nil {
		return false
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	searchCast := fromZone == zone.Library && sourcePlayer == g.Players[controllerID] && searchCastCardAllowed(g, controllerID, card)
	if !searchCast && !castSourceContains(sourcePlayer, cardID, fromZone) {
		return false
	}
	def := cardFaceOrDefault(card, game.FaceFront)
	if castFromZoneProhibited(g, controllerID, fromZone) || !freeCastFaceCastable(g, controllerID, cardID, def) {
		return false
	}
	cast := action.CastSpellAction{CardID: cardID, SourceZone: fromZone, Face: game.FaceFront}
	request := priorityCastRequest(controllerID, def, fromZone, cast,
		castPermissionsForZone(g, controllerID, cardID, fromZone, game.FaceFront), nil)
	return e.resolutionSpell(g, sourcePlayer, card, cast, request, freeCastSpell{}, true, agents, log) == actionApplied
}
