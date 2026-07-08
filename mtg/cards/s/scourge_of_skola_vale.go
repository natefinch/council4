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

// ScourgeOfSkolaVale is the card definition for Scourge of Skola Vale.
//
// Type: Creature — Hydra
// Cost: {2}{G}
//
// Oracle text:
//
//	Trample
//	This creature enters with two +1/+1 counters on it.
//	{T}, Sacrifice another creature: Put a number of +1/+1 counters on this creature equal to the sacrificed creature's toughness.
var ScourgeOfSkolaVale = newScourgeOfSkolaVale

func newScourgeOfSkolaVale() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Scourge of Skola Vale",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Hydra},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}, Sacrifice another creature: Put a number of +1/+1 counters on this creature equal to the sacrificed creature's toughness.",
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice another creature",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Creature,
							ExcludeSource:      true,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountObjectToughness,
										Multiplier: 1,
										Object:     game.SacrificedCostReference(),
									}),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with two +1/+1 counters on it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 2}),
			},
			OracleText: `
			Trample
			This creature enters with two +1/+1 counters on it.
			{T}, Sacrifice another creature: Put a number of +1/+1 counters on this creature equal to the sacrificed creature's toughness.
		`,
		},
	}
}
