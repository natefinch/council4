package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KoskunFalls is the card definition for Koskun Falls.
//
// Type: World Enchantment
// Cost: {2}{B}{B}
//
// Oracle text:
//
//	At the beginning of your upkeep, sacrifice this enchantment unless you tap an untapped creature you control.
//	Creatures can't attack you unless their controller pays {2} for each creature they control that's attacking you.
var KoskunFalls = newKoskunFalls()

func newKoskunFalls() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Koskun Falls",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.World},
			Types:      []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectAttackTax,
							AffectedPlayer:   game.PlayerYou,
							AttackTaxGeneric: 2,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Pay{
									Payment: game.ResolutionPayment{
										Prompt: "Tap an untapped creature you control?",
										AdditionalCosts: []cost.Additional{
											{
												Kind:               cost.AdditionalTapPermanents,
												Text:               "tap an untapped creature you control",
												Amount:             1,
												MatchPermanentType: true,
												PermanentType:      types.Creature,
											},
										},
									},
								},
								PublishResult: game.ResultKey("sacrifice-unless-paid"),
							},
							{
								Primitive: game.Sacrifice{
									Object: game.SourcePermanentReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "sacrifice-unless-paid",
									Succeeded: game.TriFalse,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			At the beginning of your upkeep, sacrifice this enchantment unless you tap an untapped creature you control.
			Creatures can't attack you unless their controller pays {2} for each creature they control that's attacking you.
		`,
		},
	}
}
