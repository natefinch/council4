package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CaribouRange is the card definition for Caribou Range.
//
// Type: Enchantment — Aura
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	Enchant land you control
//	Enchanted land has "{W}{W}, {T}: Create a 0/1 white Caribou creature token."
//	Sacrifice a Caribou token: You gain 1 life.
var CaribouRange = newCaribouRange

func newCaribouRange() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Caribou Range",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "land you control",
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypesAny: []types.Card{types.Land},
						Controller:       game.ControllerYou,
					}),
				}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddAbilities: []game.Ability{
								new(game.ActivatedAbility{
									Text:            "{W}{W}, {T}: Create a 0/1 white Caribou creature token.",
									ManaCost:        opt.Val(cost.Mana{cost.W, cost.W}),
									AdditionalCosts: cost.Tap,
									ZoneOfFunction:  zone.Battlefield,
									Content: game.Mode{
										Sequence: []game.Instruction{
											{
												Primitive: game.CreateToken{
													Amount: game.Fixed(1),
													Source: game.TokenDef(caribouRangeToken),
												},
											},
										},
									}.Ability(),
								}),
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Sacrifice a Caribou token: You gain 1 life.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalSacrifice,
							Text:        "Sacrifice a Caribou token",
							Amount:      1,
							SubtypesAny: cost.SubtypeSet{types.Caribou},
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Enchant land you control
			Enchanted land has "{W}{W}, {T}: Create a 0/1 white Caribou creature token."
			Sacrifice a Caribou token: You gain 1 life.
		`,
		},
	}
}

var caribouRangeToken = newCaribouRangeToken()

func newCaribouRangeToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Caribou",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Caribou},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
