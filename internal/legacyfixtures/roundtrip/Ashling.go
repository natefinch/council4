// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Ashling is the card definition for Ashling.
//
// Type: Legendary Creature - Elemental Shaman
//
// Oracle text:
//
//	{1}{R}: Put a +1/+1 counter on Ashling. If this is the third time this ability has resolved this turn, remove all +1/+1 counters from Ashling, and it deals that much damage to each creature and each player.
var Ashling = newAshling

func newAshling() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:       "Ashling",
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}{R}: Put a +1/+1 counter on Ashling. If this is the third time this ability has resolved this turn, remove all +1/+1 counters from Ashling, and it deals that much damage to each creature and each player.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
						cost.R,
					}),
					ZoneOfFunction:            zone.Battlefield,
					CountsResolutionsThisTurn: true,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.PlusOnePlusOne,
										},
									},
									game.Instruction{
										Primitive: game.RemoveCounter{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:        game.DynamicAmountObjectCounters,
												CounterKind: counter.PlusOnePlusOne,
												Object:      game.SourcePermanentReference(),
											}),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.PlusOnePlusOne,
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:                                   "If this is the third time this ability has resolved this turn",
												SourceCounterKind:                      counter.PlusOnePlusOne,
												SourceAbilityResolutionOrdinalThisTurn: 3,
											}),
										}),
										PublishCondition: game.ConditionKey("condition-0"),
										PublishResult:    game.ResultKey("removed-counter-quantity-2"),
									},
									game.Instruction{
										Primitive: game.Damage{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountPreviousEffectResult,
												Multiplier: 1,
												ResultKey:  game.ResultKey("removed-counter-quantity-2"),
											}),
											Recipient:    game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})),
											DamageSource: opt.Val(game.SourcePermanentReference()),
										},
										ConditionGate: game.ConditionKey("condition-0"),
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:             game.ResultKey("removed-counter-quantity-2"),
											AmountAvailable: true,
										}),
									},
									game.Instruction{
										Primitive: game.Damage{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountPreviousEffectResult,
												Multiplier: 1,
												ResultKey:  game.ResultKey("removed-counter-quantity-2"),
											}),
											Recipient:    game.PlayerGroupDamageRecipient(game.AllPlayersReference()),
											DamageSource: opt.Val(game.SourcePermanentReference()),
										},
										ConditionGate: game.ConditionKey("condition-0"),
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:             game.ResultKey("removed-counter-quantity-2"),
											AmountAvailable: true,
										}),
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
			{1}{R}: Put a +1/+1 counter on Ashling. If this is the third time this ability has resolved this turn, remove all +1/+1 counters from Ashling, and it deals that much damage to each creature and each player.
		`,
		},
	}
}
