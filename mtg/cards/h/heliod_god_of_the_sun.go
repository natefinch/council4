package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// HeliodGodOfTheSun is the card definition for Heliod, God of the Sun.
//
// Type: Legendary Enchantment Creature — God
// Cost: {3}{W}
//
// Oracle text:
//
//	Indestructible
//	As long as your devotion to white is less than five, Heliod isn't a creature. (Each {W} in the mana costs of permanents you control counts toward your devotion to white.)
//	Other creatures you control have vigilance.
//	{2}{W}{W}: Create a 2/1 white Cleric enchantment creature token.
var HeliodGodOfTheSun = newHeliodGodOfTheSun

func newHeliodGodOfTheSun() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Heliod, God of the Sun",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Enchantment, types.Creature},
			Subtypes:   []types.Sub{types.God},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.IndestructibleStaticBody,
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerDevotion, Op: compare.LessThan, Value: 5, Colors: []color.Color{color.White}}},
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
								game.Vigilance,
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{2}{W}{W}: Create a 2/1 white Cleric enchantment creature token.",
					ManaCost:       opt.Val(cost.Mana{cost.O(2), cost.W, cost.W}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(heliodGodOfTheSunToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Indestructible
			As long as your devotion to white is less than five, Heliod isn't a creature. (Each {W} in the mana costs of permanents you control counts toward your devotion to white.)
			Other creatures you control have vigilance.
			{2}{W}{W}: Create a 2/1 white Cleric enchantment creature token.
		`,
		},
	}
}

var heliodGodOfTheSunToken = newHeliodGodOfTheSunToken()

func newHeliodGodOfTheSunToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Cleric",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Cleric},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
