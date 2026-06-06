package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NeyithOfTheDireHunt is the card definition for Neyith of the Dire Hunt.
//
// Type: Legendary Creature — Human Warrior
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	Whenever one or more creatures you control fight or become blocked, draw a card.
//	At the beginning of combat on your turn, you may pay {2}{R/G}. If you do, double target creature's power until end of turn. That creature must be blocked this combat if able. ({R/G} can be paid with either {R} or {G}.)
var NeyithOfTheDireHunt = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green, color.Red),
	CardFace: game.CardFace{
		Name: "Neyith of the Dire Hunt",
		ManaCost: opt.Val(cost.Mana{
			cost.O(2),
			cost.G,
			cost.G,
		}),
		Colors:     []color.Color{color.Green},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Human, types.Warrior},
		Power:      opt.Val(game.PT{Value: 3}),
		Toughness:  opt.Val(game.PT{Value: 3}),
		OracleText: `
			Whenever one or more creatures you control fight or become blocked, draw a card.
			At the beginning of combat on your turn, you may pay {2}{R/G}. If you do, double target creature's power until end of turn. That creature must be blocked this combat if able. ({R/G} can be paid with either {R} or {G}.)
		`,
		TriggeredAbilities: []game.TriggeredAbilityBody{
			{
				Text: `
					Whenever one or more creatures you control fight or become blocked, draw a card.
				`,
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhenever,
					Pattern: game.TriggerPattern{
						Event:      game.EventFight,
						Controller: game.TriggerControllerYou,
						RequirePermanentTypes: []types.Card{
							types.Creature,
						},
						OneOrMore: true,
					},
				},
				Content: game.PlainAbilityContent{
					Sequence: []game.Instruction{
						{
							Primitive: game.Draw{
								Amount:      game.Fixed(1),
								TargetIndex: game.TargetIndexController,
							},
						},
					},
				},
			},
			{
				Text: `
					Whenever one or more creatures you control fight or become blocked, draw a card.
				`,
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhenever,
					Pattern: game.TriggerPattern{
						Event:      game.EventBlockerDeclared,
						Subject:    game.TriggerSubjectBlockedAttacker,
						Controller: game.TriggerControllerYou,
						RequirePermanentTypes: []types.Card{
							types.Creature,
						},
						OneOrMore: true,
					},
				},
				Content: game.PlainAbilityContent{
					Sequence: []game.Instruction{
						{
							Primitive: game.Draw{
								Amount:      game.Fixed(1),
								TargetIndex: game.TargetIndexController,
							},
						},
					},
				},
			},
			{
				Text: `
					At the beginning of combat on your turn, you may pay {2}{R/G}. If you do, double target creature's power until end of turn. That creature must be blocked this combat if able.
				`,
				Trigger: game.TriggerCondition{
					Type: game.TriggerAt,
					Pattern: game.TriggerPattern{
						Event:      game.EventBeginningOfStep,
						Controller: game.TriggerControllerYou,
						Step:       game.StepBeginningOfCombat,
					},
				},
				Content: game.PlainAbilityContent{
					Targets: []game.TargetSpec{
						{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "creature",
							Allow:      game.TargetAllowPermanent,
							Predicate: game.TargetPredicate{
								PermanentTypes: []types.Card{
									types.Creature,
								},
							},
						},
					},
					Sequence: []game.Instruction{
						{
							Primitive: game.Pay{
								Payment: game.ResolutionPayment{
									Prompt: "Pay {2}{R/G}?",
									ManaCost: opt.Val(cost.Mana{
										cost.O(2),
										cost.HybridMana(mana.R, mana.G),
									}),
								},
							},
							Optional:      true,
							PublishResult: game.ResultKey("neyith-combat-pay"),
						},
						{
							Primitive: game.ModifyPT{
								TargetIndex: 0,
								PowerDelta: game.Dynamic(game.DynamicAmount{
									Kind: game.DynamicAmountObjectPower,
									Object: game.ObjectReference{
										Kind:        game.ObjectReferenceTargetPermanent,
										TargetIndex: 0,
									},
								}),
								Duration: game.DurationUntilEndOfTurn,
							},
							ResultGate: opt.Val(game.InstructionResultGate{
								Key:       game.ResultKey("neyith-combat-pay"),
								Accepted:  game.TriTrue,
								Succeeded: game.TriTrue,
							}),
						},
						{
							Primitive: game.ApplyRule{
								TargetIndex: 0,
								RuleEffects: []game.RuleEffect{
									{
										Kind: game.RuleEffectMustBeBlocked,
									},
								},
								Duration: game.DurationUntilEndOfTurn,
							},
							ResultGate: opt.Val(game.InstructionResultGate{
								Key:       game.ResultKey("neyith-combat-pay"),
								Accepted:  game.TriTrue,
								Succeeded: game.TriTrue,
							}),
						},
					},
				},
			},
		},
	},
}
