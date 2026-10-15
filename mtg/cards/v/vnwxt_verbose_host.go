package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VnwxtVerboseHost is the card definition for Vnwxt, Verbose Host.
//
// Type: Legendary Creature — Homunculus
// Cost: {1}{U}
//
// Oracle text:
//
//	Start your engines! (If you have no speed, it starts at 1. It increases once on each of your turns when an opponent loses life. Max speed is 4.)
//	You have no maximum hand size.
//	Max speed — If you would draw a card, draw two cards instead.
var VnwxtVerboseHost = newVnwxtVerboseHost

func newVnwxtVerboseHost() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Vnwxt, Verbose Host",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Homunculus},
			Power:      opt.Val(game.PT{Value: 0}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.NoMaximumHandSizeStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.StartEnginesTriggeredBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.MaxSpeedDrawCardMultiplierReplacement("Max speed — If you would draw a card, draw two cards instead.", 2, false),
			},
			OracleText: `
			Start your engines! (If you have no speed, it starts at 1. It increases once on each of your turns when an opponent loses life. Max speed is 4.)
			You have no maximum hand size.
			Max speed — If you would draw a card, draw two cards instead.
		`,
		},
	}
}
