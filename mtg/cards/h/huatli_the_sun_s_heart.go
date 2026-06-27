package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HuatliTheSunSHeart is the card definition for Huatli, the Sun's Heart.
//
// Type: Legendary Planeswalker — Huatli
// Cost: {2}{G/W}
//
// Oracle text:
//
//	Each creature you control assigns combat damage equal to its toughness rather than its power.
//	−3: You gain life equal to the greatest toughness among creatures you control.
var HuatliTheSunSHeart = newHuatliTheSunSHeart()

func newHuatliTheSunSHeart() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Huatli, the Sun's Heart",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.HybridMana(mana.G, mana.W),
			}),
			Colors:     []color.Color{color.Green, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Huatli},
			Loyalty:    opt.Val(7),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:               game.RuleEffectAssignCombatDamageUsingToughness,
							AffectedController: game.ControllerYou,
							PermanentTypes:     []types.Card{types.Creature},
						},
					},
				},
			},
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					LoyaltyCost: -3,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.GainLife{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountGreatestToughnessInGroup,
										Multiplier: 1,
										Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									}),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Each creature you control assigns combat damage equal to its toughness rather than its power.
			−3: You gain life equal to the greatest toughness among creatures you control.
		`,
		},
	}
}
