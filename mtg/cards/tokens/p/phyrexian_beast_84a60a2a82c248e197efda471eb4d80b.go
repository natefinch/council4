package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PhyrexianBeast is the card definition for Phyrexian Beast.
//
// Type: Token Creature — Phyrexian Beast
//
// Oracle text:
//   Toxic 1 (Players dealt combat damage by this creature also get a poison counter.)

// PhyrexianBeastToken84a60a2a82c248e197efda471eb4d80b is the card definition for Phyrexian Beast.
var PhyrexianBeastToken84a60a2a82c248e197efda471eb4d80b = newPhyrexianBeastToken84a60a2a82c248e197efda471eb4d80b()

func newPhyrexianBeastToken84a60a2a82c248e197efda471eb4d80b() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:      "Phyrexian Beast",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Beast},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
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
			Toxic 1 (Players dealt combat damage by this creature also get a poison counter.)
		`,
		},
	}
}
