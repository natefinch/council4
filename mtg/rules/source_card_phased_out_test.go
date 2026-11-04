package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// Source review (#3614): pinned corpus Sea Serpent (Oracle
// c16495fc-784d-4bac-9a68-ed437008df73, printing
// 1621438d-717a-477c-93ff-ff7415da70c5): "This creature can't attack unless
// defending player controls an Island. When you control no Islands, sacrifice
// this creature." Gemrazer (Oracle 3dfb0c0a-b68f-43b9-8475-28d0192fc4ed,
// printing 0095245c-a30e-4e2a-88c9-632c678e9f03): "Mutate {1}{G}{G} ... Reach,
// trample. Whenever this creature mutates, destroy target artifact or
// enchantment an opponent controls." Scryfall returned no rulings for either
// printing. Official CR effective 2026-09-25 (media.wizards.com
// MagicCompRules 20260925.txt) read: 113.7a (an ability on the stack exists
// independently of its source), 603.8 (a state trigger triggers again only
// after it has left the stack, if its object is still in the same zone and the
// condition still matches), 702.26b (a phased-out permanent is treated as
// though it does not exist and can't be affected by anything), 702.26c (a
// permanent that phases in is treated as existing again), 702.26d (phasing
// doesn't change zones; a phased-out permanent is treated as not on the
// battlefield), 702.140c and 730.2c (a mutated permanent is the same object,
// represented by more than one card). Sea Serpent's pending trigger therefore
// can't sacrifice it, as the top card or a lower component, while it is phased
// out, and acts on it normally once it has phased in.

func newPhasedOutSerpentFixture(t *testing.T) mutateStateTriggerFixture {
	t.Helper()
	gm := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	fx := mutateStateTriggerFixture{
		game:    gm,
		engine:  NewEngine(nil),
		serpent: addCombatPermanent(gm, game.Player1, registryCard(t, "Sea Serpent")),
	}
	fx.serpentCardID = fx.serpent.CardInstanceID
	gm.Turn.ActivePlayer = game.Player1
	gm.Turn.Phase = game.PhasePrecombatMain
	gm.Turn.Step = game.StepNone
	gm.Turn.PriorityPlayer = game.Player1
	fx.checkTriggers(t)
	if got := fx.serpentTriggers(); len(got) != 1 {
		t.Fatalf("Sea Serpent triggers with no Islands = %d, want 1", len(got))
	}
	return fx
}

// serpentFixtures yields Sea Serpent alone and mutated with Gemrazer over it
// (Sea Serpent a lower component) or under it (Sea Serpent the top card), each
// with Sea Serpent's original no-Island trigger pending.
func serpentFixtures(t *testing.T, run func(t *testing.T, fx mutateStateTriggerFixture)) {
	t.Run("Sea Serpent alone", func(t *testing.T) {
		run(t, newPhasedOutSerpentFixture(t))
	})
	for _, over := range []bool{true, false} {
		name := "Gemrazer mutated under"
		if over {
			name = "Gemrazer mutated over"
		}
		t.Run(name, func(t *testing.T) {
			fx := newMutateStateTriggerFixture(t)
			fx.mutate(t, over)
			fx.checkTriggers(t)
			if got := fx.serpentTriggers(); len(got) != 1 {
				t.Fatalf("Sea Serpent triggers after mutate = %d, want the original only", len(got))
			}
			run(t, fx)
		})
	}
}

func (fx mutateStateTriggerFixture) resolveSerpentTrigger(t *testing.T) {
	t.Helper()
	pending := fx.serpentTriggers()
	top, ok := fx.game.Stack.Peek()
	if len(pending) != 1 || !ok || top.ID != pending[0].ID || top.SourceCardID != fx.serpentCardID {
		t.Fatal("Sea Serpent's trigger is not alone on top of the stack")
	}
	fx.engine.resolveTopOfStack(fx.game, &TurnLog{})
	if got := fx.serpentTriggers(); len(got) != 0 {
		t.Fatalf("Sea Serpent triggers after resolution = %d, want 0", len(got))
	}
}

func (fx mutateStateTriggerFixture) requireSerpentSacrificed(t *testing.T) {
	t.Helper()
	if _, ok := permanentByObjectID(fx.game, fx.serpent.ObjectID); ok {
		t.Fatal("Sea Serpent is still on the battlefield")
	}
	graveyard := fx.game.Players[game.Player1].Graveyard
	if !graveyard.Contains(fx.serpentCardID) {
		t.Fatal("Sea Serpent was not put into the graveyard")
	}
	if fx.gemrazerID != 0 && !graveyard.Contains(fx.gemrazerID) {
		t.Fatal("Gemrazer was not put into the graveyard with Sea Serpent")
	}
}

func (fx mutateStateTriggerFixture) requireSerpentSurvived(t *testing.T) {
	t.Helper()
	serpent, ok := permanentByObjectID(fx.game, fx.serpent.ObjectID)
	if !ok || serpent != fx.serpent {
		t.Fatal("Sea Serpent left the battlefield")
	}
	graveyard := fx.game.Players[game.Player1].Graveyard
	if graveyard.Contains(fx.serpentCardID) || fx.gemrazerID != 0 && graveyard.Contains(fx.gemrazerID) {
		t.Fatal("a Sea Serpent component was put into the graveyard")
	}
}

// runBeginningPhase runs Player1's next beginning phase, where Sea Serpent
// phases in during the untap step (CR 702.26a, 702.26c). The no-Island state
// trigger then goes on the stack at the first priority, in the upkeep, and
// resolves when every player passes.
func (fx mutateStateTriggerFixture) runBeginningPhase(t *testing.T) {
	t.Helper()
	if !fx.game.Stack.IsEmpty() {
		t.Fatal("stack is not empty before the untap step")
	}
	addCardToLibrary(fx.game, game.Player1, registryCard(t, "Island"))
	fx.engine.runBeginningPhase(fx.game, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if fx.game.Players[game.Player1].Eliminated {
		t.Fatal("Player1 was eliminated during the beginning phase")
	}
	if fx.serpent.PhasedOut {
		t.Fatal("Sea Serpent did not phase in during its controller's untap step")
	}
}

func TestSourceCardReferenceSacrificesInPhaseSource(t *testing.T) {
	serpentFixtures(t, func(t *testing.T, fx mutateStateTriggerFixture) {
		fx.resolveSerpentTrigger(t)
		fx.requireSerpentSacrificed(t)
	})
}

func TestSourceCardReferenceIgnoresPhasedOutSource(t *testing.T) {
	serpentFixtures(t, func(t *testing.T, fx mutateStateTriggerFixture) {
		phaseOutForTest(t, fx.game, fx.serpent)
		fx.resolveSerpentTrigger(t)
		fx.requireSerpentSurvived(t)
		if !fx.serpent.PhasedOut {
			t.Fatal("Sea Serpent phased in when its trigger resolved")
		}

		// A phased-out Sea Serpent's ability doesn't exist, so it doesn't
		// trigger again until Sea Serpent phases in (CR 702.26b, 603.8).
		fx.checkTriggers(t)
		if got := fx.serpentTriggers(); len(got) != 0 {
			t.Fatalf("phased-out Sea Serpent triggered %d times, want 0", len(got))
		}

		fx.runBeginningPhase(t)
		fx.requireSerpentSacrificed(t)
	})
}
