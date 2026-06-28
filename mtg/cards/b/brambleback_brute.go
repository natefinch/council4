package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BramblebackBrute is the card definition for Brambleback Brute.
//
// Type: Creature — Giant Warrior
// Cost: {2}{R}
//
// Oracle text:
//
//	This creature enters with two -1/-1 counters on it.
//	{1}{R}, Remove a counter from this creature: Target creature can't block this turn. Activate only as a sorcery.
var BramblebackBrute = newBramblebackBrute()

func newBramblebackBrute() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Brambleback Brute",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Giant, types.Warrior},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 5}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{1}{R}, Remove a counter from this creature: Target creature can't block this turn. Activate only as a sorcery.",
					ManaCost: opt.Val(cost.Mana{cost.O(1), cost.R}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:           cost.AdditionalRemoveCounter,
							Text:           "Remove a counter from this creature",
							Amount:         1,
							AnyCounterKind: true,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyRule{
									Object: opt.Val(game.TargetPermanentReference(0)),
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind: game.RuleEffectCantBlock,
										},
									},
									Duration: game.DurationThisTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with two -1/-1 counters on it.", game.CounterPlacement{Kind: counter.MinusOneMinusOne, Amount: 2}),
			},
			OracleText: `
			This creature enters with two -1/-1 counters on it.
			{1}{R}, Remove a counter from this creature: Target creature can't block this turn. Activate only as a sorcery.
		`,
		},
	}
}
