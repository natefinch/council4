package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TocasiaSDigSite is the card definition for Tocasia's Dig Site.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{3}, {T}: Surveil 1. (Look at the top card of your library. You may put it into your graveyard.)
var TocasiaSDigSite = newTocasiaSDigSite

func newTocasiaSDigSite() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Tocasia's Dig Site",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{3}, {T}: Surveil 1. (Look at the top card of your library. You may put it into your graveyard.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
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
										Primitive: game.Surveil{
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
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			OracleText: `
			{T}: Add {C}.
			{3}, {T}: Surveil 1. (Look at the top card of your library. You may put it into your graveyard.)
		`,
		},
	}
}
