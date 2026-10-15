package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AscendedLawmage is the card definition for Ascended Lawmage.
//
// Type: Creature — Vedalken Wizard
// Cost: {2}{W}{U}
//
// Oracle text:
//
//	Flying
//	Hexproof (This creature can't be the target of spells or abilities your opponents control.)
var AscendedLawmage = newAscendedLawmage

func newAscendedLawmage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Ascended Lawmage",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vedalken, types.Wizard},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.HexproofStaticBody,
			},
			OracleText: `
			Flying
			Hexproof (This creature can't be the target of spells or abilities your opponents control.)
		`,
		},
	}
}
