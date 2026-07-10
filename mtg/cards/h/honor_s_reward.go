package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HonorSReward is the card definition for Honor's Reward.
//
// Type: Instant
// Cost: {2}{W}
//
// Oracle text:
//
//	You gain 4 life. Bolster 2. (Choose a creature with the least toughness among creatures you control and put two +1/+1 counters on it.)
var HonorSReward = newHonorSReward

func newHonorSReward() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Honor's Reward",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(4),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.Bolster{
							Amount: game.Fixed(2),
						},
					},
				},
			}.Ability()),
			OracleText: `
			You gain 4 life. Bolster 2. (Choose a creature with the least toughness among creatures you control and put two +1/+1 counters on it.)
		`,
		},
	}
}
