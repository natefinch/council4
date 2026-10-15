package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VestigeOfEmrakul is the card definition for Vestige of Emrakul.
//
// Type: Creature — Eldrazi Drone
// Cost: {3}{R}
//
// Oracle text:
//
//	Devoid (This card has no color.)
//	Trample
var VestigeOfEmrakul = newVestigeOfEmrakul

func newVestigeOfEmrakul() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Vestige of Emrakul",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi, types.Drone},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.DevoidStaticBody,
				game.TrampleStaticBody,
			},
			OracleText: `
			Devoid (This card has no color.)
			Trample
		`,
		},
	}
}
