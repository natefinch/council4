package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CyclonusTheSaboteur is the card definition for Cyclonus, the Saboteur // Cyclonus, Cybertronian Fighter.
//
// Type: Legendary Artifact Creature — Robot // Legendary Artifact — Vehicle
// Face: Cyclonus, Cybertronian Fighter — Legendary Artifact — Vehicle
//
// Oracle text:
//
//	More Than Meets the Eye {5}{U}{B} (You may cast this card converted for {5}{U}{B}.)
//	Flying
//	Whenever Cyclonus deals combat damage to a player, it connives. Then if Cyclonus's power is 5 or greater, convert it. (To have a creature connive, draw a card, then discard a card. If you discarded a nonland card, put a +1/+1 counter on that creature.)
var CyclonusTheSaboteur = newCyclonusTheSaboteur()

func newCyclonusTheSaboteur() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Cyclonus, the Saboteur",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.B,
			}),
			Colors:     []color.Color{color.Black, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact, types.Creature},
			Subtypes:   []types.Sub{types.Robot},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:               game.EventDamageDealt,
							Source:              game.TriggerSourceSelf,
							Subject:             game.TriggerSubjectDamageSource,
							RequireCombatDamage: true,
							DamageRecipient:     game.DamageRecipientPlayer,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Connive{
									Object: game.EventPermanentReference(),
									Player: game.ObjectControllerReference(game.EventPermanentReference()),
									Amount: game.Fixed(1),
								},
							},
							{
								Primitive: game.Transform{
									Object: game.SourcePermanentReference(),
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										Object:        opt.Val(game.SourcePermanentReference()),
										ObjectMatches: opt.Val(game.Selection{Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 5})}),
									}),
								}),
							},
						},
					}.Ability(),
				},
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "More Than Meets the Eye",
					ManaCost: opt.Val(cost.Mana{cost.O(5), cost.U, cost.B}),
					Mechanic: cost.AlternativeMechanicMoreThanMeetsTheEye,
				},
			},
			OracleText: `
			More Than Meets the Eye {5}{U}{B} (You may cast this card converted for {5}{U}{B}.)
			Flying
			Whenever Cyclonus deals combat damage to a player, it connives. Then if Cyclonus's power is 5 or greater, convert it. (To have a creature connive, draw a card, then discard a card. If you discarded a nonland card, put a +1/+1 counter on that creature.)
		`,
		},
		Layout: game.LayoutTransform,
		Back: opt.Val(game.CardFace{
			Name:       "Cyclonus, Cybertronian Fighter",
			Colors:     []color.Color{color.Black, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			Subtypes:   []types.Sub{types.Vehicle},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.LivingMetalStaticBody,
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:               game.EventDamageDealt,
							Source:              game.TriggerSourceSelf,
							Subject:             game.TriggerSubjectDamageSource,
							RequireCombatDamage: true,
							DamageRecipient:     game.DamageRecipientPlayer,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Transform{
									Object: game.EventPermanentReference(),
								},
								PublishResult: game.ResultKey("if-you-do"),
							},
							{
								Primitive: game.AddExtraPhases{
									Beginning: true,
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "if-you-do",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Living metal (During your turn, this Vehicle is also a creature.)
			Flying
			Whenever Cyclonus deals combat damage to a player, convert it. If you do, there is an additional beginning phase after this phase. (The beginning phase includes the untap, upkeep, and draw steps.)
		`,
		}),
	}
}
