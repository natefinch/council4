package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DarigaazSCaldera is the card definition for Darigaaz's Caldera.
//
// Type: Land — Lair
//
// Oracle text:
//
//	When this land enters, sacrifice it unless you return a non-Lair land you control to its owner's hand.
//	{T}: Add {B}, {R}, or {G}.
var DarigaazSCaldera = newDarigaazSCaldera

func newDarigaazSCaldera() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name:     "Darigaaz's Caldera",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Lair},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.B, mana.R, mana.G),
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
										Prompt: "Return a non-Lair land you control to its owner's hand?",
										AdditionalCosts: []cost.Additional{
											{
												Kind:               cost.AdditionalReturnToHand,
												Text:               "return a non-Lair land you control to its owner's hand",
												Amount:             1,
												MatchPermanentType: true,
												PermanentType:      types.Land,
												ExcludeSubtype:     types.Lair,
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
			OracleText: `
			When this land enters, sacrifice it unless you return a non-Lair land you control to its owner's hand.
			{T}: Add {B}, {R}, or {G}.
		`,
		},
	}
}
