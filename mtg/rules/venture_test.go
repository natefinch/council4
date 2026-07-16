package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ventureChoiceAgent answers dungeon and branch choices by preferring the first
// option whose label appears in prefer, falling back to the request default.
type ventureChoiceAgent struct {
	prefer []string
}

func (*ventureChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return actionBuild.pass()
}

func (a *ventureChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	for _, want := range a.prefer {
		for _, option := range request.Options {
			if option.Label == want {
				return []int{option.Index}
			}
		}
	}
	if len(request.DefaultSelection) > 0 {
		return request.DefaultSelection
	}
	return []int{0}
}

// drainDungeonStack puts queued room abilities and initiative ventures on the
// stack and resolves everything, so a venture's room ability resolves.
func drainDungeonStack(engine *Engine, g *game.Game, agents [game.NumPlayers]PlayerAgent) {
	log := &TurnLog{}
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log)
	for {
		if _, ok := g.Stack.Peek(); !ok {
			break
		}
		engine.resolveTopOfStackWithChoices(g, agents, log)
		engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log)
	}
}

func mainPhaseGame(active game.PlayerID) *game.Game {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = active
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	return g
}

func TestVentureIntoDungeonEntersFirstRoom(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	stockLibrary(g, game.Player1, 5)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &ventureChoiceAgent{prefer: []string{"Lost Mine of Phandelver"}}}

	if !engine.ventureIntoDungeon(g, game.Player1, agents, &TurnLog{}) {
		t.Fatal("venture into the dungeon failed")
	}
	state := g.Players[game.Player1].Dungeon
	if !state.Exists {
		t.Fatal("player is not in a dungeon after venturing")
	}
	if state.Val.Dungeon != game.DungeonLostMineOfPhandelver || state.Val.Room != 0 {
		t.Fatalf("dungeon state = %+v, want Lost Mine room 0", state.Val)
	}
	if countEvents(g, game.EventVenturedIntoDungeon) != 1 {
		t.Fatalf("venture events = %d, want 1", countEvents(g, game.EventVenturedIntoDungeon))
	}
	// The Cave Entrance room ability (Scry 1) is queued and put on the stack.
	drainDungeonStack(engine, g, agents)
}

func TestVentureIntoDungeonNeverOffersUndercity(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	// An agent that records the offered dungeon labels.
	var offered []string
	agent := &labelRecordingAgent{onChoice: func(request game.ChoiceRequest) {
		for _, opt := range request.Options {
			offered = append(offered, opt.Label)
		}
	}}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	engine.ventureIntoDungeon(g, game.Player1, agents, &TurnLog{})
	for _, label := range offered {
		if label == "Undercity" {
			t.Fatal("venture into the dungeon offered Undercity")
		}
	}
	if len(offered) != 4 {
		t.Fatalf("offered %d dungeons, want 4 ordinary dungeons", len(offered))
	}
}

func TestVentureAdvancesAndCompletesDungeon(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	stockLibrary(g, game.Player1, 10)
	// A creature so the Storeroom "+1/+1 counter on target creature" room resolves.
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Bear", Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2}),
	}})
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &ventureChoiceAgent{
		prefer: []string{"Lost Mine of Phandelver", "Goblin Lair", "Storeroom", "Temple of Dumathoin"},
	}}

	// Path: Cave Entrance (0) -> Goblin Lair (1) -> Storeroom (3) -> Temple of Dumathoin (6, final).
	for range 4 {
		if !engine.ventureIntoDungeon(g, game.Player1, agents, &TurnLog{}) {
			t.Fatal("venture failed")
		}
		drainDungeonStack(engine, g, agents)
	}

	if g.Players[game.Player1].Dungeon.Exists {
		t.Fatal("player still in a dungeon after completing it")
	}
	if got := g.Players[game.Player1].DungeonsCompleted; got != 1 {
		t.Fatalf("DungeonsCompleted = %d, want 1", got)
	}
	if got := countEvents(g, game.EventCompletedDungeon); got != 1 {
		t.Fatalf("completion events = %d, want 1", got)
	}
}

func TestVentureIntoUndercityEntersUndercity(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	stockLibrary(g, game.Player1, 5)
	agents := [game.NumPlayers]PlayerAgent{}

	if !engine.ventureIntoUndercity(g, game.Player1, agents, &TurnLog{}) {
		t.Fatal("venture into Undercity failed")
	}
	state := g.Players[game.Player1].Dungeon
	if !state.Exists || state.Val.Dungeon != game.DungeonUndercity || state.Val.Room != 0 {
		t.Fatalf("dungeon state = %+v, want Undercity room 0", state.Val)
	}
	drainDungeonStack(engine, g, agents)
}

func TestVentureIntoUndercityAdvancesExistingDungeon(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	stockLibrary(g, game.Player1, 5)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &ventureChoiceAgent{prefer: []string{"Dungeon of the Mad Mage"}}}

	// Enter Mad Mage first with an ordinary venture.
	engine.ventureIntoDungeon(g, game.Player1, agents, &TurnLog{})
	drainDungeonStack(engine, g, agents)
	// venture into Undercity now advances the current (Mad Mage) dungeon, not Undercity.
	engine.ventureIntoUndercity(g, game.Player1, agents, &TurnLog{})
	state := g.Players[game.Player1].Dungeon
	if !state.Exists || state.Val.Dungeon != game.DungeonDungeonOfTheMadMage {
		t.Fatalf("dungeon = %+v, want still in Mad Mage (advanced, not switched to Undercity)", state.Val)
	}
	if state.Val.Room == 0 {
		t.Fatal("venture into Undercity did not advance the current dungeon")
	}
}

// labelRecordingAgent records every choice request without changing the answer.
type labelRecordingAgent struct {
	onChoice func(request game.ChoiceRequest)
}

func (*labelRecordingAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return actionBuild.pass()
}

func (a *labelRecordingAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if a.onChoice != nil {
		a.onChoice(request)
	}
	if len(request.DefaultSelection) > 0 {
		return request.DefaultSelection
	}
	return []int{0}
}

func TestVentureRoomAbilityDoublerStacks(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	stockLibrary(g, game.Player1, 5)
	// A Hama Pashar / Ruin Seeker style room-ability doubler the venturing player
	// controls.
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:       "Ruin Seeker",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{
			Text:        "Room abilities of dungeons you own trigger an additional time.",
			RuleEffects: []game.RuleEffect{{Kind: game.RuleEffectAdditionalTriggerForRoomAbility}},
		}},
	}})
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &ventureChoiceAgent{prefer: []string{"Lost Mine of Phandelver"}}}

	engine.ventureIntoDungeon(g, game.Player1, agents, &TurnLog{})
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want 2 (room ability doubled once)", got)
	}
}

func TestVentureStateClones(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	stockLibrary(g, game.Player1, 5)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &ventureChoiceAgent{prefer: []string{"Tomb of Annihilation"}}}
	engine.ventureIntoDungeon(g, game.Player1, agents, &TurnLog{})
	g.Players[game.Player1].DungeonsCompleted = 2
	g.Players[game.Player1].HasInitiative = true

	clone := g.Clone()
	cs := clone.Players[game.Player1].Dungeon
	if !cs.Exists || cs.Val != g.Players[game.Player1].Dungeon.Val {
		t.Fatalf("clone dungeon state = %+v, want %+v", cs, g.Players[game.Player1].Dungeon)
	}
	if clone.Players[game.Player1].DungeonsCompleted != 2 || !clone.Players[game.Player1].HasInitiative {
		t.Fatal("clone did not copy DungeonsCompleted / HasInitiative")
	}
	// Mutating the clone must not affect the original.
	clone.Players[game.Player1].Dungeon = opt.V[game.DungeonState]{}
	if !g.Players[game.Player1].Dungeon.Exists {
		t.Fatal("mutating clone dungeon state affected the original")
	}
}

func TestObservationExposesDungeonAndInitiative(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	stockLibrary(g, game.Player1, 5)
	g.Players[game.Player1].HasInitiative = true
	g.Players[game.Player1].DungeonsCompleted = 1
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &ventureChoiceAgent{prefer: []string{"Undercity"}}}
	engine.ventureIntoUndercity(g, game.Player1, agents, &TurnLog{})

	view := observe(g, game.Player1).PlayerState(game.Player1)
	if !view.HasInitiative || view.DungeonsCompleted != 1 {
		t.Fatalf("view initiative/completed = %v/%d, want true/1", view.HasInitiative, view.DungeonsCompleted)
	}
	if !view.InDungeon || view.DungeonName != "Undercity" || view.DungeonRoom != "Secret Entrance" {
		t.Fatalf("view dungeon = %+v, want Undercity / Secret Entrance", view)
	}
}

// permanentNamed reports whether the given player controls a battlefield
// permanent with the given name.
func permanentNamed(g *game.Game, controller game.PlayerID, name string) bool {
	for _, p := range g.Battlefield {
		if p.Owner == controller && effectiveController(g, p) == controller {
			if def, ok := g.GetCardInstance(p.CardInstanceID); ok && def.Def != nil && def.Def.Name == name {
				return true
			}
			if p.TokenDef != nil && p.TokenDef.Name == name {
				return true
			}
		}
	}
	return false
}

func TestTombOfAnnihilationTraversalCompletes(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	stockLibrary(g, game.Player1, 10)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &ventureChoiceAgent{
		prefer: []string{"Tomb of Annihilation", "Veils of Fear", "Sandfall Cell", "Cradle of the Death God"},
	}}
	// Path: Trapped Entry (0) -> Veils of Fear (1) -> Sandfall Cell (2) -> Cradle (4, final).
	for range 4 {
		if !engine.ventureIntoDungeon(g, game.Player1, agents, &TurnLog{}) {
			t.Fatal("tomb venture failed")
		}
		drainDungeonStack(engine, g, agents)
	}
	if g.Players[game.Player1].DungeonsCompleted != 1 {
		t.Fatalf("DungeonsCompleted = %d, want 1", g.Players[game.Player1].DungeonsCompleted)
	}
	// Trapped Entry (each player loses 1), Veils (2), Sandfall (2) = 5 total for a
	// player with no hand or permanents to spare.
	if g.Players[game.Player1].Life != 35 {
		t.Fatalf("life = %d, want 35 after Tomb life-loss rooms", g.Players[game.Player1].Life)
	}
	if !permanentNamed(g, game.Player1, "The Atropal") {
		t.Fatal("Cradle of the Death God did not create The Atropal")
	}
}

func TestMadMageTraversalCompletes(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	stockLibrary(g, game.Player1, 20)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &ventureChoiceAgent{
		prefer: []string{"Dungeon of the Mad Mage", "Goblin Bazaar", "Runestone Caverns"},
	}}
	// Path: Yawning Portal (0) -> Dungeon Level (1) -> Goblin Bazaar (2) ->
	// Lost Level (4) -> Runestone Caverns (5) -> Deep Mines (7) -> Mad Wizard's Lair (8, final).
	for range 7 {
		if !engine.ventureIntoDungeon(g, game.Player1, agents, &TurnLog{}) {
			t.Fatal("mad mage venture failed")
		}
		drainDungeonStack(engine, g, agents)
	}
	if g.Players[game.Player1].DungeonsCompleted != 1 {
		t.Fatalf("DungeonsCompleted = %d, want 1", g.Players[game.Player1].DungeonsCompleted)
	}
	if g.Players[game.Player1].Life != 41 {
		t.Fatalf("life = %d, want 41 (Yawning Portal gain 1)", g.Players[game.Player1].Life)
	}
	if !permanentNamed(g, game.Player1, "Treasure") {
		t.Fatal("Goblin Bazaar did not create a Treasure token")
	}
}

func TestUndercityTraversalCompletes(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	// A basic land in the library so Secret Entrance's tutor finds one.
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Forest", Supertypes: []types.Super{types.Basic}, Types: []types.Card{types.Land}, Subtypes: []types.Sub{types.Forest},
	}})
	stockLibrary(g, game.Player1, 10)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &ventureChoiceAgent{
		prefer: []string{"Lost Well", "Stash"},
	}}
	// Path: Secret Entrance (0) -> Lost Well (2) -> Stash (5) -> Catacombs (7) -> Throne (8, final).
	for range 5 {
		if !engine.ventureIntoUndercity(g, game.Player1, agents, &TurnLog{}) {
			t.Fatal("undercity venture failed")
		}
		drainDungeonStack(engine, g, agents)
	}
	if g.Players[game.Player1].DungeonsCompleted != 1 {
		t.Fatalf("DungeonsCompleted = %d, want 1", g.Players[game.Player1].DungeonsCompleted)
	}
	if !permanentNamed(g, game.Player1, "Skeleton") {
		t.Fatal("Catacombs did not create a Skeleton token")
	}
}
