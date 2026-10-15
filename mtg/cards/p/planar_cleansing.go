package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PlanarCleansing is the card definition for Planar Cleansing.
//
// Type: Sorcery
// Cost: {3}{W}{W}{W}
//
// Oracle text:
//
//	Destroy all nonland permanents.
var PlanarCleansing = newPlanarCleansing

func newPlanarCleansing() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Planar Cleansing",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{ExcludedTypes: []types.Card{types.Land}}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all nonland permanents.
		`,
		},
	}
}
