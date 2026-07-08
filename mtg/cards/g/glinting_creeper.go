package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GlintingCreeper is the card definition for Glinting Creeper.
//
// Type: Creature — Plant
// Cost: {4}{G}
//
// Oracle text:
//
//	Converge — This creature enters with two +1/+1 counters on it for each color of mana spent to cast it.
//	This creature can't be blocked by creatures with power 2 or less.
var GlintingCreeper = newGlintingCreeper

func newGlintingCreeper() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Glinting Creeper",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Plant},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBeBlockedByCreaturesWith,
							AffectedSource: true,
							BlockerRestriction: game.BlockerRestriction{
								Kind:  game.BlockerRestrictionPowerLessOrEqual,
								Power: 2,
							},
						},
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("Converge — This creature enters with two +1/+1 counters on it for each color of mana spent to cast it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Dynamic: opt.Val(&game.DynamicAmount{
					Kind:       game.DynamicAmountColorsOfManaSpentToCast,
					Multiplier: 2,
				})}),
			},
			OracleText: `
			Converge — This creature enters with two +1/+1 counters on it for each color of mana spent to cast it.
			This creature can't be blocked by creatures with power 2 or less.
		`,
		},
	}
}
