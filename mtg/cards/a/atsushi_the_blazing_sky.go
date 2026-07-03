package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AtsushiTheBlazingSky is the card definition for Atsushi, the Blazing Sky.
//
// Type: Legendary Creature — Dragon Spirit
// Cost: {2}{R}{R}
//
// Oracle text:
//
//	Flying, trample
//	When Atsushi dies, choose one —
//	• Exile the top two cards of your library. Until the end of your next turn, you may play those cards.
//	• Create three Treasure tokens.
var AtsushiTheBlazingSky = newAtsushiTheBlazingSky()

func newAtsushiTheBlazingSky() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Atsushi, the Blazing Sky",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Dragon, types.Spirit},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.TrampleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Exile the top two cards of your library. Until the end of your next turn, you may play those cards.",
								Sequence: []game.Instruction{
									{
										Primitive: game.ImpulseExile{
											Player:   game.ControllerReference(),
											Amount:   game.Fixed(2),
											Duration: game.DurationUntilEndOfYourNextTurn,
										},
									},
								},
							},
							game.Mode{
								Text: "Create three Treasure tokens.",
								Sequence: []game.Instruction{
									{
										Primitive: game.CreateToken{
											Amount: game.Fixed(3),
											Source: game.TokenDef(atsushiTheBlazingSkyToken),
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
			Flying, trample
			When Atsushi dies, choose one —
			• Exile the top two cards of your library. Until the end of your next turn, you may play those cards.
			• Create three Treasure tokens.
		`,
		},
	}
}

var atsushiTheBlazingSkyToken = newAtsushiTheBlazingSkyToken()

func newAtsushiTheBlazingSkyToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Treasure",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Treasure},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:               cost.AdditionalSacrificeSource,
							Text:               "Sacrifice this artifact",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Artifact,
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
		},
	}
}
