package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SwoopingProtector is the card definition for Swooping Protector.
//
// Type: Creature — Bird Citizen
// Cost: {3}{W}
//
// Oracle text:
//
//	Flash
//	Flying
//	This creature enters with a shield counter on it. (If it would be dealt damage or destroyed, remove a shield counter from it instead.)
var SwoopingProtector = newSwoopingProtector

func newSwoopingProtector() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Swooping Protector",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird, types.Citizen},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.FlyingStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with a shield counter on it. (If it would be dealt damage or destroyed, remove a shield counter from it instead.)", game.CounterPlacement{Kind: counter.Shield, Amount: 1}),
			},
			OracleText: `
			Flash
			Flying
			This creature enters with a shield counter on it. (If it would be dealt damage or destroyed, remove a shield counter from it instead.)
		`,
		},
	}
}
