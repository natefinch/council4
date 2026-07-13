package game

import "math/bits"

// PlayerSet is a comparable bitmask over the seats in a game, one bit per
// PlayerID. It records a set of players in a value that can live inside the
// comparable InstructionResolutionResult (and thus the cloneable resolution-
// results map) without a slice. It backs the accepted-actor publication of a
// group offer: which players accepted, how many, and (for future group-offer
// consequences) their identities.
type PlayerSet uint8

// NewPlayerSet returns a PlayerSet containing exactly the given players.
func NewPlayerSet(players ...PlayerID) PlayerSet {
	var set PlayerSet
	for _, player := range players {
		set = set.With(player)
	}
	return set
}

// With returns the set with player added. Out-of-range seats are ignored.
func (s PlayerSet) With(player PlayerID) PlayerSet {
	if player < 0 || int(player) >= NumPlayers {
		return s
	}
	return s | PlayerSet(1<<uint(player))
}

// Contains reports whether player is in the set.
func (s PlayerSet) Contains(player PlayerID) bool {
	if player < 0 || int(player) >= NumPlayers {
		return false
	}
	return s&PlayerSet(1<<uint(player)) != 0
}

// Count returns the number of players in the set.
func (s PlayerSet) Count() int {
	return bits.OnesCount8(uint8(s))
}

// Members returns the players in the set in seat order.
func (s PlayerSet) Members() []PlayerID {
	var members []PlayerID
	for seat := range NumPlayers {
		player := PlayerID(seat)
		if s.Contains(player) {
			members = append(members, player)
		}
	}
	return members
}
