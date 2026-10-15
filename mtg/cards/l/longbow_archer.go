package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LongbowArcher is the card definition for Longbow Archer.
//
// Type: Creature — Human Soldier Archer
// Cost: {W}{W}
//
// Oracle text:
//
//	Reach (This creature can block creatures with flying.)
//	First strike
var LongbowArcher = newLongbowArcher

func newLongbowArcher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Longbow Archer",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier, types.Archer},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.FirstStrikeStaticBody,
			},
			OracleText: `
			Reach (This creature can block creatures with flying.)
			First strike
		`,
		},
	}
}
