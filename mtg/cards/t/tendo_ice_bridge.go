package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// TendoIceBridge is the card definition for Tendo Ice Bridge.
//
// Type: Land
//
// Oracle text:
//
//	This land enters with a charge counter on it.
//	{T}: Add {C}.
//	{T}, Remove a charge counter from this land: Add one mana of any color.
var TendoIceBridge = newTendoIceBridge

func newTendoIceBridge() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Tendo Ice Bridge",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
				game.ManaAbility{
					Text: "{T}, Remove a charge counter from this land: Add one mana of any color.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove a charge counter from this land",
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This land enters with a charge counter on it.", game.CounterPlacement{Kind: counter.Charge, Amount: 1}),
			},
			OracleText: `
			This land enters with a charge counter on it.
			{T}: Add {C}.
			{T}, Remove a charge counter from this land: Add one mana of any color.
		`,
		},
	}
}
