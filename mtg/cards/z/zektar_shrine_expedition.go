package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ZektarShrineExpedition is the card definition for Zektar Shrine Expedition.
//
// Type: Enchantment
// Cost: {1}{R}
//
// Oracle text:
//
//	Landfall — Whenever a land you control enters, you may put a quest counter on this enchantment.
//	Remove three quest counters from this enchantment and sacrifice it: Create a 7/1 red Elemental creature token with trample and haste. Exile it at the beginning of the next end step.
var ZektarShrineExpedition = newZektarShrineExpedition

func newZektarShrineExpedition() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Zektar Shrine Expedition",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Remove three quest counters from this enchantment and sacrifice it: Create a 7/1 red Elemental creature token with trample and haste. Exile it at the beginning of the next end step.",
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
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(zektarShrineExpeditionToken),
								},
							},
							{
								Primitive: game.CreateDelayedTrigger{
									Trigger: game.DelayedTriggerDef{
										Timing: game.DelayedAtBeginningOfNextEndStep,
										Content: game.Mode{
											Sequence: []game.Instruction{
												{
													Primitive: game.Exile{
														Object: game.SourceCardPermanentReference(),
													},
												},
											},
										}.Ability(),
									},
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
			Remove three quest counters from this enchantment and sacrifice it: Create a 7/1 red Elemental creature token with trample and haste. Exile it at the beginning of the next end step.
		`,
		},
	}
}

var zektarShrineExpeditionToken = newZektarShrineExpeditionToken()

func newZektarShrineExpeditionToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Elemental",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.HasteStaticBody,
			},
		},
	}
}
