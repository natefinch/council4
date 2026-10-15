package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PhyrexianWurm is the card definition for Phyrexian Wurm.
//
// Type: Token Creature — Phyrexian Wurm
//
// Oracle text:
//   Trample
//   Toxic 1 (Players dealt combat damage by this creature also get a poison counter.)

// PhyrexianWurmTokenba0ac1b82873448d861e6e975381673d is the card definition for Phyrexian Wurm.
var PhyrexianWurmTokenba0ac1b82873448d861e6e975381673d = newPhyrexianWurmTokenba0ac1b82873448d861e6e975381673d()

func newPhyrexianWurmTokenba0ac1b82873448d861e6e975381673d() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:      "Phyrexian Wurm",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Wurm},
			Power:     opt.Val(game.PT{IsStar: true}),
			Toughness: opt.Val(game.PT{IsStar: true}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
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
			Trample
			Toxic 1 (Players dealt combat damage by this creature also get a poison counter.)
		`,
		},
	}
}
