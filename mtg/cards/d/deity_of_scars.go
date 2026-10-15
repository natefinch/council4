package d

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

// DeityOfScars is the card definition for Deity of Scars.
//
// Type: Creature — Spirit Avatar
// Cost: {B/G}{B/G}{B/G}{B/G}{B/G}
//
// Oracle text:
//
//	Trample
//	This creature enters with two -1/-1 counters on it.
//	{B/G}, Remove a -1/-1 counter from this creature: Regenerate this creature.
var DeityOfScars = newDeityOfScars

func newDeityOfScars() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Deity of Scars",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.B, mana.G),
				cost.HybridMana(mana.B, mana.G),
				cost.HybridMana(mana.B, mana.G),
				cost.HybridMana(mana.B, mana.G),
				cost.HybridMana(mana.B, mana.G),
			}),
			Colors:    []color.Color{color.Black, color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit, types.Avatar},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{B/G}, Remove a -1/-1 counter from this creature: Regenerate this creature.",
					ManaCost: opt.Val(cost.Mana{
						cost.HybridMana(mana.B, mana.G),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove a -1/-1 counter from this creature",
							Amount:      1,
							CounterKind: counter.MinusOneMinusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Regenerate{
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
				game.EntersWithCountersReplacement("This creature enters with two -1/-1 counters on it.", game.CounterPlacement{Kind: counter.MinusOneMinusOne, Amount: 2}),
			},
			OracleText: `
			Trample
			This creature enters with two -1/-1 counters on it.
			{B/G}, Remove a -1/-1 counter from this creature: Regenerate this creature.
		`,
		},
	}
}
