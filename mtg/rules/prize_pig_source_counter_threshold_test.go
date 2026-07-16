package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/opt"
)

// prizePigTriggerContent builds the resolution content of Prize Pig's "Whenever
// you gain life, put that many ribbon counters on this creature. Then if there
// are three or more ribbon counters on this creature, remove those counters and
// untap it." trigger: an ungated placement of ribbon counters equal to the life
// gained (read from the triggering event), followed by a remove-all-ribbon and an
// untap, both gated on the source holding three or more ribbon counters.
func prizePigTriggerContent() game.AbilityContent {
	source := game.SourcePermanentReference()
	gate := opt.Val(game.EffectCondition{
		Condition: opt.Val(game.Condition{
			Object: opt.Val(source),
			ObjectMatches: opt.Val(game.Selection{
				RequiredCounter: counter.Ribbon,
				RequiredCounterCount: opt.Val(compare.Int{
					Op:    compare.GreaterOrEqual,
					Value: 3,
				}),
			}),
		}),
	})
	return game.Mode{
		Sequence: []game.Instruction{
			{Primitive: game.AddCounter{
				Amount:      game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountEventLifeChange}),
				Object:      source,
				CounterKind: counter.Ribbon,
			}},
			{
				Primitive: game.RemoveCounter{
					Amount: game.Dynamic(game.DynamicAmount{
						Kind:        game.DynamicAmountObjectCounters,
						CounterKind: counter.Ribbon,
						Object:      source,
					}),
					Object:      source,
					CounterKind: counter.Ribbon,
				},
				Condition:     gate,
				PublishResult: game.ResultKey("counter-threshold-cleared"),
			},
			{
				Primitive: game.Untap{Object: source},
				ResultGate: opt.Val(game.InstructionResultGate{
					Key:       game.ResultKey("counter-threshold-cleared"),
					Succeeded: game.TriTrue,
				}),
			},
		},
	}.Ability()
}

// resolvePrizePigTrigger resolves one Prize Pig trigger for a life-gain of
// lifeGained against the given source permanent.
func resolvePrizePigTrigger(engine *Engine, g *game.Game, pig *game.Permanent, lifeGained int) {
	obj := &game.StackObject{
		ID:              g.IDGen.Next(),
		Controller:      pig.Controller,
		SourceID:        pig.ObjectID,
		TriggerEvent:    game.Event{Kind: game.EventLifeGained, Player: pig.Controller, Amount: lifeGained},
		HasTriggerEvent: true,
	}
	engine.resolveAbilityContentWithChoices(g, obj, prizePigTriggerContent(), [game.NumPlayers]PlayerAgent{}, &TurnLog{})
}

// TestPrizePigSourceCounterThresholdRuntime exercises the composed runtime
// behavior of Prize Pig's life-gain trigger across the edge cases the sequence
// must handle: placing ribbons from the life gained, keeping them below the
// threshold, removing every ribbon and untapping at or above the threshold
// (including already-above and the multi-kind irrelevant case), accumulating
// ribbons across multiple separate life-gain events, and the zero-life no-op.
func TestPrizePigSourceCounterThresholdRuntime(t *testing.T) {
	newPig := func() (*Engine, *game.Game, *game.Permanent) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		pig := addCombatCreaturePermanent(g, game.Player1)
		pig.Tapped = true
		return engine, g, pig
	}

	t.Run("below threshold places ribbons and keeps them tapped", func(t *testing.T) {
		engine, g, pig := newPig()
		resolvePrizePigTrigger(engine, g, pig, 2)
		if got := pig.Counters.Get(counter.Ribbon); got != 2 {
			t.Fatalf("ribbon counters = %d, want 2 (placed, below threshold)", got)
		}
		if !pig.Tapped {
			t.Fatal("source untapped below the threshold, want it to stay tapped")
		}
	})

	t.Run("reaching threshold removes all ribbons and untaps", func(t *testing.T) {
		engine, g, pig := newPig()
		resolvePrizePigTrigger(engine, g, pig, 3)
		if got := pig.Counters.Get(counter.Ribbon); got != 0 {
			t.Fatalf("ribbon counters = %d, want 0 (all removed at threshold)", got)
		}
		if pig.Tapped {
			t.Fatal("source stayed tapped at the threshold, want it untapped")
		}
	})

	t.Run("already above threshold removes every ribbon not just three", func(t *testing.T) {
		engine, g, pig := newPig()
		pig.Counters.Add(counter.Ribbon, 2)
		// Gaining 3 more reaches 5 ribbons; the removal clears all five.
		resolvePrizePigTrigger(engine, g, pig, 3)
		if got := pig.Counters.Get(counter.Ribbon); got != 0 {
			t.Fatalf("ribbon counters = %d, want 0 (remove those counters clears all, not exactly three)", got)
		}
		if pig.Tapped {
			t.Fatal("source stayed tapped above the threshold, want it untapped")
		}
	})

	t.Run("threshold removal leaves other counter kinds untouched", func(t *testing.T) {
		engine, g, pig := newPig()
		pig.Counters.Add(counter.PlusOnePlusOne, 1)
		resolvePrizePigTrigger(engine, g, pig, 3)
		if got := pig.Counters.Get(counter.Ribbon); got != 0 {
			t.Fatalf("ribbon counters = %d, want 0", got)
		}
		if got := pig.Counters.Get(counter.PlusOnePlusOne); got != 1 {
			t.Fatalf("+1/+1 counters = %d, want 1 (remove those ribbon counters must not touch other kinds)", got)
		}
	})

	t.Run("multiple separate life events accumulate then clear", func(t *testing.T) {
		engine, g, pig := newPig()
		resolvePrizePigTrigger(engine, g, pig, 1)
		if got := pig.Counters.Get(counter.Ribbon); got != 1 || !pig.Tapped {
			t.Fatalf("after first gain: ribbons=%d tapped=%v, want 1 and tapped", got, pig.Tapped)
		}
		resolvePrizePigTrigger(engine, g, pig, 1)
		if got := pig.Counters.Get(counter.Ribbon); got != 2 || !pig.Tapped {
			t.Fatalf("after second gain: ribbons=%d tapped=%v, want 2 and tapped", got, pig.Tapped)
		}
		resolvePrizePigTrigger(engine, g, pig, 1)
		if got := pig.Counters.Get(counter.Ribbon); got != 0 {
			t.Fatalf("after third gain: ribbons=%d, want 0 (threshold reached, all removed)", got)
		}
		if pig.Tapped {
			t.Fatal("source stayed tapped after reaching the threshold, want it untapped")
		}
	})

	t.Run("zero life gain is a no-op", func(t *testing.T) {
		engine, g, pig := newPig()
		resolvePrizePigTrigger(engine, g, pig, 0)
		if got := pig.Counters.Get(counter.Ribbon); got != 0 {
			t.Fatalf("ribbon counters = %d, want 0 (no life gained)", got)
		}
		if !pig.Tapped {
			t.Fatal("source untapped on a zero-life trigger, want it to stay tapped")
		}
	})
}
