package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PlagueWind is the card definition for Plague Wind.
//
// Type: Sorcery
// Cost: {7}{B}{B}
//
// Oracle text:
//
//	Destroy all creatures you don't control. They can't be regenerated.
var PlagueWind = newPlagueWind

func newPlagueWind() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Plague Wind",
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
							Group:               game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerNotYou}),
							PreventRegeneration: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all creatures you don't control. They can't be regenerated.
		`,
		},
	}
}
