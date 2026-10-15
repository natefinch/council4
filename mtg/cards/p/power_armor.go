package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PowerArmor is the card definition for Power Armor.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	Domain — {3}, {T}: Target creature gets +1/+1 until end of turn for each basic land type among lands you control.
var PowerArmor = newPowerArmor

func newPowerArmor() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Power Armor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Domain — {3}, {T}: Target creature gets +1/+1 until end of turn for each basic land type among lands you control.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ModifyPT{
											Object: game.TargetPermanentReference(0),
											PowerDelta: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountControllerBasicLandTypeCount,
												Multiplier: 1,
											}),
											ToughnessDelta: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountControllerBasicLandTypeCount,
												Multiplier: 1,
											}),
											Duration: game.DurationUntilEndOfTurn,
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
			Domain — {3}, {T}: Target creature gets +1/+1 until end of turn for each basic land type among lands you control.
		`,
		},
	}
}
