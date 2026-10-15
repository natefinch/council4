package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BloodCelebrant is the card definition for Blood Celebrant.
//
// Type: Creature — Human Cleric
// Cost: {B}
//
// Oracle text:
//
//	{B}, Pay 1 life: Add one mana of any color.
var BloodCelebrant = newBloodCelebrant

func newBloodCelebrant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Blood Celebrant",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Cleric},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "{B}, Pay 1 life: Add one mana of any color.",
					ManaCost: opt.Val(cost.Mana{
						cost.B,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalPayLife,
							Text:        "Pay 1 life",
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
			{B}, Pay 1 life: Add one mana of any color.
		`,
		},
	}
}
