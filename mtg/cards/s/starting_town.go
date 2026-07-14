package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// StartingTown is the card definition for Starting Town.
//
// Type: Land — Town
//
// Oracle text:
//
//	This land enters tapped unless it's your first, second, or third turn of the game.
//	{T}: Add {C}.
//	{T}, Pay 1 life: Add one mana of any color.
var StartingTown = newStartingTown

func newStartingTown() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Starting Town",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Town},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
				game.ManaAbility{
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:   cost.AdditionalPayLife,
							Text:   "Pay 1 life",
							Amount: 1,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Choose{
									Choice: game.ResolutionChoice{
										Kind:   game.ResolutionChoiceMana,
										Prompt: "Choose a color",
										Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
									},
									PublishChoice: game.ChoiceKey("oracle-mana-color"),
								},
							},
							{
								Primitive: game.AddMana{
									Amount:     game.Fixed(1),
									ChoiceFrom: game.ChoiceKey("oracle-mana-color"),
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless it's your first, second, or third turn of the game.", &game.Condition{
					Negate:                     true,
					ControllerTurnOfGameAtMost: 3,
				}),
			},
			OracleText: `
			This land enters tapped unless it's your first, second, or third turn of the game.
			{T}: Add {C}.
			{T}, Pay 1 life: Add one mana of any color.
		`,
		},
	}
}
