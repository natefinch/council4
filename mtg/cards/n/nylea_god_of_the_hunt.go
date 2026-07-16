package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// NyleaGodOfTheHunt is the card definition for Nylea, God of the Hunt.
//
// Type: Legendary Enchantment Creature — God
// Cost: {3}{G}
//
// Oracle text:
//
//	Indestructible
//	As long as your devotion to green is less than five, Nylea isn't a creature. (Each {G} in the mana costs of permanents you control counts toward your devotion to green.)
//	Other creatures you control have trample.
//	{3}{G}: Target creature gets +2/+2 until end of turn.
var NyleaGodOfTheHunt = newNyleaGodOfTheHunt

func newNyleaGodOfTheHunt() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Nylea, God of the Hunt",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Enchantment, types.Creature},
			Subtypes:   []types.Sub{types.God},
			Power:      opt.Val(game.PT{Value: 6}),
			Toughness:  opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.IndestructibleStaticBody,
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerDevotion, Op: compare.LessThan, Value: 5, Colors: []color.Color{color.Green}}},
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerType,
							AffectedSource: true,
							RemoveTypes:    []types.Card{types.Creature},
						},
					},
				},
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}, game.SourcePermanentReference()),
							AddKeywords: []game.Keyword{
								game.Trample,
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{3}{G}: Target creature gets +2/+2 until end of turn.",
					ManaCost:       opt.Val(cost.Mana{cost.O(3), cost.G}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object:         game.TargetPermanentReference(0),
									PowerDelta:     game.Fixed(2),
									ToughnessDelta: game.Fixed(2),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Indestructible
			As long as your devotion to green is less than five, Nylea isn't a creature. (Each {G} in the mana costs of permanents you control counts toward your devotion to green.)
			Other creatures you control have trample.
			{3}{G}: Target creature gets +2/+2 until end of turn.
		`,
		},
	}
}
