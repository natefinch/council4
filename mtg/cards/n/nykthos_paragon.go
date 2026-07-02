package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NykthosParagon is the card definition for Nykthos Paragon.
//
// Type: Enchantment Creature — Human Soldier
// Cost: {4}{W}{W}
//
// Oracle text:
//
//	Whenever you gain life, you may put that many +1/+1 counters on each creature you control. Do this only once each turn.
var NykthosParagon = newNykthosParagon()

func newNykthosParagon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Nykthos Paragon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 6}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventLifeGained,
							Player: game.TriggerPlayerYou,
						},
					},
					Optional:           true,
					MaxTriggersPerTurn: 1,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountEventLifeChange,
										Multiplier: 1,
									}),
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever you gain life, you may put that many +1/+1 counters on each creature you control. Do this only once each turn.
		`,
		},
	}
}
