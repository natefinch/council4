package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UntamedKavu is the card definition for Untamed Kavu.
//
// Type: Creature — Kavu
// Cost: {1}{G}
//
// Oracle text:
//
//	Kicker {3} (You may pay an additional {3} as you cast this spell.)
//	Vigilance, trample
//	If this creature was kicked, it enters with three +1/+1 counters on it.
var UntamedKavu = newUntamedKavu()

func newUntamedKavu() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Untamed Kavu",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Kavu},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{Cost: cost.Mana{cost.O(3)}},
					},
				},
				game.VigilanceStaticBody,
				game.TrampleStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("If this creature was kicked, it enters with three +1/+1 counters on it.", &game.Condition{
					EventPermanentWasKicked: true,
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 3}),
			},
			OracleText: `
			Kicker {3} (You may pay an additional {3} as you cast this spell.)
			Vigilance, trample
			If this creature was kicked, it enters with three +1/+1 counters on it.
		`,
		},
	}
}
