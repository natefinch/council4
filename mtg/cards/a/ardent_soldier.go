package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ArdentSoldier is the card definition for Ardent Soldier.
//
// Type: Creature — Human Soldier
// Cost: {1}{W}
//
// Oracle text:
//
//	Kicker {2} (You may pay an additional {2} as you cast this spell.)
//	Vigilance
//	If this creature was kicked, it enters with a +1/+1 counter on it.
var ArdentSoldier = newArdentSoldier

func newArdentSoldier() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Ardent Soldier",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{Cost: cost.Mana{cost.O(2)}},
					},
				},
				game.VigilanceStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("If this creature was kicked, it enters with a +1/+1 counter on it.", &game.Condition{
					EventPermanentWasKicked: true,
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
			},
			OracleText: `
			Kicker {2} (You may pay an additional {2} as you cast this spell.)
			Vigilance
			If this creature was kicked, it enters with a +1/+1 counter on it.
		`,
		},
	}
}
