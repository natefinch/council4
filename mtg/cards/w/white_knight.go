package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WhiteKnight is the card definition for White Knight.
//
// Type: Creature — Human Knight
// Cost: {W}{W}
//
// Oracle text:
//
//	First strike (This creature deals combat damage before creatures without first strike.)
//	Protection from black (This creature can't be blocked, targeted, dealt damage, or enchanted by anything black.)
var WhiteKnight = newWhiteKnight

func newWhiteKnight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "White Knight",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
				game.ProtectionFromColorsStaticAbility(color.Black),
			},
			OracleText: `
			First strike (This creature deals combat damage before creatures without first strike.)
			Protection from black (This creature can't be blocked, targeted, dealt damage, or enchanted by anything black.)
		`,
		},
	}
}
