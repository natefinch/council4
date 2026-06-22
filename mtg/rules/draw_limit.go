package rules

import "github.com/natefinch/council4/mtg/game"

// cardsDrawnThisTurn reports how many cards playerID has already drawn during
// the current turn, counting the EventCardDrawn events emitted this turn.
func cardsDrawnThisTurn(g *game.Game, playerID game.PlayerID) int {
	return nextPlayerEventOrdinalThisTurn(g, game.EventCardDrawn, playerID) - 1
}

// playerDrawLimitThisTurn reports the tightest active per-turn draw limit that
// applies to playerID, if any (RuleEffectDrawLimitPerTurn; "Each opponent can't
// draw more than one card each turn.", Narset). When several limits apply the
// smallest one governs.
func playerDrawLimitThisTurn(g *game.Game, playerID game.PlayerID) (int, bool) {
	limit := 0
	limited := false
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectDrawLimitPerTurn ||
			!playerRelationMatches(effect.Controller, playerID, effect.AffectedPlayer) ||
			!actionRestrictionTurnActive(g, effect) {
			continue
		}
		if !limited || effect.DrawLimitPerTurn < limit {
			limit = effect.DrawLimitPerTurn
			limited = true
		}
	}
	return limit, limited
}

// playerAtDrawLimit reports whether an active per-turn draw limit forbids
// playerID from drawing another card right now because they have already drawn
// at least the limit this turn. The over-limit draw is replaced by drawing
// nothing (CR 120.3).
func playerAtDrawLimit(g *game.Game, playerID game.PlayerID) bool {
	limit, limited := playerDrawLimitThisTurn(g, playerID)
	if !limited {
		return false
	}
	return cardsDrawnThisTurn(g, playerID) >= limit
}
