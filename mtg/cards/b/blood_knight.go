package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BloodKnight is the card definition for Blood Knight.
//
// Type: Creature — Human Knight
// Cost: {R}{R}
//
// Oracle text:
//
//	First strike, protection from white
var BloodKnight = newBloodKnight

func newBloodKnight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Blood Knight",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
				game.ProtectionFromColorsStaticAbility(color.White),
			},
			OracleText: `
			First strike, protection from white
		`,
		},
	}
}
