package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Lotus is the card definition for Lotus.
//
// Type: Token Artifact
//
// Oracle text:
//   {T}, Sacrifice this token: Add three mana of any one color.

// LotusToken9eafaa49cf0b470db8a3c51cff3d26e0 is the card definition for Lotus.
var LotusToken9eafaa49cf0b470db8a3c51cff3d26e0 = newLotusToken9eafaa49cf0b470db8a3c51cff3d26e0()

func newLotusToken9eafaa49cf0b470db8a3c51cff3d26e0() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Lotus",
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "{T}, Sacrifice this token: Add three mana of any one color.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalSacrificeSource,
							Text:        "Sacrifice this token",
							Amount:      1,
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
												Kind:   game.ResolutionChoiceMana,
												Prompt: "Choose a color",
												Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
											},
											PublishChoice: game.ChoiceKey("oracle-mana-color"),
										},
									},
									game.Instruction{
										Primitive: game.AddMana{
											Amount:     game.Fixed(3),
											ChoiceFrom: game.ChoiceKey("oracle-mana-color"),
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
			{T}, Sacrifice this token: Add three mana of any one color.
		`,
		},
	}
}
