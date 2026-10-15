package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SaguArcher is the card definition for Sagu Archer.
//
// Type: Creature — Snake Archer
// Cost: {4}{G}
//
// Oracle text:
//
//	Reach (This creature can block creatures with flying.)
//	Morph {4}{G} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
var SaguArcher = newSaguArcher

func newSaguArcher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Sagu Archer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Snake, types.Archer},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.StaticAbility{
					Text: "Morph {4}{G}",
					KeywordAbilities: []game.KeywordAbility{
						game.MorphKeyword{
							Cost: cost.Mana{
								cost.O(4),
								cost.G,
							},
						},
					},
				},
			},
			OracleText: `
			Reach (This creature can block creatures with flying.)
			Morph {4}{G} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
		`,
		},
	}
}
