package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BogTatters is the card definition for Bog Tatters.
//
// Type: Creature — Wraith
// Cost: {4}{B}
//
// Oracle text:
//
//	Swampwalk (This creature can't be blocked as long as defending player controls a Swamp.)
var BogTatters = newBogTatters

func newBogTatters() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Bog Tatters",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wraith},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.SwampwalkStaticBody,
			},
			OracleText: `
			Swampwalk (This creature can't be blocked as long as defending player controls a Swamp.)
		`,
		},
	}
}
