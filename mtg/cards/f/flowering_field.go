package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FloweringField is the card definition for Flowering Field.
//
// Type: Enchantment — Aura
// Cost: {1}{W}
//
// Oracle text:
//
//	Enchant land
//	Enchanted land has "{T}: Prevent the next 1 damage that would be dealt to any target this turn."
var FloweringField = newFloweringField()

func newFloweringField() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Flowering Field",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
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
									Text:            "{T}: Prevent the next 1 damage that would be dealt to any target this turn.",
									AdditionalCosts: cost.Tap,
									ZoneOfFunction:  zone.Battlefield,
									Content: game.Mode{
										Targets: []game.TargetSpec{
											game.TargetSpec{
												MinTargets: 1,
												MaxTargets: 1,
												Constraint: "any target",
												Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
											},
										},
										Sequence: []game.Instruction{
											{
												Primitive: game.PreventDamage{
													AnyTarget: game.AnyTargetDamageRecipient(0),
													Amount:    game.Fixed(1),
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
			Enchanted land has "{T}: Prevent the next 1 damage that would be dealt to any target this turn."
		`,
		},
	}
}
