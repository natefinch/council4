package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ScalebaneSElite is the card definition for Scalebane's Elite.
//
// Type: Creature — Human Soldier
// Cost: {3}{G}{W}
//
// Oracle text:
//
//	Protection from black
var ScalebaneSElite = newScalebaneSElite

func newScalebaneSElite() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Scalebane's Elite",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.W,
			}),
			Colors:    []color.Color{color.Green, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.ProtectionFromColorsStaticAbility(color.Black),
			},
			OracleText: `
			Protection from black
		`,
		},
	}
}
