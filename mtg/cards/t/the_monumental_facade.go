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

// TheMonumentalFacade is the card definition for The Monumental Facade.
//
// Type: Land — Sphere
//
// Oracle text:
//
//	This land enters with two oil counters on it.
//	{T}: Add {C}.
//	{T}, Remove an oil counter from this land: Put an oil counter on target artifact or creature you control. Activate only as a sorcery.
var TheMonumentalFacade = newTheMonumentalFacade

func newTheMonumentalFacade() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "The Monumental Facade",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Sphere},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}, Remove an oil counter from this land: Put an oil counter on target artifact or creature you control. Activate only as a sorcery.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove an oil counter from this land",
							Amount:      1,
							CounterKind: counter.Oil,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target artifact or creature you control",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature}, Controller: game.ControllerYou}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.TargetPermanentReference(0),
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
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This land enters with two oil counters on it.", game.CounterPlacement{Kind: counter.Oil, Amount: 2}),
			},
			OracleText: `
			This land enters with two oil counters on it.
			{T}: Add {C}.
			{T}, Remove an oil counter from this land: Put an oil counter on target artifact or creature you control. Activate only as a sorcery.
		`,
		},
	}
}
