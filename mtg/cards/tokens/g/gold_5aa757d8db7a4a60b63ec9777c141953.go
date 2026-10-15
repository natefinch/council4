package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Gold is the card definition for Gold.
//
// Type: Token Artifact — Gold
//
// Oracle text:
//   Sacrifice this token: Add one mana of any color.

// GoldToken5aa757d8db7a4a60b63ec9777c141953 is the card definition for Gold.
var GoldToken5aa757d8db7a4a60b63ec9777c141953 = newGoldToken5aa757d8db7a4a60b63ec9777c141953()

func newGoldToken5aa757d8db7a4a60b63ec9777c141953() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Gold",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Gold},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "Sacrifice this token: Add one mana of any color.",
					AdditionalCosts: []cost.Additional{
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
											Amount:     game.Fixed(1),
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
			Sacrifice this token: Add one mana of any color.
		`,
		},
	}
}
