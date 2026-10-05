package rules

import "github.com/natefinch/council4/mtg/game"

func linkedCardInstance(g *game.Game, ref game.LinkedObjectRef) (*game.CardInstance, bool) {
	card, ok := g.GetCardInstance(ref.CardID)
	if !ok || (ref.CardZoneVersionSet || ref.CardZoneVersion != 0) && card.ZoneVersion != ref.CardZoneVersion {
		return nil, false
	}
	return card, true
}
