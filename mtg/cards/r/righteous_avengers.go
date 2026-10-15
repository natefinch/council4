package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RighteousAvengers is the card definition for Righteous Avengers.
//
// Type: Creature — Human Soldier
// Cost: {4}{W}
//
// Oracle text:
//
//	Plainswalk (This creature can't be blocked as long as defending player controls a Plains.)
var RighteousAvengers = newRighteousAvengers

func newRighteousAvengers() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Righteous Avengers",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.PlainswalkStaticBody,
			},
			OracleText: `
			Plainswalk (This creature can't be blocked as long as defending player controls a Plains.)
		`,
		},
	}
}
