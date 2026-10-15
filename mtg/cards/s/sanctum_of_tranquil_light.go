package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SanctumOfTranquilLight is the card definition for Sanctum of Tranquil Light.
//
// Type: Legendary Enchantment — Shrine
// Cost: {W}
//
// Oracle text:
//
//	{5}{W}: Tap target creature. This ability costs {1} less to activate for each Shrine you control.
var SanctumOfTranquilLight = newSanctumOfTranquilLight

func newSanctumOfTranquilLight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Sanctum of Tranquil Light",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Enchantment},
			Subtypes:   []types.Sub{types.Shrine},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{5}{W}: Tap target creature. This ability costs {1} less to activate for each Shrine you control.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(5),
						cost.W,
					}),
					CostModifiers: []game.CostModifier{
						game.CostModifier{
							Kind:               game.CostModifierAbility,
							PerObjectReduction: 1,
							CountSelection:     &game.Selection{SubtypesAny: []types.Sub{types.Shrine}, Controller: game.ControllerYou},
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
										Primitive: game.Tap{
											Object: game.TargetPermanentReference(0),
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
			{5}{W}: Tap target creature. This ability costs {1} less to activate for each Shrine you control.
		`,
		},
	}
}
