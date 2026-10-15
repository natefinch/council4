package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GoblinFirebomb is the card definition for Goblin Firebomb.
//
// Type: Artifact
// Cost: {1}
//
// Oracle text:
//
//	Flash
//	{7}, {T}, Sacrifice this artifact: Destroy target permanent.
var GoblinFirebomb = newGoblinFirebomb

func newGoblinFirebomb() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Goblin Firebomb",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{7}, {T}, Sacrifice this artifact: Destroy target permanent.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(7),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalSacrificeSource,
							Text:        "Sacrifice this artifact",
							Amount:      1,
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
										Constraint: "target permanent",
										Allow:      game.TargetAllowPermanent,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Destroy{
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
			Flash
			{7}, {T}, Sacrifice this artifact: Destroy target permanent.
		`,
		},
	}
}
