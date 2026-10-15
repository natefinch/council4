package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StormFleetSprinter is the card definition for Storm Fleet Sprinter.
//
// Type: Creature — Human Pirate
// Cost: {1}{U}{R}
//
// Oracle text:
//
//	Haste
//	This creature can't be blocked.
var StormFleetSprinter = newStormFleetSprinter

func newStormFleetSprinter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Storm Fleet Sprinter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.R,
			}),
			Colors:    []color.Color{color.Red, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Pirate},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
				game.CantBeBlockedStaticBody,
			},
			OracleText: `
			Haste
			This creature can't be blocked.
		`,
		},
	}
}
