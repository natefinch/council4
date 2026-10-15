package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SapseepForest is the card definition for Sapseep Forest.
//
// Type: Land — Forest
//
// Oracle text:
//
//	({T}: Add {G}.)
//	This land enters tapped.
//	{G}, {T}: You gain 1 life. Activate only if you control two or more green permanents.
var SapseepForest = newSapseepForest

func newSapseepForest() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:     "Sapseep Forest",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Forest},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{G}, {T}: You gain 1 life. Activate only if you control two or more green permanents.",
					ManaCost: opt.Val(cost.Mana{
						cost.G,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					ActivationCondition: opt.Val(game.Condition{
						Text: "only if you control two or more green permanents",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{ColorsAny: []color.Color{color.Green}},
							MinCount:  2,
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.GainLife{
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
				game.TapManaAbility(mana.G),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {G}.)
			This land enters tapped.
			{G}, {T}: You gain 1 life. Activate only if you control two or more green permanents.
		`,
		},
	}
}
