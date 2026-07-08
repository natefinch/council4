package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RisenReef is the card definition for Risen Reef.
//
// Type: Creature — Elemental
// Cost: {1}{G}{U}
//
// Oracle text:
//
//	Whenever this creature or another Elemental you control enters, look at the top card of your library. If it's a land card, you may put it onto the battlefield tapped. If you don't put the card onto the battlefield, put it into your hand.
var RisenReef = newRisenReef

func newRisenReef() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Risen Reef",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.U,
			}),
			Colors:    []color.Color{color.Green, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                  game.EventPermanentEnteredBattlefield,
							Controller:             game.TriggerControllerYou,
							SubjectSelectionOrSelf: true,
							SubjectSelection:       game.Selection{SubtypesAny: []types.Sub{types.Sub("Elemental")}},
						},
					},
					Content: game.Mode{
						Text: "Whenever this creature or another Elemental you control enters, look at the top card of your library. If it's a land card, you may put it onto the battlefield tapped. If you don't put the card onto the battlefield, put it into your hand.",
						Sequence: []game.Instruction{
							{
								Primitive: game.LookAtLibraryTop{
									Player:        game.ControllerReference(),
									PublishLinked: game.LinkedKey("look-at-top-battlefield-card"),
								},
							},
							{
								Primitive: game.ConditionalDestinationPlace{
									Card:     game.CardReference{Kind: game.CardReferenceLinked, LinkID: "look-at-top-battlefield-card"},
									FromZone: zone.Library,
									CardCondition: opt.Val(game.CardSelection{
										Card:      game.CardReference{Kind: game.CardReferenceLinked, LinkID: "look-at-top-battlefield-card"},
										Selection: game.Selection{RequiredTypesAny: []types.Card{types.Land}},
									}),
									EntryTapped: true,
									Else:        zone.Hand,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever this creature or another Elemental you control enters, look at the top card of your library. If it's a land card, you may put it onto the battlefield tapped. If you don't put the card onto the battlefield, put it into your hand.
		`,
		},
	}
}
