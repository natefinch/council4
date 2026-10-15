package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BlackKnight is the card definition for Black Knight.
//
// Type: Creature — Human Knight
// Cost: {B}{B}
//
// Oracle text:
//
//	First strike (This creature deals combat damage before creatures without first strike.)
//	Protection from white (This creature can't be blocked, targeted, dealt damage, or enchanted by anything white.)
var BlackKnight = newBlackKnight

func newBlackKnight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Black Knight",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
				game.ProtectionFromColorsStaticAbility(color.White),
			},
			OracleText: `
			First strike (This creature deals combat damage before creatures without first strike.)
			Protection from white (This creature can't be blocked, targeted, dealt damage, or enchanted by anything white.)
		`,
		},
	}
}
