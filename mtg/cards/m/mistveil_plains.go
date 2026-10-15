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

// MistveilPlains is the card definition for Mistveil Plains.
//
// Type: Land — Plains
//
// Oracle text:
//
//	({T}: Add {W}.)
//	This land enters tapped.
//	{W}, {T}: Put target card from your graveyard on the bottom of your library. Activate only if you control two or more white permanents.
var MistveilPlains = newMistveilPlains

func newMistveilPlains() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:     "Mistveil Plains",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Plains},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{W}, {T}: Put target card from your graveyard on the bottom of your library. Activate only if you control two or more white permanents.",
					ManaCost: opt.Val(cost.Mana{
						cost.W,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					ActivationCondition: opt.Val(game.Condition{
						Text: "only if you control two or more white permanents",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{ColorsAny: []color.Color{color.White}},
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
										Constraint: "target card from your graveyard",
										Allow:      game.TargetAllowCard,
										Selection:  opt.Val(game.Selection{Controller: game.ControllerYou}),
										TargetZone: zone.Graveyard,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind: game.CardReferenceTarget,
											},
											FromZone:          zone.Graveyard,
											Destination:       zone.Library,
											DestinationBottom: true,
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
				game.TapManaAbility(mana.W),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {W}.)
			This land enters tapped.
			{W}, {T}: Put target card from your graveyard on the bottom of your library. Activate only if you control two or more white permanents.
		`,
		},
	}
}
