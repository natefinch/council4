package y

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// YargleAndMultani is the card definition for Yargle and Multani.
//
// Type: Legendary Creature — Frog Spirit Elemental
// Cost: {3}{B}{B}{G}
//
// Oracle text:
var YargleAndMultani = newYargleAndMultani

func newYargleAndMultani() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Yargle and Multani",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.B,
				cost.G,
			}),
			Colors:     []color.Color{color.Black, color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Frog, types.Spirit, types.Elemental},
			Power:      opt.Val(game.PT{Value: 18}),
			Toughness:  opt.Val(game.PT{Value: 6}),
		},
	}
}
