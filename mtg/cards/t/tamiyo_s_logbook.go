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

// TamiyoSLogbook is the card definition for Tamiyo's Logbook.
//
// Type: Artifact — Book
// Cost: {2}{U}
//
// Oracle text:
//
//	{5}{U}, {T}: Draw a card. This ability costs {1} less to activate for each other artifact you control.
var TamiyoSLogbook = newTamiyoSLogbook

func newTamiyoSLogbook() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Tamiyo's Logbook",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Book},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{5}{U}, {T}: Draw a card. This ability costs {1} less to activate for each other artifact you control.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(5),
						cost.U,
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
							CountSelection:     &game.Selection{RequiredTypes: []types.Card{types.Artifact}, ExcludeSource: true, Controller: game.ControllerYou},
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
			{5}{U}, {T}: Draw a card. This ability costs {1} less to activate for each other artifact you control.
		`,
		},
	}
}
