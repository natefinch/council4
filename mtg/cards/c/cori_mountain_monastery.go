package c

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

// CoriMountainMonastery is the card definition for Cori Mountain Monastery.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control a Plains or an Island.
//	{T}: Add {R}.
//	{3}{R}, {T}: Exile the top card of your library. Until the end of your next turn, you may play that card.
var CoriMountainMonastery = newCoriMountainMonastery

func newCoriMountainMonastery() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:  "Cori Mountain Monastery",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{3}{R}, {T}: Exile the top card of your library. Until the end of your next turn, you may play that card.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
						cost.R,
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
										Primitive: game.ImpulseExile{
											Player:   game.ControllerReference(),
											Amount:   game.Fixed(1),
											Duration: game.DurationUntilEndOfYourNextTurn,
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
				game.TapManaAbility(mana.R),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless you control a Plains or an Island.", &game.Condition{
					Negate: true,
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{RequiredTypes: []types.Card{types.Land}, SubtypesAny: []types.Sub{types.Sub("Plains"), types.Sub("Island")}},
					}),
				}),
			},
			OracleText: `
			This land enters tapped unless you control a Plains or an Island.
			{T}: Add {R}.
			{3}{R}, {T}: Exile the top card of your library. Until the end of your next turn, you may play that card.
		`,
		},
	}
}
