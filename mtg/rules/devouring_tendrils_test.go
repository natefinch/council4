package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: issue #3415. Oracle ID c0112586-3826-452e-9e44-7e6edd197f78
// from the pinned corpus 6da9ced0…: "Target creature you control deals damage
// equal to its power to target creature or planeswalker you don't control.
// When the permanent you don't control dies this turn, you gain 2 life."
// https://api.scryfall.com/cards/f7444555-da19-4d2d-ad81-233c54fbb78e/rulings
// returned one WotC ruling (2021-04-16): the delayed ability triggers when that
// permanent dies this turn even if it didn't die because of the damage. CR
// 2026-09-25 (verified official copy): 120.3c (damage removes loyalty),
// 603.7a-b ("this turn" delayed trigger created on resolution), 603.7c (bound
// to that object), 603.7d (spell is the source; its controller controls it),
// 608.2b (illegal targets; missing information means that part doesn't
// happen), 700.4 (dies) and 704.5i (0 loyalty). Raw corpus card and rulings
// are retained under .cardwork/gap-3415.

type devouringTendrilsScenario struct {
	g       *game.Game
	engine  *Engine
	agents  [game.NumPlayers]PlayerAgent
	spellID id.ID
}

func newDevouringTendrilsScenario(t *testing.T) *devouringTendrilsScenario {
	t.Helper()
	// The legacy registration is name-keyed; the corpus has one card with
	// this name (Oracle ID c0112586-3826-452e-9e44-7e6edd197f78).
	matches := cards.NewDefaultRegistry().LookupAll("Devouring Tendrils")
	if len(matches) != 1 {
		t.Fatalf("default registry has %d Devouring Tendrils definitions, want 1", len(matches))
	}
	def := matches[0]
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	spellID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.G, 2)
	return &devouringTendrilsScenario{
		g:       g,
		engine:  NewEngine(nil),
		agents:  [game.NumPlayers]PlayerAgent{game.Player1: monarchNoChoiceAgent{}},
		spellID: spellID,
	}
}

// cast casts the real card through the ordinary cast action, paying {1}{G}.
func (s *devouringTendrilsScenario) cast(t *testing.T, mine, theirs *game.Permanent) {
	t.Helper()
	setMainPhasePriority(s.g, game.Player1)
	if !s.engine.applyAction(s.g, game.Player1, action.CastSpell(s.spellID, []game.Target{
		game.PermanentTarget(mine.ObjectID),
		game.PermanentTarget(theirs.ObjectID),
	}, 0, nil)) {
		t.Fatal("Devouring Tendrils cast failed")
	}
	if s.g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("mana left after casting = %d, want 0", s.g.Players[game.Player1].ManaPool.Total())
	}
}

func (s *devouringTendrilsScenario) resolve(t *testing.T) {
	t.Helper()
	s.engine.resolveTopOfStackWithChoices(s.g, s.agents, &TurnLog{})
	if s.g.Stack.Size() != 0 {
		t.Fatalf("stack size after resolution = %d, want 0", s.g.Stack.Size())
	}
	if !s.g.Players[game.Player1].Graveyard.Contains(s.spellID) {
		t.Fatal("Devouring Tendrils did not go to its owner's graveyard")
	}
}

func (s *devouringTendrilsScenario) castAndResolve(t *testing.T, mine, theirs *game.Permanent) {
	t.Helper()
	s.cast(t, mine, theirs)
	s.resolve(t)
}

func (s *devouringTendrilsScenario) life(player game.PlayerID) int {
	return s.g.Players[player].Life
}

func (s *devouringTendrilsScenario) assertLife(t *testing.T, want int, context string) {
	t.Helper()
	if got := s.life(game.Player1); got != want {
		t.Fatalf("%s: Player1 life = %d, want %d", context, got, want)
	}
}

// fire runs state-based actions, puts the single delayed trigger on the stack
// and resolves it, returning the trigger as it existed on the stack.
func (s *devouringTendrilsScenario) fire(t *testing.T) game.StackObject {
	t.Helper()
	s.engine.applyStateBasedActions(s.g)
	log := TurnLog{}
	if !s.engine.putTriggeredAbilitiesOnStackWithChoices(s.g, s.agents, &log) || s.g.Stack.Size() != 1 {
		t.Fatalf("delayed dies trigger did not fire alone: stack size %d", s.g.Stack.Size())
	}
	obj, _ := s.g.Stack.Peek()
	fired := *obj
	s.engine.resolveTopOfStackWithChoices(s.g, s.agents, &log)
	if s.g.Stack.Size() != 0 {
		t.Fatalf("stack size after delayed trigger = %d, want 0", s.g.Stack.Size())
	}
	return fired
}

func (s *devouringTendrilsScenario) assertNoFire(t *testing.T) {
	t.Helper()
	s.engine.applyStateBasedActions(s.g)
	if s.engine.putTriggeredAbilitiesOnStackWithChoices(s.g, s.agents, &TurnLog{}) || s.g.Stack.Size() != 0 {
		t.Fatalf("delayed dies trigger fired unexpectedly: stack size %d", s.g.Stack.Size())
	}
}

func (s *devouringTendrilsScenario) assertBoundTo(t *testing.T, permanent *game.Permanent) {
	t.Helper()
	if len(s.g.DelayedTriggers) != 1 || s.g.DelayedTriggers[0].BoundDyingObjectID != permanent.ObjectID {
		t.Fatalf("delayed triggers = %+v, want one bound to %v", s.g.DelayedTriggers, permanent.ObjectID)
	}
}

// destroy kills permanent through an ordinary destroy spell resolution.
func (s *devouringTendrilsScenario) destroy(t *testing.T, permanent *game.Permanent) {
	t.Helper()
	addEffectSpellToStack(s.g, game.Player2, game.Destroy{Object: game.TargetPermanentReference(0)},
		[]game.Target{game.PermanentTarget(permanent.ObjectID)})
	s.engine.resolveTopOfStack(s.g, &TurnLog{})
	if _, ok := permanentByObjectID(s.g, permanent.ObjectID); ok {
		t.Fatal("destroy effect did not remove the permanent")
	}
}

// TestDevouringTendrilsLethalDamageGainsLifeOnlyFromDelayedTrigger proves the
// spell resolves only the damage and schedules the trigger; the 2 life comes
// from the delayed trigger after state-based actions kill the target.
func TestDevouringTendrilsLethalDamageGainsLifeOnlyFromDelayedTrigger(t *testing.T) {
	s := newDevouringTendrilsScenario(t)
	mine := addCombatCreaturePermanentWithPower(s.g, game.Player1, 3)
	theirs := addCombatCreaturePermanentWithPower(s.g, game.Player2, 3)

	s.cast(t, mine, theirs)
	stackObj, _ := s.g.Stack.Peek()
	stackID := stackObj.ID
	s.resolve(t)

	if theirs.MarkedDamage != 3 {
		t.Fatalf("target marked damage = %d, want 3", theirs.MarkedDamage)
	}
	assertEvent(t, s.g.Events, game.EventDamageDealt, func(event game.Event) bool {
		return event.Amount == 3 && event.SourceID == mine.CardInstanceID && event.PermanentID == theirs.ObjectID
	})
	s.assertLife(t, 40, "after resolution")
	s.assertBoundTo(t, theirs)

	fired := s.fire(t)

	if fired.Controller != game.Player1 || fired.SourceID != stackID || fired.SourceCardID != s.spellID {
		t.Fatalf("delayed trigger controller/source/card = %v/%v/%v, want %v/%v/%v",
			fired.Controller, fired.SourceID, fired.SourceCardID, game.Player1, stackID, s.spellID)
	}
	if !s.g.Players[game.Player2].Graveyard.Contains(theirs.CardInstanceID) {
		t.Fatal("target did not die")
	}
	s.assertLife(t, 42, "after the target died")
	if got := s.life(game.Player2); got != 40 {
		t.Fatalf("Player2 life = %d, want 40", got)
	}
	s.assertNoFire(t)
	expireCleanupDurations(s.g)
	if len(s.g.DelayedTriggers) != 0 {
		t.Fatalf("this-turn delayed trigger survived cleanup: %+v", s.g.DelayedTriggers)
	}
}

// TestDevouringTendrilsSurvivingTargetGainsNoLife is the issue's probe: a 1/1
// damaging a 10/10 must not gain life.
func TestDevouringTendrilsSurvivingTargetGainsNoLife(t *testing.T) {
	s := newDevouringTendrilsScenario(t)
	mine := addCombatCreaturePermanentWithPower(s.g, game.Player1, 1)
	theirs := addCombatCreaturePermanentWithPower(s.g, game.Player2, 10)

	s.castAndResolve(t, mine, theirs)

	if theirs.MarkedDamage != 1 {
		t.Fatalf("target marked damage = %d, want 1", theirs.MarkedDamage)
	}
	s.assertNoFire(t)
	if _, ok := permanentByObjectID(s.g, theirs.ObjectID); !ok {
		t.Fatal("10/10 target died to 1 damage")
	}
	s.assertLife(t, 40, "surviving target")
	s.assertBoundTo(t, theirs)
}

// TestDevouringTendrilsLaterDeathThisTurnGainsLife proves the ruling: the
// trigger fires when that permanent dies later this turn for any reason, and
// another permanent dying does not fire it.
func TestDevouringTendrilsLaterDeathThisTurnGainsLife(t *testing.T) {
	s := newDevouringTendrilsScenario(t)
	mine := addCombatCreaturePermanentWithPower(s.g, game.Player1, 1)
	theirs := addCombatCreaturePermanentWithPower(s.g, game.Player2, 10)
	bystander := addCombatCreaturePermanentWithPower(s.g, game.Player2, 1)

	s.castAndResolve(t, mine, theirs)
	s.destroy(t, bystander)
	s.assertNoFire(t)
	s.destroy(t, mine)
	s.assertNoFire(t)
	s.assertLife(t, 40, "after other permanents died")

	s.destroy(t, theirs)
	s.fire(t)
	s.assertLife(t, 42, "after the target was destroyed later this turn")
}

// TestDevouringTendrilsPlaneswalkerTarget proves damage removes loyalty and a
// planeswalker put into the graveyard for 0 loyalty dies and fires the trigger;
// a surviving planeswalker gains nothing.
func TestDevouringTendrilsPlaneswalkerTarget(t *testing.T) {
	t.Run("dies", func(t *testing.T) {
		s := newDevouringTendrilsScenario(t)
		mine := addCombatCreaturePermanentWithPower(s.g, game.Player1, 4)
		walker := addLoyaltyPlaneswalker(s.g, game.Player2, 4)

		s.castAndResolve(t, mine, walker)
		s.assertLife(t, 40, "after resolution")
		s.assertBoundTo(t, walker)
		s.fire(t)

		if !s.g.Players[game.Player2].Graveyard.Contains(walker.CardInstanceID) {
			t.Fatal("planeswalker with 0 loyalty did not die")
		}
		s.assertLife(t, 42, "after the planeswalker died")
	})

	t.Run("survives", func(t *testing.T) {
		s := newDevouringTendrilsScenario(t)
		mine := addCombatCreaturePermanentWithPower(s.g, game.Player1, 2)
		walker := addLoyaltyPlaneswalker(s.g, game.Player2, 5)

		s.castAndResolve(t, mine, walker)
		if got := walker.Counters.Get(counter.Loyalty); got != 3 {
			t.Fatalf("planeswalker loyalty = %d, want 3", got)
		}
		s.assertNoFire(t)
		s.assertLife(t, 40, "surviving planeswalker")
	})
}

// TestDevouringTendrilsNonGraveyardDepartureGainsNoLife proves leaving the
// battlefield for exile is not dying, and a returned card is a new object.
func TestDevouringTendrilsNonGraveyardDepartureGainsNoLife(t *testing.T) {
	s := newDevouringTendrilsScenario(t)
	mine := addCombatCreaturePermanentWithPower(s.g, game.Player1, 1)
	theirs := addCombatCreaturePermanentWithPower(s.g, game.Player2, 10)

	s.castAndResolve(t, mine, theirs)
	if !movePermanentToZone(s.g, theirs, zone.Exile) {
		t.Fatal("exile failed")
	}
	s.assertNoFire(t)
	s.assertLife(t, 40, "after the target was exiled")
}

// TestDevouringTendrilsTriggerExpiresAtCleanup proves the "this turn" window:
// the target dying on a later turn gains nothing.
func TestDevouringTendrilsTriggerExpiresAtCleanup(t *testing.T) {
	s := newDevouringTendrilsScenario(t)
	mine := addCombatCreaturePermanentWithPower(s.g, game.Player1, 1)
	theirs := addCombatCreaturePermanentWithPower(s.g, game.Player2, 10)

	s.castAndResolve(t, mine, theirs)
	expireCleanupDurations(s.g)
	s.g.Turn.TurnNumber++
	if len(s.g.DelayedTriggers) != 0 {
		t.Fatalf("this-turn delayed trigger survived cleanup: %+v", s.g.DelayedTriggers)
	}
	s.destroy(t, theirs)
	s.assertNoFire(t)
	s.assertLife(t, 40, "target died on a later turn")
}

// TestDevouringTendrilsIllegalTargets applies CR 608.2b to each target: an
// illegal source creature deals no damage but the trigger is still created for
// the legal permanent; an illegal permanent you don't control is neither
// damaged nor tracked.
func TestDevouringTendrilsIllegalTargets(t *testing.T) {
	t.Run("creature you control illegal", func(t *testing.T) {
		s := newDevouringTendrilsScenario(t)
		mine := addCombatCreaturePermanentWithPower(s.g, game.Player1, 5)
		theirs := addCombatCreaturePermanentWithPower(s.g, game.Player2, 3)

		s.cast(t, mine, theirs)
		mine.Controller = game.Player3
		s.resolve(t)

		if theirs.MarkedDamage != 0 {
			t.Fatalf("illegal source dealt %d damage", theirs.MarkedDamage)
		}
		s.assertBoundTo(t, theirs)
		s.destroy(t, theirs)
		s.fire(t)
		s.assertLife(t, 42, "legal target died")
	})

	t.Run("permanent you don't control illegal", func(t *testing.T) {
		s := newDevouringTendrilsScenario(t)
		mine := addCombatCreaturePermanentWithPower(s.g, game.Player1, 5)
		theirs := addCombatCreaturePermanentWithPower(s.g, game.Player2, 3)

		s.cast(t, mine, theirs)
		theirs.Controller = game.Player1
		s.resolve(t)

		if theirs.MarkedDamage != 0 {
			t.Fatalf("illegal target was dealt %d damage", theirs.MarkedDamage)
		}
		for _, trigger := range s.g.DelayedTriggers {
			if trigger.BoundDyingObjectID != 0 {
				t.Fatalf("illegal target was tracked: %+v", trigger)
			}
		}
		s.destroy(t, theirs)
		s.assertNoFire(t)
		s.assertLife(t, 40, "illegal target died")
	})
}

// TestDevouringTendrilsDelayedTriggerSurvivesClone proves the scheduled
// binding is game-owned state: a clone fires independently of the original.
func TestDevouringTendrilsDelayedTriggerSurvivesClone(t *testing.T) {
	s := newDevouringTendrilsScenario(t)
	mine := addCombatCreaturePermanentWithPower(s.g, game.Player1, 1)
	theirs := addCombatCreaturePermanentWithPower(s.g, game.Player2, 10)
	s.castAndResolve(t, mine, theirs)

	original := s.g
	s.g = original.Clone()
	clonedTarget, ok := permanentByObjectID(s.g, theirs.ObjectID)
	if !ok {
		t.Fatal("cloned target missing")
	}
	s.destroy(t, clonedTarget)
	s.fire(t)
	s.assertLife(t, 42, "clone after target died")

	if got := original.Players[game.Player1].Life; got != 40 {
		t.Fatalf("original game life = %d, want 40", got)
	}
	if len(original.DelayedTriggers) != 1 {
		t.Fatalf("original delayed triggers = %d, want 1", len(original.DelayedTriggers))
	}
}
