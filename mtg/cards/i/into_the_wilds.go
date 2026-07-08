package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// IntoTheWilds is the card definition for Into the Wilds.
//
// Type: Enchantment
// Cost: {3}{G}
//
// Oracle text:
//
//	At the beginning of your upkeep, look at the top card of your library. If it's a land card, you may put it onto the battlefield.
var IntoTheWilds = newIntoTheWilds

func newIntoTheWilds() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Into the Wilds",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
						},
					},
					Content: game.Mode{
						Text: "At the beginning of your upkeep, look at the top card of your library. If it's a land card, you may put it onto the battlefield.",
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
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			At the beginning of your upkeep, look at the top card of your library. If it's a land card, you may put it onto the battlefield.
		`,
		},
	}
}
