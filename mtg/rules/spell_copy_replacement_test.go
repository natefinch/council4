package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestAdditionalSpellCopyReplacementAddsDirectAndStormCopies(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	staff := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Twinning Staff",
		Types: nil,
		ReplacementAbilities: []game.ReplacementAbility{
			game.AdditionalSpellCopyReplacement("copy one additional time", 1, true),
		},
	}})
	registerPermanentReplacementEffects(g, staff)

	original := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   g.IDGen.Next(),
		Controller: game.Player1,
	}
	resolveInstruction(engine, g, original, game.CopyStackObject{
		Object: game.ResolvingStackObjectReference(),
	}, &TurnLog{})
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("direct copy stack size = %d, want 2", got)
	}

	g.Stack = game.Stack{}
	resolveInstruction(engine, g, original, game.CopyStackObject{
		Object: game.ResolvingStackObjectReference(),
		Count:  2,
	}, &TurnLog{})
	if got := g.Stack.Size(); got != 3 {
		t.Fatalf("two-copy batch stack size = %d, want 3", got)
	}

	g.Stack = game.Stack{}
	createStormCopies(g, original, &game.CardDef{}, 2)
	if got := g.Stack.Size(); got != 3 {
		t.Fatalf("storm copy stack size = %d, want 3", got)
	}
}

func TestAdditionalSpellCopyReplacementDoesNotCopyAbilities(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	staff := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Twinning Staff",
		ReplacementAbilities: []game.ReplacementAbility{
			game.AdditionalSpellCopyReplacement("copy one additional time", 1, true),
		},
	}})
	registerPermanentReplacementEffects(g, staff)
	original := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackTriggeredAbility,
		SourceID:   g.IDGen.Next(),
		Controller: game.Player1,
	}

	resolveInstruction(engine, g, original, game.CopyStackObject{
		Object: game.ResolvingStackObjectReference(),
	}, &TurnLog{})

	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("ability copy stack size = %d, want 1", got)
	}
}

func TestAdditionalSpellCopyReplacementFollowsCurrentController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	staff := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Twinning Staff",
		ReplacementAbilities: []game.ReplacementAbility{
			game.AdditionalSpellCopyReplacement("copy one additional time", 1, true),
		},
	}})
	registerPermanentReplacementEffects(g, staff)
	staff.Controller = game.Player2

	for _, test := range []struct {
		player game.PlayerID
		want   int
	}{
		{player: game.Player1, want: 1},
		{player: game.Player2, want: 2},
	} {
		g.Stack = game.Stack{}
		original := &game.StackObject{
			ID:         g.IDGen.Next(),
			Kind:       game.StackSpell,
			SourceID:   g.IDGen.Next(),
			Controller: test.player,
		}
		resolveInstruction(engine, g, original, game.CopyStackObject{
			Object: game.ResolvingStackObjectReference(),
		}, &TurnLog{})
		if got := g.Stack.Size(); got != test.want {
			t.Fatalf("%v copy stack size = %d, want %d", test.player, got, test.want)
		}
	}
}
