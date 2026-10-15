package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AccursedDuneyard is the card definition for Accursed Duneyard.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{2}, {T}: Regenerate target Shade, Skeleton, Specter, Spirit, Vampire, Wraith, or Zombie. (The next time it would be destroyed this turn, instead tap it, remove it from combat, and heal all damage on it.)
var AccursedDuneyard = newAccursedDuneyard

func newAccursedDuneyard() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Accursed Duneyard",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}, {T}: Regenerate target Shade, Skeleton, Specter, Spirit, Vampire, Wraith, or Zombie. (The next time it would be destroyed this turn, instead tap it, remove it from combat, and heal all damage on it.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
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
										Constraint: "target Shade, Skeleton, Specter, Spirit, Vampire, Wraith, or Zombie",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Shade, types.Skeleton, types.Specter, types.Spirit, types.Vampire, types.Wraith, types.Zombie}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Regenerate{
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
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			OracleText: `
			{T}: Add {C}.
			{2}, {T}: Regenerate target Shade, Skeleton, Specter, Spirit, Vampire, Wraith, or Zombie. (The next time it would be destroyed this turn, instead tap it, remove it from combat, and heal all damage on it.)
		`,
		},
	}
}
