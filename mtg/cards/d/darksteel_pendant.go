package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DarksteelPendant is the card definition for Darksteel Pendant.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	Indestructible (Effects that say "destroy" don't destroy this artifact.)
//	{1}, {T}: Scry 1. (Look at the top card of your library. You may put that card on the bottom.)
var DarksteelPendant = newDarksteelPendant

func newDarksteelPendant() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Darksteel Pendant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.IndestructibleStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}, {T}: Scry 1. (Look at the top card of your library. You may put that card on the bottom.)",
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
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Indestructible (Effects that say "destroy" don't destroy this artifact.)
			{1}, {T}: Scry 1. (Look at the top card of your library. You may put that card on the bottom.)
		`,
		},
	}
}
