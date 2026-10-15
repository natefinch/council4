package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GemstoneArray is the card definition for Gemstone Array.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	{2}: Put a charge counter on this artifact.
//	Remove a charge counter from this artifact: Add one mana of any color.
var GemstoneArray = newGemstoneArray

func newGemstoneArray() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Gemstone Array",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}: Put a charge counter on this artifact.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.Charge,
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
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "Remove a charge counter from this artifact: Add one mana of any color.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove a charge counter from this artifact",
							Amount:      1,
							CounterKind: counter.Charge,
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
			{2}: Put a charge counter on this artifact.
			Remove a charge counter from this artifact: Add one mana of any color.
		`,
		},
	}
}
