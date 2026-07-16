package game

import "testing"

// dungeonRoomExpectation lists a dungeon's rooms in order with each room's
// expected next-room indices, taken verbatim from the dungeon card's Scryfall
// Oracle text ("Leads to: ..."). A nil Next marks a final room.
type dungeonRoomExpectation struct {
	name string
	next []int
}

func TestDungeonGraphsMatchOracleText(t *testing.T) {
	cases := map[DungeonID][]dungeonRoomExpectation{
		DungeonLostMineOfPhandelver: {
			{"Cave Entrance", []int{1, 2}},
			{"Goblin Lair", []int{3, 4}},
			{"Mine Tunnels", []int{4, 5}},
			{"Storeroom", []int{6}},
			{"Dark Pool", []int{6}},
			{"Fungi Cavern", []int{6}},
			{"Temple of Dumathoin", nil},
		},
		DungeonTombOfAnnihilation: {
			{"Trapped Entry", []int{1, 3}},
			{"Veils of Fear", []int{2}},
			{"Sandfall Cell", []int{4}},
			{"Oubliette", []int{4}},
			{"Cradle of the Death God", nil},
		},
		DungeonDungeonOfTheMadMage: {
			{"Yawning Portal", []int{1}},
			{"Dungeon Level", []int{2, 3}},
			{"Goblin Bazaar", []int{4}},
			{"Twisted Caverns", []int{4}},
			{"Lost Level", []int{5, 6}},
			{"Runestone Caverns", []int{7}},
			{"Muiral's Graveyard", []int{7}},
			{"Deep Mines", []int{8}},
			{"Mad Wizard's Lair", nil},
		},
		DungeonUndercity: {
			{"Secret Entrance", []int{1, 2}},
			{"Forge", []int{3, 4}},
			{"Lost Well", []int{4, 5}},
			{"Trap!", []int{6}},
			{"Arena", []int{6, 7}},
			{"Stash", []int{7}},
			{"Archives", []int{8}},
			{"Catacombs", []int{8}},
			{"Throne of the Dead Three", nil},
		},
		DungeonBaldursGateWilderness: {
			{"Crash Landing", nil},
			{"Goblin Camp", nil},
			{"Emerald Grove", nil},
			{"Auntie's Teahouse", nil},
			{"Defiled Temple", nil},
			{"Mountain Pass", nil},
			{"Ebonlake Grotto", nil},
			{"Grymforge", nil},
			{"Githyanki Crèche", nil},
			{"Last Light Inn", nil},
			{"Reithwin Tollhouse", nil},
			{"Moonrise Towers", nil},
			{"Gauntlet of Shar", nil},
			{"Balthazar's Lab", nil},
			{"Circus of the Last Days", nil},
			{"Undercity Ruins", nil},
			{"Steel Watch Foundry", nil},
			{"Ansur's Sanctum", nil},
			{"Temple of Bhaal", nil},
		},
	}
	for dungeonID, rooms := range cases {
		def, ok := DungeonByID(dungeonID)
		if !ok {
			t.Fatalf("dungeon %d not registered", dungeonID)
		}
		if len(def.Rooms) != len(rooms) {
			t.Fatalf("%s: %d rooms, want %d", def.Name, len(def.Rooms), len(rooms))
		}
		finalRooms := 0
		for i, want := range rooms {
			room := def.Rooms[i]
			if room.Name != want.name {
				t.Errorf("%s room %d name = %q, want %q", def.Name, i, room.Name, want.name)
			}
			if len(room.Next) != len(want.next) {
				t.Errorf("%s room %q next = %v, want %v", def.Name, room.Name, room.Next, want.next)
			} else {
				for j := range want.next {
					if room.Next[j] != want.next[j] {
						t.Errorf("%s room %q next[%d] = %d, want %d", def.Name, room.Name, j, room.Next[j], want.next[j])
					}
				}
			}
			for _, n := range room.Next {
				if n < 0 || n >= len(def.Rooms) {
					t.Errorf("%s room %q has out-of-range next index %d", def.Name, room.Name, n)
				}
			}
			if room.Final() {
				finalRooms++
				if !def.FreeTraversal && i != len(rooms)-1 {
					t.Errorf("%s: final room %q is not the last room", def.Name, room.Name)
				}
			}
		}
		// A free-traversal dungeon (Baldur's Gate Wilderness) has no Next edges on
		// any room; its completion is by the visited-room set, verified separately.
		if !def.FreeTraversal && finalRooms != 1 {
			t.Errorf("%s: %d final rooms, want exactly 1", def.Name, finalRooms)
		}
	}
}

func TestDungeonRegistryAndOrdinaryDungeons(t *testing.T) {
	ordinary := OrdinaryDungeons()
	if len(ordinary) != 4 {
		t.Fatalf("OrdinaryDungeons() = %v, want 4 dungeons", ordinary)
	}
	for _, id := range ordinary {
		if id == DungeonUndercity {
			t.Fatal("OrdinaryDungeons() includes Undercity, which is enterable only via venture into Undercity")
		}
		if _, ok := DungeonByID(id); !ok {
			t.Fatalf("ordinary dungeon %d not registered", id)
		}
	}
	if _, ok := DungeonByID(DungeonUndercity); !ok {
		t.Fatal("Undercity not registered")
	}
	if _, ok := DungeonByID(DungeonNone); ok {
		t.Fatal("DungeonNone should not resolve to a dungeon")
	}
}

// TestDungeonRoomAbilitiesValidate checks that every room ability is composed of
// primitives that validate against the room's own target specs, so a malformed
// target reference or primitive is caught statically.
func TestDungeonRoomAbilitiesValidate(t *testing.T) {
	for _, def := range dungeonRegistry {
		for _, room := range def.Rooms {
			for _, mode := range room.Ability.Modes {
				for i := range mode.Sequence {
					instr := mode.Sequence[i]
					if instr.Primitive == nil {
						t.Errorf("%s room %q has a nil primitive", def.Name, room.Name)
						continue
					}
					if err := instr.Primitive.validatePrimitive(mode.Targets, true); err != nil {
						t.Errorf("%s room %q primitive %T invalid: %v", def.Name, room.Name, instr.Primitive, err)
					}
				}
			}
		}
	}
}

// TestDungeonCompletionStructure verifies the completion model: an ordinary
// dungeon has exactly one final (terminal) room, and it is the last room; a
// free-traversal dungeon has no terminal edges (every room is a leaf) and
// completes via the visited-room set instead.
func TestDungeonCompletionStructure(t *testing.T) {
	for _, def := range dungeonRegistry {
		if def.FreeTraversal {
			for _, room := range def.Rooms {
				if !room.Final() {
					t.Errorf("%s free-traversal room %q must have no Next edges", def.Name, room.Name)
				}
			}
			continue
		}
		finalRooms := 0
		for i, room := range def.Rooms {
			if room.Final() {
				finalRooms++
				if i != len(def.Rooms)-1 {
					t.Errorf("%s: final room %q is not the last room", def.Name, room.Name)
				}
			}
		}
		if finalRooms != 1 {
			t.Errorf("%s: %d final rooms, want exactly 1", def.Name, finalRooms)
		}
	}
}
