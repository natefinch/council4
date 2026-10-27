// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RTNestingGrounds is the card definition for RT Nesting Grounds.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{1}, {T}: Move a counter from target permanent you control onto a second target permanent. Activate only as a sorcery.
var RTNestingGrounds = newRTNestingGrounds

func newRTNestingGrounds() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Nesting Grounds",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}, {T}: Move a counter from target permanent you control onto a second target permanent. Activate only as a sorcery.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target permanent you control",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{Controller: game.ControllerYou}),
									},
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target permanent",
										Allow:      game.TargetAllowPermanent,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveCounters{
											Amount:      game.Fixed(1),
											Object:      game.TargetPermanentReference(1),
											CounterKind: counter.PlusOnePlusOne,
											Source: game.CounterSourceSpec{
												Kind:   game.CounterSourceTarget,
												Object: game.TargetPermanentReference(0),
											},
											ChooseKind: true,
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
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			OracleText: `
			{T}: Add {C}.
			{1}, {T}: Move a counter from target permanent you control onto a second target permanent. Activate only as a sorcery.
		`,
		},
	}
}
