package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SnowhornRider is the card definition for Snowhorn Rider.
//
// Type: Creature — Human Warrior
// Cost: {3}{G}{U}{R}
//
// Oracle text:
//
//	Trample
//	Morph {2}{G}{U}{R} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
var SnowhornRider = newSnowhornRider

func newSnowhornRider() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Snowhorn Rider",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.U,
				cost.R,
			}),
			Colors:    []color.Color{color.Green, color.Red, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Warrior},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.StaticAbility{
					Text: "Morph {2}{G}{U}{R}",
					KeywordAbilities: []game.KeywordAbility{
						game.MorphKeyword{
							Cost: cost.Mana{
								cost.O(2),
								cost.G,
								cost.U,
								cost.R,
							},
						},
					},
				},
			},
			OracleText: `
			Trample
			Morph {2}{G}{U}{R} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
		`,
		},
	}
}
