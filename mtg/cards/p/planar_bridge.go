package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PlanarBridge is the card definition for Planar Bridge.
//
// Type: Legendary Artifact
// Cost: {6}
//
// Oracle text:
//
//	{8}, {T}: Search your library for a permanent card, put it onto the battlefield, then shuffle.
var PlanarBridge = newPlanarBridge

func newPlanarBridge() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Planar Bridge",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{8}, {T}: Search your library for a permanent card, put it onto the battlefield, then shuffle.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(8),
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
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Search{
											Player: game.ControllerReference(),
											Spec: game.SearchSpec{
												SourceZone:  zone.Library,
												Destination: zone.Battlefield,
												Filter:      game.Selection{RequirePermanentCard: true},
											},
											Amount: game.Fixed(1),
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
			{8}, {T}: Search your library for a permanent card, put it onto the battlefield, then shuffle.
		`,
		},
	}
}
