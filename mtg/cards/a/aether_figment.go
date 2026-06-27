package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AetherFigment is the card definition for Aether Figment.
//
// Type: Creature — Illusion
// Cost: {1}{U}
//
// Oracle text:
//
//	Kicker {3} (You may pay an additional {3} as you cast this spell.)
//	If this creature was kicked, it enters with two +1/+1 counters on it.
//	This creature can't be blocked.
var AetherFigment = newAetherFigment()

func newAetherFigment() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Aether Figment",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Illusion},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{Cost: cost.Mana{cost.O(3)}},
					},
				},
				game.CantBeBlockedStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("If this creature was kicked, it enters with two +1/+1 counters on it.", &game.Condition{
					EventPermanentWasKicked: true,
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 2}),
			},
			OracleText: `
			Kicker {3} (You may pay an additional {3} as you cast this spell.)
			If this creature was kicked, it enters with two +1/+1 counters on it.
			This creature can't be blocked.
		`,
		},
	}
}
