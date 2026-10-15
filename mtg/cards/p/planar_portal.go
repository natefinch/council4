package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PlanarPortal is the card definition for Planar Portal.
//
// Type: Artifact
// Cost: {6}
//
// Oracle text:
//
//	{6}, {T}: Search your library for a card, put that card into your hand, then shuffle.
var PlanarPortal = newPlanarPortal

func newPlanarPortal() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Planar Portal",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{6}, {T}: Search your library for a card, put that card into your hand, then shuffle.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(6),
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
												SourceZone:       zone.Library,
												Destination:      zone.Hand,
												FailToFindPolicy: game.SearchMustFindIfAvailable,
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
			{6}, {T}: Search your library for a card, put that card into your hand, then shuffle.
		`,
		},
	}
}
