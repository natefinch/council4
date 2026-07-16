package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// roomAbilityDoublerPermanent adds a Hama Pashar / Ruin Seeker style room-ability
// doubler the given player controls.
func roomAbilityDoublerPermanent(g *game.Game, controller game.PlayerID) {
	addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:       "Ruin Seeker",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{
			Text:        "Room abilities of dungeons you own trigger an additional time.",
			RuleEffects: []game.RuleEffect{{Kind: game.RuleEffectAdditionalTriggerForRoomAbility}},
		}},
	}})
}

func TestDoubledFinalRoomCompletesOnce(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	stockLibrary(g, game.Player1, 12)
	roomAbilityDoublerPermanent(g, game.Player1)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &ventureChoiceAgent{
		prefer: []string{"Lost Mine of Phandelver", "Goblin Lair", "Dark Pool", "Temple of Dumathoin"},
	}}
	// Path: Cave Entrance (0) -> Goblin Lair (1) -> Dark Pool (4) -> Temple of Dumathoin (6, final).
	for range 4 {
		if !engine.ventureIntoDungeon(g, game.Player1, agents, &TurnLog{}) {
			t.Fatal("venture failed")
		}
		drainDungeonStack(engine, g, agents)
	}
	if got := g.Players[game.Player1].DungeonsCompleted; got != 1 {
		t.Fatalf("DungeonsCompleted = %d, want 1 (doubled final room completes once)", got)
	}
	if got := countEvents(g, game.EventCompletedDungeon); got != 1 {
		t.Fatalf("completion events = %d, want 1", got)
	}
	if g.Players[game.Player1].Dungeon.Exists {
		t.Fatal("player still in a dungeon after completing")
	}
}

func TestDoubledFinalRoomPutsTwoAbilitiesOnStack(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	stockLibrary(g, game.Player1, 12)
	roomAbilityDoublerPermanent(g, game.Player1)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &ventureChoiceAgent{
		prefer: []string{"Lost Mine of Phandelver", "Goblin Lair", "Dark Pool", "Temple of Dumathoin"},
	}}
	// Advance to just before entering the final room.
	for range 3 {
		engine.ventureIntoDungeon(g, game.Player1, agents, &TurnLog{})
		drainDungeonStack(engine, g, agents)
	}
	// Enter the final room and put its (doubled) ability on the stack without
	// resolving. The doubler yields two copies; completion has not happened yet.
	engine.ventureIntoDungeon(g, game.Player1, agents, &TurnLog{})
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want 2 (final room ability doubled)", got)
	}
	if g.Players[game.Player1].DungeonsCompleted != 0 {
		t.Fatal("completed before the final room ability left the stack")
	}
	// Resolve both copies; completion fires exactly once.
	drainDungeonStack(engine, g, agents)
	if got := g.Players[game.Player1].DungeonsCompleted; got != 1 {
		t.Fatalf("DungeonsCompleted = %d, want 1", got)
	}
}

func TestCounteredFinalRoomStillCompletes(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	stockLibrary(g, game.Player1, 12)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &ventureChoiceAgent{
		prefer: []string{"Lost Mine of Phandelver", "Goblin Lair", "Dark Pool", "Temple of Dumathoin"},
	}}
	// Advance to just before entering the final room.
	for range 3 {
		engine.ventureIntoDungeon(g, game.Player1, agents, &TurnLog{})
		drainDungeonStack(engine, g, agents)
	}
	// Enter the final room and put its ability on the stack.
	engine.ventureIntoDungeon(g, game.Player1, agents, &TurnLog{})
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("final room ability was not put on the stack")
	}
	// Counter (remove) the final room ability before it resolves. Completion must
	// still occur because the ability left the stack (CR 309.7).
	if !counterStackObject(g, obj.ID) {
		t.Fatal("failed to counter the final room ability")
	}
	if got := g.Players[game.Player1].DungeonsCompleted; got != 1 {
		t.Fatalf("DungeonsCompleted = %d, want 1 (countered final room still completes)", got)
	}
	if g.Players[game.Player1].Dungeon.Exists {
		t.Fatal("player still in a dungeon after a countered final room ability")
	}
}

// roomIndexByName returns the index of a named room in a dungeon.
func roomIndexByName(def *game.DungeonDef, name string) int {
	for i, room := range def.Rooms {
		if room.Name == name {
			return i
		}
	}
	return -1
}

func TestBaldursGateGithyankiFinalRoomNoCreaturesCompletes(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	def, _ := game.DungeonByID(game.DungeonBaldursGateWilderness)
	githyanki := roomIndexByName(def, "Githyanki Crèche")
	// Put the player in Baldur's Gate Wilderness with every room visited except
	// Githyanki Crèche, sitting on an already-visited room, with no creatures.
	visited := allRoomsVisitedMask(def) &^ roomBit(githyanki)
	g.Players[game.Player1].Dungeon = opt.Val(game.DungeonState{
		ObjectID: g.IDGen.Next(),
		Dungeon:  game.DungeonBaldursGateWilderness,
		Room:     0,
		Visited:  visited,
	})
	agents := [game.NumPlayers]PlayerAgent{}
	// Venturing advances to the only unvisited room (Githyanki), the final room.
	if !engine.ventureIntoDungeon(g, game.Player1, agents, &TurnLog{}) {
		t.Fatal("venture into the final room failed")
	}
	drainDungeonStack(engine, g, agents)
	if g.Players[game.Player1].Dungeon.Exists {
		t.Fatal("still in the dungeon after the final room with no creatures")
	}
	if g.Players[game.Player1].DungeonsCompleted != 1 {
		t.Fatalf("DungeonsCompleted = %d, want 1", g.Players[game.Player1].DungeonsCompleted)
	}
}

func TestDroppedFinalRoomTriggerCompletesDungeon(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	// The player is in a dungeon at a room whose (synthetic) final ability targets
	// a creature. With no creatures on the battlefield the ability has no legal
	// targets and is dropped before reaching the stack (CR 603.3c); the dungeon
	// must still complete.
	objID := g.IDGen.Next()
	g.Players[game.Player1].Dungeon = opt.Val(game.DungeonState{
		ObjectID: objID,
		Dungeon:  game.DungeonLostMineOfPhandelver,
		Room:     3,
	})
	g.PendingRoomAbilities = append(g.PendingRoomAbilities, game.RoomAbilityTrigger{
		Controller:      game.Player1,
		DungeonObjectID: objID,
		Dungeon:         game.DungeonLostMineOfPhandelver,
		Room:            3,
		Final:           true,
		Ability: game.TriggeredAbility{Content: game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1, MaxTargets: 1, Constraint: "target creature",
				Allow:     game.TargetAllowPermanent,
				Selection: opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
			}},
			Sequence: []game.Instruction{{Primitive: game.AddCounter{
				Amount: game.Fixed(1), Object: game.TargetPermanentReference(0), CounterKind: counter.PlusOnePlusOne,
			}}},
		}.Ability()},
	})
	placed := engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if placed {
		t.Fatal("the no-legal-target final room ability should not have been placed on the stack")
	}
	if g.Players[game.Player1].Dungeon.Exists {
		t.Fatal("still in the dungeon after the dropped final room ability")
	}
	if g.Players[game.Player1].DungeonsCompleted != 1 {
		t.Fatalf("DungeonsCompleted = %d, want 1 (dropped final room still completes)", g.Players[game.Player1].DungeonsCompleted)
	}
}
