package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// distributeCounters builds a Distribute AddCounter instruction whose object
// addresses every permanent chosen for the spec at index 0.
func distributeCounters(total int) game.AddCounter {
	return game.AddCounter{
		Amount:      game.Fixed(total),
		Object:      game.AllTargetPermanentsReference(0),
		CounterKind: counter.PlusOnePlusOne,
		Distribute:  true,
	}
}

// TestDistributeCountersHonorsControllerAllocation proves the controller's
// resolution-time allocation is honored: each chosen target receives the number
// of counters the player assigned and the allocations sum to the fixed total.
func TestDistributeCountersHonorsControllerAllocation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCreaturePermanent(g, game.Player1)
	second := addCreaturePermanent(g, game.Player1)
	addEffectSpellToStack(g, game.Player1, distributeCounters(3), []game.Target{
		game.PermanentTarget(first.ObjectID),
		game.PermanentTarget(second.ObjectID),
	})
	// Allocate 1 to the first target, 2 to the second (multiset of option indices).
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0, 1, 1}}},
	}
	log := TurnLog{}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	gotFirst, ok := permanentByObjectID(g, first.ObjectID)
	if !ok {
		t.Fatal("first target left the battlefield")
	}
	gotSecond, ok := permanentByObjectID(g, second.ObjectID)
	if !ok {
		t.Fatal("second target left the battlefield")
	}
	if got := gotFirst.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("first target +1/+1 counters = %d, want 1", got)
	}
	if got := gotSecond.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("second target +1/+1 counters = %d, want 2", got)
	}
	total := gotFirst.Counters.Get(counter.PlusOnePlusOne) + gotSecond.Counters.Get(counter.PlusOnePlusOne)
	if total != 3 {
		t.Fatalf("total counters placed = %d, want 3 (conserved)", total)
	}
	if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoiceCounterAllocation {
		t.Fatalf("choices = %+v, want one ChoiceCounterAllocation", log.Choices)
	}
}

// TestDistributeCountersDropsShareOfIllegalTarget proves that a target which has
// become illegal since the division was chosen receives no counters and its
// assigned share is lost, never redistributed to the surviving targets. The
// controller assigned 1 to the first target and 2 to the second; the second is
// removed before resolution, so only 1 counter is placed (on the first), NOT the
// full 3 concentrated onto the survivor.
func TestDistributeCountersDropsShareOfIllegalTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCreaturePermanent(g, game.Player1)
	second := addCreaturePermanent(g, game.Player1)
	addEffectSpellToStack(g, game.Player1, distributeCounters(3), []game.Target{
		game.PermanentTarget(first.ObjectID),
		game.PermanentTarget(second.ObjectID),
	})
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0, 1, 1}}},
	}
	if _, ok := removePermanentFromBattlefield(g, second.ObjectID); !ok {
		t.Fatal("failed to remove the second target before resolution")
	}
	log := TurnLog{}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	gotFirst, ok := permanentByObjectID(g, first.ObjectID)
	if !ok {
		t.Fatal("first target left the battlefield")
	}
	if got := gotFirst.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("surviving target +1/+1 counters = %d, want 1 (illegal share dropped)", got)
	}
}
