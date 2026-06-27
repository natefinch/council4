package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SecurityBlockade is the card definition for Security Blockade.
//
// Type: Enchantment — Aura
// Cost: {2}{W}
//
// Oracle text:
//
//	Enchant land
//	When this Aura enters, create a 2/2 white Knight creature token with vigilance.
//	Enchanted land has "{T}: Prevent the next 1 damage that would be dealt to you this turn."
var SecurityBlockade = newSecurityBlockade()

func newSecurityBlockade() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Security Blockade",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
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
									Text:            "{T}: Prevent the next 1 damage that would be dealt to you this turn.",
									AdditionalCosts: cost.Tap,
									ZoneOfFunction:  zone.Battlefield,
									Content: game.Mode{
										Sequence: []game.Instruction{
											{
												Primitive: game.PreventDamage{
													Player: game.ControllerReference(),
													Amount: game.Fixed(1),
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
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(securityBlockadeToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Enchant land
			When this Aura enters, create a 2/2 white Knight creature token with vigilance.
			Enchanted land has "{T}: Prevent the next 1 damage that would be dealt to you this turn."
		`,
		},
	}
}

var securityBlockadeToken = newSecurityBlockadeToken()

func newSecurityBlockadeToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Knight",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
		},
	}
}
