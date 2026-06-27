package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PalliationAccord is the card definition for Palliation Accord.
//
// Type: Enchantment
// Cost: {3}{W}{U}
//
// Oracle text:
//
//	Whenever a creature an opponent controls becomes tapped, put a palliation counter on this enchantment.
//	Remove a palliation counter from this enchantment: Prevent the next 1 damage that would be dealt to you this turn.
var PalliationAccord = newPalliationAccord()

func newPalliationAccord() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Palliation Accord",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.U,
			}),
			Colors: []color.Color{color.Blue, color.White},
			Types:  []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Remove a palliation counter from this enchantment: Prevent the next 1 damage that would be dealt to you this turn.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove a palliation counter from this enchantment",
							Amount:      1,
							CounterKind: counter.Palliation,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PreventDamage{
									Player: game.ControllerReference(),
									Amount: game.Fixed(1),
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
							Event:            game.EventPermanentTapped,
							Controller:       game.TriggerControllerOpponent,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.Palliation,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever a creature an opponent controls becomes tapped, put a palliation counter on this enchantment.
			Remove a palliation counter from this enchantment: Prevent the next 1 damage that would be dealt to you this turn.
		`,
		},
	}
}
