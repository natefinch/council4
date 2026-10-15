package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CitanulFlute is the card definition for Citanul Flute.
//
// Type: Artifact
// Cost: {5}
//
// Oracle text:
//
//	{X}, {T}: Search your library for a creature card with mana value X or less, reveal it, put it into your hand, then shuffle.
var CitanulFlute = newCitanulFlute

func newCitanulFlute() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Citanul Flute",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{X}, {T}: Search your library for a creature card with mana value X or less, reveal it, put it into your hand, then shuffle.",
					ManaCost: opt.Val(cost.Mana{
						cost.X,
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
												SourceZone:        zone.Library,
												Destination:       zone.Hand,
												Filter:            game.Selection{RequiredTypes: []types.Card{types.Creature}},
												MaxManaValueFromX: true,
												Reveal:            true,
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
			{X}, {T}: Search your library for a creature card with mana value X or less, reveal it, put it into your hand, then shuffle.
		`,
		},
	}
}
