package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Hecatomb is the card definition for Hecatomb.
//
// Type: Enchantment
// Cost: {1}{B}{B}
//
// Oracle text:
//
//	When this enchantment enters, sacrifice this enchantment unless you sacrifice four creatures.
//	Tap an untapped Swamp you control: This enchantment deals 1 damage to any target.
var Hecatomb = newHecatomb()

func newHecatomb() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Hecatomb",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Tap an untapped Swamp you control: This enchantment deals 1 damage to any target.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalTapPermanents,
							Text:        "Tap an untapped Swamp you control",
							Amount:      1,
							SubtypesAny: cost.SubtypeSet{types.Swamp},
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "any target",
								Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:       game.Fixed(1),
									Recipient:    game.AnyTargetDamageRecipient(0),
									DamageSource: opt.Val(game.SourcePermanentReference()),
								},
							},
						},
					}.Ability(),
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
										Prompt: "Sacrifice four creatures?",
										AdditionalCosts: []cost.Additional{
											{
												Kind:               cost.AdditionalSacrifice,
												Text:               "sacrifice four creatures",
												Amount:             4,
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
			When this enchantment enters, sacrifice this enchantment unless you sacrifice four creatures.
			Tap an untapped Swamp you control: This enchantment deals 1 damage to any target.
		`,
		},
	}
}
