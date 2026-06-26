package rules

import "github.com/natefinch/council4/mtg/game"

// tryDrawCardDigReplacement applies the first active draw-replacement dig owned
// by playerID (CR 614), replacing an imminent draw with looking at the top N
// cards of their library, putting the take count into their hand, and routing
// the rest to the recorded remainder (Underrealm Lich). It reports true when a
// dig replacement applied so the caller skips the draw; a missing or empty
// library still counts as replaced because the draw event itself was replaced.
func (e *Engine) tryDrawCardDigReplacement(g *game.Game, playerID game.PlayerID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	for i := range g.ReplacementEffects {
		replacement := &g.ReplacementEffects[i]
		if replacement.DrawCardDigLook <= 0 {
			continue
		}
		if !replacementSourceIsActive(g, replacement) {
			continue
		}
		if replacementCurrentController(g, replacement) != playerID {
			continue
		}
		e.digCards(g, agents, log, nil, playerID, replacement.DrawCardDigLook, replacement.DrawCardDigTake, replacement.DrawCardDigRemainder, digFilter{})
		return true
	}
	return false
}
