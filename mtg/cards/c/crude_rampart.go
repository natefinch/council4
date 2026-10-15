package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CrudeRampart is the card definition for Crude Rampart.
//
// Type: Creature — Wall
// Cost: {3}{W}
//
// Oracle text:
//
//	Defender (This creature can't attack.)
//	Morph {4}{W} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
var CrudeRampart = newCrudeRampart

func newCrudeRampart() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Crude Rampart",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wall},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
				game.StaticAbility{
					Text: "Morph {4}{W}",
					KeywordAbilities: []game.KeywordAbility{
						game.MorphKeyword{
							Cost: cost.Mana{
								cost.O(4),
								cost.W,
							},
						},
					},
				},
			},
			OracleText: `
			Defender (This creature can't attack.)
			Morph {4}{W} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
		`,
		},
	}
}
