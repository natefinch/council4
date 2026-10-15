package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SwarmOfRats is the card definition for Swarm of Rats.
//
// Type: Creature — Rat
// Cost: {1}{B}
//
// Oracle text:
//
//	Swarm of Rats's power is equal to the number of Rats you control.
var SwarmOfRats = newSwarmOfRats

func newSwarmOfRats() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Swarm of Rats",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:       []color.Color{color.Black},
			Types:        []types.Card{types.Creature},
			Subtypes:     []types.Sub{types.Rat},
			Power:        opt.Val(game.PT{IsStar: true}),
			Toughness:    opt.Val(game.PT{Value: 1}),
			DynamicPower: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerSubtypeCount, Subtype: types.Sub("Rat")}),
			OracleText: `
			Swarm of Rats's power is equal to the number of Rats you control.
		`,
		},
	}
}
