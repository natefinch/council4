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
	// PlayerGroupReferenceOpponentsDealtCombatDamageThisGameByNamed denotes every
	// opponent of the resolving controller who has been dealt combat damage this
	// game by a creature with the name recorded in Name ("each opponent dealt
	// combat damage this game by a creature named Gollum, Obsessed Stalker loses
	// life equal to the amount of life you gained this turn"). The runtime scans
	// the full game event log for combat damage dealt to a player by a source
	// whose name matches Name, so it is reusable by any ability keyed on that
	// wording. Only combat damage dealt to players counts; damage to permanents
	// does not qualify.
	PlayerGroupReferenceOpponentsDealtCombatDamageThisGameByNamed
)

// PlayerGroupReference is a small sealed pure-data type describing a group of players.
// Use OpponentsReference or AllPlayersReference to construct.
type PlayerGroupReference struct {
	Kind PlayerGroupReferenceKind
	// Name is the required creature name for
	// PlayerGroupReferenceOpponentsDealtCombatDamageThisGameByNamed. It must be
	// empty for every other kind.
	Name string
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

// OpponentsDealtCombatDamageThisGameByNamedReference returns a reference to every
// opponent of the resolving controller who has been dealt combat damage this game
// by a creature named name ("each opponent dealt combat damage this game by a
// creature named Gollum, Obsessed Stalker loses ..." — Gollum, Obsessed Stalker).
func OpponentsDealtCombatDamageThisGameByNamedReference(name string) PlayerGroupReference {
	return PlayerGroupReference{
		Kind: PlayerGroupReferenceOpponentsDealtCombatDamageThisGameByNamed,
		Name: name,
	}
}

// Validate reports structural problems with a PlayerGroupReference.
func (r PlayerGroupReference) Validate() []string {
	switch r.Kind {
	case PlayerGroupReferenceOpponents, PlayerGroupReferenceAllPlayers,
		PlayerGroupReferenceTargetedPlayers,
		PlayerGroupReferenceOpponentsAttackingTriggerPlayer:
		if r.Name != "" {
			return []string{fmt.Sprintf("player group reference kind %d must not set a name", r.Kind)}
		}
		return nil
	case PlayerGroupReferenceOpponentsDealtCombatDamageThisGameByNamed:
		if r.Name == "" {
			return []string{"player group reference for opponents dealt combat damage this game by named creature requires a name"}
		}
		return nil
	case PlayerGroupReferenceNone:
		return []string{"player group reference has no kind"}
	default:
		return []string{fmt.Sprintf("unknown player group reference kind %d", r.Kind)}
	}
}
