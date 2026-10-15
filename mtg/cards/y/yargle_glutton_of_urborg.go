package y

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// YargleGluttonOfUrborg is the card definition for Yargle, Glutton of Urborg.
//
// Type: Legendary Creature — Frog Spirit
// Cost: {4}{B}
//
// Oracle text:
var YargleGluttonOfUrborg = newYargleGluttonOfUrborg

func newYargleGluttonOfUrborg() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Yargle, Glutton of Urborg",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Frog, types.Spirit},
			Power:      opt.Val(game.PT{Value: 9}),
			Toughness:  opt.Val(game.PT{Value: 3}),
		},
	}
}
