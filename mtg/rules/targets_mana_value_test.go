package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func spellWithManaValue(g *game.Game, controller game.PlayerID, manaCost cost.Mana) *game.StackObject {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Def:   &game.CardDef{CardFace: game.CardFace{Name: "Test Spell", ManaCost: opt.Val(manaCost), Types: []types.Card{types.Instant}}},
		Owner: controller,
	}
	obj := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: cardID, Controller: controller}
	g.Stack.Push(obj)
	return obj
}

func TestStackObjectTargetMatchesManaValuePredicate(t *testing.T) {
	manaValueSpec := func(c compare.Int) *game.TargetSpec {
		return &game.TargetSpec{
			MinTargets: 1,
			MaxTargets: 1,
			Allow:      game.TargetAllowStackObject,
			Predicate: game.TargetPredicate{
				StackObjectKinds: []game.StackObjectKind{game.StackSpell},
				ManaValue:        opt.Val(c),
			},
		}
	}

	tests := []struct {
		name    string
		compare compare.Int
		cost    cost.Mana
		want    bool
	}{
		{name: "equal matches", compare: compare.Int{Op: compare.Equal, Value: 1}, cost: cost.Mana{cost.U}, want: true},
		{name: "equal rejects higher", compare: compare.Int{Op: compare.Equal, Value: 1}, cost: cost.Mana{cost.O(3)}, want: false},
		{name: "or less matches lower", compare: compare.Int{Op: compare.LessOrEqual, Value: 2}, cost: cost.Mana{cost.U}, want: true},
		{name: "or less rejects higher", compare: compare.Int{Op: compare.LessOrEqual, Value: 2}, cost: cost.Mana{cost.O(3)}, want: false},
		{name: "or greater matches higher", compare: compare.Int{Op: compare.GreaterOrEqual, Value: 3}, cost: cost.Mana{cost.O(4)}, want: true},
		{name: "or greater rejects lower", compare: compare.Int{Op: compare.GreaterOrEqual, Value: 3}, cost: cost.Mana{cost.U}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			obj := spellWithManaValue(g, game.Player2, tt.cost)
			spec := manaValueSpec(tt.compare)

			got := stackObjectTargetMatchesSpec(g, game.Player1, 0, spec, obj.ID)
			if got != tt.want {
				t.Fatalf("mana value match = %v, want %v", got, tt.want)
			}
		})
	}
}
