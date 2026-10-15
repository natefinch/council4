package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TumbleMagnet is the card definition for Tumble Magnet.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	This artifact enters with three charge counters on it.
//	{T}, Remove a charge counter from this artifact: Tap target artifact or creature.
var TumbleMagnet = newTumbleMagnet

func newTumbleMagnet() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Tumble Magnet",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}, Remove a charge counter from this artifact: Tap target artifact or creature.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove a charge counter from this artifact",
							Amount:      1,
							CounterKind: counter.Charge,
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
										Constraint: "target artifact or creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature}}),
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This artifact enters with three charge counters on it.", game.CounterPlacement{Kind: counter.Charge, Amount: 3}),
			},
			OracleText: `
			This artifact enters with three charge counters on it.
			{T}, Remove a charge counter from this artifact: Tap target artifact or creature.
		`,
		},
	}
}
