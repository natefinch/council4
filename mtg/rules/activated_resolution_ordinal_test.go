package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

func ordinalFixture(t *testing.T, text string) (*game.Game, *Engine, *game.Permanent) {
	t.Helper()
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Ordinal", Layout: "normal", TypeLine: "Artifact", OracleText: text,
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, def)
	for player := range game.NumPlayers {
		g.Players[player].ManaPool.Add(mana.C, 20)
		for range 10 {
			addCardToLibrary(g, game.PlayerID(player), evidenceCard("Drawn", 1))
		}
	}
	return g, NewEngine(nil), source
}

func activateOrdinal(t *testing.T, g *game.Game, engine *Engine, sourceID id.ID, index int, targets []game.Target, modes []int) *game.StackObject {
	t.Helper()
	if !engine.applyAction(g, g.Turn.PriorityPlayer, action.ActivateAbilityWithModes(sourceID, index, targets, 0, modes)) {
		t.Fatalf("activation failed: source=%d index=%d player=%d mana=%v stack=%d", sourceID, index, g.Turn.PriorityPlayer, g.Players[g.Turn.PriorityPlayer].ManaPool.Units(), g.Stack.Size())
	}
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("non-mana activation did not use the stack")
	}
	return obj
}

func TestCompiledActivatedResolutionOrdinal(t *testing.T) {
	t.Parallel()
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Ordinal", Layout: "normal", TypeLine: "Artifact",
		OracleText: "{1}: If this is the second time this ability has resolved this turn, draw a card.",
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, def)
	addCardToLibrary(g, game.Player1, evidenceCard("Drawn", 1))
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	engine := NewEngine(nil)
	for range 3 {
		if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
			t.Fatal("activation failed")
		}
	}
	if !g.Players[game.Player1].ManaPool.IsEmpty() || g.Players[game.Player1].Hand.Size() != 0 {
		t.Fatal("costs must be paid without resolving the gated draw")
	}
	for i, want := range []int{0, 1, 1} {
		engine.resolveTopOfStack(g, &TurnLog{})
		if got := g.Players[game.Player1].Hand.Size(); got != want {
			t.Fatalf("resolution %d: hand=%d, want %d", i+1, got, want)
		}
	}
}

func TestActivatedOrdinalFirstSecondThird(t *testing.T) {
	t.Parallel()
	for ordinal, word := range []string{"first", "second", "third"} {
		t.Run(word, func(t *testing.T) {
			t.Parallel()
			g, engine, source := ordinalFixture(t, fmt.Sprintf(
				"{1}: If this is the %s time this ability has resolved this turn, draw a card.", word))
			for i := range 3 {
				activateOrdinal(t, g, engine, source.ObjectID, 0, nil, nil)
				engine.resolveTopOfStack(g, &TurnLog{})
				want := 0
				if i >= ordinal {
					want = 1
				}
				if got := g.Players[game.Player1].Hand.Size(); got != want {
					t.Fatalf("resolution %d: hand=%d, want %d", i+1, got, want)
				}
			}
		})
	}
}

func TestActivatedOrdinalCounteredAndFizzled(t *testing.T) {
	t.Parallel()
	g, engine, source := ordinalFixture(t, "{1}: If this is the second time this ability has resolved this turn, destroy target creature.")
	target := addCombatPermanent(g, game.Player2, creatureDef("Victim"))
	chosen := []game.Target{game.PermanentTarget(target.ObjectID)}
	countered := activateOrdinal(t, g, engine, source.ObjectID, 0, chosen, nil)
	if !counterStackObject(g, countered.ID) {
		t.Fatal("counter failed")
	}
	fizzled := activateOrdinal(t, g, engine, source.ObjectID, 0, chosen, nil)
	if !movePermanentToZone(g, target, zone.Graveyard) {
		t.Fatal("target did not leave")
	}
	obj, _ := g.Stack.Pop()
	if got := engine.resolveStackObject(g, obj, &TurnLog{}); got != "countered by rules" {
		t.Fatalf("fizzled resolution=%q", got)
	}
	if len(g.ResolvedActivatedAbilitiesThisTurn) != 0 || fizzled.ResolutionOrdinalThisTurn != 0 {
		t.Fatal("countered/fizzled objects must not count")
	}
	target = addCombatPermanent(g, game.Player2, creatureDef("Live Victim"))
	chosen[0] = game.PermanentTarget(target.ObjectID)
	for i := range 2 {
		activateOrdinal(t, g, engine, source.ObjectID, 0, chosen, nil)
		engine.resolveTopOfStack(g, &TurnLog{})
		_, live := permanentByObjectID(g, target.ObjectID)
		if live != (i == 0) {
			t.Fatalf("resolution %d: target live=%v", i+1, live)
		}
	}
}

func TestActivatedOrdinalCopiesControllersAndResolutionOrder(t *testing.T) {
	t.Parallel()
	g, engine, source := ordinalFixture(t, "{1}: If this is the second time this ability has resolved this turn, draw a card.")
	first := activateOrdinal(t, g, engine, source.ObjectID, 0, nil, nil)
	second := activateOrdinal(t, g, engine, source.ObjectID, 0, nil, nil)
	copyObj := game.NewStackObjectCopy(first, g.IDGen.Next())
	copyObj.Controller = game.Player2
	g.Stack.Push(copyObj)
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player2].Hand.Size() != 0 {
		t.Fatal("first resolving copy must not see an activation count")
	}
	source.Controller = game.Player2
	g.Turn.PriorityPlayer = game.Player2
	third := activateOrdinal(t, g, engine, source.ObjectID, 0, nil, nil)
	// Resolve the earliest announcement next, not the newest one.
	if _, ok := g.Stack.RemoveByID(first.ID); !ok {
		t.Fatal("original missing")
	}
	g.Stack.Push(first)
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Hand.Size() != 1 {
		t.Fatal("second actual resolution must draw for its captured controller")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player2].Hand.Size() != 0 || g.ResolvedActivatedAbilitiesThisTurn[first.ActivatedResolutionUse.Val] != 4 {
		t.Fatal("copy and changed controller must share the original ability tally")
	}
	if first.ActivatedResolutionUse != second.ActivatedResolutionUse || first.ActivatedResolutionUse != third.ActivatedResolutionUse {
		t.Fatal("announcement order and controller are not ability identity")
	}
	first.ResolutionOrdinalThisTurn = 2
	if game.NewStackObjectCopy(first, g.IDGen.Next()).ResolutionOrdinalThisTurn != 0 {
		t.Fatal("a new copy inherited an already-resolved ordinal")
	}
}

func TestActivatedOrdinalModalAndInstructionAccounting(t *testing.T) {
	t.Parallel()
	g, engine, source := ordinalFixture(t,
		"{1}: Choose two \u2014\n\u2022 If this is the second time this ability has resolved this turn, draw a card, then draw a card.\n\u2022 If this is the second time this ability has resolved this turn, you gain 2 life.")
	for i := range 3 {
		obj := activateOrdinal(t, g, engine, source.ObjectID, 0, nil, []int{0, 1})
		engine.resolveTopOfStack(g, &TurnLog{})
		wantHand, wantLife := 0, 40
		if i >= 1 {
			wantHand, wantLife = 2, 42
		}
		if got := g.Players[game.Player1].Hand.Size(); got != wantHand || g.Players[game.Player1].Life != wantLife {
			t.Fatalf("resolution %d: hand=%d life=%d", i+1, got, g.Players[game.Player1].Life)
		}
		if g.ResolvedActivatedAbilitiesThisTurn[obj.ActivatedResolutionUse.Val] != i+1 {
			t.Fatal("modes or Instructions counted separately")
		}
	}
}

func TestActivatedOrdinalCountsUngatedChosenMode(t *testing.T) {
	t.Parallel()
	g, engine, source := ordinalFixture(t,
		"{1}: Choose one \u2014\n\u2022 If this is the second time this ability has resolved this turn, draw a card.\n\u2022 You gain 2 life.")
	activateOrdinal(t, g, engine, source.ObjectID, 0, nil, []int{1})
	engine.resolveTopOfStack(g, &TurnLog{})
	activateOrdinal(t, g, engine, source.ObjectID, 0, nil, []int{0})
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Hand.Size() != 1 || g.Players[game.Player1].Life != 42 {
		t.Fatal("all resolutions of the whole ability must count, regardless of chosen mode")
	}
}

func TestActivatedOrdinalSourceBlinkAndDisappearance(t *testing.T) {
	t.Parallel()
	g, engine, source := ordinalFixture(t, "{1}: If this is the second time this ability has resolved this turn, draw a card.")
	old := activateOrdinal(t, g, engine, source.ObjectID, 0, nil, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	activateOrdinal(t, g, engine, source.ObjectID, 0, nil, nil)
	if !movePermanentToZone(g, source, zone.Exile) {
		t.Fatal("source did not leave")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Hand.Size() != 1 {
		t.Fatal("captured ability must resolve after source disappearance")
	}
	card, _ := g.GetCardInstance(source.CardInstanceID)
	returned, ok := createCardPermanent(g, card, game.Player1, zone.Exile)
	if !ok || returned.ObjectID == source.ObjectID {
		t.Fatal("blink did not create a fresh source incarnation")
	}
	obj := activateOrdinal(t, g, engine, returned.ObjectID, 0, nil, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Hand.Size() != 1 || obj.ActivatedResolutionUse == old.ActivatedResolutionUse {
		t.Fatal("new source incarnation reused an old tally")
	}
}

func TestActivatedOrdinalCloneAndTurnReset(t *testing.T) {
	t.Parallel()
	g, engine, source := ordinalFixture(t, "{1}: If this is the second time this ability has resolved this turn, draw a card.")
	first := activateOrdinal(t, g, engine, source.ObjectID, 0, nil, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	activateOrdinal(t, g, engine, source.ObjectID, 0, nil, nil)
	cloned := g.Clone()
	engine.resolveTopOfStack(cloned, &TurnLog{})
	if cloned.Players[game.Player1].Hand.Size() != 1 ||
		g.Players[game.Player1].Hand.Size() != 0 ||
		g.ResolvedActivatedAbilitiesThisTurn[first.ActivatedResolutionUse.Val] != 1 {
		t.Fatal("clone lost the captured identity or shared mutable tally state")
	}
	engine.advanceToNextTurn(g)
	if len(g.ResolvedActivatedAbilitiesThisTurn) != 0 || len(g.ResolvedTriggeredAbilitiesThisTurn) != 0 {
		t.Fatal("turn advancement did not reset resolution tallies")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Hand.Size() != 0 ||
		g.ResolvedActivatedAbilitiesThisTurn[first.ActivatedResolutionUse.Val] != 1 ||
		cloned.ResolvedActivatedAbilitiesThisTurn[first.ActivatedResolutionUse.Val] != 2 {
		t.Fatal("pending ability must start the new turn's tally without changing the clone")
	}
}
