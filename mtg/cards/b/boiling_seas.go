package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BoilingSeas is the card definition for Boiling Seas.
//
// Type: Sorcery
// Cost: {3}{R}
//
// Oracle text:
//
//	Destroy all Islands.
var BoilingSeas = newBoilingSeas

func newBoilingSeas() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Boiling Seas",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Island")}}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all Islands.
		`,
		},
	}
}
