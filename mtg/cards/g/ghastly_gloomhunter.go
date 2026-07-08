package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GhastlyGloomhunter is the card definition for Ghastly Gloomhunter.
//
// Type: Creature — Zombie Bat
// Cost: {1}{B}
//
// Oracle text:
//
//	Kicker {3}{B} (You may pay an additional {3}{B} as you cast this spell.)
//	Flying, lifelink
//	If this creature was kicked, it enters with two +1/+1 counters on it.
var GhastlyGloomhunter = newGhastlyGloomhunter

func newGhastlyGloomhunter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Ghastly Gloomhunter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Bat},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{Cost: cost.Mana{cost.O(3), cost.B}},
					},
				},
				game.FlyingStaticBody,
				game.LifelinkStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("If this creature was kicked, it enters with two +1/+1 counters on it.", &game.Condition{
					EventPermanentWasKicked: true,
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 2}),
			},
			OracleText: `
			Kicker {3}{B} (You may pay an additional {3}{B} as you cast this spell.)
			Flying, lifelink
			If this creature was kicked, it enters with two +1/+1 counters on it.
		`,
		},
	}
}
