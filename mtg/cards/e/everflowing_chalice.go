package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EverflowingChalice is the card definition for Everflowing Chalice.
//
// Type: Artifact
// Cost: {0}
//
// Oracle text:
//
//	Multikicker {2} (You may pay an additional {2} any number of times as you cast this spell.)
//	This artifact enters with a charge counter on it for each time it was kicked.
//	{T}: Add {C} for each charge counter on this artifact.
var EverflowingChalice = newEverflowingChalice

func newEverflowingChalice() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Everflowing Chalice",
			ManaCost: opt.Val(cost.Mana{
				cost.O(0),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Multikicker {2}",
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{
							Cost: cost.Mana{
								cost.O(2),
							},
							Multi: true,
						},
					},
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "{T}: Add {C} for each charge counter on this artifact.",
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
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:        game.DynamicAmountObjectCounters,
												Multiplier:  1,
												CounterKind: counter.Charge,
												Object:      game.SourcePermanentReference(),
											}),
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This artifact enters with a charge counter on it for each time it was kicked.", game.CounterPlacement{Kind: counter.Charge, Dynamic: opt.Val(&game.DynamicAmount{
					Kind:       game.DynamicAmountTimesKicked,
					Multiplier: 1,
				})}),
			},
			OracleText: `
			Multikicker {2} (You may pay an additional {2} any number of times as you cast this spell.)
			This artifact enters with a charge counter on it for each time it was kicked.
			{T}: Add {C} for each charge counter on this artifact.
		`,
		},
	}
}
