package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NightwhorlHermit is the card definition for Nightwhorl Hermit.
//
// Type: Creature — Rat Rogue
// Cost: {2}{U}
//
// Oracle text:
//
//	Vigilance
//	Threshold — As long as there are seven or more cards in your graveyard, this creature gets +1/+0 and can't be blocked.
var NightwhorlHermit = newNightwhorlHermit()

func newNightwhorlHermit() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Nightwhorl Hermit",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Rat, types.Rogue},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerGraveyardCardCount, Op: compare.GreaterOrEqual, Value: 7}},
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							AffectedSource: true,
							PowerDelta:     1,
						},
					},
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBeBlocked,
							AffectedSource: true,
						},
					},
				},
			},
			OracleText: `
			Vigilance
			Threshold — As long as there are seven or more cards in your graveyard, this creature gets +1/+0 and can't be blocked.
		`,
		},
	}
}
