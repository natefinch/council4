package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TamiyoSNotebook is the card definition for Tamiyo's Notebook.
//
// Type: Token Legendary Artifact
//
// Oracle text:
//   Spells you cast cost {2} less to cast.
//   {T}: Draw a card.

// TamiyoSNotebookToken94de5091d6ba43c495b6cec223be2fe1 is the card definition for Tamiyo's Notebook.
var TamiyoSNotebookToken94de5091d6ba43c495b6cec223be2fe1 = newTamiyoSNotebookToken94de5091d6ba43c495b6cec223be2fe1()

func newTamiyoSNotebookToken94de5091d6ba43c495b6cec223be2fe1() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:       "Tamiyo's Notebook",
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Spells you cast cost {2} less to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 2,
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}: Draw a card.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
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
			Spells you cast cost {2} less to cast.
			{T}: Draw a card.
		`,
		},
	}
}
