package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TarmogoyfNest is the card definition for Tarmogoyf Nest.
//
// Type: Kindred Enchantment — Lhurgoyf Aura
// Cost: {2}{G}
//
// Oracle text:
//
//	Enchant land
//	Enchanted land has "{1}{G}, {T}: Create a Tarmogoyf token." (It's a {1}{G} Lhurgoyf creature with "Tarmogoyf's power is equal to the number of card types among cards in all graveyards and its toughness is equal to that number plus 1.")
var TarmogoyfNest = newTarmogoyfNest

func newTarmogoyfNest() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Tarmogoyf Nest",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Kindred, types.Enchantment},
			Subtypes: []types.Sub{types.Lhurgoyf, types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "land",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}}),
				}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddAbilities: []game.Ability{
								new(game.ActivatedAbility{
									Text:            "{1}{G}, {T}: Create a Tarmogoyf token.",
									ManaCost:        opt.Val(cost.Mana{cost.O(1), cost.G}),
									AdditionalCosts: cost.Tap,
									ZoneOfFunction:  zone.Battlefield,
									Content: game.Mode{
										Sequence: []game.Instruction{
											{
												Primitive: game.CreateToken{
													Amount: game.Fixed(1),
													Source: game.TokenDef(tarmogoyfNestToken),
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
			OracleText: `
			Enchant land
			Enchanted land has "{1}{G}, {T}: Create a Tarmogoyf token." (It's a {1}{G} Lhurgoyf creature with "Tarmogoyf's power is equal to the number of card types among cards in all graveyards and its toughness is equal to that number plus 1.")
		`,
		},
	}
}

var tarmogoyfNestToken = newTarmogoyfNestToken()

func newTarmogoyfNestToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Tarmogoyf",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:           []color.Color{color.Green},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Lhurgoyf},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueCardTypesAmongAllGraveyards}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueCardTypesAmongAllGraveyards, Offset: 1}),
		},
	}
}
