package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TheIrencrag is the card definition for The Irencrag.
//
// Type: Legendary Artifact
// Cost: {2}
//
// Oracle text:
//
//	{T}: Add {C}.
//	Whenever a legendary creature you control enters, you may have The Irencrag become a legendary Equipment artifact named Everflame, Heroes' Legacy. If you do, it gains equip {3} and "Equipped creature gets +3/+3" and loses all other abilities.
var TheIrencrag = newTheIrencrag

func newTheIrencrag() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "The Irencrag",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Supertypes: []types.Super{types.Legendary}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.SourcePermanentReference()),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:              game.LayerAbility,
											RemoveAllAbilities: true,
										},
										game.ContinuousEffect{
											Layer:   game.LayerText,
											SetName: "Everflame, Heroes' Legacy",
										},
										game.ContinuousEffect{
											Layer:         game.LayerType,
											AddSupertypes: []types.Super{types.Legendary},
											SetTypes:      []types.Card{types.Artifact},
											SetSubtypes:   []types.Sub{types.Equipment},
										},
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddAbilities: []game.Ability{
												new(game.EquipActivatedAbility(cost.Mana{cost.O(3)})),
												new(game.StaticAbility{
													ContinuousEffects: []game.ContinuousEffect{
														game.ContinuousEffect{
															Layer:          game.LayerPowerToughnessModify,
															Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
															PowerDelta:     3,
															ToughnessDelta: 3,
														},
													},
												}),
											},
										},
									},
									Duration: game.DurationPermanent,
								},
								Optional: true,
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{T}: Add {C}.
			Whenever a legendary creature you control enters, you may have The Irencrag become a legendary Equipment artifact named Everflame, Heroes' Legacy. If you do, it gains equip {3} and "Equipped creature gets +3/+3" and loses all other abilities.
		`,
		},
	}
}
