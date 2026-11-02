package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review (#3538):
//   - Sea Serpent (1621438d-717a-477c-93ff-ff7415da70c5): "This creature can't
//     attack unless defending player controls an Island. When you control no
//     Islands, sacrifice this creature." No Scryfall rulings.
//   - Gemrazer (0095245c-a30e-4e2a-88c9-632c678e9f03): "Mutate {1}{G}{G} ...
//     Reach, trample. Whenever this creature mutates, destroy target artifact
//     or enchantment an opponent controls." No Scryfall rulings.
//   - Vedalken Orrery (44067327-6fe3-4222-b591-6d14d0e360d7): "You may cast
//     spells as though they had flash." Rulings: applies only to casting
//     spells; its controller may cast creature cards any time they could cast
//     an instant.
//   - CR 603.8: a state-triggered ability doesn't trigger again until the
//     ability has resolved, has been countered, or has otherwise left the
//     stack. CR 702.140c/e: a mutating spell merges with its target, and the
//     mutated permanent has all abilities of each component. CR 730.2c: a
//     merged permanent is the same object it was before. CR 113.7a: a
//     triggered ability on the stack exists independently of its source.
//
// Gemrazer cast with flash over Sea Serpent while Sea Serpent's state trigger
// is pending moves Serpent's trigger to a later ability index on the same
// object. The trigger must stay latched until that instance leaves the stack.

func TestMutateOverDoesNotRetriggerPendingStateTrigger(t *testing.T) {
	fx := newMutateStateTriggerFixture(t)
	fx.mutate(t, true)

	fx.checkTriggers(t)
	if got := fx.serpentTriggers(); len(got) != 1 {
		t.Fatalf("Sea Serpent triggers on stack after mutate-over = %d, want the original only", len(got))
	}

	// The condition stops and starts matching again while the trigger is still
	// pending; it still doesn't trigger again (CR 603.8).
	island := addCombatPermanent(fx.game, game.Player1, registryCard(t, "Island"))
	fx.checkTriggers(t)
	if _, ok := destroyPermanent(fx.game, island.ObjectID); !ok {
		t.Fatal("Island was not destroyed")
	}
	fx.checkTriggers(t)
	if got := fx.serpentTriggers(); len(got) != 1 {
		t.Fatalf("Sea Serpent triggers on stack after Island came and went = %d, want 1", len(got))
	}

	fx.counterAndExpectOneRetrigger(t)

	// Resolving the retriggered ability sacrifices the whole mutated permanent.
	fx.engine.resolveTopOfStack(fx.game, &TurnLog{})
	if _, ok := permanentByObjectID(fx.game, fx.serpent.ObjectID); ok {
		t.Fatal("mutated Sea Serpent is still on the battlefield after its trigger resolved")
	}
	graveyard := fx.game.Players[game.Player1].Graveyard
	if !graveyard.Contains(fx.serpentCardID) || !graveyard.Contains(fx.gemrazerID) {
		t.Fatal("Sea Serpent and Gemrazer were not both put into the graveyard")
	}
	if len(fx.game.StateTriggerLatches) != 0 {
		t.Fatalf("state trigger latches after resolution = %+v, want released", fx.game.StateTriggerLatches)
	}
}

// The original trigger still refers to the same object after Gemrazer is put
// over Sea Serpent (CR 730.2c), so resolving it sacrifices the whole mutated
// permanent even though Sea Serpent is no longer its top card.
func TestMutateOverOriginalStateTriggerSacrificesMergedPermanent(t *testing.T) {
	for _, islandArrives := range []bool{false, true} {
		name := "no Island"
		if islandArrives {
			name = "Island arrives before resolution"
		}
		t.Run(name, func(t *testing.T) {
			fx := newMutateStateTriggerFixture(t)
			fx.mutate(t, true)
			if islandArrives {
				// The condition isn't rechecked on resolution (CR 603.8).
				addCombatPermanent(fx.game, game.Player1, registryCard(t, "Island"))
			}
			fx.checkTriggers(t)
			pending := fx.serpentTriggers()
			if len(pending) != 1 {
				t.Fatalf("Sea Serpent triggers on stack = %d, want the original only", len(pending))
			}
			if top, ok := fx.game.Stack.Peek(); !ok || top.ID != pending[0].ID || top.SourceCardID != fx.serpentCardID {
				t.Fatal("Sea Serpent's original trigger is not on top of the stack")
			}
			fx.engine.resolveTopOfStack(fx.game, &TurnLog{})
			if _, ok := permanentByObjectID(fx.game, fx.serpent.ObjectID); ok {
				t.Fatal("mutated Sea Serpent survived its original trigger")
			}
			graveyard := fx.game.Players[game.Player1].Graveyard
			if !graveyard.Contains(fx.serpentCardID) || !graveyard.Contains(fx.gemrazerID) {
				t.Fatal("Sea Serpent and Gemrazer were not both put into the graveyard")
			}
			fx.checkTriggers(t)
			if len(fx.game.StateTriggerLatches) != 0 {
				t.Fatalf("state trigger latches after resolution = %+v, want released", fx.game.StateTriggerLatches)
			}
		})
	}
}

func TestMutateUnderDoesNotRetriggerPendingStateTrigger(t *testing.T) {
	fx := newMutateStateTriggerFixture(t)
	fx.mutate(t, false)

	fx.checkTriggers(t)
	if got := fx.serpentTriggers(); len(got) != 1 {
		t.Fatalf("Sea Serpent triggers on stack after mutate-under = %d, want the original only", len(got))
	}
	fx.counterAndExpectOneRetrigger(t)
}

func TestMutateOverStateTriggerLatchSurvivesGameClone(t *testing.T) {
	fx := newMutateStateTriggerFixture(t)
	fx.mutate(t, true)

	clone := fx
	clone.game = fx.game.Clone()
	clone.checkTriggers(t)
	if got := clone.serpentTriggers(); len(got) != 1 {
		t.Fatalf("Sea Serpent triggers on cloned stack = %d, want the original only", len(got))
	}
	clone.counterAndExpectOneRetrigger(t)

	fx.checkTriggers(t)
	if got := fx.serpentTriggers(); len(got) != 1 {
		t.Fatalf("original game Sea Serpent triggers = %d after clone counter, want 1", len(got))
	}
}

type mutateStateTriggerFixture struct {
	game          *game.Game
	engine        *Engine
	serpent       *game.Permanent
	serpentCardID id.ID
	gemrazerID    id.ID
}

// newMutateStateTriggerFixture puts Sea Serpent's state trigger on the stack
// and casts Gemrazer with flash from Vedalken Orrery targeting Sea Serpent.
func newMutateStateTriggerFixture(t *testing.T) mutateStateTriggerFixture {
	t.Helper()
	gm := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	fx := mutateStateTriggerFixture{
		game:       gm,
		engine:     NewEngine(nil),
		serpent:    addCombatPermanent(gm, game.Player1, registryCard(t, "Sea Serpent")),
		gemrazerID: addCardToHand(gm, game.Player1, registryCard(t, "Gemrazer")),
	}
	fx.serpentCardID = fx.serpent.CardInstanceID
	addCombatPermanent(gm, game.Player1, registryCard(t, "Vedalken Orrery"))
	for range 3 {
		addBasicLandPermanent(gm, game.Player1, types.Forest)
	}
	gm.Turn.ActivePlayer = game.Player1
	gm.Turn.Phase = game.PhasePrecombatMain
	gm.Turn.Step = game.StepNone
	gm.Turn.PriorityPlayer = game.Player1

	fx.checkTriggers(t)
	if got := fx.serpentTriggers(); len(got) != 1 {
		t.Fatalf("Sea Serpent triggers with no Islands = %d, want 1", len(got))
	}
	if !fx.engine.applyAction(gm, game.Player1, action.CastMutateSpell(fx.gemrazerID, fx.serpent.ObjectID)) {
		t.Fatal("Gemrazer mutate cast with flash failed")
	}
	if top, ok := gm.Stack.Peek(); !ok || top.Kind != game.StackSpell || top.SourceID != fx.gemrazerID {
		t.Fatal("Gemrazer is not on top of Sea Serpent's trigger")
	}
	return fx
}

func (fx mutateStateTriggerFixture) mutate(t *testing.T, over bool) {
	t.Helper()
	choice := 1
	if over {
		choice = 0
	}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{choice}}}}
	fx.engine.resolveTopOfStackWithChoices(fx.game, agents, &TurnLog{})
	serpent, ok := permanentByObjectID(fx.game, fx.serpent.ObjectID)
	if !ok {
		t.Fatal("mutated Sea Serpent left the battlefield")
	}
	wantTop := fx.serpentCardID
	if over {
		wantTop = fx.gemrazerID
	}
	if serpent.CardInstanceID != wantTop || len(serpent.MergedCards) != 1 {
		t.Fatalf("mutated permanent top = %v with %d merged cards, want top %v", serpent.CardInstanceID, len(serpent.MergedCards), wantTop)
	}
	if fx.game.Players[game.Player1].Graveyard.Contains(fx.gemrazerID) {
		t.Fatal("Gemrazer was put into the graveyard instead of merging")
	}
}

// checkTriggers performs the priority check that puts pending triggers on the
// stack (CR 117.5).
func (fx mutateStateTriggerFixture) checkTriggers(t *testing.T) {
	t.Helper()
	fx.engine.putTriggeredAbilitiesOnStack(fx.game)
}

func (fx mutateStateTriggerFixture) serpentTriggers() []*game.StackObject {
	var triggers []*game.StackObject
	for _, obj := range fx.game.Stack.Objects() {
		if obj.Kind == game.StackTriggeredAbility && obj.SourceID == fx.serpent.ObjectID &&
			obj.InlineTrigger != nil && obj.InlineTrigger.Trigger.State.Exists {
			triggers = append(triggers, obj)
		}
	}
	return triggers
}

// counterAndExpectOneRetrigger counters the pending trigger and checks that
// the still-matching condition triggers it exactly once more.
func (fx mutateStateTriggerFixture) counterAndExpectOneRetrigger(t *testing.T) {
	t.Helper()
	pending := fx.serpentTriggers()
	if len(pending) != 1 {
		t.Fatalf("pending Sea Serpent triggers = %d, want 1", len(pending))
	}
	if !counterStackObject(fx.game, pending[0].ID) {
		t.Fatal("Sea Serpent trigger was not countered")
	}
	if got := fx.serpentTriggers(); len(got) != 0 {
		t.Fatalf("Sea Serpent triggers after counter = %d, want 0", len(got))
	}
	fx.checkTriggers(t)
	fx.checkTriggers(t)
	if got := fx.serpentTriggers(); len(got) != 1 {
		t.Fatalf("Sea Serpent triggers after its countered instance left the stack = %d, want exactly 1", len(got))
	}
}
