package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SturdyHatchling is the card definition for Sturdy Hatchling.
//
// Type: Creature — Elemental
// Cost: {3}{G/U}
//
// Oracle text:
//
//	This creature enters with four -1/-1 counters on it.
//	{G/U}: This creature gains shroud until end of turn. (It can't be the target of spells or abilities.)
//	Whenever you cast a green spell, remove a -1/-1 counter from this creature.
//	Whenever you cast a blue spell, remove a -1/-1 counter from this creature.
var SturdyHatchling = newSturdyHatchling

func newSturdyHatchling() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Sturdy Hatchling",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.HybridMana(mana.G, mana.U),
			}),
			Colors:    []color.Color{color.Green, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{G/U}: This creature gains shroud until end of turn. (It can't be the target of spells or abilities.)",
					ManaCost:       opt.Val(cost.Mana{cost.HybridMana(mana.G, mana.U)}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.SourceCardPermanentReference()),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddKeywords: []game.Keyword{
												game.Shroud,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
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
							Event:         game.EventSpellCast,
							Controller:    game.TriggerControllerYou,
							CardSelection: game.Selection{ColorsAny: []color.Color{color.Green}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.RemoveCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.MinusOneMinusOne,
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:         game.EventSpellCast,
							Controller:    game.TriggerControllerYou,
							CardSelection: game.Selection{ColorsAny: []color.Color{color.Blue}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.RemoveCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.MinusOneMinusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with four -1/-1 counters on it.", game.CounterPlacement{Kind: counter.MinusOneMinusOne, Amount: 4}),
			},
			OracleText: `
			This creature enters with four -1/-1 counters on it.
			{G/U}: This creature gains shroud until end of turn. (It can't be the target of spells or abilities.)
			Whenever you cast a green spell, remove a -1/-1 counter from this creature.
			Whenever you cast a blue spell, remove a -1/-1 counter from this creature.
		`,
		},
	}
}
