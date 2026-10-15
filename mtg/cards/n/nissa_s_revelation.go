package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NissaSRevelation is the card definition for Nissa's Revelation.
//
// Type: Sorcery
// Cost: {5}{G}{G}
//
// Oracle text:
//
//	Scry 5, then reveal the top card of your library. If it's a creature card, you draw cards equal to its power and you gain life equal to its toughness.
var NissaSRevelation = newNissaSRevelation

func newNissaSRevelation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Nissa's Revelation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Scry{
							Amount: game.Fixed(5),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.Reveal{
							Amount:        game.Fixed(1),
							Player:        game.ControllerReference(),
							PublishLinked: game.LinkedKey("sequence-effect-1-product"),
							PublishCharacteristics: game.LibraryCardCharacteristics{
								Power:     game.ResultKey("sequence-effect-1-power"),
								Toughness: game.ResultKey("sequence-effect-1-toughness"),
							},
						},
						LocalProducts: game.LocalProducts{
							Results: []game.ResultKey{game.ResultKey("sequence-effect-1-power"), game.ResultKey("sequence-effect-1-toughness")},
							Links:   []game.LinkedKey{game.LinkedKey("sequence-effect-1-product")},
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountPreviousEffectResult,
								Multiplier: 1,
								ResultKey:  game.ResultKey("sequence-effect-1-power"),
							}),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Object:        opt.Val(game.LinkedObjectReference("sequence-effect-1-product")),
								ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							}),
						}),
						PublishCondition: "condition-0",
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:             "sequence-effect-1-power",
							AmountAvailable: true,
						}),
					},
					{
						Primitive: game.GainLife{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountPreviousEffectResult,
								Multiplier: 1,
								ResultKey:  game.ResultKey("sequence-effect-1-toughness"),
							}),
							Player: game.ControllerReference(),
						},
						ConditionGate: "condition-0",
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:             "sequence-effect-1-toughness",
							AmountAvailable: true,
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Scry 5, then reveal the top card of your library. If it's a creature card, you draw cards equal to its power and you gain life equal to its toughness.
		`,
		},
	}
}
