package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MaskOfImmolation is the card definition for Mask of Immolation.
//
// Type: Artifact — Equipment
// Cost: {1}{R}
//
// Oracle text:
//
//	When this Equipment enters, create a 1/1 red Elemental creature token, then attach this Equipment to it.
//	Equipped creature has "Sacrifice this creature: It deals 1 damage to any target."
//	Equip {2} ({2}: Attach to target creature you control. Equip only as a sorcery.)
var MaskOfImmolation = newMaskOfImmolation()

func newMaskOfImmolation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Mask of Immolation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddAbilities: []game.Ability{
								new(game.ActivatedAbility{
									Text: "Sacrifice this creature: It deals 1 damage to any target.",
									AdditionalCosts: []cost.Additional{
										{
											Kind:   cost.AdditionalSacrificeSource,
											Text:   "Sacrifice this creature",
											Amount: 1,
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
								}),
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(2)}),
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
								Primitive: game.CreateToken{
									Amount:        game.Fixed(1),
									Source:        game.TokenDef(maskOfImmolationToken),
									PublishLinked: game.LinkedKey("created-token"),
								},
							},
							{
								Primitive: game.Attach{
									Attachment: game.SourcePermanentReference(),
									Target:     game.LinkedObjectReference("created-token"),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this Equipment enters, create a 1/1 red Elemental creature token, then attach this Equipment to it.
			Equipped creature has "Sacrifice this creature: It deals 1 damage to any target."
			Equip {2} ({2}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		},
	}
}

var maskOfImmolationToken = newMaskOfImmolationToken()

func newMaskOfImmolationToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Elemental",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
