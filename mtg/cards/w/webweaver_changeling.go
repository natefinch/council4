package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WebweaverChangeling is the card definition for Webweaver Changeling.
//
// Type: Creature — Shapeshifter
// Cost: {3}{G}{G}
//
// Oracle text:
//
//	Changeling (This card is every creature type.)
//	Reach
//	When this creature enters, if there are three or more creature cards in your graveyard, you gain 5 life.
var WebweaverChangeling = newWebweaverChangeling

func newWebweaverChangeling() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Webweaver Changeling",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.ChangelingStaticBody,
				game.ReachStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, if there are three or more creature cards in your graveyard, you gain 5 life.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
						InterveningIf: "if there are three or more creature cards in your graveyard",
						InterveningCondition: opt.Val(game.Condition{
							Text:              "if there are three or more creature cards in your graveyard",
							SourceCounterKind: counter.PlusOnePlusOne,
							ControllerGraveyardCardOfTypeCountAtLeast: 3,
							ControllerGraveyardCountCardType:          types.Creature,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(5),
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
			Changeling (This card is every creature type.)
			Reach
			When this creature enters, if there are three or more creature cards in your graveyard, you gain 5 life.
		`,
		},
	}
}
