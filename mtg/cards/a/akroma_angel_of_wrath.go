package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AkromaAngelOfWrath is the card definition for Akroma, Angel of Wrath.
//
// Type: Legendary Creature — Angel
// Cost: {5}{W}{W}{W}
//
// Oracle text:
//
//	Flying, first strike, vigilance, trample, haste, protection from black and from red
var AkromaAngelOfWrath = newAkromaAngelOfWrath

func newAkromaAngelOfWrath() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Akroma, Angel of Wrath",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
				cost.W,
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Angel},
			Power:      opt.Val(game.PT{Value: 6}),
			Toughness:  opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.FirstStrikeStaticBody,
				game.VigilanceStaticBody,
				game.TrampleStaticBody,
				game.HasteStaticBody,
				game.ProtectionFromColorsStaticAbility(color.Black, color.Red),
			},
			OracleText: `
			Flying, first strike, vigilance, trample, haste, protection from black and from red
		`,
		},
	}
}
