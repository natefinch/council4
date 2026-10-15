package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GlistenerSeer is the card definition for Glistener Seer.
//
// Type: Creature — Phyrexian Advisor
// Cost: {U}
//
// Oracle text:
//
//	This creature enters with three oil counters on it.
//	{T}, Remove an oil counter from this creature: Scry 1. (Look at the top card of your library. You may put that card on the bottom.)
var GlistenerSeer = newGlistenerSeer

func newGlistenerSeer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Glistener Seer",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Advisor},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}, Remove an oil counter from this creature: Scry 1. (Look at the top card of your library. You may put that card on the bottom.)",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove an oil counter from this creature",
							Amount:      1,
							CounterKind: counter.Oil,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Scry{
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with three oil counters on it.", game.CounterPlacement{Kind: counter.Oil, Amount: 3}),
			},
			OracleText: `
			This creature enters with three oil counters on it.
			{T}, Remove an oil counter from this creature: Scry 1. (Look at the top card of your library. You may put that card on the bottom.)
		`,
		},
	}
}
