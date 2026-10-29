package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Retaliation is the card definition for Retaliation.
//
// Type: Enchantment
// Cost: {2}{G}
//
// Oracle text:
//
//	Creatures you control have "Whenever this creature becomes blocked by a creature, this creature gets +1/+1 until end of turn."
var Retaliation = newRetaliation

func newRetaliation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Retaliation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures you control have \"Whenever this creature becomes blocked by a creature, this creature gets +1/+1 until end of turn.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.TriggeredAbility{
									Text: "Whenever this creature becomes blocked by a creature, this creature gets +1/+1 until end of turn.",
									Trigger: game.TriggerCondition{
										Type: game.TriggerWhenever,
										Pattern: game.TriggerPattern{
											Event:           game.EventAttackerBecameBlocked,
											BlockedScope:    game.BlockedTriggerEachBlocker,
											Source:          game.TriggerSourceSelf,
											StackObjectKind: game.StackSpell,
											CounterKind:     counter.PlusOnePlusOne,
										},
									},
									Content: game.AbilityContent{
										Modes: []game.Mode{
											game.Mode{
												Sequence: []game.Instruction{
													game.Instruction{
														Primitive: game.ModifyPT{
															Object:         game.SourcePermanentReference(),
															PowerDelta:     game.Fixed(1),
															ToughnessDelta: game.Fixed(1),
															Duration:       game.DurationUntilEndOfTurn,
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
			Creatures you control have "Whenever this creature becomes blocked by a creature, this creature gets +1/+1 until end of turn."
		`,
		},
	}
}
