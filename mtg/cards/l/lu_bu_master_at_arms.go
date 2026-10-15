package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LuBuMasterAtArms is the card definition for Lu Bu, Master-at-Arms.
//
// Type: Legendary Creature — Human Soldier Warrior
// Cost: {5}{R}
//
// Oracle text:
//
//	Haste; horsemanship (This creature can't be blocked except by creatures with horsemanship.)
var LuBuMasterAtArms = newLuBuMasterAtArms

func newLuBuMasterAtArms() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Lu Bu, Master-at-Arms",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Soldier, types.Warrior},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
				game.HorsemanshipStaticBody,
			},
			OracleText: `
			Haste; horsemanship (This creature can't be blocked except by creatures with horsemanship.)
		`,
		},
	}
}
