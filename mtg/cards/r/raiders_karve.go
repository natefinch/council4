package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RaidersKarve is the card definition for Raiders' Karve.
//
// Type: Artifact — Vehicle
// Cost: {3}
//
// Oracle text:
//
//	Whenever this Vehicle attacks, look at the top card of your library. If it's a land card, you may put it onto the battlefield tapped.
//	Crew 3 (Tap any number of creatures you control with total power 3 or more: This Vehicle becomes an artifact creature until end of turn.)
var RaidersKarve = newRaidersKarve

func newRaidersKarve() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Raiders' Karve",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact},
			Subtypes:  []types.Sub{types.Vehicle},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.CrewActivatedAbility(3),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Text: "Whenever this Vehicle attacks, look at the top card of your library. If it's a land card, you may put it onto the battlefield tapped.",
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
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever this Vehicle attacks, look at the top card of your library. If it's a land card, you may put it onto the battlefield tapped.
			Crew 3 (Tap any number of creatures you control with total power 3 or more: This Vehicle becomes an artifact creature until end of turn.)
		`,
		},
	}
}
