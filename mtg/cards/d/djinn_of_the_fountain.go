package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DjinnOfTheFountain is the card definition for Djinn of the Fountain.
//
// Type: Creature — Djinn
// Cost: {4}{U}{U}
//
// Oracle text:
//
//	Flying
//	Whenever you cast an instant or sorcery spell, choose one —
//	• This creature gets +1/+1 until end of turn.
//	• Exile this creature. Return it to the battlefield under its owner's control at the beginning of the next end step.
//	• Scry 1.
var DjinnOfTheFountain = newDjinnOfTheFountain

func newDjinnOfTheFountain() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Djinn of the Fountain",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Djinn},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:         game.EventSpellCast,
							Controller:    game.TriggerControllerYou,
							CardSelection: game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}},
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "This creature gets +1/+1 until end of turn.",
								Sequence: []game.Instruction{
									{
										Primitive: game.ModifyPT{
											Object:         game.SourcePermanentReference(),
											PowerDelta:     game.Fixed(1),
											ToughnessDelta: game.Fixed(1),
											Duration:       game.DurationUntilEndOfTurn,
										},
									},
								},
							},
							game.Mode{
								Text: "Exile this creature. Return it to the battlefield under its owner's control at the beginning of the next end step.",
								Sequence: []game.Instruction{
									{
										Primitive: game.Exile{
											Object:         game.SourcePermanentReference(),
											ExileLinkedKey: game.LinkedKey("delayed-self-blink"),
										},
									},
									{
										Primitive: game.CreateDelayedTrigger{
											Trigger: game.DelayedTriggerDef{
												Timing: game.DelayedAtBeginningOfNextEndStep,
												Content: game.Mode{
													Sequence: []game.Instruction{
														{
															Primitive: game.PutOnBattlefield{
																Source: game.LinkedBattlefieldSource(game.LinkedKey("delayed-self-blink")),
															},
														},
													},
												}.Ability(),
											},
										},
									},
								},
							},
							game.Mode{
								Text: "Scry 1.",
								Sequence: []game.Instruction{
									{
										Primitive: game.Scry{
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
			OracleText: `
			Flying
			Whenever you cast an instant or sorcery spell, choose one —
			• This creature gets +1/+1 until end of turn.
			• Exile this creature. Return it to the battlefield under its owner's control at the beginning of the next end step.
			• Scry 1.
		`,
		},
	}
}
