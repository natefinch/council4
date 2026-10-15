package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GiganticBigBear is the card definition for Gigantic Big Bear.
//
// Type: Creature — Bear
// Cost: {5}{G}{G}
//
// Oracle text:
//
//	This spell can't be countered.
//	Hexproof, haste
var GiganticBigBear = newGiganticBigBear

func newGiganticBigBear() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Gigantic Big Bear",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bear},
			Power:     opt.Val(game.PT{Value: 10}),
			Toughness: opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.CantBeCounteredStaticBody,
				game.HexproofStaticBody,
				game.HasteStaticBody,
			},
			OracleText: `
			This spell can't be countered.
			Hexproof, haste
		`,
		},
	}
}
