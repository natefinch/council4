// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// RTReaverCleaver is the card definition for RT Reaver Cleaver.
//
// Type: Legendary Artifact — Equipment
//
// Oracle text:
//
//	Equipped creature gets +1/+1 and has trample and "Whenever this creature deals combat damage to a player or planeswalker, create that many Treasure tokens."
//	Equip {3}
var RTReaverCleaver = newRTReaverCleaver

func newRTReaverCleaver() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:       "RT Reaver Cleaver",
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			Subtypes:   []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +1/+1 and has trample and \"Whenever this creature deals combat damage to a player or planeswalker, create that many Treasure tokens.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Trample},
						},
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.TriggeredAbility{
									Text: "Whenever this creature deals combat damage to a player or planeswalker, create that many Treasure tokens.",
									Trigger: game.TriggerCondition{
										Type: game.TriggerWhenever,
										Pattern: game.TriggerPattern{
											Event:                    game.EventDamageDealt,
											Source:                   game.TriggerSourceSelf,
											Subject:                  game.TriggerSubjectDamageSource,
											StackObjectKind:          game.StackSpell,
											DamageRecipient:          game.DamageRecipientPlayer | game.DamageRecipientPermanent,
											DamageRecipientSelection: game.Selection{RequiredTypes: []types.Card{types.Planeswalker}},
											RequireCombatDamage:      true,
											CounterKind:              counter.PlusOnePlusOne,
										},
									},
									Content: game.AbilityContent{
										Modes: []game.Mode{
											game.Mode{
												Sequence: []game.Instruction{
													game.Instruction{
														Primitive: game.CreateToken{
															Amount: game.Dynamic(game.DynamicAmount{
																Kind:       game.DynamicAmountEventDamage,
																Multiplier: 1,
															}),
															Source: game.TokenDef(rTReaverCleaverToken),
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
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(3)}),
			},
			OracleText: `
			Equipped creature gets +1/+1 and has trample and "Whenever this creature deals combat damage to a player or planeswalker, create that many Treasure tokens."
			Equip {3}
		`,
		},
	}
}

var rTReaverCleaverToken = newRTReaverCleaverToken()

func newRTReaverCleaverToken() *game.CardDef {
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
