package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MoltenTailMasticore is the card definition for Molten-Tail Masticore.
//
// Type: Artifact Creature — Masticore
// Cost: {4}
//
// Oracle text:
//
//	At the beginning of your upkeep, sacrifice this creature unless you discard a card.
//	{4}, Exile a creature card from your graveyard: This creature deals 4 damage to any target.
//	{2}: Regenerate this creature.
var MoltenTailMasticore = newMoltenTailMasticore()

func newMoltenTailMasticore() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Molten-Tail Masticore",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Masticore},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{4}, Exile a creature card from your graveyard: This creature deals 4 damage to any target.",
					ManaCost: opt.Val(cost.Mana{cost.O(4)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:          cost.AdditionalExile,
							Text:          "Exile a creature card from your graveyard",
							Amount:        1,
							Source:        zone.Graveyard,
							MatchCardType: true,
							CardType:      types.Creature,
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
									Amount:       game.Fixed(4),
									Recipient:    game.AnyTargetDamageRecipient(0),
									DamageSource: opt.Val(game.SourcePermanentReference()),
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:           "{2}: Regenerate this creature.",
					ManaCost:       opt.Val(cost.Mana{cost.O(2)}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Regenerate{
									Object: game.SourcePermanentReference(),
								},
							},
						},
					}.Ability(),
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
										Prompt: "Discard a card?",
										AdditionalCosts: []cost.Additional{
											{
												Kind:   cost.AdditionalDiscard,
												Text:   "discard a card",
												Amount: 1,
												Source: zone.Hand,
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
			At the beginning of your upkeep, sacrifice this creature unless you discard a card.
			{4}, Exile a creature card from your graveyard: This creature deals 4 damage to any target.
			{2}: Regenerate this creature.
		`,
		},
	}
}
