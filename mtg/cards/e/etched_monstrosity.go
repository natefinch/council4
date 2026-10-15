package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EtchedMonstrosity is the card definition for Etched Monstrosity.
//
// Type: Artifact Creature — Phyrexian Golem
// Cost: {5}
//
// Oracle text:
//
//	This creature enters with five -1/-1 counters on it.
//	{W}{U}{B}{R}{G}, Remove five -1/-1 counters from this creature: Target player draws three cards.
var EtchedMonstrosity = newEtchedMonstrosity

func newEtchedMonstrosity() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Etched Monstrosity",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Golem},
			Power:     opt.Val(game.PT{Value: 10}),
			Toughness: opt.Val(game.PT{Value: 10}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{W}{U}{B}{R}{G}, Remove five -1/-1 counters from this creature: Target player draws three cards.",
					ManaCost: opt.Val(cost.Mana{
						cost.W,
						cost.U,
						cost.B,
						cost.R,
						cost.G,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove five -1/-1 counters from this creature",
							Amount:      5,
							CounterKind: counter.MinusOneMinusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
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
										Primitive: game.Draw{
											Amount: game.Fixed(3),
											Player: game.TargetPlayerReference(0),
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
				game.EntersWithCountersReplacement("This creature enters with five -1/-1 counters on it.", game.CounterPlacement{Kind: counter.MinusOneMinusOne, Amount: 5}),
			},
			OracleText: `
			This creature enters with five -1/-1 counters on it.
			{W}{U}{B}{R}{G}, Remove five -1/-1 counters from this creature: Target player draws three cards.
		`,
		},
	}
}
