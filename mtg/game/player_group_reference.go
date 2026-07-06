package game

import "fmt"

// PlayerGroupReferenceKind identifies the set of players in a player-group reference.
type PlayerGroupReferenceKind int

// PlayerGroupReferenceNone is the zero value indicating no player group is set.
const (
	PlayerGroupReferenceNone PlayerGroupReferenceKind = iota
	PlayerGroupReferenceOpponents
	PlayerGroupReferenceAllPlayers
	// PlayerGroupReferenceTargetedPlayers denotes every player chosen as a target
	// of the resolving spell or ability ("any number of target players each mill
	// two cards" — Court of Cunning). The runtime resolves it to each player
	// target still legal as the effect resolves, so a group effect applies to the
	// whole chosen set at once.
	PlayerGroupReferenceTargetedPlayers
)

// PlayerGroupReference is a small sealed pure-data type describing a group of players.
// Use OpponentsReference or AllPlayersReference to construct.
type PlayerGroupReference struct {
	Kind PlayerGroupReferenceKind
}

// OpponentsReference returns a reference to all opponents of the resolving controller.
func OpponentsReference() PlayerGroupReference {
	return PlayerGroupReference{Kind: PlayerGroupReferenceOpponents}
}

// AllPlayersReference returns a reference to all players.
func AllPlayersReference() PlayerGroupReference {
	return PlayerGroupReference{Kind: PlayerGroupReferenceAllPlayers}
}

// TargetedPlayersReference returns a reference to every player targeted by the
// resolving spell or ability.
func TargetedPlayersReference() PlayerGroupReference {
	return PlayerGroupReference{Kind: PlayerGroupReferenceTargetedPlayers}
}

// Validate reports structural problems with a PlayerGroupReference.
func (r PlayerGroupReference) Validate() []string {
	switch r.Kind {
	case PlayerGroupReferenceOpponents, PlayerGroupReferenceAllPlayers,
		PlayerGroupReferenceTargetedPlayers:
		return nil
	case PlayerGroupReferenceNone:
		return []string{"player group reference has no kind"}
	default:
		return []string{fmt.Sprintf("unknown player group reference kind %d", r.Kind)}
	}
}
