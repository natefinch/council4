package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestCopyStackObjectEffectCopiesResolvingSpell verifies that a spell whose own
// resolution copies itself ("Copy this spell.", Sevinne's Reclamation) puts an
// independent copy on the stack. The resolving spell is popped off the stack
// before its effects run, so the copy source is the resolving object itself
// (game.ResolvingStackObjectReference) rather than a by-ID stack lookup, which
// would fail once the original is gone.
func TestCopyStackObjectEffectCopiesResolvingSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	addEffectSpellToStack(g, game.Player1,
		game.CopyStackObject{Object: game.ResolvingStackObjectReference()}, nil)
	original, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("stack empty after pushing self-copy spell")
	}
	originalID := original.ID

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size after resolving self-copy spell = %d, want 1 (original popped, copy pushed)", got)
	}
	top, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("stack empty after copy effect")
	}
	if !top.Copy {
		t.Fatal("top stack object is not marked as a copy")
	}
	if top.ID == originalID {
		t.Fatal("copy shares the original's ID, want a distinct object")
	}
	if top.Kind != game.StackSpell || top.SourceID != original.SourceID {
		t.Fatalf("copy = %+v, want a spell from the same source", top)
	}
}
