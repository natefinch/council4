package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ResplendentGriffin is the card definition for Resplendent Griffin.
//
// Type: Creature — Griffin
// Cost: {1}{W}{U}
//
// Oracle text:
//
//	Flying
//	Ascend (If you control ten or more permanents, you get the city's blessing for the rest of the game.)
//	Whenever this creature attacks, if you have the city's blessing, put a +1/+1 counter on it.
var ResplendentGriffin = newResplendentGriffin

func newResplendentGriffin() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Resplendent Griffin",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Griffin},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.AscendStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
						InterveningIf: "if you have the city's blessing",
						InterveningCondition: opt.Val(game.Condition{
							ControllerHasCityBlessing: true,
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.EventPermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			Ascend (If you control ten or more permanents, you get the city's blessing for the rest of the game.)
			Whenever this creature attacks, if you have the city's blessing, put a +1/+1 counter on it.
		`,
		},
	}
}
