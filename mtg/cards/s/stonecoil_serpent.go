package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StonecoilSerpent is the card definition for Stonecoil Serpent.
//
// Type: Artifact Creature — Snake
// Cost: {X}
//
// Oracle text:
//
//	Reach, trample, protection from multicolored
//	This creature enters with X +1/+1 counters on it.
var StonecoilSerpent = newStonecoilSerpent

func newStonecoilSerpent() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Stonecoil Serpent",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Snake},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.TrampleStaticBody,
				game.ProtectionFromMulticoloredStaticAbility(),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with X +1/+1 counters on it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, AmountFromX: true}),
			},
			OracleText: `
			Reach, trample, protection from multicolored
			This creature enters with X +1/+1 counters on it.
		`,
		},
	}
}
