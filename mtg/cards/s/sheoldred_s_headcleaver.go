package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SheoldredSHeadcleaver is the card definition for Sheoldred's Headcleaver.
//
// Type: Creature — Phyrexian Warrior
// Cost: {3}{B}
//
// Oracle text:
//
//	Menace
//	Toxic 2 (Players dealt combat damage by this creature also get two poison counters.)
var SheoldredSHeadcleaver = newSheoldredSHeadcleaver

func newSheoldredSHeadcleaver() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Sheoldred's Headcleaver",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
				game.StaticAbility{
					Text: "Toxic 2",
					KeywordAbilities: []game.KeywordAbility{
						game.ToxicKeyword{
							Amount: 2,
						},
					},
				},
			},
			OracleText: `
			Menace
			Toxic 2 (Players dealt combat damage by this creature also get two poison counters.)
		`,
		},
	}
}
