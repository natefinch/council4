package game

import "github.com/natefinch/council4/mtg/game/id"

type attachedDeathDepartureKey struct {
	hostID   id.ID
	sourceID id.ID
}

type attachedDeathDeparture struct {
	card  ZoneCardSnapshot
	owner PlayerID
}

// RememberAttachedDeathDeparture retains a producer-certified completed Aura
// departure until its original host's captured death watchers are queued.
func (g *Game) RememberAttachedDeathDeparture(hostID, sourceID id.ID, card ZoneCardSnapshot, owner PlayerID) {
	key := attachedDeathDepartureKey{hostID: hostID, sourceID: sourceID}
	if g.attachedDeathDepartures == nil {
		g.attachedDeathDepartures = make(map[attachedDeathDepartureKey]attachedDeathDeparture)
	}
	if _, exists := g.attachedDeathDepartures[key]; !exists {
		g.attachedDeathDepartures[key] = attachedDeathDeparture{card: card, owner: owner}
	}
}

// AttachedDeathDeparture reads a certificate without consuming other captured
// abilities of the same source and host occurrence.
func (g *Game) AttachedDeathDeparture(hostID, sourceID id.ID) (ZoneCardSnapshot, PlayerID, bool) {
	departure, ok := g.attachedDeathDepartures[attachedDeathDepartureKey{hostID: hostID, sourceID: sourceID}]
	return departure.card, departure.owner, ok
}

// ForgetAttachedDeathDepartures releases certificates after the host event has
// been consumed. Queued stack objects retain their own immutable zone witness.
func (g *Game) ForgetAttachedDeathDepartures(hostID id.ID) {
	for key := range g.attachedDeathDepartures {
		if key.hostID == hostID {
			delete(g.attachedDeathDepartures, key)
		}
	}
}
