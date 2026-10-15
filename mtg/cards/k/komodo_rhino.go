package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KomodoRhino is the card definition for Komodo Rhino.
//
// Type: Creature — Lizard Rhino
// Cost: {3}{R}
//
// Oracle text:
//
//	Trample (This creature can deal excess combat damage to the player it's attacking.)
var KomodoRhino = newKomodoRhino

func newKomodoRhino() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Komodo Rhino",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Lizard, types.Rhino},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			OracleText: `
			Trample (This creature can deal excess combat damage to the player it's attacking.)
		`,
		},
	}
}
