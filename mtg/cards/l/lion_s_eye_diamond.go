package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LionSEyeDiamond is the card definition for Lion's Eye Diamond.
//
// Type: Artifact
// Cost: {0}
//
// Oracle text:
//
//	Discard your hand, Sacrifice this artifact: Add three mana of any one color. Activate only as an instant.
var LionSEyeDiamond = newLionSEyeDiamond

func newLionSEyeDiamond() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Lion's Eye Diamond",
			ManaCost: opt.Val(cost.Mana{
				cost.O(0),
			}),
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "Discard your hand, Sacrifice this artifact: Add three mana of any one color. Activate only as an instant.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:          cost.AdditionalDiscard,
							Text:          "Discard your hand",
							AmountDynamic: cost.AdditionalDynamicHandSize,
							Source:        zone.Hand,
							CounterKind:   counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalSacrificeSource,
							Text:        "Sacrifice this artifact",
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
			Discard your hand, Sacrifice this artifact: Add three mana of any one color. Activate only as an instant.
		`,
		},
	}
}
