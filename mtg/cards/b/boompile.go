package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Boompile is the card definition for Boompile.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	{T}: Flip a coin. If you win the flip, destroy all nonland permanents.
var Boompile = newBoompile

func newBoompile() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Boompile",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}: Flip a coin. If you win the flip, destroy all nonland permanents.",
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
										Primitive: game.RollDie{
											Sides: 2,
										},
										PublishResult: game.ResultKey("coin-flip-result"),
									},
									game.Instruction{
										Primitive: game.Destroy{
											Group: game.BattlefieldGroup(game.Selection{ExcludedTypes: []types.Card{types.Land}}),
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key: game.ResultKey("coin-flip-result"),
											AmountRange: opt.Val(game.IntRange{
												Min: 2,
												Max: 2,
											}),
										}),
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
			{T}: Flip a coin. If you win the flip, destroy all nonland permanents.
		`,
		},
	}
}
