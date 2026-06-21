package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// castSpellTrigger builds a spell on the stack and a "copy that spell" triggered
// ability whose triggering event carries the spell when withEvent is set. It
// returns the cast spell's stack object.
func castSpellTrigger(g *game.Game, withEvent bool) *game.StackObject {
	source := addCreaturePermanent(g, game.Player1)

	spellSourceID := g.IDGen.Next()
	g.CardInstances[spellSourceID] = &game.CardInstance{
		ID: spellSourceID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:  "Cast Spell",
			Types: []types.Card{types.Instant},
		}},
		Owner: game.Player1,
	}
	castSpell := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   spellSourceID,
		Controller: game.Player1,
	}
	g.Stack.Push(castSpell)

	trigger := game.TriggeredAbility{
		Content: game.Mode{
			Sequence: []game.Instruction{{
				Primitive: game.CopyStackObject{Object: game.EventStackObjectReference()},
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
	if withEvent {
		original.HasTriggerEvent = true
		original.TriggerEvent = game.Event{
			Kind:          game.EventSpellCast,
			Controller:    game.Player1,
			StackObjectID: castSpell.ID,
		}
	}
	g.Stack.Push(original)
	return castSpell
}

// TestCopyStackObjectEffectCopiesTriggeringSpell verifies that a "copy that
// spell" trigger (Reflections of Littjara) copies the spell named by its
// triggering event onto the stack via the EventStackObject reference.
func TestCopyStackObjectEffectCopiesTriggeringSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	castSpell := castSpellTrigger(g, true)

	depthBefore := g.Stack.Size()
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Stack.Size(); got != depthBefore {
		t.Fatalf("stack size after copy = %d, want %d (copy replaces resolved trigger)", got, depthBefore)
	}
	top, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("stack empty after copy effect")
	}
	if !top.Copy {
		t.Fatal("top stack object is not marked as a copy")
	}
	if top.ID == castSpell.ID {
		t.Fatal("copy shares the cast spell's ID, want a distinct object")
	}
	if top.Kind != game.StackSpell || top.SourceID != castSpell.SourceID {
		t.Fatalf("copy = %+v, want a copy of the cast spell", top)
	}
}

// TestCopyStackObjectEffectWithoutTriggeringSpellDoesNothing verifies that a
// "copy that spell" trigger with no spell-cast event on hand pushes no copy, so
// the effect fails closed rather than copying an unrelated object.
func TestCopyStackObjectEffectWithoutTriggeringSpellDoesNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	castSpellTrigger(g, false)

	depthBefore := g.Stack.Size()
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Stack.Size(); got != depthBefore-1 {
		t.Fatalf("stack size = %d, want %d (trigger resolved, no copy pushed)", got, depthBefore-1)
	}
	top, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("stack empty, want the cast spell still present")
	}
	if top.Copy {
		t.Fatal("a copy was pushed without a triggering spell")
	}
}
