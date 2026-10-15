package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TabletOfCompleation is the card definition for Tablet of Compleation.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	{T}: Put an oil counter on this artifact.
//	{T}: Add {C}. Activate only if this artifact has two or more oil counters on it.
//	{1}, {T}: Draw a card. Activate only if this artifact has five or more oil counters on it.
var TabletOfCompleation = newTabletOfCompleation

func newTabletOfCompleation() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Tablet of Compleation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}: Put an oil counter on this artifact.",
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
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.Oil,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.ActivatedAbility{
					Text: "{1}, {T}: Draw a card. Activate only if this artifact has five or more oil counters on it.",
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
					ActivationCondition: opt.Val(game.Condition{
						Text:   "only if this artifact has five or more oil counters on it",
						Object: opt.Val(game.SourcePermanentReference()),
						ObjectMatches: opt.Val(game.Selection{RequiredCounter: counter.Oil, RequiredCounterCount: opt.Val(compare.Int{
							Op:    compare.GreaterOrEqual,
							Value: 5,
						})}),
						SourceCounterKind: counter.PlusOnePlusOne,
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
					Text: "{T}: Add {C}. Activate only if this artifact has two or more oil counters on it.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ActivationCondition: opt.Val(game.Condition{
						Text:   "only if this artifact has two or more oil counters on it",
						Object: opt.Val(game.SourcePermanentReference()),
						ObjectMatches: opt.Val(game.Selection{RequiredCounter: counter.Oil, RequiredCounterCount: opt.Val(compare.Int{
							Op:    compare.GreaterOrEqual,
							Value: 2,
						})}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
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
			{T}: Put an oil counter on this artifact.
			{T}: Add {C}. Activate only if this artifact has two or more oil counters on it.
			{1}, {T}: Draw a card. Activate only if this artifact has five or more oil counters on it.
		`,
		},
	}
}
