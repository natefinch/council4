package game

import "github.com/natefinch/council4/mtg/game/id"

// DungeonID identifies a specific dungeon card a player can venture into
// (CR 309). Dungeons are shared objects that exist outside the game until a
// player ventures into one; each player is in at most one dungeon at a time.
type DungeonID int

// Dungeon identifiers. DungeonNone is the zero value used when a player is not
// in any dungeon.
const (
	DungeonNone DungeonID = iota
	// DungeonLostMineOfPhandelver, DungeonTombOfAnnihilation, and
	// DungeonDungeonOfTheMadMage are the three ordinary dungeons a player may
	// choose the first time they "venture into the dungeon" (CR 309.5).
	DungeonLostMineOfPhandelver
	DungeonTombOfAnnihilation
	DungeonDungeonOfTheMadMage
	// DungeonUndercity is entered only through the "venture into Undercity"
	// action (the Undercity dungeon card reads "You can't enter this dungeon
	// unless you 'venture into Undercity.'").
	DungeonUndercity
	// DungeonBaldursGateWilderness is a free-traversal dungeon a player may choose
	// the first time they venture into the dungeon.
	DungeonBaldursGateWilderness
)

// RoomDef is one immutable room within a dungeon graph (CR 309.4). Ability is
// the room's effect, which resolves as a triggered ability using the normal
// stack machinery when a venturing player enters the room, so room-ability
// trigger multipliers (RuleEffectAdditionalTriggerForRoomAbility) apply. Next
// holds the indices of the rooms this room leads to; an empty Next marks a final
// room, and the venturing player completes the dungeon after that room's ability
// leaves the stack (CR 309.7).
type RoomDef struct {
	// Name is the room's name (e.g. "Cave Entrance").
	Name string

	// Ability is the room's effect, resolved as a triggered ability. Dungeon
	// completion is recorded by the runtime when a final room's ability leaves the
	// stack (a stack-object marker), so a final room's Ability is just its effect.
	Ability AbilityContent

	// Next lists the indices (into the owning DungeonDef.Rooms) of the rooms this
	// room leads to. When it holds more than one room the venturing player
	// chooses which to advance to; when empty the room is a final room.
	Next []int
}

// Final reports whether this room is a final room of its dungeon (no room leads
// out of it), meaning entering it and resolving its ability completes the
// dungeon.
func (r RoomDef) Final() bool { return len(r.Next) == 0 }

// DungeonDef is an immutable dungeon graph definition (CR 309.3). Rooms[0] is
// the entrance (the topmost room) a player enters first when they venture into
// this dungeon.
type DungeonDef struct {
	// ID identifies which dungeon this is.
	ID DungeonID

	// Name is the dungeon card's name (e.g. "Lost Mine of Phandelver").
	Name string

	// FreeTraversal reports a dungeon a player traverses freely (Baldur's Gate
	// Wilderness): each venture enters any room the player has not yet visited
	// rather than a connected room, and the dungeon is completed once every room
	// has been visited. Rooms in a free-traversal dungeon carry no Next edges.
	FreeTraversal bool

	// Rooms holds every room, entrance first. Room indices are stable and are
	// referenced by RoomDef.Next and by a player's live DungeonState.Room.
	Rooms []RoomDef
}

// Room returns the room at the given index, reporting whether the index is
// valid for this dungeon.
func (d *DungeonDef) Room(index int) (RoomDef, bool) {
	if d == nil || index < 0 || index >= len(d.Rooms) {
		return RoomDef{}, false
	}
	return d.Rooms[index], true
}

// DungeonState is a player's live position in a dungeon (CR 309.4). A player is
// in at most one dungeon at a time, so a player has a single optional
// DungeonState. Every field is a value type, so it clones by struct copy with
// the rest of the player.
type DungeonState struct {
	// ObjectID is the synthetic object id assigned to the dungeon card when the
	// player entered it. It is the source of every room-ability triggered ability
	// so room-ability trigger multipliers can identify the dungeon and its owner.
	ObjectID id.ID

	// Dungeon identifies which dungeon graph the player is in.
	Dungeon DungeonID

	// Room is the index of the player's current room within the dungeon's Rooms.
	Room int

	// Visited is a bitmask of the rooms this player has entered in this dungeon
	// (bit i set means room i has been visited). It is set for every dungeon but
	// consulted only for free-traversal dungeons, whose venture chooses any
	// unvisited room and whose completion occurs once every room's bit is set.
	Visited uint32
}

// DungeonRoomMark tags a room-ability stack object so dungeon completion can be
// recorded exactly once when a final room's ability leaves the stack (resolved,
// countered, or removed), including when the ability is copied or doubled.
type DungeonRoomMark struct {
	// Owner is the venturing player whose dungeon this room belongs to.
	Owner PlayerID

	// ObjectID is the dungeon entry's synthetic object id (DungeonState.ObjectID),
	// so completion only fires for the specific dungeon entry that queued the
	// ability, never a later re-entry that reused the room index.
	ObjectID id.ID

	// Room is the room's index within the dungeon.
	Room int

	// Final reports that entering this room completed the dungeon's rooms (the
	// terminal room of an ordinary dungeon, or the last unvisited room of a
	// free-traversal dungeon), so its ability leaving the stack completes the
	// dungeon.
	Final bool
}

// RoomAbilityTrigger is a queued dungeon room-entry triggered ability (CR
// 309.6) waiting to be put on the stack. A venture creates one when a player
// enters a room; the trigger pass drains it, turning it into an ordinary
// event-backed triggered ability whose source is the dungeon so room-ability
// trigger multipliers apply. It mirrors ReflexiveTrigger, but a room ability
// captures nothing from a creating spell: its content is self-contained.
type RoomAbilityTrigger struct {
	// Controller is the venturing player. It becomes the resolving ability's
	// controller, so "you", "your", and targeting relations resolve to them.
	Controller PlayerID

	// DungeonObjectID is the synthetic object id of the dungeon the player is in
	// (their DungeonState.ObjectID). It is the room ability's source, letting
	// room-ability trigger multipliers identify the dungeon and its owner.
	DungeonObjectID id.ID

	// Dungeon and Room identify which room's ability this is, for observation and
	// completion bookkeeping.
	Dungeon DungeonID
	Room    int

	// Final reports that entering this room completed the dungeon's rooms, so the
	// room ability's stack object is marked to complete the dungeon when it leaves
	// the stack.
	Final bool

	// Ability is the room's effect, resolved from the stack.
	Ability TriggeredAbility
}
