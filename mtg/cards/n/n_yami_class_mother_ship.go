package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// NYamiClassMotherShip is the card definition for N'Yami-Class Mother Ship.
//
// Type: Artifact — Vehicle
// Cost: {6}
//
// Oracle text:
//
//	Flying, vigilance, haste
//	Whenever this Vehicle deals combat damage to a player, look at the top card of your library. If it's a permanent card, you may put it onto the battlefield. If you don't put it onto the battlefield, put it into your hand.
//	Crew 3
var NYamiClassMotherShip = newNYamiClassMotherShip

func newNYamiClassMotherShip() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "N'Yami-Class Mother Ship",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types:     []types.Card{types.Artifact},
			Subtypes:  []types.Sub{types.Vehicle},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.VigilanceStaticBody,
				game.HasteStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CrewActivatedAbility(3),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:               game.EventDamageDealt,
							Source:              game.TriggerSourceSelf,
							Subject:             game.TriggerSubjectDamageSource,
							RequireCombatDamage: true,
							DamageRecipient:     game.DamageRecipientPlayer,
						},
					},
					Content: game.Mode{
						Text: "Whenever this Vehicle deals combat damage to a player, look at the top card of your library. If it's a permanent card, you may put it onto the battlefield. If you don't put it onto the battlefield, put it into your hand.",
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
										Selection: game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Battle, types.Creature, types.Enchantment, types.Land, types.Planeswalker}},
									}),
									Else: zone.Hand,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying, vigilance, haste
			Whenever this Vehicle deals combat damage to a player, look at the top card of your library. If it's a permanent card, you may put it onto the battlefield. If you don't put it onto the battlefield, put it into your hand.
			Crew 3
		`,
		},
	}
}
