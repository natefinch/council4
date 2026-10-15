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

// TavernSwindler is the card definition for Tavern Swindler.
//
// Type: Creature — Human Rogue
// Cost: {1}{B}
//
// Oracle text:
//
//	{T}, Pay 3 life: Flip a coin. If you win the flip, you gain 6 life.
var TavernSwindler = newTavernSwindler

func newTavernSwindler() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Tavern Swindler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Rogue},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}, Pay 3 life: Flip a coin. If you win the flip, you gain 6 life.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalPayLife,
							Text:        "Pay 3 life",
							Amount:      3,
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
										Primitive: game.GainLife{
											Amount: game.Fixed(6),
											Player: game.ControllerReference(),
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
			{T}, Pay 3 life: Flip a coin. If you win the flip, you gain 6 life.
		`,
		},
	}
}
