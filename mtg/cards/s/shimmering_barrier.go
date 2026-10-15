package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ShimmeringBarrier is the card definition for Shimmering Barrier.
//
// Type: Creature — Wall
// Cost: {1}{W}
//
// Oracle text:
//
//	Defender (This creature can't attack.)
//	First strike
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var ShimmeringBarrier = newShimmeringBarrier

func newShimmeringBarrier() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Shimmering Barrier",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wall},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
				game.FirstStrikeStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			Defender (This creature can't attack.)
			First strike
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
