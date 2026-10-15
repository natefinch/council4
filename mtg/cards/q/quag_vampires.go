package q

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// QuagVampires is the card definition for Quag Vampires.
//
// Type: Creature — Vampire Rogue
// Cost: {B}
//
// Oracle text:
//
//	Multikicker {1}{B} (You may pay an additional {1}{B} any number of times as you cast this spell.)
//	Swampwalk (This creature can't be blocked as long as defending player controls a Swamp.)
//	This creature enters with a +1/+1 counter on it for each time it was kicked.
var QuagVampires = newQuagVampires

func newQuagVampires() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Quag Vampires",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire, types.Rogue},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Multikicker {1}{B}",
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{
							Cost: cost.Mana{
								cost.O(1),
								cost.B,
							},
							Multi: true,
						},
					},
				},
				game.SwampwalkStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with a +1/+1 counter on it for each time it was kicked.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Dynamic: opt.Val(&game.DynamicAmount{
					Kind:       game.DynamicAmountTimesKicked,
					Multiplier: 1,
				})}),
			},
			OracleText: `
			Multikicker {1}{B} (You may pay an additional {1}{B} any number of times as you cast this spell.)
			Swampwalk (This creature can't be blocked as long as defending player controls a Swamp.)
			This creature enters with a +1/+1 counter on it for each time it was kicked.
		`,
		},
	}
}
