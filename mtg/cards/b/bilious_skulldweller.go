package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BiliousSkulldweller is the card definition for Bilious Skulldweller.
//
// Type: Creature — Phyrexian Insect
// Cost: {B}
//
// Oracle text:
//
//	Deathtouch
//	Toxic 1 (Players dealt combat damage by this creature also get a poison counter.)
var BiliousSkulldweller = newBiliousSkulldweller

func newBiliousSkulldweller() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Bilious Skulldweller",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Insect},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.DeathtouchStaticBody,
				game.StaticAbility{
					Text: "Toxic 1",
					KeywordAbilities: []game.KeywordAbility{
						game.ToxicKeyword{
							Amount: 1,
						},
					},
				},
			},
			OracleText: `
			Deathtouch
			Toxic 1 (Players dealt combat damage by this creature also get a poison counter.)
		`,
		},
	}
}
