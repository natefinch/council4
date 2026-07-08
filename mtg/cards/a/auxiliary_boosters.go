package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AuxiliaryBoosters is the card definition for Auxiliary Boosters.
//
// Type: Artifact — Equipment
// Cost: {4}{W}
//
// Oracle text:
//
//	When this Equipment enters, create a 2/2 colorless Robot artifact creature token and attach this Equipment to it.
//	Equipped creature gets +1/+2 and has flying.
//	Equip {3} ({3}: Attach to target creature you control. Equip only as a sorcery.)
var AuxiliaryBoosters = newAuxiliaryBoosters

func newAuxiliaryBoosters() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Auxiliary Boosters",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     1,
							ToughnessDelta: 2,
						},
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddKeywords: []game.Keyword{
								game.Flying,
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(3)}),
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
									Source:        game.TokenDef(auxiliaryBoostersToken),
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
			When this Equipment enters, create a 2/2 colorless Robot artifact creature token and attach this Equipment to it.
			Equipped creature gets +1/+2 and has flying.
			Equip {3} ({3}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		},
	}
}

var auxiliaryBoostersToken = newAuxiliaryBoostersToken()

func newAuxiliaryBoostersToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Robot",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Robot},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
