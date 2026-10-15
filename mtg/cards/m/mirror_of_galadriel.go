package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MirrorOfGaladriel is the card definition for Mirror of Galadriel.
//
// Type: Legendary Artifact
// Cost: {2}
//
// Oracle text:
//
//	{5}, {T}: Scry 1, then draw a card. This ability costs {1} less to activate for each legendary creature you control.
var MirrorOfGaladriel = newMirrorOfGaladriel

func newMirrorOfGaladriel() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Mirror of Galadriel",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{5}, {T}: Scry 1, then draw a card. This ability costs {1} less to activate for each legendary creature you control.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(5),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					CostModifiers: []game.CostModifier{
						game.CostModifier{
							Kind:               game.CostModifierAbility,
							PerObjectReduction: 1,
							CountSelection:     &game.Selection{RequiredTypes: []types.Card{types.Creature}, Supertypes: []types.Super{types.Legendary}, Controller: game.ControllerYou},
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Scry{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
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
			{5}, {T}: Scry 1, then draw a card. This ability costs {1} less to activate for each legendary creature you control.
		`,
		},
	}
}
