package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// KarnSBastion is the card definition for Karn's Bastion.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{4}, {T}: Proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)
var KarnSBastion = newKarnSBastion

func newKarnSBastion() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Karn's Bastion",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{4}, {T}: Proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(4),
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
										Primitive: game.Proliferate{
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
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			OracleText: `
			{T}: Add {C}.
			{4}, {T}: Proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)
		`,
		},
	}
}
