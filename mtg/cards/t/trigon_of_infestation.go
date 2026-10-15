package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TrigonOfInfestation is the card definition for Trigon of Infestation.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	This artifact enters with three charge counters on it.
//	{G}{G}, {T}: Put a charge counter on this artifact.
//	{2}, {T}, Remove a charge counter from this artifact: Create a 1/1 green Phyrexian Insect creature token with infect.
var TrigonOfInfestation = newTrigonOfInfestation

func newTrigonOfInfestation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Trigon of Infestation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{G}{G}, {T}: Put a charge counter on this artifact.",
					ManaCost: opt.Val(cost.Mana{
						cost.G,
						cost.G,
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
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.Charge,
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
					Text: "{2}, {T}, Remove a charge counter from this artifact: Create a 1/1 green Phyrexian Insect creature token with infect.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove a charge counter from this artifact",
							Amount:      1,
							CounterKind: counter.Charge,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(trigonOfInfestationToken),
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
				game.EntersWithCountersReplacement("This artifact enters with three charge counters on it.", game.CounterPlacement{Kind: counter.Charge, Amount: 3}),
			},
			OracleText: `
			This artifact enters with three charge counters on it.
			{G}{G}, {T}: Put a charge counter on this artifact.
			{2}, {T}, Remove a charge counter from this artifact: Create a 1/1 green Phyrexian Insect creature token with infect.
		`,
		},
	}
}

var trigonOfInfestationToken = newTrigonOfInfestationToken()

func newTrigonOfInfestationToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Phyrexian Insect",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Insect},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.InfectStaticBody,
			},
		},
	}
}
