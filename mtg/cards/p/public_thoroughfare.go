package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PublicThoroughfare is the card definition for Public Thoroughfare.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped.
//	When this land enters, sacrifice it unless you tap an untapped artifact or land you control.
//	{T}: Add one mana of any color.
var PublicThoroughfare = newPublicThoroughfare

func newPublicThoroughfare() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Public Thoroughfare",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Pay{
									Payment: game.ResolutionPayment{
										Prompt: "Tap an untapped artifact or land you control?",
										AdditionalCosts: []cost.Additional{
											{
												Kind:               cost.AdditionalTapPermanents,
												Text:               "tap an untapped artifact or land you control",
												Amount:             1,
												MatchPermanentType: true,
												PermanentType:      types.Artifact,
												PermanentTypeAlt:   types.Land,
											},
										},
									},
								},
								PublishResult: game.ResultKey("sacrifice-unless-paid"),
							},
							{
								Primitive: game.Sacrifice{
									Object: game.EventPermanentReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "sacrifice-unless-paid",
									Succeeded: game.TriFalse,
								}),
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			When this land enters, sacrifice it unless you tap an untapped artifact or land you control.
			{T}: Add one mana of any color.
		`,
		},
	}
}
