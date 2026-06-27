package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KavuPrimarch is the card definition for Kavu Primarch.
//
// Type: Creature — Kavu
// Cost: {3}{G}
//
// Oracle text:
//
//	Kicker {4} (You may pay an additional {4} as you cast this spell.)
//	Convoke (Your creatures can help cast this spell. Each creature you tap while casting this spell pays for {1} or one mana of that creature's color.)
//	If this creature was kicked, it enters with four +1/+1 counters on it.
var KavuPrimarch = newKavuPrimarch()

func newKavuPrimarch() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Kavu Primarch",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Kavu},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{Cost: cost.Mana{cost.O(4)}},
					},
				},
				game.ConvokeStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("If this creature was kicked, it enters with four +1/+1 counters on it.", &game.Condition{
					EventPermanentWasKicked: true,
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 4}),
			},
			OracleText: `
			Kicker {4} (You may pay an additional {4} as you cast this spell.)
			Convoke (Your creatures can help cast this spell. Each creature you tap while casting this spell pays for {1} or one mana of that creature's color.)
			If this creature was kicked, it enters with four +1/+1 counters on it.
		`,
		},
	}
}
