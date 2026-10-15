package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NiblisOfDusk is the card definition for Niblis of Dusk.
//
// Type: Creature — Spirit
// Cost: {2}{U}
//
// Oracle text:
//
//	Flying
//	Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
var NiblisOfDusk = newNiblisOfDusk

func newNiblisOfDusk() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Niblis of Dusk",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.ProwessStaticBody,
			},
			OracleText: `
			Flying
			Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
		`,
		},
	}
}
