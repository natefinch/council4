package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ExplorerSScope is the card definition for Explorer's Scope.
//
// Type: Artifact — Equipment
// Cost: {1}
//
// Oracle text:
//
//	Whenever equipped creature attacks, look at the top card of your library. If it's a land card, you may put it onto the battlefield tapped.
//	Equip {1} ({1}: Attach to target creature you control. Equip only as a sorcery.)
var ExplorerSScope = newExplorerSScope

func newExplorerSScope() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Explorer's Scope",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(1)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventAttackerDeclared,
							Source:           game.TriggerSourceAttachedPermanent,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Text: "Whenever equipped creature attacks, look at the top card of your library. If it's a land card, you may put it onto the battlefield tapped.",
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
			Whenever equipped creature attacks, look at the top card of your library. If it's a land card, you may put it onto the battlefield tapped.
			Equip {1} ({1}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		},
	}
}
