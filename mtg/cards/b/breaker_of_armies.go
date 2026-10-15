package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BreakerOfArmies is the card definition for Breaker of Armies.
//
// Type: Creature — Eldrazi
// Cost: {8}
//
// Oracle text:
//
//	All creatures able to block this creature do so.
var BreakerOfArmies = newBreakerOfArmies

func newBreakerOfArmies() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Breaker of Armies",
			ManaCost: opt.Val(cost.Mana{
				cost.O(8),
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi},
			Power:     opt.Val(game.PT{Value: 10}),
			Toughness: opt.Val(game.PT{Value: 8}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "All creatures able to block this creature do so.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectMustBeBlockedByAllAble,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
						},
					},
				},
			},
			OracleText: `
			All creatures able to block this creature do so.
		`,
		},
	}
}
