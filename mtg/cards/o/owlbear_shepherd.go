package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OwlbearShepherd is the card definition for Owlbear Shepherd.
//
// Type: Creature — Goblin Druid
// Cost: {2}{G}
//
// Oracle text:
//
//	At the beginning of your end step, if creatures you control have total power 8 or greater, draw a card.
var OwlbearShepherd = newOwlbearShepherd()

func newOwlbearShepherd() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Owlbear Shepherd",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Druid},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 4}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepEnd,
						},
						InterveningIf: "if creatures you control have total power 8 or greater",
						InterveningCondition: opt.Val(game.Condition{
							ControlsMatching: opt.Val(game.SelectionCount{
								Selection:  game.Selection{RequiredTypes: []types.Card{types.Creature}},
								TotalPower: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 8}),
							}),
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			At the beginning of your end step, if creatures you control have total power 8 or greater, draw a card.
		`,
		},
	}
}
