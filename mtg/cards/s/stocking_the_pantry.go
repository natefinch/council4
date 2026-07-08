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

// StockingThePantry is the card definition for Stocking the Pantry.
//
// Type: Enchantment
// Cost: {G}
//
// Oracle text:
//
//	Whenever you put one or more +1/+1 counters on a creature you control, put a supply counter on this enchantment.
//	{2}, Remove a supply counter from this enchantment: Draw a card.
var StockingThePantry = newStockingThePantry

func newStockingThePantry() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Stocking the Pantry",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{2}, Remove a supply counter from this enchantment: Draw a card.",
					ManaCost: opt.Val(cost.Mana{cost.O(2)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove a supply counter from this enchantment",
							Amount:      1,
							CounterKind: counter.Supply,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
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
							Event:            game.EventCountersAdded,
							Controller:       game.TriggerControllerYou,
							CauseController:  game.TriggerControllerYou,
							OneOrMore:        true,
							MatchCounterKind: true,
							CounterKind:      counter.PlusOnePlusOne,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.Supply,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever you put one or more +1/+1 counters on a creature you control, put a supply counter on this enchantment.
			{2}, Remove a supply counter from this enchantment: Draw a card.
		`,
		},
	}
}
