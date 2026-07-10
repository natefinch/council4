package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EchoesOfTheKinTree is the card definition for Echoes of the Kin Tree.
//
// Type: Enchantment
// Cost: {1}{W}
//
// Oracle text:
//
//	{2}{W}: Bolster 1. (Choose a creature with the least toughness among creatures you control and put a +1/+1 counter on it.)
var EchoesOfTheKinTree = newEchoesOfTheKinTree

func newEchoesOfTheKinTree() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Echoes of the Kin Tree",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{2}{W}: Bolster 1. (Choose a creature with the least toughness among creatures you control and put a +1/+1 counter on it.)",
					ManaCost:       opt.Val(cost.Mana{cost.O(2), cost.W}),
					ZoneOfFunction: zone.Battlefield,
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
			{2}{W}: Bolster 1. (Choose a creature with the least toughness among creatures you control and put a +1/+1 counter on it.)
		`,
		},
	}
}
