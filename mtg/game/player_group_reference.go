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
	// PlayerGroupReferenceOpponentsAttackingTriggerPlayer denotes every opponent
	// of the resolving controller who has a creature attacking the player the
	// resolving triggered ability's attack event was declared against ("Each
	// opponent attacking that player does the same" — Curse of Opulence). "That
	// player" is read from the trigger event's attack target, and only creatures
	// attacking that player directly count, so an attack on that player's
	// planeswalker or battle does not add its controller (CR 508.1). The runtime
	// derives the group from the recorded combat attackers, so it is reusable by
	// any "whenever a player is attacked" ability.
	PlayerGroupReferenceOpponentsAttackingTriggerPlayer
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

// OpponentsAttackingTriggerPlayerReference returns a reference to every opponent
// of the resolving controller attacking the player the resolving triggered
// ability's attack event was declared against ("Each opponent attacking that
// player does the same" — Curse of Opulence).
func OpponentsAttackingTriggerPlayerReference() PlayerGroupReference {
	return PlayerGroupReference{Kind: PlayerGroupReferenceOpponentsAttackingTriggerPlayer}
}

// Validate reports structural problems with a PlayerGroupReference.
func (r PlayerGroupReference) Validate() []string {
	switch r.Kind {
	case PlayerGroupReferenceOpponents, PlayerGroupReferenceAllPlayers,
		PlayerGroupReferenceTargetedPlayers,
		PlayerGroupReferenceOpponentsAttackingTriggerPlayer:
		return nil
	case PlayerGroupReferenceNone:
		return []string{"player group reference has no kind"}
	default:
		return []string{fmt.Sprintf("unknown player group reference kind %d", r.Kind)}
	}
}
