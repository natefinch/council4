package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FaerieSwarm is the card definition for Faerie Swarm.
//
// Type: Creature — Faerie
// Cost: {3}{U}
//
// Oracle text:
//
//	Flying
//	Faerie Swarm's power and toughness are each equal to the number of blue permanents you control.
var FaerieSwarm = newFaerieSwarm

func newFaerieSwarm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Faerie Swarm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:           []color.Color{color.Blue},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Faerie},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerColorPermanentCount, Color: color.Blue}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerColorPermanentCount, Color: color.Blue}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			OracleText: `
			Flying
			Faerie Swarm's power and toughness are each equal to the number of blue permanents you control.
		`,
		},
	}
}
