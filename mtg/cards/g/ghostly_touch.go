package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GhostlyTouch is the card definition for Ghostly Touch.
//
// Type: Enchantment — Aura
// Cost: {1}{U}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature has "Whenever this creature attacks, you may tap or untap target permanent."
var GhostlyTouch = newGhostlyTouch

func newGhostlyTouch() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Ghostly Touch",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
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
					Text: "Enchanted creature has \"Whenever this creature attacks, you may tap or untap target permanent.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.TriggeredAbility{
									Text: "Whenever this creature attacks, you may tap or untap target permanent.",
									Trigger: game.TriggerCondition{
										Type: game.TriggerWhenever,
										Pattern: game.TriggerPattern{
											Event:           game.EventAttackerDeclared,
											Source:          game.TriggerSourceSelf,
											StackObjectKind: game.StackSpell,
											CounterKind:     counter.PlusOnePlusOne,
										},
									},
									Optional: true,
									Content: game.AbilityContent{
										Modes: []game.Mode{
											game.Mode{
												Targets: []game.TargetSpec{
													game.TargetSpec{
														MinTargets: 1,
														MaxTargets: 1,
														Constraint: "target permanent",
														Allow:      game.TargetAllowPermanent,
													},
												},
												Sequence: []game.Instruction{
													game.Instruction{
														Primitive: game.TapOrUntap{
															Object: game.TargetPermanentReference(0),
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
			},
			OracleText: `
			Enchant creature
			Enchanted creature has "Whenever this creature attacks, you may tap or untap target permanent."
		`,
		},
	}
}
