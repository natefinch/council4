package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KnightOfGlory is the card definition for Knight of Glory.
//
// Type: Creature — Human Knight
// Cost: {1}{W}
//
// Oracle text:
//
//	Protection from black (This creature can't be blocked, targeted, dealt damage, or enchanted by anything black.)
//	Exalted (Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.)
var KnightOfGlory = newKnightOfGlory

func newKnightOfGlory() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Knight of Glory",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ProtectionFromColorsStaticAbility(color.Black),
				game.ExaltedStaticBody,
			},
			OracleText: `
			Protection from black (This creature can't be blocked, targeted, dealt damage, or enchanted by anything black.)
			Exalted (Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.)
		`,
		},
	}
}
