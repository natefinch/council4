package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SkyhunterProwler is the card definition for Skyhunter Prowler.
//
// Type: Creature — Cat Knight
// Cost: {2}{W}
//
// Oracle text:
//
//	Flying, vigilance (This creature can't be blocked except by creatures with flying or reach, and attacking doesn't cause this creature to tap.)
var SkyhunterProwler = newSkyhunterProwler

func newSkyhunterProwler() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Skyhunter Prowler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat, types.Knight},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.VigilanceStaticBody,
			},
			OracleText: `
			Flying, vigilance (This creature can't be blocked except by creatures with flying or reach, and attacking doesn't cause this creature to tap.)
		`,
		},
	}
}
