package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CatWarriors is the card definition for Cat Warriors.
//
// Type: Creature — Cat Warrior
// Cost: {1}{G}{G}
//
// Oracle text:
//
//	Forestwalk (This creature can't be blocked as long as defending player controls a Forest.)
var CatWarriors = newCatWarriors

func newCatWarriors() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Cat Warriors",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.ForestwalkStaticBody,
			},
			OracleText: `
			Forestwalk (This creature can't be blocked as long as defending player controls a Forest.)
		`,
		},
	}
}
