package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// namedCreatureDef returns a simple castable creature card with the given name.
func namedCreatureDef(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}
}

// controlledCreatures returns the battlefield creatures a player controls.
func controlledCreatures(g *game.Game, controller game.PlayerID) []*game.Permanent {
	var out []*game.Permanent
	for _, p := range g.Battlefield {
		if effectiveController(g, p) != controller {
			continue
		}
		if permanentHasType(g, p, types.Creature) {
			out = append(out, p)
		}
	}
	return out
}

func TestThroneOfTheDeadThreePutsCreatureWithCountersAndHexproof(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	// A basic land for Secret Entrance's tutor, plus creatures so Throne's top-ten
	// reveal always contains a creature to put onto the battlefield.
	addCardToLibrary(g, game.Player1, basicForestDef())
	for range 8 {
		addCardToLibrary(g, game.Player1, namedCreatureDef("Cultist"))
	}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &ventureChoiceAgent{prefer: []string{"Lost Well", "Stash"}}}
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
	var put *game.Permanent
	for _, p := range controlledCreatures(g, game.Player1) {
		if card, ok := g.GetCardInstance(p.CardInstanceID); ok && card.Def != nil && card.Def.Name == "Cultist" {
			put = p
			break
		}
	}
	if put == nil {
		t.Fatal("Throne of the Dead Three did not put a creature onto the battlefield")
	}
	if got := put.Counters.Get(counter.PlusOnePlusOne); got != 3 {
		t.Fatalf("put creature has %d +1/+1 counters, want 3", got)
	}
	if !hasKeyword(g, put, game.Hexproof) {
		t.Fatal("put creature does not have hexproof")
	}
}

func TestThroneOfTheDeadThreeNoCreatureStillCompletes(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	// Only a basic land and non-creature cards: Throne finds no creature to put.
	addCardToLibrary(g, game.Player1, basicForestDef())
	for range 8 {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Relic", Types: []types.Card{types.Artifact}}})
	}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &ventureChoiceAgent{prefer: []string{"Lost Well", "Stash"}}}
	for range 5 {
		engine.ventureIntoUndercity(g, game.Player1, agents, &TurnLog{})
		drainDungeonStack(engine, g, agents)
	}
	if g.Players[game.Player1].DungeonsCompleted != 1 {
		t.Fatalf("DungeonsCompleted = %d, want 1 (completes even with no creature)", g.Players[game.Player1].DungeonsCompleted)
	}
	// The only battlefield creatures come from Catacombs' Skeleton token on the
	// path; Throne itself put no library creature card.
	for _, p := range controlledCreatures(g, game.Player1) {
		if p.TokenDef == nil {
			t.Fatal("Throne put a library creature card despite none being revealed")
		}
	}
}

func TestMadWizardsLairDrawsThreeAndCastsOneFree(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	// A deep library of castable creatures so the final room draws three castable
	// spells regardless of the scry/impulse rooms along the path.
	for range 25 {
		addCardToLibrary(g, game.Player1, namedCreatureDef("Apparition"))
	}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &ventureChoiceAgent{
		prefer: []string{"Dungeon of the Mad Mage", "Goblin Bazaar", "Runestone Caverns"},
	}}
	handBefore := g.Players[game.Player1].Hand.Size()
	// Path: Yawning Portal (0) -> Dungeon Level (1) -> Goblin Bazaar (2) ->
	// Lost Level (4) -> Runestone Caverns (5) -> Deep Mines (7) -> Mad Wizard's Lair (8).
	for range 7 {
		if !engine.ventureIntoDungeon(g, game.Player1, agents, &TurnLog{}) {
			t.Fatal("mad mage venture failed")
		}
		drainDungeonStack(engine, g, agents)
	}
	if g.Players[game.Player1].DungeonsCompleted != 1 {
		t.Fatalf("DungeonsCompleted = %d, want 1", g.Players[game.Player1].DungeonsCompleted)
	}
	// Drew three, then cast one of them for free (which resolved to a creature on
	// the battlefield). Net hand gain is two: three drawn minus the one cast.
	if got := g.Players[game.Player1].Hand.Size() - handBefore; got != 2 {
		t.Fatalf("hand delta = %d, want 2 (drew 3, cast 1 for free)", got)
	}
	if len(controlledCreatures(g, game.Player1)) != 1 {
		t.Fatalf("controlled creatures = %d, want 1 (the free-cast creature)", len(controlledCreatures(g, game.Player1)))
	}
}
