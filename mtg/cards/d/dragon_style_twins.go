package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DragonStyleTwins is the card definition for Dragon-Style Twins.
//
// Type: Creature — Human Monk
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	Double strike
//	Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
var DragonStyleTwins = newDragonStyleTwins

func newDragonStyleTwins() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Dragon-Style Twins",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Monk},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DoubleStrikeStaticBody,
				game.ProwessStaticBody,
			},
			OracleText: `
			Double strike
			Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
		`,
		},
	}
}
