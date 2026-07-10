package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AbzanSkycaptain is the card definition for Abzan Skycaptain.
//
// Type: Creature — Bird Soldier
// Cost: {3}{W}
//
// Oracle text:
//
//	Flying
//	When this creature dies, bolster 2. (Choose a creature with the least toughness among creatures you control and put two +1/+1 counters on it.)
var AbzanSkycaptain = newAbzanSkycaptain

func newAbzanSkycaptain() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Abzan Skycaptain",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Bolster{
									Amount: game.Fixed(2),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			When this creature dies, bolster 2. (Choose a creature with the least toughness among creatures you control and put two +1/+1 counters on it.)
		`,
		},
	}
}
