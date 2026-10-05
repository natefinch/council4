package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

func ashlingQuantityFixture(t *testing.T) (*game.Game, *Engine, *game.Permanent) {
	t.Helper()
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Ashling", Layout: "normal", TypeLine: "Legendary Creature - Elemental Shaman",
		ManaCost: "{1}{R}", Power: new("1"), Toughness: new("1"),
		OracleText: "{1}{R}: Put a +1/+1 counter on Ashling. If this is the third time this ability has resolved this turn, remove all +1/+1 counters from Ashling, and it deals that much damage to each creature and each player.",
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, def)
	source.Counters.Add(counter.Indestructible, 1)
	source.Counters.Add(counter.Charge, 2)
	source.Counters.Add(counter.Loyalty, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 20)
	g.Players[game.Player1].ManaPool.Add(mana.R, 20)
	return g, NewEngine(nil), source
}

func TestCompiledAshlingCounterQuantityWholeResolution(t *testing.T) {
	t.Parallel()
	g, engine, source := ashlingQuantityFixture(t)
	victims := []*game.Permanent{
		addCombatPermanent(g, game.Player1, creatureWithPT("Friendly", 10, 10)),
		addCombatPermanent(g, game.Player2, creatureWithPT("Enemy", 10, 10)),
	}
	decoy := addCombatPermanent(g, game.Player2, evidenceCard("Not a Creature", 1))
	var third *game.StackObject
	for i := range 3 {
		obj := activateOrdinal(t, g, engine, source.ObjectID, 0, nil, nil)
		if i == 2 {
			third = obj
			clone := g.Clone()
			engine.resolveTopOfStack(clone, &TurnLog{})
			if clone.Players[game.Player1].Life != 37 || g.Players[game.Player1].Life != 40 {
				t.Fatal("clone lost the pending ordinal or shared mutable quantity")
			}
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		wantCounters, wantLife, wantDamage := i+1, 40, 0
		if i == 2 {
			wantCounters, wantLife, wantDamage = 0, 37, 3
		}
		if source.Counters.Get(counter.PlusOnePlusOne) != wantCounters ||
			source.Counters.Get(counter.Charge) != 2 || source.Counters.Get(counter.Loyalty) != 1 ||
			source.Counters.Get(counter.Indestructible) != 1 || source.MarkedDamage != wantDamage {
			t.Fatalf("resolution %d: counters=%v damage=%d", i+1, source.Counters.All(), source.MarkedDamage)
		}
		for _, victim := range victims {
			if victim.MarkedDamage != wantDamage {
				t.Fatalf("resolution %d: object %d damage=%d", i+1, victim.ObjectID, victim.MarkedDamage)
			}
		}
		for _, player := range g.Players {
			if player.Life != wantLife {
				t.Fatalf("resolution %d: player %d life=%d", i+1, player.ID, player.Life)
			}
		}
	}
	if third.ResolvedAmounts["removed-counter-quantity-2"] != 3 || decoy.MarkedDamage != 0 {
		t.Fatal("actual named-counter quantity or creature domain was lost")
	}
	// A fresh resolving copy keeps the ability identity, not the old scalar.
	copyObj := game.NewStackObjectCopy(third, g.IDGen.Next())
	g.Stack.Push(copyObj)
	engine.resolveTopOfStack(g, &TurnLog{})
	if source.Counters.Get(counter.PlusOnePlusOne) != 1 || g.Players[game.Player1].Life != 37 {
		t.Fatal("fourth resolution copied the third resolution's damage")
	}
	if _, available := copyObj.ResolvedAmounts["removed-counter-quantity-2"]; available {
		t.Fatal("skipped copy publisher retained its old quantity")
	}
	engine.advanceToNextTurn(g)
	g.Turn.PriorityPlayer = game.Player1
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)
	obj := activateOrdinal(t, g, engine, source.ObjectID, 0, nil, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.ResolvedActivatedAbilitiesThisTurn[obj.ActivatedResolutionUse.Val] != 1 || source.Counters.Get(counter.PlusOnePlusOne) != 2 ||
		g.Players[game.Player1].Life != 37 {
		t.Fatal("new turn reused the old ordinal/quantity")
	}
}

func TestCompiledAshlingCounterQuantitySourceIncarnation(t *testing.T) {
	t.Parallel()
	g, engine, source := ashlingQuantityFixture(t)
	for range 2 {
		activateOrdinal(t, g, engine, source.ObjectID, 0, nil, nil)
		engine.resolveTopOfStack(g, &TurnLog{})
	}
	third := activateOrdinal(t, g, engine, source.ObjectID, 0, nil, nil)
	if !movePermanentToZone(g, source, zone.Exile) {
		t.Fatal("source did not leave")
	}
	card, _ := g.GetCardInstance(source.CardInstanceID)
	returned, ok := createCardPermanent(g, card, game.Player1, zone.Exile)
	if !ok {
		t.Fatal("source did not return")
	}
	returned.Counters.Add(counter.PlusOnePlusOne, 8)
	engine.resolveTopOfStack(g, &TurnLog{})
	if returned.Counters.Get(counter.PlusOnePlusOne) != 8 || g.ResolvedActivatedAbilitiesThisTurn[third.ActivatedResolutionUse.Val] != 3 {
		t.Fatal("old ability substituted the new source incarnation")
	}
	if _, available := third.ResolvedAmounts["removed-counter-quantity-2"]; available {
		t.Fatal("unavailable source published a nominal quantity")
	}
	for _, player := range g.Players {
		if player.Life != 40 {
			t.Fatal("unavailable producer caused damage")
		}
	}
	obj := activateOrdinal(t, g, engine, returned.ObjectID, 0, nil, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.ResolvedActivatedAbilitiesThisTurn[obj.ActivatedResolutionUse.Val] != 1 || returned.Counters.Get(counter.PlusOnePlusOne) != 9 {
		t.Fatal("new source incarnation inherited the old ordinal")
	}
}

func TestCompiledCounterQuantityCapturesGroupBeforeRemoval(t *testing.T) {
	t.Parallel()
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Capture", Layout: "normal", TypeLine: "Artifact",
		OracleText: "{1}: If there are three or more charge counters on this artifact, remove all charge counters from this artifact, then you gain that much life.",
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, def)
	source.Counters.Add(counter.Charge, 3)
	obj := &game.StackObject{SourceID: source.ObjectID, Controller: game.Player1}
	NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.ActivatedAbilities[0].Content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if g.Players[game.Player1].Life != 43 || source.Counters.Get(counter.Charge) != 0 {
		t.Fatal("group condition was re-evaluated after counter removal")
	}
}
