package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CopperHostCrusher is the card definition for Copper Host Crusher.
//
// Type: Creature — Phyrexian Bear Rhino
// Cost: {6}{G}{G}
//
// Oracle text:
//
//	Trample
//	Hexproof (This creature can't be the target of spells or abilities your opponents control.)
var CopperHostCrusher = newCopperHostCrusher

func newCopperHostCrusher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Copper Host Crusher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Bear, types.Rhino},
			Power:     opt.Val(game.PT{Value: 8}),
			Toughness: opt.Val(game.PT{Value: 8}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.HexproofStaticBody,
			},
			OracleText: `
			Trample
			Hexproof (This creature can't be the target of spells or abilities your opponents control.)
		`,
		},
	}
}
