package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AzoriusFirstWing is the card definition for Azorius First-Wing.
//
// Type: Creature — Griffin
// Cost: {W}{U}
//
// Oracle text:
//
//	Flying, protection from enchantments
var AzoriusFirstWing = newAzoriusFirstWing

func newAzoriusFirstWing() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Azorius First-Wing",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Griffin},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.ProtectionFromTypesStaticAbility(types.Enchantment),
			},
			OracleText: `
			Flying, protection from enchantments
		`,
		},
	}
}
