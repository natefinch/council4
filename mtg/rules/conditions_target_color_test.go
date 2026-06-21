package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestCounterTargetSpellColorGate exercises the "Counter target spell if it's
// blue." resolving target-color gate (Pyroblast / Red Elemental Blast): the
// counter only resolves when the targeted spell currently has the gated color.
func TestCounterTargetSpellColorGate(t *testing.T) {
	cases := []struct {
		name          string
		spellColors   []color.Color
		wantCountered bool
	}{
		{name: "blue spell countered", spellColors: []color.Color{color.Blue}, wantCountered: true},
		{name: "red spell not countered", spellColors: []color.Color{color.Red}, wantCountered: false},
		{name: "colorless spell not countered", spellColors: nil, wantCountered: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			targetID := g.IDGen.Next()
			g.CardInstances[targetID] = &game.CardInstance{
				ID: targetID,
				Def: &game.CardDef{CardFace: game.CardFace{
					Name:   "Target Spell",
					Types:  []types.Card{types.Sorcery},
					Colors: tc.spellColors,
				}},
				Owner: game.Player2,
			}
			targetObj := &game.StackObject{
				ID:         g.IDGen.Next(),
				Kind:       game.StackSpell,
				SourceID:   targetID,
				Controller: game.Player2,
			}
			g.Stack.Push(targetObj)

			gate := game.Instruction{
				Primitive: game.CounterObject{Object: game.TargetStackObjectReference(0)},
				Condition: opt.Val(game.EffectCondition{
					Object: game.TargetStackObjectReference(0),
					Condition: opt.Val(game.Condition{
						Object:        opt.Val(game.TargetStackObjectReference(0)),
						ObjectMatches: opt.Val(game.Selection{ColorsAny: []color.Color{color.Blue}}),
					}),
				}),
			}
			addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{gate},
				[]game.Target{game.StackObjectTarget(targetObj.ID)})

			engine.resolveTopOfStack(g, &TurnLog{})

			_, stillOnStack := stackObjectByID(g, targetObj.ID)
			countered := !stillOnStack
			if countered != tc.wantCountered {
				t.Fatalf("countered = %v, want %v", countered, tc.wantCountered)
			}
			if tc.wantCountered && !g.Players[game.Player2].Graveyard.Contains(targetID) {
				t.Fatal("countered spell did not move to graveyard")
			}
		})
	}
}
