package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SnakeUmbra is the card definition for Snake Umbra.
//
// Type: Enchantment — Aura
// Cost: {2}{G}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature gets +1/+1 and has "Whenever this creature deals damage to an opponent, you may draw a card."
//	Umbra armor (If enchanted creature would be destroyed, instead remove all damage from it and destroy this Aura.)
var SnakeUmbra = newSnakeUmbra

func newSnakeUmbra() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Snake Umbra",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					Text: "Enchanted creature gets +1/+1 and has \"Whenever this creature deals damage to an opponent, you may draw a card.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.TriggeredAbility{
									Text: "Whenever this creature deals damage to an opponent, you may draw a card.",
									Trigger: game.TriggerCondition{
										Type: game.TriggerWhenever,
										Pattern: game.TriggerPattern{
											Event:           game.EventDamageDealt,
											Source:          game.TriggerSourceSelf,
											Player:          game.TriggerPlayerOpponent,
											Subject:         game.TriggerSubjectDamageSource,
											StackObjectKind: game.StackSpell,
											DamageRecipient: game.DamageRecipientPlayer,
											CounterKind:     counter.PlusOnePlusOne,
										},
									},
									Optional: true,
									Content: game.AbilityContent{
										Modes: []game.Mode{
											game.Mode{
												Sequence: []game.Instruction{
													game.Instruction{
														Primitive: game.Draw{
															Amount: game.Fixed(1),
															Player: game.ControllerReference(),
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
						},
					},
				},
				game.UmbraArmorStaticBody,
			},
			OracleText: `
			Enchant creature
			Enchanted creature gets +1/+1 and has "Whenever this creature deals damage to an opponent, you may draw a card."
			Umbra armor (If enchanted creature would be destroyed, instead remove all damage from it and destroy this Aura.)
		`,
		},
	}
}
