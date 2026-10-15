package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// HonorWornShaku is the card definition for Honor-Worn Shaku.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{T}: Add {C}.
//	Tap an untapped legendary permanent you control: Untap this artifact.
var HonorWornShaku = newHonorWornShaku

func newHonorWornShaku() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Honor-Worn Shaku",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Tap an untapped legendary permanent you control: Untap this artifact.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:             cost.AdditionalTapPermanents,
							Text:             "Tap an untapped legendary permanent you control",
							Amount:           1,
							CounterKind:      counter.PlusOnePlusOne,
							RequireSupertype: types.Legendary,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Untap{
											Object: game.SourcePermanentReference(),
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
			Tap an untapped legendary permanent you control: Untap this artifact.
		`,
		},
	}
}
