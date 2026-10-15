package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WraithViciousVigilante is the card definition for Wraith, Vicious Vigilante.
//
// Type: Legendary Creature — Human Detective Hero
// Cost: {1}{W}{U}
//
// Oracle text:
//
//	Double strike
//	Fear Gas — Wraith can't be blocked.
var WraithViciousVigilante = newWraithViciousVigilante

func newWraithViciousVigilante() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Wraith, Vicious Vigilante",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Detective, types.Hero},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.DoubleStrikeStaticBody,
				game.CantBeBlockedStaticBody,
			},
			OracleText: `
			Double strike
			Fear Gas — Wraith can't be blocked.
		`,
		},
	}
}
