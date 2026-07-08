package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StrongholdConfessor is the card definition for Stronghold Confessor.
//
// Type: Creature — Human Cleric
// Cost: {B}
//
// Oracle text:
//
//	Kicker {3} (You may pay an additional {3} as you cast this spell.)
//	Menace (This creature can't be blocked except by two or more creatures.)
//	If this creature was kicked, it enters with two +1/+1 counters on it.
var StrongholdConfessor = newStrongholdConfessor

func newStrongholdConfessor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Stronghold Confessor",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Cleric},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{Cost: cost.Mana{cost.O(3)}},
					},
				},
				game.MenaceStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("If this creature was kicked, it enters with two +1/+1 counters on it.", &game.Condition{
					EventPermanentWasKicked: true,
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 2}),
			},
			OracleText: `
			Kicker {3} (You may pay an additional {3} as you cast this spell.)
			Menace (This creature can't be blocked except by two or more creatures.)
			If this creature was kicked, it enters with two +1/+1 counters on it.
		`,
		},
	}
}
