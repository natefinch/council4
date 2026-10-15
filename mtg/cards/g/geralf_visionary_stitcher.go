package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GeralfVisionaryStitcher is the card definition for Geralf, Visionary Stitcher.
//
// Type: Legendary Creature — Human Wizard
// Cost: {2}{U}
//
// Oracle text:
//
//	Zombies you control have flying.
//	{U}, {T}, Sacrifice another nontoken creature: Create an X/X blue Zombie creature token, where X is the sacrificed creature's toughness.
var GeralfVisionaryStitcher = newGeralfVisionaryStitcher

func newGeralfVisionaryStitcher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Geralf, Visionary Stitcher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Wizard},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Zombies you control have flying.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Zombie")}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Flying},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{U}, {T}, Sacrifice another nontoken creature: Create an X/X blue Zombie creature token, where X is the sacrificed creature's toughness.",
					ManaCost: opt.Val(cost.Mana{
						cost.U,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice another nontoken creature",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Creature,
							CounterKind:        counter.PlusOnePlusOne,
							RequireNonToken:    true,
							ExcludeSource:      true,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(geralfVisionaryStitcherToken),
											Power: opt.Val(game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountObjectToughness,
												Multiplier: 1,
												Object:     game.SacrificedCostReference(),
											})),
											Toughness: opt.Val(game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountObjectToughness,
												Multiplier: 1,
												Object:     game.SacrificedCostReference(),
											})),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Zombies you control have flying.
			{U}, {T}, Sacrifice another nontoken creature: Create an X/X blue Zombie creature token, where X is the sacrificed creature's toughness.
		`,
		},
	}
}

var geralfVisionaryStitcherToken = newGeralfVisionaryStitcherToken()

func newGeralfVisionaryStitcherToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Zombie",
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Creature},
			Subtypes: []types.Sub{types.Zombie},
		},
	}
}
