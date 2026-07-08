package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WeakstoneSSubjugation is the card definition for Weakstone's Subjugation.
//
// Type: Enchantment — Aura
// Cost: {U}
//
// Oracle text:
//
//	Enchant artifact or creature
//	When this Aura enters, you may pay {3}. If you do, tap enchanted permanent.
//	Enchanted permanent doesn't untap during its controller's untap step.
var WeakstoneSSubjugation = newWeakstoneSSubjugation

func newWeakstoneSSubjugation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Weakstone's Subjugation",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "artifact or creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature}}),
				}),
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectDoesntUntap,
							AffectedAttached: true,
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
								Primitive: game.Pay{
									Payment: game.ResolutionPayment{
										Prompt: "Pay {3}?",
										ManaCost: opt.Val(cost.Mana{
											cost.O(3),
										}),
									},
								},
								PublishResult: game.ResultKey("controller-paid"),
							},
							{
								Primitive: game.Tap{
									Object: game.SourceAttachedPermanentReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "controller-paid",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Enchant artifact or creature
			When this Aura enters, you may pay {3}. If you do, tap enchanted permanent.
			Enchanted permanent doesn't untap during its controller's untap step.
		`,
		},
	}
}
