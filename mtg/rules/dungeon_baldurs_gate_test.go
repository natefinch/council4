package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestBaldursGateFreeTraversalVisitsEveryRoomOnceAndCompletes(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	// A basic land for Crash Landing plus filler for scry/draw rooms.
	addCardToLibrary(g, game.Player1, basicForestDef())
	stockLibrary(g, game.Player1, 40)
	def, _ := game.DungeonByID(game.DungeonBaldursGateWilderness)
	roomCount := len(def.Rooms)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &ventureChoiceAgent{prefer: []string{"Baldur's Gate Wilderness"}}}

	visited := map[int]bool{}
	for i := range roomCount {
		if !engine.ventureIntoDungeon(g, game.Player1, agents, &TurnLog{}) {
			t.Fatalf("venture %d failed", i)
		}
		state := g.Players[game.Player1].Dungeon
		if !state.Exists {
			// The final venture completes the dungeon before we can read the room;
			// that is only valid on the last iteration.
			if i != roomCount-1 {
				t.Fatalf("dungeon cleared after venture %d before all rooms visited", i)
			}
		} else {
			if visited[state.Val.Room] {
				t.Fatalf("venture %d re-entered already-visited room %d", i, state.Val.Room)
			}
			visited[state.Val.Room] = true
		}
		drainDungeonStack(engine, g, agents)
	}
	if g.Players[game.Player1].Dungeon.Exists {
		t.Fatal("still in the dungeon after visiting every room")
	}
	if g.Players[game.Player1].DungeonsCompleted != 1 {
		t.Fatalf("DungeonsCompleted = %d, want 1", g.Players[game.Player1].DungeonsCompleted)
	}
	if got := countEvents(g, game.EventCompletedDungeon); got != 1 {
		t.Fatalf("completion events = %d, want 1", got)
	}
}

func TestBaldursGateNeverRepeatsARoom(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, basicForestDef())
	stockLibrary(g, game.Player1, 40)
	def, _ := game.DungeonByID(game.DungeonBaldursGateWilderness)
	// An agent that always tries to re-enter the very first room by name; free
	// traversal must never offer an already-visited room.
	firstRoom := def.Rooms[0].Name
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &ventureChoiceAgent{prefer: []string{"Baldur's Gate Wilderness", firstRoom}}}
	seen := map[int]int{}
	for range len(def.Rooms) {
		engine.ventureIntoDungeon(g, game.Player1, agents, &TurnLog{})
		if state := g.Players[game.Player1].Dungeon; state.Exists {
			seen[state.Val.Room]++
		}
		drainDungeonStack(engine, g, agents)
	}
	for room, count := range seen {
		if count > 1 {
			t.Fatalf("room %d visited %d times", room, count)
		}
	}
	if g.Players[game.Player1].DungeonsCompleted != 1 {
		t.Fatalf("DungeonsCompleted = %d, want 1", g.Players[game.Player1].DungeonsCompleted)
	}
}

// ventureIntoRoom enters a free-traversal dungeon and advances to a specific
// named room, resolving each room's ability. It returns after the target room
// resolves.
func ventureToBaldursGateRoom(t *testing.T, engine *Engine, g *game.Game, roomName string) {
	t.Helper()
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &ventureChoiceAgent{prefer: []string{"Baldur's Gate Wilderness", roomName}}}
	if !engine.ventureIntoDungeon(g, game.Player1, agents, &TurnLog{}) {
		t.Fatal("venture failed")
	}
	drainDungeonStack(engine, g, agents)
}

func TestBaldursGateReithwinTollhouseCreatesTreasures(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	ventureToBaldursGateRoom(t, engine, g, "Reithwin Tollhouse")
	treasures := 0
	for _, p := range g.Battlefield {
		if p.TokenDef != nil && p.TokenDef.Name == "Treasure" {
			treasures++
		}
	}
	if treasures < 2 || treasures > 8 {
		t.Fatalf("Reithwin Tollhouse created %d Treasures, want 2..8 (2d4)", treasures)
	}
}

func TestBaldursGateEmeraldGroveCreatesKnight(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	ventureToBaldursGateRoom(t, engine, g, "Emerald Grove")
	found := false
	for _, p := range g.Battlefield {
		if p.TokenDef != nil && p.TokenDef.Name == "Knight" {
			found = true
		}
	}
	if !found {
		t.Fatal("Emerald Grove did not create a Knight token")
	}
}

func TestBaldursGateGauntletOfSharDrainsOpponents(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	ventureToBaldursGateRoom(t, engine, g, "Gauntlet of Shar")
	for _, opponentID := range aliveOpponents(g, game.Player1) {
		if g.Players[opponentID].Life != 35 {
			t.Fatalf("opponent %d life = %d, want 35 (lost 5)", opponentID, g.Players[opponentID].Life)
		}
	}
}

func TestBaldursGateTempleOfBhaalDebuffsOpponentCreatures(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	opp := addCombatPermanent(g, game.Player2, namedCreatureDef("Ogre"))
	ownID := opp.ObjectID
	ventureToBaldursGateRoom(t, engine, g, "Temple of Bhaal")
	permanent, ok := permanentByObjectID(g, ownID)
	if !ok {
		t.Fatal("opponent creature disappeared")
	}
	values := effectivePermanentValues(g, permanent)
	if values.power != -4 || values.toughness != -4 {
		t.Fatalf("opponent creature = %d/%d, want -4/-4 (1/1 minus 5/5)", values.power, values.toughness)
	}
}

func TestBaldursGateGrymforgeGoadsOpponentCreature(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	opp := addCombatPermanent(g, game.Player2, namedCreatureDef("Bandit"))
	ventureToBaldursGateRoom(t, engine, g, "Grymforge")
	permanent, ok := permanentByObjectID(g, opp.ObjectID)
	if !ok {
		t.Fatal("opponent creature disappeared")
	}
	if _, goaded := permanent.Goaded[game.Player1]; !goaded {
		t.Fatal("Grymforge did not goad the opponent's creature")
	}
}

func TestBaldursGateGithyankiCrecheDistributesCounters(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	own := addCombatPermanent(g, game.Player1, namedCreatureDef("Recruit"))
	ventureToBaldursGateRoom(t, engine, g, "Githyanki Crèche")
	permanent, ok := permanentByObjectID(g, own.ObjectID)
	if !ok {
		t.Fatal("own creature disappeared")
	}
	// With a single legal target the three counters all land on it.
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 3 {
		t.Fatalf("distributed counters = %d, want 3", got)
	}
}

func TestBaldursGateCircusCopiesCommanderNonLegendary(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	// A legendary commander for Player1.
	cmdID := g.IDGen.Next()
	g.CardInstances[cmdID] = &game.CardInstance{ID: cmdID, Owner: game.Player1, Def: &game.CardDef{CardFace: game.CardFace{
		Name:       "Commander Zed",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Power:      opt.Val(game.PT{Value: 3}),
		Toughness:  opt.Val(game.PT{Value: 3}),
	}}}
	g.Players[game.Player1].CommanderInstanceID = cmdID

	ventureToBaldursGateRoom(t, engine, g, "Circus of the Last Days")
	var copyToken *game.Permanent
	for _, p := range g.Battlefield {
		if p.TokenDef != nil && p.TokenDef.Name == "Commander Zed" {
			copyToken = p
		}
	}
	if copyToken == nil {
		t.Fatal("Circus did not create a commander copy token")
	}
	for _, super := range copyToken.TokenDef.Supertypes {
		if super == types.Legendary {
			t.Fatal("commander copy token is still legendary")
		}
	}
}

func TestBaldursGateAnsursSanctumRevealsAndDrains(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	// Four cards of mana value 2 each on top: total mana value 8.
	for range 4 {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:     "Spell",
			Types:    []types.Card{types.Sorcery},
			ManaCost: opt.Val(cost.Mana{cost.O(2)}),
		}})
	}
	handBefore := g.Players[game.Player1].Hand.Size()
	ventureToBaldursGateRoom(t, engine, g, "Ansur's Sanctum")
	if got := g.Players[game.Player1].Hand.Size() - handBefore; got != 4 {
		t.Fatalf("hand delta = %d, want 4 (revealed top four to hand)", got)
	}
	for _, opponentID := range aliveOpponents(g, game.Player1) {
		if g.Players[opponentID].Life != 32 {
			t.Fatalf("opponent %d life = %d, want 32 (lost 8 total mana value)", opponentID, g.Players[opponentID].Life)
		}
	}
}

func TestBaldursGateSteelWatchFoundryEmblemAnthem(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	own := addCombatPermanent(g, game.Player1, namedCreatureDef("Trooper")) // 1/1
	ventureToBaldursGateRoom(t, engine, g, "Steel Watch Foundry")
	if len(g.Emblems) != 1 || g.Emblems[0].Owner != game.Player1 {
		t.Fatalf("emblems = %+v, want one Player1 emblem", g.Emblems)
	}
	permanent, ok := permanentByObjectID(g, own.ObjectID)
	if !ok {
		t.Fatal("own creature disappeared")
	}
	values := effectivePermanentValues(g, permanent)
	if values.power != 3 || values.toughness != 3 {
		t.Fatalf("creature = %d/%d, want 3/3 (1/1 + emblem +2/+2)", values.power, values.toughness)
	}
	if !hasKeyword(g, permanent, game.Trample) {
		t.Fatal("emblem did not grant trample")
	}
}

func TestBaldursGateSteelWatchEmblemDoesNotBuffOpponents(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	opp := addCombatPermanent(g, game.Player2, namedCreatureDef("Rival")) // 1/1
	ventureToBaldursGateRoom(t, engine, g, "Steel Watch Foundry")
	permanent, ok := permanentByObjectID(g, opp.ObjectID)
	if !ok {
		t.Fatal("opponent creature disappeared")
	}
	values := effectivePermanentValues(g, permanent)
	if values.power != 1 || values.toughness != 1 {
		t.Fatalf("opponent creature = %d/%d, want 1/1 (emblem only buffs its controller's creatures)", values.power, values.toughness)
	}
	if hasKeyword(g, permanent, game.Trample) {
		t.Fatal("emblem granted trample to an opponent's creature")
	}
}

func TestBaldursGateEbonlakeGrottoCreatesTwoFaerieDragons(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	ventureToBaldursGateRoom(t, engine, g, "Ebonlake Grotto")
	dragons := 0
	for _, p := range g.Battlefield {
		if p.TokenDef != nil && p.TokenDef.Name == "Faerie Dragon" {
			dragons++
			if !hasKeyword(g, p, game.Flying) {
				t.Fatal("Faerie Dragon lacks flying")
			}
		}
	}
	if dragons != 2 {
		t.Fatalf("created %d Faerie Dragons, want 2", dragons)
	}
}

func TestBaldursGateDefiledTempleGatesDrawOnSacrifice(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, namedCreatureDef("Chaff"))
	stockLibrary(g, game.Player1, 5)
	ownedBefore := controlledPermanentCount(g, game.Player1)
	handBefore := g.Players[game.Player1].Hand.Size()
	ventureToBaldursGateRoom(t, engine, g, "Defiled Temple")
	sacrificed := ownedBefore - controlledPermanentCount(g, game.Player1)
	drew := g.Players[game.Player1].Hand.Size() - handBefore
	// The draw is gated on the sacrifice: they happen together or not at all.
	if drew != sacrificed {
		t.Fatalf("drew %d but sacrificed %d; draw must be gated on the sacrifice", drew, sacrificed)
	}
}

func controlledPermanentCount(g *game.Game, controller game.PlayerID) int {
	count := 0
	for _, p := range g.Battlefield {
		if effectiveController(g, p) == controller {
			count++
		}
	}
	return count
}
