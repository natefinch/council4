package game

import "github.com/natefinch/council4/mtg/game/id"

// ObjectID is a unique identifier for a game object (permanent, stack object,
// card instance, etc.). It is an alias for id.ID to keep all game types in
// one package while using the leaf id package for generation.
type ObjectID = id.ID

// PlayerID identifies a player by their seat position (0–3).
type PlayerID int

const (
	Player1 PlayerID = 0
	Player2 PlayerID = 1
	Player3 PlayerID = 2
	Player4 PlayerID = 3
)

// NumPlayers is the number of players in a Commander game.
const NumPlayers = 4
