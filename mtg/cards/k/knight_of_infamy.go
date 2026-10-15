package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KnightOfInfamy is the card definition for Knight of Infamy.
//
// Type: Creature — Human Knight
// Cost: {1}{B}
//
// Oracle text:
//
//	Protection from white (This creature can't be blocked, targeted, dealt damage, or enchanted by anything white.)
//	Exalted (Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.)
var KnightOfInfamy = newKnightOfInfamy

func newKnightOfInfamy() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Knight of Infamy",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ProtectionFromColorsStaticAbility(color.White),
				game.ExaltedStaticBody,
			},
			OracleText: `
			Protection from white (This creature can't be blocked, targeted, dealt damage, or enchanted by anything white.)
			Exalted (Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.)
		`,
		},
	}
}
