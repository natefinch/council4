package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AnafenzaKinTreeSpirit is the card definition for Anafenza, Kin-Tree Spirit.
//
// Type: Legendary Creature — Spirit Soldier
// Cost: {W}{W}
//
// Oracle text:
//
//	Whenever another nontoken creature you control enters, bolster 1. (Choose a creature with the least toughness among creatures you control and put a +1/+1 counter on it.)
var AnafenzaKinTreeSpirit = newAnafenzaKinTreeSpirit

func newAnafenzaKinTreeSpirit() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Anafenza, Kin-Tree Spirit",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Spirit, types.Soldier},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							ExcludeSelf:      true,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, NonToken: true},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Bolster{
									Amount: game.Fixed(1),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever another nontoken creature you control enters, bolster 1. (Choose a creature with the least toughness among creatures you control and put a +1/+1 counter on it.)
		`,
		},
	}
}
