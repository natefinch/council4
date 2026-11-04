package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: issue #3564. Pinned corpus 6da9ced0…: Fight for the Throne
// (Oracle ID 53129de9-4809-4c8c-9d11-37369899e70a) "... When the creature an
// opponent controls dies this turn, if you control your commander, you become
// the monarch." and Devouring Tendrils (Oracle ID
// c0112586-3826-452e-9e44-7e6edd197f78) "... When the permanent you don't
// control dies this turn, you gain 2 life." CR 2026-09-25 (verified copy):
// 514.2 (cleanup removes damage and ends "until end of turn" and "this turn"
// effects), 514.3a (cleanup then performs state-based actions, puts triggered
// abilities on the stack and grants priority; another cleanup step follows),
// 603.7a-b (a delayed triggered ability is created on resolution and with a
// stated duration such as "this turn" may trigger more than once). The delayed
// ability is not a "this turn" effect, so it still sees the cleanup deaths and
// cleanup-priority events of its own turn, and never a later turn's events. Raw
// CR text is retained under .cardwork/cr-20260925.txt.

func endingPhasePassPolicies() [game.NumPlayers]PlayerAgent {
	var policies [game.NumPlayers]PlayerAgent
	for i := range policies {
		policies[i] = monarchNoChoiceAgent{}
	}
	return policies
}

func runEndingPhaseForTest(engine *Engine, g *game.Game, policies [game.NumPlayers]PlayerAgent) *TurnLog {
	log := &TurnLog{}
	engine.setReplacementChoiceContext(g, policies, log)
	engine.runEndingPhaseWithLog(g, policies, log)
	g.ClearChoiceContext()
	return log
}

func cleanupStepCount(log *TurnLog) int {
	count := 0
	for _, step := range log.Steps {
		if step.Step == game.StepCleanup {
			count++
		}
	}
	return count
}

// castRegistryFightForThrone casts the default-registry Fight for the Throne
// as Player1 and resolves it through the real spell-resolution path.
func castRegistryFightForThrone(t *testing.T, g *game.Game, engine *Engine, mine, opp *game.Permanent) {
	t.Helper()
	addImplementationSpellToStack(g, game.Player1, registryCard(t, "Fight for the Throne"), []game.Target{
		game.PermanentTarget(mine.ObjectID),
		game.PermanentTarget(opp.ObjectID),
	})
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("Fight for the Throne was not put on the stack")
	}
	obj.TargetCounts = []int{1, 1}
	engine.resolveTopOfStack(g, &TurnLog{})
	if len(g.DelayedTriggers) != 1 || g.DelayedTriggers[0].BoundDyingObjectID != opp.ObjectID {
		t.Fatalf("delayed triggers = %+v, want one bound to %v", g.DelayedTriggers, opp.ObjectID)
	}
}

// newFightForThroneCleanupDeath stages the issue's scenario: the opposing 0/0
// creature is sustained by a temporary +0/+3 bonus, survives the fight, and
// will die as a cleanup state-based action once the bonus ends (CR 514.2).
func newFightForThroneCleanupDeath(t *testing.T) (*game.Game, *Engine, *game.Permanent) {
	t.Helper()
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	addCommanderPermanent(g, game.Player1)
	mine := addCombatCreaturePermanentWithPower(g, game.Player1, 1) // 2/2 after the counter
	opp := addCreatureWithPT(g, game.Player2, 0, 0)
	opp.TemporaryToughnessModifier = 3
	castRegistryFightForThrone(t, g, engine, mine, opp)
	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, opp.ObjectID); !ok || opp.MarkedDamage != 2 {
		t.Fatal("the sustained opposing creature should survive the fight with 2 damage")
	}
	if engine.putTriggeredAbilitiesOnStackWithChoices(g, endingPhasePassPolicies(), &TurnLog{}) {
		t.Fatal("delayed trigger fired before the creature died")
	}
	return g, engine, opp
}

func TestFightForThroneCleanupDeathMakesMonarch(t *testing.T) {
	g, engine, opp := newFightForThroneCleanupDeath(t)

	log := runEndingPhaseForTest(engine, g, endingPhasePassPolicies())

	if !g.Players[game.Player2].Graveyard.Contains(opp.CardInstanceID) {
		t.Fatal("the fought creature did not die as a cleanup state-based action")
	}
	if !g.Players[game.Player1].IsMonarch {
		t.Fatal("the this-turn delayed trigger did not see the cleanup death")
	}
	if got := cleanupStepCount(log); got != 2 {
		t.Fatalf("cleanup steps = %d, want 2 (the trigger opens priority, then another cleanup step)", got)
	}
	if g.Turn.Step != game.StepCleanup || g.Stack.Size() != 0 {
		t.Fatalf("turn ended at step %v with %d stack objects, want an empty cleanup", g.Turn.Step, g.Stack.Size())
	}
	// The window lasts the whole turn; ending the turn removes the trigger.
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed triggers before the turn ended = %d, want 1", len(g.DelayedTriggers))
	}
	engine.advanceToNextTurn(g)
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("this-turn delayed trigger survived the end of its turn: %+v", g.DelayedTriggers)
	}
}

// TestFightForThroneCleanupDeathClonedGame proves a clone taken before cleanup
// carries the trigger and plays its own cleanup without touching the original,
// and both games reach the same result.
func TestFightForThroneCleanupDeathClonedGame(t *testing.T) {
	g, engine, opp := newFightForThroneCleanupDeath(t)
	clone := g.Clone()

	runEndingPhaseForTest(NewEngine(nil), clone, endingPhasePassPolicies())

	if !clone.Players[game.Player1].IsMonarch || !clone.Players[game.Player2].Graveyard.Contains(opp.CardInstanceID) {
		t.Fatal("the cloned game's cleanup death did not fire the delayed trigger")
	}
	if g.Players[game.Player1].IsMonarch || len(g.DelayedTriggers) != 1 {
		t.Fatal("running the clone's cleanup changed the original game")
	}
	if _, ok := permanentByObjectID(g, opp.ObjectID); !ok {
		t.Fatal("running the clone's cleanup killed the original game's creature")
	}

	runEndingPhaseForTest(engine, g, endingPhasePassPolicies())
	if !g.Players[game.Player1].IsMonarch {
		t.Fatal("the original game's cleanup death did not fire the delayed trigger")
	}
}

func TestDevouringTendrilsCleanupDeathGainsLife(t *testing.T) {
	s := newDevouringTendrilsScenario(t)
	mine := addCombatCreaturePermanentWithPower(s.g, game.Player1, 1)
	theirs := addCreatureWithPT(s.g, game.Player2, 0, 0)
	theirs.TemporaryToughnessModifier = 3

	s.castAndResolve(t, mine, theirs)
	s.assertNoFire(t)
	s.assertBoundTo(t, theirs)

	runEndingPhaseForTest(s.engine, s.g, endingPhasePassPolicies())

	if !s.g.Players[game.Player2].Graveyard.Contains(theirs.CardInstanceID) {
		t.Fatal("the target did not die as a cleanup state-based action")
	}
	s.assertLife(t, 42, "after the target died during cleanup")
}

// cleanupDestroyPolicy casts its destroy spell at the target during the first
// cleanup priority window it receives (CR 514.3a).
type cleanupDestroyPolicy struct {
	monarchNoChoiceAgent
	spell  id.ID
	target id.ID
	cast   bool
}

func (p *cleanupDestroyPolicy) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	if obs.Turn.Step != game.StepCleanup || p.cast {
		return action.Pass()
	}
	for _, candidate := range legal {
		cast, ok := candidate.CastSpellPayload()
		if ok && cast.CardID == p.spell && len(cast.Targets) == 1 && cast.Targets[0].PermanentID == p.target {
			p.cast = true
			return candidate
		}
	}
	return action.Pass()
}

func TestDevouringTendrilsDeathDuringCleanupPriorityGainsLife(t *testing.T) {
	s := newDevouringTendrilsScenario(t)
	mine := addCombatCreaturePermanentWithPower(s.g, game.Player1, 1)
	theirs := addCombatCreaturePermanentWithPower(s.g, game.Player2, 10)
	// A different creature dying in cleanup opens the cleanup priority window.
	casualty := addCreatureWithPT(s.g, game.Player2, 0, 0)
	casualty.TemporaryToughnessModifier = 1

	s.castAndResolve(t, mine, theirs)
	s.assertNoFire(t)

	destroy := addCardToHand(s.g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Cleanup Destroy", Types: []types.Card{types.Instant},
		ManaCost: opt.Val(cost.Mana{cost.B}),
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1, MaxTargets: 1, Constraint: "target creature",
				Allow:     game.TargetAllowPermanent,
				Selection: opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
			}},
			Sequence: []game.Instruction{{Primitive: game.Destroy{Object: game.TargetPermanentReference(0)}}},
		}.Ability()),
	}})
	s.g.Players[game.Player1].ManaPool.AddPersistent(mana.B, 1)
	policy := &cleanupDestroyPolicy{spell: destroy, target: theirs.ObjectID}
	policies := endingPhasePassPolicies()
	policies[game.Player1] = policy

	log := runEndingPhaseForTest(s.engine, s.g, policies)

	if !policy.cast {
		t.Fatal("the destroy spell was not cast during cleanup priority")
	}
	if !s.g.Players[game.Player2].Graveyard.Contains(casualty.CardInstanceID) ||
		!s.g.Players[game.Player2].Graveyard.Contains(theirs.CardInstanceID) {
		t.Fatal("the cleanup casualty and the destroyed target should both have died")
	}
	s.assertLife(t, 42, "after the target was destroyed during cleanup priority")
	if got := cleanupStepCount(log); got < 2 {
		t.Fatalf("cleanup steps = %d, want at least 2", got)
	}
}

// TestDevouringTendrilsQuietCleanupThenLaterTurn proves an ordinary cleanup with
// no matching event fires nothing and keeps the trigger until the turn ends,
// and the target dying on a later turn fires nothing.
func TestDevouringTendrilsQuietCleanupThenLaterTurn(t *testing.T) {
	setup := func(t *testing.T) (*devouringTendrilsScenario, *game.Permanent) {
		t.Helper()
		s := newDevouringTendrilsScenario(t)
		mine := addCombatCreaturePermanentWithPower(s.g, game.Player1, 1)
		theirs := addCombatCreaturePermanentWithPower(s.g, game.Player2, 10)
		s.castAndResolve(t, mine, theirs)

		log := runEndingPhaseForTest(s.engine, s.g, endingPhasePassPolicies())

		if got := cleanupStepCount(log); got != 1 {
			t.Fatalf("cleanup steps = %d, want 1", got)
		}
		if theirs.MarkedDamage != 0 {
			t.Fatalf("cleanup left %d damage on the target", theirs.MarkedDamage)
		}
		s.assertLife(t, 40, "after a cleanup with no death")
		s.assertBoundTo(t, theirs)
		return s, theirs
	}

	t.Run("turn ends", func(t *testing.T) {
		s, theirs := setup(t)
		s.engine.advanceToNextTurn(s.g)
		if len(s.g.DelayedTriggers) != 0 {
			t.Fatalf("this-turn delayed trigger survived the end of its turn: %+v", s.g.DelayedTriggers)
		}
		s.destroy(t, theirs)
		s.assertNoFire(t)
		s.assertLife(t, 40, "target died on a later turn")
	})

	t.Run("stale trigger on a later turn", func(t *testing.T) {
		s, theirs := setup(t)
		s.g.Turn.TurnNumber++
		s.destroy(t, theirs)
		s.assertNoFire(t)
		s.assertLife(t, 40, "target died on a later turn")
		if len(s.g.DelayedTriggers) != 0 {
			t.Fatalf("a stale this-turn delayed trigger was kept: %+v", s.g.DelayedTriggers)
		}
	})
}
