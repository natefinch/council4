package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SaguMauler is the card definition for Sagu Mauler.
//
// Type: Creature — Beast
// Cost: {4}{G}{U}
//
// Oracle text:
//
//	Trample, hexproof
//	Morph {3}{G}{U} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
var SaguMauler = newSaguMauler

func newSaguMauler() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Sagu Mauler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.U,
			}),
			Colors:    []color.Color{color.Green, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.HexproofStaticBody,
				game.StaticAbility{
					Text: "Morph {3}{G}{U}",
					KeywordAbilities: []game.KeywordAbility{
						game.MorphKeyword{
							Cost: cost.Mana{
								cost.O(3),
								cost.G,
								cost.U,
							},
						},
					},
				},
			},
			OracleText: `
			Trample, hexproof
			Morph {3}{G}{U} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
		`,
		},
	}
}
