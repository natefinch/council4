package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SarinthSteelseeker is the card definition for Sarinth Steelseeker.
//
// Type: Creature — Human Artificer Scout
// Cost: {1}{G}
//
// Oracle text:
//
//	Whenever an artifact you control enters, look at the top card of your library. If it's a land card, you may reveal it and put it into your hand. If you don't put the card into your hand, you may put it into your graveyard.
var SarinthSteelseeker = newSarinthSteelseeker

func newSarinthSteelseeker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Sarinth Steelseeker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Artificer, types.Scout},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever an artifact you control enters, look at the top card of your library. If it's a land card, you may reveal it and put it into your hand. If you don't put the card into your hand, you may put it into your graveyard.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Whenever an artifact you control enters, look at the top card of your library. If it's a land card, you may reveal it and put it into your hand. If you don't put the card into your hand, you may put it into your graveyard.",
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.LookAtLibraryTop{
											Player:        game.ControllerReference(),
											PublishLinked: game.LinkedKey("look-at-top-card"),
										},
									},
									game.Instruction{
										Primitive: game.ConditionalDestinationPlace{
											Card: game.CardReference{
												Kind:   game.CardReferenceLinked,
												LinkID: "look-at-top-card",
											},
											FromZone: zone.Library,
											CardCondition: opt.Val(game.CardSelection{
												Card: game.CardReference{
													Kind:   game.CardReferenceLinked,
													LinkID: "look-at-top-card",
												},
												Selection: game.Selection{RequiredTypesAny: []types.Card{types.Land}},
											}),
											Then:         zone.Hand,
											ThenReveal:   true,
											Else:         zone.Graveyard,
											ElseOptional: true,
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
			Whenever an artifact you control enters, look at the top card of your library. If it's a land card, you may reveal it and put it into your hand. If you don't put the card into your hand, you may put it into your graveyard.
		`,
		},
	}
}
