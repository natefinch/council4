package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KrenkoSEnforcer is the card definition for Krenko's Enforcer.
//
// Type: Creature — Goblin Warrior
// Cost: {1}{R}{R}
//
// Oracle text:
//
//	Intimidate (This creature can't be blocked except by artifact creatures and/or creatures that share a color with it.)
var KrenkoSEnforcer = newKrenkoSEnforcer

func newKrenkoSEnforcer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Krenko's Enforcer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.IntimidateStaticBody,
			},
			OracleText: `
			Intimidate (This creature can't be blocked except by artifact creatures and/or creatures that share a color with it.)
		`,
		},
	}
}
