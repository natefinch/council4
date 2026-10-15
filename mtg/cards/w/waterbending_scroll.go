package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// WaterbendingScroll is the card definition for Waterbending Scroll.
//
// Type: Artifact
// Cost: {1}{U}
//
// Oracle text:
//
//	{6}, {T}: Draw a card. This ability costs {1} less to activate for each Island you control.
var WaterbendingScroll = newWaterbendingScroll

func newWaterbendingScroll() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Waterbending Scroll",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{6}, {T}: Draw a card. This ability costs {1} less to activate for each Island you control.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(6),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					CostModifiers: []game.CostModifier{
						game.CostModifier{
							Kind:               game.CostModifierAbility,
							PerObjectReduction: 1,
							CountSelection:     &game.Selection{SubtypesAny: []types.Sub{types.Island}, Controller: game.ControllerYou},
						},
					},
					ZoneOfFunction: zone.Battlefield,
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
			OracleText: `
			{6}, {T}: Draw a card. This ability costs {1} less to activate for each Island you control.
		`,
		},
	}
}
