package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// HotSprings is the card definition for Hot Springs.
//
// Type: Enchantment — Aura
// Cost: {1}{G}
//
// Oracle text:
//
//	Enchant land you control
//	Enchanted land has "{T}: Prevent the next 1 damage that would be dealt to any target this turn."
var HotSprings = newHotSprings

func newHotSprings() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Hot Springs",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "land you control",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}, Controller: game.ControllerYou}),
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
			Enchant land you control
			Enchanted land has "{T}: Prevent the next 1 damage that would be dealt to any target this turn."
		`,
		},
	}
}
