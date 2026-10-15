package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PathOfAncestry is the card definition for Path of Ancestry.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add one mana of any color in your commander's color identity. When that mana is spent to cast a creature spell that shares a creature type with your commander, scry 1. (Look at the top card of your library. You may put that card on the bottom.)
var PathOfAncestry = newPathOfAncestry

func newPathOfAncestry() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Path of Ancestry",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "{T}: Add one mana of any color in your commander's color identity. When that mana is spent to cast a creature spell that shares a creature type with your commander, scry 1. (Look at the top card of your library. You may put that card on the bottom.)",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Choose{
											Choice: game.ResolutionChoice{
												Kind:        game.ResolutionChoiceMana,
												Prompt:      "Choose a color in your commander's color identity",
												ColorSource: game.ResolutionChoiceColorSourceCommanderIdentity,
											},
											PublishChoice: game.ChoiceKey("oracle-commander-color"),
										},
									},
									game.Instruction{
										Primitive: game.AddMana{
											Amount:     game.Fixed(1),
											ChoiceFrom: game.ChoiceKey("oracle-commander-color"),
											SpendRider: opt.Val(game.ManaSpendRider{
												Condition: game.ManaSpendCastCommanderCreatureType,
												Effect: game.Mode{
													Sequence: []game.Instruction{
														game.Instruction{
															Primitive: game.Scry{
																Amount: game.Fixed(1),
																Player: game.ControllerReference(),
															},
														},
													},
												},
											}),
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
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add one mana of any color in your commander's color identity. When that mana is spent to cast a creature spell that shares a creature type with your commander, scry 1. (Look at the top card of your library. You may put that card on the bottom.)
		`,
		},
	}
}
