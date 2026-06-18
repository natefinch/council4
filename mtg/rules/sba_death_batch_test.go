package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestStateBasedDeathsShareSimultaneousBatch verifies that every creature
// dying in one state-based-action pass is tagged with one shared simultaneous
// event ID, so a "whenever one or more creatures die" trigger coalesces the
// whole batch into a single stack object.
func TestStateBasedDeathsShareSimultaneousBatch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                 game.EventPermanentDied,
		Controller:            game.TriggerControllerYou,
		RequirePermanentTypes: []types.Card{types.Creature},
		OneOrMore:             true,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	first.MarkedDamage = 2
	second := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	second.MarkedDamage = 2

	engine.applyStateBasedActions(g)

	var batchIDs []id.ID
	for _, event := range g.Events {
		if event.Kind != game.EventPermanentDied {
			continue
		}
		if event.PermanentID == first.ObjectID || event.PermanentID == second.ObjectID {
			batchIDs = append(batchIDs, event.SimultaneousID)
		}
	}
	if len(batchIDs) != 2 {
		t.Fatalf("died events = %d, want 2", len(batchIDs))
	}
	if batchIDs[0] == 0 || batchIDs[0] != batchIDs[1] {
		t.Fatalf("died batch IDs = %v, want one shared nonzero ID", batchIDs)
	}

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("one-or-more dies trigger was not put on stack")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want one coalesced trigger for the simultaneous death batch", got)
	}
}

// TestStateBasedDeathFiresNonSelfDiesTriggerForSimultaneousDeath verifies that
// a permanent dying to a state-based action still sees another permanent that
// dies in the same pass, so "whenever another creature dies" triggers on a
// departed source fire for its simultaneous companions.
func TestStateBasedDeathFiresNonSelfDiesTriggerForSimultaneousDeath(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:       game.EventPermanentDied,
		ExcludeSelf: true,
		SubjectSelection: game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		},
	}, []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	source.MarkedDamage = 1
	other := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	other.MarkedDamage = 2

	engine.applyStateBasedActions(g)

	if _, ok := permanentByObjectID(g, source.ObjectID); ok {
		t.Fatal("source creature with lethal damage remained on battlefield")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("departed source did not trigger for another simultaneous SBA death")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != source.ObjectID || obj.TriggerEvent.PermanentID != other.ObjectID {
		t.Fatalf("top of stack = %+v, want source %v triggered by %v", obj, source.ObjectID, other.ObjectID)
	}
}
