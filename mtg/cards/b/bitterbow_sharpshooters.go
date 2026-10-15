package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BitterbowSharpshooters is the card definition for Bitterbow Sharpshooters.
//
// Type: Creature — Jackal Archer
// Cost: {4}{G}
//
// Oracle text:
//
//	Reach, vigilance
var BitterbowSharpshooters = newBitterbowSharpshooters

func newBitterbowSharpshooters() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Bitterbow Sharpshooters",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Jackal, types.Archer},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.VigilanceStaticBody,
			},
			OracleText: `
			Reach, vigilance
		`,
		},
	}
}
