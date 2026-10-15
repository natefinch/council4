package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DerangedWhelp is the card definition for Deranged Whelp.
//
// Type: Creature — Wolf
// Cost: {1}{R}
//
// Oracle text:
//
//	Menace (This creature can't be blocked except by two or more creatures.)
var DerangedWhelp = newDerangedWhelp

func newDerangedWhelp() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Deranged Whelp",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wolf},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
			},
			OracleText: `
			Menace (This creature can't be blocked except by two or more creatures.)
		`,
		},
	}
}
