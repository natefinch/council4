package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// IorRuinExpedition is the card definition for Ior Ruin Expedition.
//
// Type: Enchantment
// Cost: {1}{U}
//
// Oracle text:
//
//	Landfall — Whenever a land you control enters, you may put a quest counter on this enchantment.
//	Remove three quest counters from this enchantment and sacrifice it: Draw two cards.
var IorRuinExpedition = newIorRuinExpedition()

func newIorRuinExpedition() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Ior Ruin Expedition",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Remove three quest counters from this enchantment and sacrifice it: Draw two cards.",
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
								Primitive: game.Draw{
									Amount: game.Fixed(2),
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
			Remove three quest counters from this enchantment and sacrifice it: Draw two cards.
		`,
		},
	}
}
