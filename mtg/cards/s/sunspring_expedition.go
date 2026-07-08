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

// SunspringExpedition is the card definition for Sunspring Expedition.
//
// Type: Enchantment
// Cost: {W}
//
// Oracle text:
//
//	Landfall — Whenever a land you control enters, you may put a quest counter on this enchantment.
//	Remove three quest counters from this enchantment and sacrifice it: You gain 8 life.
var SunspringExpedition = newSunspringExpedition

func newSunspringExpedition() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Sunspring Expedition",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Remove three quest counters from this enchantment and sacrifice it: You gain 8 life.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove three quest counters from this enchantment",
							Amount:      3,
							CounterKind: counter.Quest,
						},
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "sacrifice it",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(8),
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
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Land}},
						},
					},
					Optional: true,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.Quest,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Landfall — Whenever a land you control enters, you may put a quest counter on this enchantment.
			Remove three quest counters from this enchantment and sacrifice it: You gain 8 life.
		`,
		},
	}
}
