package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ShuEliteCompanions is the card definition for Shu Elite Companions.
//
// Type: Creature — Human Soldier
// Cost: {4}{W}
//
// Oracle text:
//
//	Horsemanship (This creature can't be blocked except by creatures with horsemanship.)
var ShuEliteCompanions = newShuEliteCompanions

func newShuEliteCompanions() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Shu Elite Companions",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.HorsemanshipStaticBody,
			},
			OracleText: `
			Horsemanship (This creature can't be blocked except by creatures with horsemanship.)
		`,
		},
	}
}
