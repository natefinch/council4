package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NoxiousHatchling is the card definition for Noxious Hatchling.
//
// Type: Creature — Elemental
// Cost: {3}{B/G}
//
// Oracle text:
//
//	This creature enters with four -1/-1 counters on it.
//	Wither (This deals damage to creatures in the form of -1/-1 counters.)
//	Whenever you cast a black spell, remove a -1/-1 counter from this creature.
//	Whenever you cast a green spell, remove a -1/-1 counter from this creature.
var NoxiousHatchling = newNoxiousHatchling

func newNoxiousHatchling() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Noxious Hatchling",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.HybridMana(mana.B, mana.G),
			}),
			Colors:    []color.Color{color.Black, color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.WitherStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:         game.EventSpellCast,
							Controller:    game.TriggerControllerYou,
							CardSelection: game.Selection{ColorsAny: []color.Color{color.Black}},
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
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with four -1/-1 counters on it.", game.CounterPlacement{Kind: counter.MinusOneMinusOne, Amount: 4}),
			},
			OracleText: `
			This creature enters with four -1/-1 counters on it.
			Wither (This deals damage to creatures in the form of -1/-1 counters.)
			Whenever you cast a black spell, remove a -1/-1 counter from this creature.
			Whenever you cast a green spell, remove a -1/-1 counter from this creature.
		`,
		},
	}
}
