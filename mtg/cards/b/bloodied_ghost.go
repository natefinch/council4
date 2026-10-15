package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BloodiedGhost is the card definition for Bloodied Ghost.
//
// Type: Creature — Spirit
// Cost: {1}{W/B}{W/B}
//
// Oracle text:
//
//	Flying
//	This creature enters with a -1/-1 counter on it.
var BloodiedGhost = newBloodiedGhost

func newBloodiedGhost() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Bloodied Ghost",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.HybridMana(mana.W, mana.B),
				cost.HybridMana(mana.W, mana.B),
			}),
			Colors:    []color.Color{color.Black, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with a -1/-1 counter on it.", game.CounterPlacement{Kind: counter.MinusOneMinusOne, Amount: 1}),
			},
			OracleText: `
			Flying
			This creature enters with a -1/-1 counter on it.
		`,
		},
	}
}
