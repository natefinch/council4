package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AmuletOfUnmaking is the card definition for Amulet of Unmaking.
//
// Type: Artifact
// Cost: {5}
//
// Oracle text:
//
//	{5}, {T}, Exile this artifact: Exile target artifact, creature, or land. Activate only as a sorcery.
var AmuletOfUnmaking = newAmuletOfUnmaking

func newAmuletOfUnmaking() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Amulet of Unmaking",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{5}, {T}, Exile this artifact: Exile target artifact, creature, or land. Activate only as a sorcery.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(5),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalExileSource,
							Text:        "Exile this artifact",
							Amount:      1,
							Source:      zone.Battlefield,
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
										Constraint: "target artifact, creature, or land",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Land}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MovePermanent{
											Object:      game.TargetPermanentReference(0),
											Destination: zone.Exile,
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
			{5}, {T}, Exile this artifact: Exile target artifact, creature, or land. Activate only as a sorcery.
		`,
		},
	}
}
