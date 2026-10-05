package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestOptionalLibraryObservationDoesNotGateIndependentRider(t *testing.T) {
	for _, accepted := range []bool{false, true} {
		for _, empty := range []bool{false, true} {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Players[game.Player1].Life = 20
			if !empty {
				addCardToLibrary(g, game.Player3, vanillaCreature("Observed", 2, 3))
			}
			obj := &game.StackObject{
				ID: g.IDGen.Next(), Controller: game.Player1,
				Targets: []game.Target{game.PlayerTarget(game.Player3)},
			}
			var agents [game.NumPlayers]PlayerAgent
			if !accepted {
				agents[game.Player1] = &declineChoiceAgent{}
			}
			engine := NewEngine(nil)
			engine.resolveInstructionWithChoices(g, obj, &game.Instruction{
				Primitive: game.Reveal{
					Player: game.TargetPlayerReference(0), Amount: game.Fixed(1), PublishLinked: "observed",
				},
				Optional: true,
			}, agents, &TurnLog{})
			engine.resolveInstructionWithChoices(g, obj, &game.Instruction{
				Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(4)},
				Condition: opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
					Object:        opt.Val(game.LinkedObjectReference("observed")),
					ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
				})}),
			}, agents, &TurnLog{})
			engine.resolveInstructionWithChoices(g, obj, &game.Instruction{
				Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(2)},
			}, agents, &TurnLog{})
			want := 22
			if accepted && !empty {
				want = 26
			}
			if actual := g.Players[game.Player1].Life; actual != want {
				t.Fatalf("accepted=%v empty=%v: life=%d, want %d", accepted, empty, actual, want)
			}
		}
	}
}
