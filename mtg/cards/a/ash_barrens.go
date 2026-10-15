package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AshBarrens is the card definition for Ash Barrens.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	Basic landcycling {1} ({1}, Discard this card: Search your library for a basic land card, reveal it, put it into your hand, then shuffle.)
var AshBarrens = newAshBarrens

func newAshBarrens() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Ash Barrens",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalDiscard,
							Text:        "Discard this card",
							Amount:      1,
							Source:      zone.Hand,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Hand,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Search{
											Player: game.ControllerReference(),
											Spec: game.SearchSpec{
												SourceZone:  zone.Library,
												Destination: zone.Hand,
												Filter:      game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}},
												Reveal:      true,
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
					KeywordAbilities: []game.KeywordAbility{
						game.CyclingKeyword{
							Cost: cost.Mana{
								cost.O(1),
							},
						},
					},
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			OracleText: `
			{T}: Add {C}.
			Basic landcycling {1} ({1}, Discard this card: Search your library for a basic land card, reveal it, put it into your hand, then shuffle.)
		`,
		},
	}
}
