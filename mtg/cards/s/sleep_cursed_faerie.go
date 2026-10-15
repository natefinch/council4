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

// SleepCursedFaerie is the card definition for Sleep-Cursed Faerie.
//
// Type: Creature — Faerie Wizard
// Cost: {U}
//
// Oracle text:
//
//	Flying, ward {2}
//	This creature enters tapped with three stun counters on it. (If it would become untapped, remove a stun counter from it instead.)
//	{1}{U}: Untap this creature.
var SleepCursedFaerie = newSleepCursedFaerie

func newSleepCursedFaerie() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Sleep-Cursed Faerie",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Faerie, types.Wizard},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.WardStaticAbility(cost.Mana{cost.O(2)}),
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}{U}: Untap this creature.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
						cost.U,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Untap{
											Object: game.SourcePermanentReference(),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedWithCountersReplacement("This creature enters tapped with three stun counters on it. (If it would become untapped, remove a stun counter from it instead.)", game.CounterPlacement{Kind: counter.Stun, Amount: 3}),
			},
			OracleText: `
			Flying, ward {2}
			This creature enters tapped with three stun counters on it. (If it would become untapped, remove a stun counter from it instead.)
			{1}{U}: Untap this creature.
		`,
		},
	}
}
