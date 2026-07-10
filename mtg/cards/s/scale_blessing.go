package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ScaleBlessing is the card definition for Scale Blessing.
//
// Type: Instant
// Cost: {3}{W}
//
// Oracle text:
//
//	Bolster 1, then put a +1/+1 counter on each creature you control with a +1/+1 counter on it. (To bolster 1, choose a creature with the least toughness among creatures you control and put a +1/+1 counter on it.)
var ScaleBlessing = newScaleBlessing

func newScaleBlessing() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Scale Blessing",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Bolster{
							Amount: game.Fixed(1),
						},
					},
					{
						Primitive: game.AddCounter{
							Amount:      game.Fixed(1),
							Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, MatchCounter: true, RequiredCounter: counter.PlusOnePlusOne}),
							CounterKind: counter.PlusOnePlusOne,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Bolster 1, then put a +1/+1 counter on each creature you control with a +1/+1 counter on it. (To bolster 1, choose a creature with the least toughness among creatures you control and put a +1/+1 counter on it.)
		`,
		},
	}
}
