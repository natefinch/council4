package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DauthiMarauder is the card definition for Dauthi Marauder.
//
// Type: Creature — Dauthi Minion
// Cost: {2}{B}
//
// Oracle text:
//
//	Shadow (This creature can block or be blocked by only creatures with shadow.)
var DauthiMarauder = newDauthiMarauder

func newDauthiMarauder() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Dauthi Marauder",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dauthi, types.Minion},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ShadowStaticBody,
			},
			OracleText: `
			Shadow (This creature can block or be blocked by only creatures with shadow.)
		`,
		},
	}
}
