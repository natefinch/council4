package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ShiftingWall is the card definition for Shifting Wall.
//
// Type: Artifact Creature — Wall
// Cost: {X}
//
// Oracle text:
//
//	Defender (This creature can't attack.)
//	This creature enters with X +1/+1 counters on it.
var ShiftingWall = newShiftingWall

func newShiftingWall() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Shifting Wall",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Wall},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with X +1/+1 counters on it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, AmountFromX: true}),
			},
			OracleText: `
			Defender (This creature can't attack.)
			This creature enters with X +1/+1 counters on it.
		`,
		},
	}
}
