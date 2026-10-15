package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SphinxOfJwarIsle is the card definition for Sphinx of Jwar Isle.
//
// Type: Creature — Sphinx
// Cost: {4}{U}{U}
//
// Oracle text:
//
//	Flying
//	Shroud (This creature can't be the target of spells or abilities.)
//	You may look at the top card of your library any time.
var SphinxOfJwarIsle = newSphinxOfJwarIsle

func newSphinxOfJwarIsle() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Sphinx of Jwar Isle",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Sphinx},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.ShroudStaticBody,
				game.LookAtTopCardAnyTimeStaticBody,
			},
			OracleText: `
			Flying
			Shroud (This creature can't be the target of spells or abilities.)
			You may look at the top card of your library any time.
		`,
		},
	}
}
