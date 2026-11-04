package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: issue #3598. Official CR effective 2026-09-25 (verified local
// copy, retained workflow unit evidence cr-20260925.txt), read 310.9-310.9g,
// 310.11 and 310.12a. 310.9: each battle has a player designated as its
// protector; 310.9a: its controller chooses the protector as it enters;
// 310.9e/f: a battle has one protector, replaced only when another player
// becomes its protector; 310.9g: the protector doesn't change if it stops being
// a battle; 310.11: a battle whose protector can't protect it gets a new
// protector chosen by its controller (SBA); 310.12a: only an opponent of a
// Siege's controller can protect it. The protector is a designation, not a
// characteristic hidden by fog of war, so every observer sees it. Invasion of
// Segovia (pinned corpus oracle 7b3c0d89-55ed-43d7-a00a-7b060857ec4e, {2}{U}
// Battle — Siege, defense 4): "(As a Siege enters, choose an opponent to
// protect it. You and others can attack it. When it's defeated, exile it, then
// cast it transformed.) When this Siege enters, create two 1/1 blue Kraken
// creature tokens with trample." Rulings (Scryfall, retrieved 2026-11-04, 2023-04-14
// set): the controller chooses an opponent as protector as the Siege enters; if
// its protector gains control of it, a new protector is chosen as a state-based
// action. Word of Seizing (oracle b80d2841-03b0-42d3-ac9b-51e793b1dc24) "Untap
// target permanent and gain control of it until end of turn."

// assertObservedProtector checks that every player's observation of objectID
// reports protector and controller.
func assertObservedProtector(t *testing.T, g *game.Game, objectID id.ID, protector opt.V[game.PlayerID], controller game.PlayerID) {
	t.Helper()
	for viewer := range game.PlayerID(game.NumPlayers) {
		view := findPermanentView(t, NewObservation(g, viewer), objectID)
		if view.BattleProtector != protector {
			t.Fatalf("player %d observes protector %+v, want %+v", viewer+1, view.BattleProtector, protector)
		}
		if view.Controller != controller {
			t.Fatalf("player %d observes controller %d, want %d", viewer+1, view.Controller, controller)
		}
	}
}

func TestObservedSiegeProtectorFollowsChoiceAndReassignment(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	g.Turn.PriorityPlayer = game.Player1
	engine := choicesProgramEngine(t)
	card := addCardToHand(g, game.Player1, registryCardDef(t, "Invasion of Segovia"))
	g.Players[game.Player1].ManaPool.Add(mana.U, 3)
	var offered [][]string
	agents := [game.NumPlayers]PlayerAgent{game.Player1: protectorChooser(game.Player3, &offered)}
	log := &TurnLog{}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(card, nil, 0, nil), agents, log) {
		t.Fatal("casting Invasion of Segovia failed")
	}
	engine.resolveTopOfStackWithChoices(g, agents, log)
	siege, ok := findPermanentByCardID(g, card)
	if !ok {
		t.Fatal("Invasion of Segovia did not enter the battlefield")
	}
	if want := [][]string{{"Player 2", "Player 3", "Player 4"}}; !slices.EqualFunc(offered, want, slices.Equal) {
		t.Fatalf("protector options = %v, want %v", offered, want)
	}
	assertObservedProtector(t, g, siege.ObjectID, opt.Val(game.Player3), game.Player1)

	castAndResolveWordOfSeizing(t, g, engine, game.Player3, siege)
	offered = nil
	engine.setReplacementChoiceContext(g, [game.NumPlayers]PlayerAgent{game.Player3: protectorChooser(game.Player4, &offered)}, &TurnLog{})
	defer g.ClearChoiceContext()
	engine.applyStateBasedActions(g)
	if want := [][]string{{"Player 1", "Player 2", "Player 4"}}; !slices.EqualFunc(offered, want, slices.Equal) {
		t.Fatalf("reassigned protector options = %v, want %v", offered, want)
	}
	assertObservedProtector(t, g, siege.ObjectID, opt.Val(game.Player4), game.Player3)
}

func TestObservedNonBattleHasNoProtector(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	bears := registryCombatPermanent(t, g, game.Player1, "Grizzly Bears")
	assertObservedProtector(t, g, bears.ObjectID, opt.V[game.PlayerID]{}, game.Player1)
}

func TestObservedProtectorRetainedAfterPermanentStopsBeingABattle(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	siege := registryCombatPermanent(t, g, game.Player1, "Invasion of Segovia")
	siege.Counters.Add(counter.Defense, 4)
	siege.BattleProtector = opt.Val(game.Player2)
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), Timestamp: game.Timestamp(g.IDGen.Next()),
		Layer: game.LayerType, AffectedObjectID: siege.ObjectID, RemoveTypes: []types.Card{types.Battle},
	})
	NewEngine(nil).applyStateBasedActions(g)
	if _, ok := g.PermanentByID(siege.ObjectID); !ok {
		t.Fatal("the former battle left the battlefield")
	}
	view := findPermanentView(t, NewObservation(g, game.Player3), siege.ObjectID)
	if slices.Contains(view.Types, types.Battle) {
		t.Fatalf("observed types %v still include Battle", view.Types)
	}
	assertObservedProtector(t, g, siege.ObjectID, opt.Val(game.Player2), game.Player1)
}

func TestObservedProtectorIsCloneSafe(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	siege := registryCombatPermanent(t, g, game.Player1, "Invasion of Segovia")
	siege.Counters.Add(counter.Defense, 4)
	siege.BattleProtector = opt.Val(game.Player3)
	before := findPermanentView(t, NewObservation(g, game.Player2), siege.ObjectID)

	clone := g.Clone()
	cloned, ok := clone.PermanentByID(siege.ObjectID)
	if !ok {
		t.Fatal("the clone lacks the Siege")
	}
	assertObservedProtector(t, clone, siege.ObjectID, opt.Val(game.Player3), game.Player1)
	cloned.BattleProtector = opt.Val(game.Player4)

	assertObservedProtector(t, clone, siege.ObjectID, opt.Val(game.Player4), game.Player1)
	assertObservedProtector(t, g, siege.ObjectID, opt.Val(game.Player3), game.Player1)
	siege.BattleProtector = opt.Val(game.Player2)
	if before.BattleProtector != opt.Val(game.Player3) {
		t.Fatalf("an earlier observation changed to protector %+v", before.BattleProtector)
	}
}
