package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// Source review: CR 2026-09-25 603.4 (an intervening "if" is checked again as
// the ability resolves; if false, the ability does nothing) and 702.100a/c
// (evolve compares power to power and toughness to toughness; a creature can't
// have greater power or toughness than a noncreature permanent). Fathom Mage
// rulings 2013-04-15: the comparison happens again on resolution; two 3/3s
// entering beside a 2/2 evolver trigger twice but the second ability does
// nothing; the greater stat may switch from toughness to power and the ability
// still resolves; an entered creature that has left the battlefield is compared
// using its last known power and toughness.

func evolveResolutionScenario(t *testing.T, evolverPower, evolverToughness int) (*game.Game, *Engine, *game.Permanent) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := vanillaCreature("Evolving Creature", evolverPower, evolverToughness)
	def.StaticAbilities = []game.StaticAbility{game.EvolveStaticBody}
	return g, NewEngine(nil), addCombatPermanent(g, game.Player1, def)
}

func emitEvolveEntry(g *game.Game, entered ...*game.Permanent) {
	for _, permanent := range entered {
		emitEvent(g, game.Event{
			Kind: game.EventPermanentEnteredBattlefield, Controller: game.Player1, PermanentID: permanent.ObjectID,
		})
	}
}

func TestEvolveDoesNothingWhenComparisonFailsOnResolution(t *testing.T) {
	g, engine, evolver := evolveResolutionScenario(t, 1, 1)
	entering := addCombatPermanent(g, game.Player1, vanillaCreature("Two-Two", 2, 2))
	emitEvolveEntry(g, entering)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("evolve did not trigger for a 2/2 entering beside a 1/1")
	}
	evolver.Counters.Add(counter.PlusOnePlusOne, 5)

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := evolver.Counters.Get(counter.PlusOnePlusOne); got != 5 {
		t.Fatalf("evolve counters = %d, want 5: a 6/6 evolver must not evolve off a 2/2 on resolution", got)
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want the evolve ability removed", g.Stack.Size())
	}
}

func TestEvolveSimultaneousEntriesRecheckEachResolution(t *testing.T) {
	g, engine, evolver := evolveResolutionScenario(t, 2, 2)
	first := addCombatPermanent(g, game.Player1, vanillaCreature("First Three-Three", 3, 3))
	second := addCombatPermanent(g, game.Player1, vanillaCreature("Second Three-Three", 3, 3))
	emitEvolveEntry(g, first, second)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("evolve did not trigger for two 3/3s entering beside a 2/2")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want two evolve triggers", got)
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if got := evolver.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("evolve counters after first resolution = %d, want 1", got)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := evolver.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("evolve counters after second resolution = %d, want 1: a 3/3 evolver must not evolve off another 3/3", got)
	}
}

func TestEvolveResolvesWhenGreaterStatSwitches(t *testing.T) {
	g, engine, evolver := evolveResolutionScenario(t, 2, 2)
	entering := addCombatPermanent(g, game.Player1, vanillaCreature("One-Three", 1, 3))
	emitEvolveEntry(g, entering)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("evolve did not trigger for a 1/3 entering beside a 2/2")
	}
	addEffectSpellToStack(g, game.Player1, game.ModifyPT{
		Object:         game.TargetPermanentReference(0),
		PowerDelta:     game.Fixed(2),
		ToughnessDelta: game.Fixed(-2),
		Duration:       game.DurationUntilEndOfTurn,
	}, []game.Target{game.PermanentTarget(entering.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if power := effectivePower(g, entering); power != 3 {
		t.Fatalf("entered creature power = %d, want 3 after +2/-2", power)
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := evolver.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("evolve counters = %d, want 1: a 3/1 still has greater power than a 2/2", got)
	}
}

func TestEvolveUsesEnteredCreatureLastKnownInformation(t *testing.T) {
	for _, test := range []struct {
		name         string
		shrink       bool
		wantCounters int
	}{
		{name: "left as a 3/3", wantCounters: 1},
		{name: "left after shrinking to 1/1", shrink: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			g, engine, evolver := evolveResolutionScenario(t, 2, 2)
			entering := addCombatPermanent(g, game.Player1, vanillaCreature("Three-Three", 3, 3))
			emitEvolveEntry(g, entering)
			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("evolve did not trigger for a 3/3 entering beside a 2/2")
			}
			target := []game.Target{game.PermanentTarget(entering.ObjectID)}
			if test.shrink {
				addEffectSpellToStack(g, game.Player1, game.ModifyPT{
					Object:         game.TargetPermanentReference(0),
					PowerDelta:     game.Fixed(-2),
					ToughnessDelta: game.Fixed(-2),
					Duration:       game.DurationUntilEndOfTurn,
				}, target)
				engine.resolveTopOfStack(g, &TurnLog{})
			}
			addEffectSpellToStack(g, game.Player1, game.Destroy{Object: game.TargetPermanentReference(0)}, target)
			engine.resolveTopOfStack(g, &TurnLog{})
			if _, ok := permanentByObjectID(g, entering.ObjectID); ok {
				t.Fatal("entered creature is still on the battlefield")
			}

			engine.resolveTopOfStack(g, &TurnLog{})

			if got := evolver.Counters.Get(counter.PlusOnePlusOne); got != test.wantCounters {
				t.Fatalf("evolve counters = %d, want %d from the entered creature's last known stats", got, test.wantCounters)
			}
		})
	}
}
