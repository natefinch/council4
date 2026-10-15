package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DisruptDecorum is the card definition for Disrupt Decorum.
//
// Type: Sorcery
// Cost: {2}{R}{R}
//
// Oracle text:
//
//	Goad all creatures you don't control. (Until your next turn, those creatures attack each combat if able and attack a player other than you if able.)
var DisruptDecorum = newDisruptDecorum

func newDisruptDecorum() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Disrupt Decorum",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Goad{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerNotYou}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Goad all creatures you don't control. (Until your next turn, those creatures attack each combat if able and attack a player other than you if able.)
		`,
		},
	}
}
