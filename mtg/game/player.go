package game

import (
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
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

	// CantBecomeMonarchThisTurn blocks this player from becoming the monarch for
	// the rest of the turn ("You can't become the monarch this turn.", Jared
	// Carthalion). It is cleared as each turn begins.
	CantBecomeMonarchThisTurn bool

	// HasInitiative is true if this player has the initiative (dungeon
	// mechanic). At most one living player has the initiative at a time (CR
	// 720). Taking, and combat-damage transfer of, the initiative each cause the
	// holder to venture into Undercity.
	HasInitiative bool

	// Dungeon is the player's live position in a dungeon (CR 309.4), unset when
	// the player is not in any dungeon. A player is in at most one dungeon at a
	// time; venturing enters a dungeon or advances the current one, and
	// completing a dungeon clears this state.
	Dungeon opt.V[DungeonState]

	// DungeonsCompleted is the number of dungeons this player has completed this
	// game (CR 309.7). It only ever increases, so cards that ask whether the
	// player "has completed a dungeon" (Imoen, Mystic Trickster) are satisfied
	// once it is at least one.
	DungeonsCompleted int

	// HasCityBlessing is true if this player has the city's blessing
	// (ascend mechanic — gained when controlling 10+ permanents).
	HasCityBlessing bool

	// RingLevel tracks the player's level of "The Ring tempts you"
	// (0 = not tempted, 1–4 = ring levels).
	RingLevel int

	// RingBearerID is the ObjectID of the permanent this player has designated
	// as their Ring-bearer (CR 701.51), or the zero ID when they have no
	// Ring-bearer. ObjectID identifies the specific battlefield permanent (it is
	// unique per game object and nonzero for tokens), so the designation does not
	// carry across zone changes. The Ring's level abilities apply to this
	// permanent.
	RingBearerID id.ID

	// RingTemptedCount is the number of times the Ring has tempted this player
	// this game. Cards reference it ("the Ring has tempted you two or more times
	// this game").
	RingTemptedCount int

	// EnergyCounters tracks the player's energy counter total (Kaladesh
	// mechanic).
	EnergyCounters int

	// ExperienceCounters tracks the player's experience counter total
	// (Commander 2015 mechanic).
	ExperienceCounters int

	// Speed tracks the player's speed for the "Start your engines!" subsystem
	// (CR 702.179). It starts at 0 (no speed), is set to 1 when the player
	// gains speed, increases by at most 1 on each of the player's turns the
	// first time an opponent loses life that turn, and is capped at 4.
	Speed int

	// SpeedIncreasedTurn records the turn number on which Speed was last
	// increased by the once-per-turn opponent-life-loss rule, so the increase
	// happens at most once per the player's turn (CR 702.179c).
	SpeedIncreasedTurn int

	// TurnsTaken counts how many turns this player has begun during the game,
	// incremented as each of the player's turns starts. It counts the player's
	// own turns rather than the global turn number, so in a multiplayer game a
	// later seat's first turn is TurnsTaken 1 even though it is a high global
	// turn number, and an extra turn a player takes increments their own count.
	// It backs the "your first, second, or third turn of the game" turn-ordinal
	// condition (Starting Town). It is zero until the player takes their first
	// turn.
	TurnsTaken int

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
