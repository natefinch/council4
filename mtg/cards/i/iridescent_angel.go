package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IridescentAngel is the card definition for Iridescent Angel.
//
// Type: Creature — Angel
// Cost: {5}{W}{U}
//
// Oracle text:
//
//	Flying, protection from each color
var IridescentAngel = newIridescentAngel

func newIridescentAngel() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Iridescent Angel",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Angel},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.ProtectionFromEachColorStaticAbility(),
			},
			OracleText: `
			Flying, protection from each color
		`,
		},
	}
}
