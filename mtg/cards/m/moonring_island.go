package m

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

// MoonringIsland is the card definition for Moonring Island.
//
// Type: Land — Island
//
// Oracle text:
//
//	({T}: Add {U}.)
//	This land enters tapped.
//	{U}, {T}: Look at the top card of target player's library. Activate only if you control two or more blue permanents.
var MoonringIsland = newMoonringIsland

func newMoonringIsland() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:     "Moonring Island",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Island},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{U}, {T}: Look at the top card of target player's library. Activate only if you control two or more blue permanents.",
					ManaCost: opt.Val(cost.Mana{
						cost.U,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					ActivationCondition: opt.Val(game.Condition{
						Text: "only if you control two or more blue permanents",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{ColorsAny: []color.Color{color.Blue}},
							MinCount:  2,
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target player",
										Allow:      game.TargetAllowPlayer,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.LookAtLibraryTop{
											Player:        game.TargetPlayerReference(0),
											PublishLinked: game.LinkedKey("look-at-target-library-top"),
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
				game.TapManaAbility(mana.U),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {U}.)
			This land enters tapped.
			{U}, {T}: Look at the top card of target player's library. Activate only if you control two or more blue permanents.
		`,
		},
	}
}
