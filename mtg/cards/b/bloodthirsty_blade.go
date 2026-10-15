package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BloodthirstyBlade is the card definition for Bloodthirsty Blade.
//
// Type: Artifact — Equipment
// Cost: {2}
//
// Oracle text:
//
//	Equipped creature gets +2/+0 and is goaded. (It attacks each combat if able and attacks a player other than you if able.)
//	{1}: Attach this Equipment to target creature an opponent controls. Activate only as a sorcery.
var BloodthirstyBlade = newBloodthirstyBlade

func newBloodthirstyBlade() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Bloodthirsty Blade",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +2/+0 and is goaded. (It attacks each combat if able and attacks a player other than you if able.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: 2,
						},
					},
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectGoaded,
							AffectedAttached: true,
							Duration:         game.DurationPermanent,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}: Attach this Equipment to target creature an opponent controls. Activate only as a sorcery.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
					}),
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature an opponent controls",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Attach{
											Attachment: game.SourcePermanentReference(),
											Target:     game.TargetPermanentReference(0),
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
			Equipped creature gets +2/+0 and is goaded. (It attacks each combat if able and attacks a player other than you if able.)
			{1}: Attach this Equipment to target creature an opponent controls. Activate only as a sorcery.
		`,
		},
	}
}
