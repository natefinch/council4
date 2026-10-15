package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DauthiSlayer is the card definition for Dauthi Slayer.
//
// Type: Creature — Dauthi Soldier
// Cost: {B}{B}
//
// Oracle text:
//
//	Shadow (This creature can block or be blocked by only creatures with shadow.)
//	This creature attacks each combat if able.
var DauthiSlayer = newDauthiSlayer

func newDauthiSlayer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Dauthi Slayer",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dauthi, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.ShadowStaticBody,
				game.MustAttackStaticBody,
			},
			OracleText: `
			Shadow (This creature can block or be blocked by only creatures with shadow.)
			This creature attacks each combat if able.
		`,
		},
	}
}
