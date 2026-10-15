package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RavagesOfWar is the card definition for Ravages of War.
//
// Type: Sorcery
// Cost: {3}{W}
//
// Oracle text:
//
//	Destroy all lands.
var RavagesOfWar = newRavagesOfWar

func newRavagesOfWar() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Ravages of War",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all lands.
		`,
		},
	}
}
