package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RiverwheelAerialists is the card definition for Riverwheel Aerialists.
//
// Type: Creature — Djinn Monk
// Cost: {5}{U}
//
// Oracle text:
//
//	Flying
//	Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
var RiverwheelAerialists = newRiverwheelAerialists

func newRiverwheelAerialists() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Riverwheel Aerialists",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Djinn, types.Monk},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 5}),
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
