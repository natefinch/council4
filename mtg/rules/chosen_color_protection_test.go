package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
)

// TestResolveChosenColorProtection verifies that a granted "protection from the
// color of your choice" ability is rewritten into protection from the concrete
// color the controller chooses as the granting ability resolves, and that the
// shared template the rewrite reads from is left untouched.
func TestResolveChosenColorProtection(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// The color prompt lists W,U,B,R,G; index 3 selects red.
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{3}}}}
	obj := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackActivatedAbility, Controller: game.Player1}
	r := &effectResolver{engine: engine, game: g, obj: obj, agents: agents, log: &TurnLog{}}

	chosen := game.ProtectionFromChosenColorStaticAbility()
	templates := []game.ContinuousEffect{{
		Layer:        game.LayerAbility,
		AddAbilities: []game.Ability{&chosen},
	}}

	resolved := r.resolveChosenColorProtection(templates)
	static, ok := resolved[0].AddAbilities[0].(*game.StaticAbility)
	if !ok {
		t.Fatalf("ability = %T, want *game.StaticAbility", resolved[0].AddAbilities[0])
	}
	prot, ok := game.StaticBodyProtectionKeyword(static)
	if !ok || prot.ChosenColor {
		t.Fatalf("resolved protection = %+v ok=%v, want concrete color", prot, ok)
	}
	if len(prot.FromColors) != 1 || prot.FromColors[0] != color.Red {
		t.Fatalf("resolved protection colors = %v, want [red]", prot.FromColors)
	}

	// The shared template must keep its ChosenColor marker so the next
	// resolution of the same ability prompts again rather than reusing red.
	original, ok := game.StaticBodyProtectionKeyword(&chosen)
	if !ok || !original.ChosenColor || len(original.FromColors) != 0 {
		t.Fatalf("template protection mutated: %+v", original)
	}
}
