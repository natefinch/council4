package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func pushSpell(g *game.Game, controller game.PlayerID, def *game.CardDef, faceDown bool) {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: def, Owner: controller}
	g.Stack.Push(&game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   cardID,
		Controller: controller,
		FaceDown:   faceDown,
	})
}

func TestStackObjectViewExposesSpellCharacteristics(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := &game.CardDef{CardFace: game.CardFace{
		Name:     "Hill Giant",
		Types:    []types.Card{types.Creature},
		Colors:   []color.Color{color.Red},
		ManaCost: opt.Val(cost.Mana{cost.O(3)}),
	}}
	pushSpell(g, game.Player2, def, false)

	stack := observe(g, game.Player1).Stack()
	if len(stack) != 1 {
		t.Fatalf("Stack() = %d objects, want 1", len(stack))
	}
	view := stack[0]
	if view.ManaValue != 3 {
		t.Errorf("ManaValue = %d, want 3", view.ManaValue)
	}
	if !slices.Contains(view.Types, types.Creature) {
		t.Errorf("Types = %v, want to contain Creature", view.Types)
	}
	if !slices.Contains(view.Colors, color.Red) {
		t.Errorf("Colors = %v, want to contain Red", view.Colors)
	}
}

func TestStackObjectViewRedactsFaceDownSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := &game.CardDef{CardFace: game.CardFace{
		Name:     "Hidden Morph",
		Types:    []types.Card{types.Creature},
		Colors:   []color.Color{color.Green},
		ManaCost: opt.Val(cost.Mana{cost.O(5)}),
	}}
	pushSpell(g, game.Player2, def, true)

	view := observe(g, game.Player1).Stack()[0]
	if view.Name != "" {
		t.Errorf("Name = %q, want empty for a face-down spell", view.Name)
	}
	if view.ManaValue != 0 || view.Types != nil || view.Colors != nil {
		t.Errorf("face-down spell leaked identity: MV=%d Types=%v Colors=%v", view.ManaValue, view.Types, view.Colors)
	}
}
