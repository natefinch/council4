package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StarCompass is the card definition for Star Compass.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	This artifact enters tapped.
//	{T}: Add one mana of any color that a basic land you control could produce.
var StarCompass = newStarCompass

func newStarCompass() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Star Compass",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "{T}: Add one mana of any color that a basic land you control could produce.",
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
												Kind:           game.ResolutionChoiceMana,
												Prompt:         "Choose a color a land you control could produce",
												ColorSource:    game.ResolutionChoiceColorSourceLandsProduce,
												PlayerRelation: game.PlayerYou,
												Selection:      &game.Selection{Supertypes: []types.Super{types.Basic}},
											},
											PublishChoice: game.ChoiceKey("oracle-lands-produce-color"),
										},
									},
									game.Instruction{
										Primitive: game.AddMana{
											Amount:     game.Fixed(1),
											ChoiceFrom: game.ChoiceKey("oracle-lands-produce-color"),
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
				game.EntersTappedReplacement("This artifact enters tapped."),
			},
			OracleText: `
			This artifact enters tapped.
			{T}: Add one mana of any color that a basic land you control could produce.
		`,
		},
	}
}
