package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GladecoverScout is the card definition for Gladecover Scout.
//
// Type: Creature — Elf Scout
// Cost: {G}
//
// Oracle text:
//
//	Hexproof (This creature can't be the target of spells or abilities your opponents control.)
var GladecoverScout = newGladecoverScout

func newGladecoverScout() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Gladecover Scout",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Scout},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.HexproofStaticBody,
			},
			OracleText: `
			Hexproof (This creature can't be the target of spells or abilities your opponents control.)
		`,
		},
	}
}
