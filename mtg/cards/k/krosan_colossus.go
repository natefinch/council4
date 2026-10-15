package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KrosanColossus is the card definition for Krosan Colossus.
//
// Type: Creature — Beast
// Cost: {6}{G}{G}{G}
//
// Oracle text:
//
//	Morph {6}{G}{G} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
var KrosanColossus = newKrosanColossus

func newKrosanColossus() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Krosan Colossus",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.G,
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 9}),
			Toughness: opt.Val(game.PT{Value: 9}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Morph {6}{G}{G}",
					KeywordAbilities: []game.KeywordAbility{
						game.MorphKeyword{
							Cost: cost.Mana{
								cost.O(6),
								cost.G,
								cost.G,
							},
						},
					},
				},
			},
			OracleText: `
			Morph {6}{G}{G} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
		`,
		},
	}
}
