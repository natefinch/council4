package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TempestOfLight is the card definition for Tempest of Light.
//
// Type: Instant
// Cost: {2}{W}
//
// Oracle text:
//
//	Destroy all enchantments.
var TempestOfLight = newTempestOfLight

func newTempestOfLight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Tempest of Light",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Enchantment}}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all enchantments.
		`,
		},
	}
}
