package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GurmagAngler is the card definition for Gurmag Angler.
//
// Type: Creature — Zombie Fish
// Cost: {6}{B}
//
// Oracle text:
//
//	Delve (Each card you exile from your graveyard while casting this spell pays for {1}.)
var GurmagAngler = newGurmagAngler

func newGurmagAngler() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Gurmag Angler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Fish},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.DelveStaticBody,
			},
			OracleText: `
			Delve (Each card you exile from your graveyard while casting this spell pays for {1}.)
		`,
		},
	}
}
