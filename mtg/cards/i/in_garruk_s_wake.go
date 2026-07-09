package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// InGarrukSWake is the card definition for In Garruk's Wake.
//
// Type: Sorcery
// Cost: {7}{B}{B}
//
// Oracle text:
//
//	Destroy all creatures you don't control and all planeswalkers you don't control.
var InGarrukSWake = newInGarrukSWake

func newInGarrukSWake() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "In Garruk's Wake",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}, Controller: game.ControllerNotYou}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all creatures you don't control and all planeswalkers you don't control.
		`,
		},
	}
}
