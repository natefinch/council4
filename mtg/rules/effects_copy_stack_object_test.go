package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestCopyStackObjectEffectCopiesTriggeredAbility verifies that resolving a
// CopyStackObject effect targeting a triggered ability puts an independent copy
// on the stack (CR 707) and that the copy resolves on its own, executing the
// ability body a second time.
func TestCopyStackObjectEffectCopiesTriggeredAbility(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	source := addCreaturePermanent(g, game.Player1)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Library Card"}})
	trigger := game.TriggeredAbility{
		Content: game.Mode{
			Sequence: []game.Instruction{{
				Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
			}},
		}.Ability(),
	}
	original := &game.StackObject{
		ID:            g.IDGen.Next(),
		Kind:          game.StackTriggeredAbility,
		SourceID:      source.ObjectID,
		SourceCardID:  source.CardInstanceID,
		Controller:    game.Player1,
		InlineTrigger: &trigger,
	}
	g.Stack.Push(original)

	depthBefore := g.Stack.Size()
	addEffectSpellToStack(g, game.Player1,
		game.CopyStackObject{Object: game.TargetStackObjectReference(0)},
		[]game.Target{game.StackObjectTarget(original.ID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Stack.Size(); got != depthBefore+1 {
		t.Fatalf("stack size after copy = %d, want %d (copy pushed)", got, depthBefore+1)
	}
	top, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("stack empty after copy effect")
	}
	if !top.Copy {
		t.Fatal("top stack object is not marked as a copy")
	}
	if top.ID == original.ID {
		t.Fatal("copy shares the original's ID, want a distinct object")
	}
	if top.Kind != game.StackTriggeredAbility || top.SourceID != original.SourceID {
		t.Fatalf("copy = %+v, want a triggered ability from the same source", top)
	}

	handBefore := g.Players[game.Player1].Hand.Size()
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != handBefore+1 {
		t.Fatalf("hand size after copy resolves = %d, want %d (copy drew a card)", got, handBefore+1)
	}
}

// TestCopyStackObjectEffectChoosesNewTargets verifies the "you may choose new
// targets for the copy" rider lets the resolving controller retarget the copy
// without disturbing the original ability's targets.
func TestCopyStackObjectEffectChoosesNewTargets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Trigger Source",
		Types: []types.Card{types.Artifact},
	}})
	victimA := addCreaturePermanent(g, game.Player2)
	victimB := addCreaturePermanent(g, game.Player2)

	trigger := game.TriggeredAbility{
		Content: game.Mode{
			Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature"}},
			Sequence: []game.Instruction{{
				Primitive: game.Damage{Amount: game.Fixed(1), Recipient: game.AnyTargetDamageRecipient(0)},
			}},
		}.Ability(),
	}
	original := &game.StackObject{
		ID:            g.IDGen.Next(),
		Kind:          game.StackTriggeredAbility,
		SourceID:      source.ObjectID,
		SourceCardID:  source.CardInstanceID,
		Controller:    game.Player1,
		InlineTrigger: &trigger,
		Targets:       []game.Target{game.PermanentTarget(victimA.ObjectID)},
		TargetCounts:  []int{1},
	}
	g.Stack.Push(original)

	addEffectSpellToStack(g, game.Player1,
		game.CopyStackObject{Object: game.TargetStackObjectReference(0), MayChooseNewTargets: true},
		[]game.Target{game.StackObjectTarget(original.ID)})

	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	top, ok := g.Stack.Peek()
	if !ok || !top.Copy {
		t.Fatal("copy not on top of stack")
	}
	if len(top.Targets) != 1 || top.Targets[0].PermanentID != victimB.ObjectID {
		t.Fatalf("copy targets = %+v, want victim B %v", top.Targets, victimB.ObjectID)
	}
	if len(original.Targets) != 1 || original.Targets[0].PermanentID != victimA.ObjectID {
		t.Fatalf("original retargeted to %+v, want unchanged victim A %v", original.Targets, victimA.ObjectID)
	}
}
