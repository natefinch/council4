package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// IdolOfOblivion is the card definition for Idol of Oblivion.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	{T}: Draw a card. Activate only if you created a token this turn.
//	{8}, {T}, Sacrifice this artifact: Create a 10/10 colorless Eldrazi creature token.
var IdolOfOblivion = newIdolOfOblivion

func newIdolOfOblivion() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Idol of Oblivion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}: Draw a card. Activate only if you created a token this turn.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					ActivationCondition: opt.Val(game.Condition{
						Text:                           "only if you created a token this turn",
						ControllerCreatedTokenThisTurn: true,
						SourceCounterKind:              counter.PlusOnePlusOne,
					}),
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
				game.ActivatedAbility{
					Text: "{8}, {T}, Sacrifice this artifact: Create a 10/10 colorless Eldrazi creature token.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(8),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalSacrificeSource,
							Text:        "Sacrifice this artifact",
							Amount:      1,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(idolOfOblivionToken),
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
			{T}: Draw a card. Activate only if you created a token this turn.
			{8}, {T}, Sacrifice this artifact: Create a 10/10 colorless Eldrazi creature token.
		`,
		},
	}
}

var idolOfOblivionToken = newIdolOfOblivionToken()

func newIdolOfOblivionToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Eldrazi",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi},
			Power:     opt.Val(game.PT{Value: 10}),
			Toughness: opt.Val(game.PT{Value: 10}),
		},
	}
}
