package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SyndicateInfiltrator is the card definition for Syndicate Infiltrator.
//
// Type: Creature — Vampire Wizard
// Cost: {2}{U}{B}
//
// Oracle text:
//
//	Flying
//	As long as there are five or more mana values among cards in your graveyard, this creature gets +2/+2.
var SyndicateInfiltrator = newSyndicateInfiltrator

func newSyndicateInfiltrator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Syndicate Infiltrator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.B,
			}),
			Colors:    []color.Color{color.Black, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire, types.Wizard},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerGraveyardManaValueCount, Op: compare.GreaterOrEqual, Value: 5}},
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							AffectedSource: true,
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
					},
				},
			},
			OracleText: `
			Flying
			As long as there are five or more mana values among cards in your graveyard, this creature gets +2/+2.
		`,
		},
	}
}
