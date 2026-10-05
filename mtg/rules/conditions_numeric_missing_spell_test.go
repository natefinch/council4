package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/opt"
)

func TestLiveSpellNumericInformationDoesNotRequireColors(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		setup func(*game.Game, *game.StackObject)
		known bool
		zero  bool
	}{
		{"ordinary known nonzero", nil, true, false},
		{"nil definition", func(g *game.Game, spell *game.StackObject) { g.CardInstances[spell.SourceID].Def = nil }, false, false},
		{"nil alternate definition", func(g *game.Game, spell *game.StackObject) {
			spell.Face = game.FaceAlternate
			g.CardInstances[spell.SourceID].Def = nil
		}, false, false},
		{"missing card", func(g *game.Game, spell *game.StackObject) { delete(g.CardInstances, spell.SourceID) }, false, false},
		{"face down without card", func(g *game.Game, spell *game.StackObject) {
			spell.FaceDown = true
			delete(g.CardInstances, spell.SourceID)
		}, true, true},
		{"face down nil definition", func(g *game.Game, spell *game.StackObject) {
			spell.FaceDown = true
			g.CardInstances[spell.SourceID].Def = nil
		}, true, true},
		{"nonspell", func(_ *game.Game, spell *game.StackObject) { spell.Kind = game.StackActivatedAbility }, false, false},
	} {
		for _, event := range []bool{false, true} {
			for _, negate := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/event=%t/negate=%t", tt.name, event, negate), func(t *testing.T) {
					t.Parallel()
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					spell := spellWithManaValue(g, game.Player2, cost.Mana{cost.O(4)})
					if tt.setup != nil {
						tt.setup(g, spell)
					}
					obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{game.StackObjectTarget(spell.ID)}}
					object := game.TargetStackObjectReference(0)
					if event {
						obj.HasTriggerEvent = true
						obj.TriggerEvent = game.Event{Kind: game.EventSpellCast, StackObjectID: spell.ID, Controller: game.Player2}
						object = game.EventStackObjectReference()
					}
					condition := game.Condition{Object: opt.Val(object), Negate: negate,
						ObjectMatches: opt.Val(game.Selection{ManaValue: opt.Val(compare.Int{Op: compare.Equal, Value: 0})})}
					ctx := conditionContext{controller: game.Player1, obj: obj}
					want := tt.known && (tt.zero != negate)
					if got := conditionSatisfied(g, ctx, opt.Val(condition)); got != want {
						t.Fatalf("numeric condition=%t, want %t", got, want)
					}
					if !tt.known {
						condition.ObjectMatches.Val.Colorless = true
						if conditionSatisfied(g, ctx, opt.Val(condition)) {
							t.Fatal("unavailable numeric/color information succeeded")
						}
					}
				})
			}
		}
	}
}
