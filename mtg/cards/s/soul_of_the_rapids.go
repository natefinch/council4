package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SoulOfTheRapids is the card definition for Soul of the Rapids.
//
// Type: Creature — Elemental
// Cost: {3}{U}{U}
//
// Oracle text:
//
//	Flying
//	Hexproof (This creature can't be the target of spells or abilities your opponents control.)
var SoulOfTheRapids = newSoulOfTheRapids

func newSoulOfTheRapids() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Soul of the Rapids",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
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
