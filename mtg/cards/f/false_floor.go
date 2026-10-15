package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FalseFloor is the card definition for False Floor.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	This artifact enters tapped.
//	Creatures enter tapped.
//	{2}, {T}, Exile this artifact: Exile all untapped creatures. Activate only as a sorcery.
var FalseFloor = newFalseFloor

func newFalseFloor() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "False Floor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}, {T}, Exile this artifact: Exile all untapped creatures. Activate only as a sorcery.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalExileSource,
							Text:        "Exile this artifact",
							Amount:      1,
							Source:      zone.Battlefield,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MovePermanent{
											Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Tapped: game.TriFalse}),
											Destination: zone.Exile,
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
				game.EntersTappedReplacement("This artifact enters tapped."),
				game.EntersTappedGroupReplacement("Creatures enter tapped.", game.TriggerControllerAny, types.Creature),
			},
			OracleText: `
			This artifact enters tapped.
			Creatures enter tapped.
			{2}, {T}, Exile this artifact: Exile all untapped creatures. Activate only as a sorcery.
		`,
		},
	}
}
