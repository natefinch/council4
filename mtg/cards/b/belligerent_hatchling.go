package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BelligerentHatchling is the card definition for Belligerent Hatchling.
//
// Type: Creature — Elemental
// Cost: {3}{R/W}
//
// Oracle text:
//
//	First strike
//	This creature enters with four -1/-1 counters on it.
//	Whenever you cast a red spell, remove a -1/-1 counter from this creature.
//	Whenever you cast a white spell, remove a -1/-1 counter from this creature.
var BelligerentHatchling = newBelligerentHatchling()

func newBelligerentHatchling() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Belligerent Hatchling",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.HybridMana(mana.R, mana.W),
			}),
			Colors:    []color.Color{color.Red, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:         game.EventSpellCast,
							Controller:    game.TriggerControllerYou,
							CardSelection: game.Selection{ColorsAny: []color.Color{color.Red}},
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
							CardSelection: game.Selection{ColorsAny: []color.Color{color.White}},
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
			First strike
			This creature enters with four -1/-1 counters on it.
			Whenever you cast a red spell, remove a -1/-1 counter from this creature.
			Whenever you cast a white spell, remove a -1/-1 counter from this creature.
		`,
		},
	}
}
