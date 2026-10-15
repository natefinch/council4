package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SeaGateWreckage is the card definition for Sea Gate Wreckage.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}. ({C} represents colorless mana.)
//	{2}{C}, {T}: Draw a card. Activate only if you have no cards in hand.
var SeaGateWreckage = newSeaGateWreckage

func newSeaGateWreckage() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Sea Gate Wreckage",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}{C}, {T}: Draw a card. Activate only if you have no cards in hand.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
						cost.C,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					ActivationCondition: opt.Val(game.Condition{
						Text:                "only if you have no cards in hand",
						ControllerHandEmpty: true,
						SourceCounterKind:   counter.PlusOnePlusOne,
					}),
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
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
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "{T}: Add {C}. ({C} represents colorless mana.)",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.C,
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
			{T}: Add {C}. ({C} represents colorless mana.)
			{2}{C}, {T}: Draw a card. Activate only if you have no cards in hand.
		`,
		},
	}
}
