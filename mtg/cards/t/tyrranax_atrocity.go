package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TyrranaxAtrocity is the card definition for Tyrranax Atrocity.
//
// Type: Creature — Phyrexian Dinosaur
// Cost: {3}{G}{G}
//
// Oracle text:
//
//	Haste
//	Toxic 3 (Players dealt combat damage by this creature also get three poison counters.)
var TyrranaxAtrocity = newTyrranaxAtrocity

func newTyrranaxAtrocity() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Tyrranax Atrocity",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Dinosaur},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
				game.StaticAbility{
					Text: "Toxic 3",
					KeywordAbilities: []game.KeywordAbility{
						game.ToxicKeyword{
							Amount: 3,
						},
					},
				},
			},
			OracleText: `
			Haste
			Toxic 3 (Players dealt combat damage by this creature also get three poison counters.)
		`,
		},
	}
}
