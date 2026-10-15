package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BaneslayerAngel is the card definition for Baneslayer Angel.
//
// Type: Creature — Angel
// Cost: {3}{W}{W}
//
// Oracle text:
//
//	Flying, first strike, lifelink, protection from Demons and from Dragons
var BaneslayerAngel = newBaneslayerAngel

func newBaneslayerAngel() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Baneslayer Angel",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Angel},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.FirstStrikeStaticBody,
				game.LifelinkStaticBody,
				game.ProtectionFromSubtypesStaticAbility(types.Demon, types.Dragon),
			},
			OracleText: `
			Flying, first strike, lifelink, protection from Demons and from Dragons
		`,
		},
	}
}
