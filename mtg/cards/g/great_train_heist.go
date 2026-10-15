package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GreatTrainHeist is the card definition for Great Train Heist.
//
// Type: Instant
// Cost: {R}
//
// Oracle text:
//
//	Spree (Choose one or more additional costs.)
//	+ {2}{R} — Untap all creatures you control. If it's your combat phase, there is an additional combat phase after this phase.
//	+ {2} — Creatures you control get +1/+0 and gain first strike until end of turn.
//	+ {R} — Choose target opponent. Whenever a creature you control deals combat damage to that player this turn, create a tapped Treasure token.
var GreatTrainHeist = newGreatTrainHeist

func newGreatTrainHeist() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Great Train Heist",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "{2}{R} — Untap all creatures you control. If it's your combat phase, there is an additional combat phase after this phase.",
						Cost: opt.Val(cost.Mana{cost.O(2), cost.R}),
						Sequence: []game.Instruction{
							{
								Primitive: game.Untap{
									Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
								},
							},
							{
								Primitive: game.AddExtraPhases{
									Combat: true,
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										ControllerCombatPhase: true,
									}),
								}),
							},
						},
					},
					game.Mode{
						Text: "{2} — Creatures you control get +1/+0 and gain first strike until end of turn.",
						Cost: opt.Val(cost.Mana{cost.O(2)}),
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Duration:   game.DurationPermanent,
											Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											Layer:      game.LayerPowerToughnessModify,
											PowerDelta: 1,
										},
										game.ContinuousEffect{
											Duration:    game.DurationPermanent,
											Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											Layer:       game.LayerAbility,
											AddKeywords: []game.Keyword{game.FirstStrike},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					},
					game.Mode{
						Text: "{R} — Choose target opponent. Whenever a creature you control deals combat damage to that player this turn, create a tapped Treasure token.",
						Cost: opt.Val(cost.Mana{cost.R}),
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target opponent",
								Allow:      game.TargetAllowPlayer,
								Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateDelayedTrigger{
									Trigger: game.DelayedTriggerDef{
										Content: game.AbilityContent{
											Modes: []game.Mode{
												game.Mode{
													Sequence: []game.Instruction{
														game.Instruction{
															Primitive: game.CreateToken{
																Amount:      game.Fixed(1),
																Source:      game.TokenDef(greatTrainHeistToken),
																EntryTapped: true,
															},
														},
													},
												},
											},
											MinModes: 1,
											MaxModes: 1,
										},
										EventPattern: opt.Val(game.TriggerPattern{
											Event:                 game.EventDamageDealt,
											Controller:            game.TriggerControllerYou,
											Subject:               game.TriggerSubjectDamageSource,
											StackObjectKind:       game.StackSpell,
											DamageRecipient:       game.DamageRecipientPlayer,
											DamageSourceSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
											RequireCombatDamage:   true,
											CounterKind:           counter.PlusOnePlusOne,
										}),
										Window:      game.DelayedWindowThisTurn,
										EventPlayer: opt.Val(game.TargetPlayerReference(0)),
									},
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 3,
			}),
			OracleText: `
			Spree (Choose one or more additional costs.)
			+ {2}{R} — Untap all creatures you control. If it's your combat phase, there is an additional combat phase after this phase.
			+ {2} — Creatures you control get +1/+0 and gain first strike until end of turn.
			+ {R} — Choose target opponent. Whenever a creature you control deals combat damage to that player this turn, create a tapped Treasure token.
		`,
		},
	}
}

var greatTrainHeistToken = newGreatTrainHeistToken()

func newGreatTrainHeistToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Treasure",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Treasure},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "{T}, Sacrifice this artifact: Add one mana of any color.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:               cost.AdditionalSacrificeSource,
							Text:               "Sacrifice this artifact",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Artifact,
							CounterKind:        counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Choose{
											Choice: game.ResolutionChoice{
												Kind:   game.ResolutionChoiceMana,
												Prompt: "Choose a color",
												Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
											},
											PublishChoice: game.ChoiceKey("oracle-mana-color"),
										},
									},
									game.Instruction{
										Primitive: game.AddMana{
											Amount:     game.Fixed(1),
											ChoiceFrom: game.ChoiceKey("oracle-mana-color"),
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
	}
}
