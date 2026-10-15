package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ImprovisedArsenal is the card definition for Improvised Arsenal.
//
// Type: Artifact — Equipment
// Cost: {1}{R}
//
// Oracle text:
//
//	Equipped creature gets +1/+0 for each artifact you control.
//	{4}{R}: Create a token that's a copy of this Equipment.
//	Equip {R}
var ImprovisedArsenal = newImprovisedArsenal

func newImprovisedArsenal() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Improvised Arsenal",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +1/+0 for each artifact you control.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerPowerToughnessModify,
							PowerDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountCountSelector,
								Multiplier:  1,
								CounterKind: counter.PlusOnePlusOne,
								Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Artifact}, Controller: game.ControllerYou}),
							}),
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{4}{R}: Create a token that's a copy of this Equipment.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(4),
						cost.R,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenCopyOf(game.TokenCopySpec{
												Source: game.TokenCopySourceObject,
												Object: game.SourcePermanentReference(),
											}),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.EquipActivatedAbility(cost.Mana{cost.R}),
			},
			OracleText: `
			Equipped creature gets +1/+0 for each artifact you control.
			{4}{R}: Create a token that's a copy of this Equipment.
			Equip {R}
		`,
		},
	}
}
