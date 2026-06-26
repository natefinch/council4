package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LeadBellyChimera is the card definition for Lead-Belly Chimera.
//
// Type: Artifact Creature — Chimera
// Cost: {4}
//
// Oracle text:
//
//	Trample
//	Sacrifice this creature: Put a +2/+2 counter on target Chimera creature. It gains trample. (This effect lasts indefinitely.)
var LeadBellyChimera = newLeadBellyChimera()

func newLeadBellyChimera() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Lead-Belly Chimera",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Chimera},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Sacrifice this creature: Put a +2/+2 counter on target Chimera creature. It gains trample. (This effect lasts indefinitely.)",
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this creature",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target Chimera creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Sub("Chimera")}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.TargetPermanentReference(0),
									CounterKind: counter.PlusTwoPlusTwo,
								},
							},
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.TargetPermanentReference(0)),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddKeywords: []game.Keyword{
												game.Trample,
											},
										},
									},
									Duration: game.DurationPermanent,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Trample
			Sacrifice this creature: Put a +2/+2 counter on target Chimera creature. It gains trample. (This effect lasts indefinitely.)
		`,
		},
	}
}
