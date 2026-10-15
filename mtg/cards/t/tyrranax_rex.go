package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TyrranaxRex is the card definition for Tyrranax Rex.
//
// Type: Creature — Phyrexian Dinosaur
// Cost: {4}{G}{G}{G}
//
// Oracle text:
//
//	This spell can't be countered.
//	Trample, ward {4}, haste
//	Toxic 4 (Players dealt combat damage by this creature also get four poison counters.)
var TyrranaxRex = newTyrranaxRex

func newTyrranaxRex() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Tyrranax Rex",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Dinosaur},
			Power:     opt.Val(game.PT{Value: 8}),
			Toughness: opt.Val(game.PT{Value: 8}),
			StaticAbilities: []game.StaticAbility{
				game.CantBeCounteredStaticBody,
				game.TrampleStaticBody,
				game.WardStaticAbility(cost.Mana{cost.O(4)}),
				game.HasteStaticBody,
				game.StaticAbility{
					Text: "Toxic 4",
					KeywordAbilities: []game.KeywordAbility{
						game.ToxicKeyword{
							Amount: 4,
						},
					},
				},
			},
			OracleText: `
			This spell can't be countered.
			Trample, ward {4}, haste
			Toxic 4 (Players dealt combat damage by this creature also get four poison counters.)
		`,
		},
	}
}
