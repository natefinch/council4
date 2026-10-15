package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VanquishTheFoul is the card definition for Vanquish the Foul.
//
// Type: Sorcery
// Cost: {5}{W}
//
// Oracle text:
//
//	Destroy target creature with power 4 or greater. Scry 1. (Look at the top card of your library. You may put that card on the bottom.)
var VanquishTheFoul = newVanquishTheFoul

func newVanquishTheFoul() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Vanquish the Foul",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature with power 4 or greater",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4})}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.Scry{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy target creature with power 4 or greater. Scry 1. (Look at the top card of your library. You may put that card on the bottom.)
		`,
		},
	}
}
