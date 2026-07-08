package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BorborygmosEnraged is the card definition for Borborygmos Enraged.
//
// Type: Legendary Creature — Cyclops
// Cost: {4}{R}{R}{G}{G}
//
// Oracle text:
//
//	Trample
//	Whenever Borborygmos Enraged deals combat damage to a player, reveal the top three cards of your library. Put all land cards revealed this way into your hand and the rest into your graveyard.
//	Discard a land card: Borborygmos Enraged deals 3 damage to any target.
var BorborygmosEnraged = newBorborygmosEnraged

func newBorborygmosEnraged() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Borborygmos Enraged",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
				cost.R,
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green, color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Cyclops},
			Power:      opt.Val(game.PT{Value: 7}),
			Toughness:  opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Discard a land card: Borborygmos Enraged deals 3 damage to any target.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:          cost.AdditionalDiscard,
							Text:          "Discard a land card",
							Amount:        1,
							Source:        zone.Hand,
							MatchCardType: true,
							CardType:      types.Land,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "any target",
								Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(3),
									Recipient: game.AnyTargetDamageRecipient(0),
								},
							},
						},
					}.Ability(),
				},
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
						Sequence: []game.Instruction{
							{
								Primitive: game.RevealTopPartition{
									Player:    game.ControllerReference(),
									Amount:    game.Fixed(3),
									Selection: game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Trample
			Whenever Borborygmos Enraged deals combat damage to a player, reveal the top three cards of your library. Put all land cards revealed this way into your hand and the rest into your graveyard.
			Discard a land card: Borborygmos Enraged deals 3 damage to any target.
		`,
		},
	}
}
