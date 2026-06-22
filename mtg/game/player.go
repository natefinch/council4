package game

import (
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Player represents a single player in a Commander game with all their
// game state — life total, zones, mana pool, and commander-specific
// tracking.
type Player struct {
	// ID is this player's seat position (0–3).
	ID PlayerID

	// Name is the player's display name.
	Name string

	// --- Life and damage ---

	// Life is the player's current life total. Starts at 40 in Commander.
	Life int

	// StartingLife is the player's starting life total, captured at game setup
	// (40 in Commander). It anchors conditions phrased relative to the starting
	// life total, such as "you have at least N life more than your starting life
	// total".
	StartingLife int

	// PoisonCounters tracks the player's poison counter total. A player
	// with 10 or more poison counters loses the game (CR 704.5c).
	PoisonCounters int

	// CommanderDamage tracks combat damage received from each commander,
	// keyed by the commander's CardInstance ID. A player who has received
	// 21 or more combat damage from a single commander loses the game
	// (CR 903.10).
	CommanderDamage map[id.ID]int

	// --- Commander state ---

	// CommanderInstanceID is the CardInstance ID of this player's commander.
	CommanderInstanceID id.ID

	// CommanderCastCount is the number of times this player has cast their
	// commander from the command zone. The commander tax is +{2} generic
	// per previous cast (CR 903.8).
	CommanderCastCount int

	// CommanderMulligansTaken is the number of mulligans this player has taken
	// during the current game's Commander mulligan procedure.
	CommanderMulligansTaken int

	// --- Mana ---

	// ManaPool is the player's current available mana.
	ManaPool mana.Pool

	// ManaRiders tracks one-shot spend riders attached to units of mana in
	// ManaPool (e.g. Path of Ancestry's "When that mana is spent ..." rider).
	// One entry exists per tagged mana unit. Entries are removed when the tagged
	// mana is spent or when the pool empties between steps (CR 500.4). The slice
	// is empty for the overwhelming majority of game states, so rider handling
	// is gated behind a length check and adds no cost to ordinary play.
	ManaRiders []ManaRiderInstance

	// --- Zones ---
	// Each player has their own library, hand, graveyard, exile, and
	// command zone. The battlefield is shared (in Game.Battlefield).

	// Library is the player's draw deck — hidden, ordered.
	Library zone.Zone

	// Hand is the player's hand — hidden from opponents.
	Hand zone.Zone

	// Graveyard is the player's discard pile — public, ordered.
	Graveyard zone.Zone

	// Exile is the player's exile zone — usually public.
	Exile zone.Zone

	// CommandZone is the player's command zone (commander starts here).
	CommandZone zone.Zone

	// --- Game status ---

	// Eliminated is true if the player has lost and been removed from
	// the game.
	Eliminated bool

	// --- Designations and special states ---

	// IsMonarch is true if this player is the current monarch (draws an
	// extra card at end step).
	IsMonarch bool

	// HasInitiative is true if this player has the initiative (dungeon
	// mechanic).
	HasInitiative bool

	// HasCityBlessing is true if this player has the city's blessing
	// (ascend mechanic — gained when controlling 10+ permanents).
	HasCityBlessing bool

	// RingLevel tracks the player's level of "The Ring tempts you"
	// (0 = not tempted, 1–4 = ring levels).
	RingLevel int

	// EnergyCounters tracks the player's energy counter total (Kaladesh
	// mechanic).
	EnergyCounters int

	// ExperienceCounters tracks the player's experience counter total
	// (Commander 2015 mechanic).
	ExperienceCounters int
	Speed              int
	SpeedIncreasedTurn int

	// PowerBracket and PowerLevel are optional deck metadata carried from setup
	// for later simulation/reporting. They do not affect rules behavior.
	PowerBracket string
	PowerLevel   int
}

// NewPlayer creates a new player with the given seat and name,
// initialized for a Commander game (40 life, empty zones).
func NewPlayer(seat PlayerID, name string) *Player {
	return &Player{
		ID:              seat,
		Name:            name,
		Life:            40,
		StartingLife:    40,
		CommanderDamage: make(map[id.ID]int),
		ManaPool:        mana.NewPool(),
		Library:         zone.New(zone.Library),
		Hand:            zone.New(zone.Hand),
		Graveyard:       zone.New(zone.Graveyard),
		Exile:           zone.New(zone.Exile),
		CommandZone:     zone.New(zone.Command),
	}
}

// CommanderTax returns the additional generic mana that must be paid
// to cast this player's commander from the command zone, based on
// how many times it has been cast previously (CR 903.8).
func (p *Player) CommanderTax() int {
	return p.CommanderCastCount * 2
}

// IsAlive reports whether this player is still in the game.
func (p *Player) IsAlive() bool {
	return !p.Eliminated
}

// HasLethalPoison reports whether this player has enough poison counters
// to lose the game (10 or more, CR 704.5c).
func (p *Player) HasLethalPoison() bool {
	return p.PoisonCounters >= 10
}

// HasLethalCommanderDamage reports whether this player has received
// 21 or more combat damage from any single commander (CR 903.10).
func (p *Player) HasLethalCommanderDamage() bool {
	for _, dmg := range p.CommanderDamage {
		if dmg >= 21 {
			return true
		}
	}
	return false
}
