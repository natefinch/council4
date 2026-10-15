package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GiantSSkewer is the card definition for Giant's Skewer.
//
// Type: Artifact — Equipment
// Cost: {1}{B}
//
// Oracle text:
//
//	Equipped creature gets +2/+1.
//	Whenever equipped creature deals combat damage to a creature, create a Food token. (It's an artifact with "{2}, {T}, Sacrifice this token: You gain 3 life.")
//	Equip {3} ({3}: Attach to target creature you control. Equip only as a sorcery.)
var GiantSSkewer = newGiantSSkewer

func newGiantSSkewer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Giant's Skewer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:   []color.Color{color.Black},
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +2/+1.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     2,
							ToughnessDelta: 1,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(3)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever equipped creature deals combat damage to a creature, create a Food token. (It's an artifact with \"{2}, {T}, Sacrifice this token: You gain 3 life.\")",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                 game.EventDamageDealt,
							Source:                game.TriggerSourceAttachedPermanent,
							Subject:               game.TriggerSubjectDamageSource,
							StackObjectKind:       game.StackSpell,
							DamageRecipient:       game.DamageRecipientPermanent,
							DamageRecipientTypes:  []types.Card{types.Creature},
							DamageSourceSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							RequireCombatDamage:   true,
							CounterKind:           counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(giantSSkewerToken),
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
			Equipped creature gets +2/+1.
			Whenever equipped creature deals combat damage to a creature, create a Food token. (It's an artifact with "{2}, {T}, Sacrifice this token: You gain 3 life.")
			Equip {3} ({3}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		},
	}
}

var giantSSkewerToken = newGiantSSkewerToken()

func newGiantSSkewerToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Food",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Food},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}, {T}, Sacrifice this artifact: You gain 3 life.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
					}),
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
										Primitive: game.GainLife{
											Amount: game.Fixed(3),
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
		},
	}
}
