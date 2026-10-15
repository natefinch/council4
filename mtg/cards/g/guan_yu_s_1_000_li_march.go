package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GuanYuS1000LiMarch is the card definition for Guan Yu's 1,000-Li March.
//
// Type: Sorcery
// Cost: {4}{W}{W}
//
// Oracle text:
//
//	Destroy all tapped creatures.
var GuanYuS1000LiMarch = newGuanYuS1000LiMarch

func newGuanYuS1000LiMarch() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Guan Yu's 1,000-Li March",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Tapped: game.TriTrue}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all tapped creatures.
		`,
		},
	}
}
