package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KalonianBehemoth is the card definition for Kalonian Behemoth.
//
// Type: Creature — Beast
// Cost: {5}{G}{G}
//
// Oracle text:
//
//	Shroud (This creature can't be the target of spells or abilities.)
var KalonianBehemoth = newKalonianBehemoth

func newKalonianBehemoth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Kalonian Behemoth",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 9}),
			Toughness: opt.Val(game.PT{Value: 9}),
			StaticAbilities: []game.StaticAbility{
				game.ShroudStaticBody,
			},
			OracleText: `
			Shroud (This creature can't be the target of spells or abilities.)
		`,
		},
	}
}
