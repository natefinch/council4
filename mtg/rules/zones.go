package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

func removePermanentFromBattlefield(g *game.Game, objectID id.ID) *game.Permanent {
	if g == nil {
		return nil
	}
	for i, permanent := range g.Battlefield {
		if permanent == nil || permanent.ObjectID != objectID {
			continue
		}
		g.Battlefield = append(g.Battlefield[:i], g.Battlefield[i+1:]...)
		return permanent
	}
	return nil
}

func movePermanentToZone(g *game.Game, permanent *game.Permanent, destination game.ZoneType) bool {
	if g == nil || permanent == nil {
		return false
	}
	removed := removePermanentFromBattlefield(g, permanent.ObjectID)
	if removed == nil {
		return false
	}
	if removed.Token {
		return true
	}
	zone := destinationZone(g, removed.Owner, destination)
	if zone == nil {
		return false
	}
	zone.Add(removed.CardInstanceID)
	return true
}

func destroyPermanent(g *game.Game, objectID id.ID) (*game.Permanent, bool) {
	permanent := permanentByObjectID(g, objectID)
	if permanent == nil {
		return nil, false
	}
	if hasKeyword(g, permanent, game.Indestructible) {
		return nil, false
	}
	if !movePermanentToZone(g, permanent, game.ZoneGraveyard) {
		return nil, false
	}
	return permanent, true
}

func destinationZone(g *game.Game, owner game.PlayerID, destination game.ZoneType) *game.Zone {
	if g == nil || owner < 0 || int(owner) >= len(g.Players) {
		return nil
	}
	player := g.Players[owner]
	if player == nil {
		return nil
	}
	switch destination {
	case game.ZoneLibrary:
		return &player.Library
	case game.ZoneHand:
		return &player.Hand
	case game.ZoneGraveyard:
		return &player.Graveyard
	case game.ZoneExile:
		return &player.Exile
	case game.ZoneCommand:
		return &player.CommandZone
	default:
		return nil
	}
}
