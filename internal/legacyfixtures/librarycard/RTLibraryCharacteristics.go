// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RTLibraryCharacteristics is the card definition for RT Library Characteristics.
//
// Type: Sorcery
//
// Oracle text:
//
//	Reveal the top card of your library. If it's a creature card, you draw cards equal to its power and you gain life equal to its toughness.
var RTLibraryCharacteristics = newRTLibraryCharacteristics

func newRTLibraryCharacteristics() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Library Characteristics",
			Types: []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Reveal{
							Amount:        game.Fixed(1),
							Player:        game.ControllerReference(),
							PublishLinked: game.LinkedKey("sequence-effect-0-product"),
							PublishCharacteristics: game.LibraryCardCharacteristics{
								Power:     game.ResultKey("sequence-effect-0-power"),
								Toughness: game.ResultKey("sequence-effect-0-toughness"),
							},
						},
						LocalProducts: game.LocalProducts{
							Results: []game.ResultKey{game.ResultKey("sequence-effect-0-power"), game.ResultKey("sequence-effect-0-toughness")},
							Links:   []game.LinkedKey{game.LinkedKey("sequence-effect-0-product")},
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountPreviousEffectResult,
								Multiplier: 1,
								ResultKey:  game.ResultKey("sequence-effect-0-power"),
							}),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Object:        opt.Val(game.LinkedObjectReference("sequence-effect-0-product")),
								ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							}),
						}),
						PublishCondition: "condition-0",
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:             "sequence-effect-0-power",
							AmountAvailable: true,
						}),
					},
					{
						Primitive: game.GainLife{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountPreviousEffectResult,
								Multiplier: 1,
								ResultKey:  game.ResultKey("sequence-effect-0-toughness"),
							}),
							Player: game.ControllerReference(),
						},
						ConditionGate: "condition-0",
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:             "sequence-effect-0-toughness",
							AmountAvailable: true,
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Reveal the top card of your library. If it's a creature card, you draw cards equal to its power and you gain life equal to its toughness.
		`,
		},
	}
}
