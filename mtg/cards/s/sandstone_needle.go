package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SandstoneNeedle is the card definition for Sandstone Needle.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped with two depletion counters on it.
//	{T}, Remove a depletion counter from this land: Add {R}{R}. If there are no depletion counters on this land, sacrifice it.
var SandstoneNeedle = newSandstoneNeedle

func newSandstoneNeedle() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:  "Sandstone Needle",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}, Remove a depletion counter from this land: Add {R}{R}. If there are no depletion counters on this land, sacrifice it.",
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove a depletion counter from this land",
							Amount:      1,
							CounterKind: counter.Depletion,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.R,
								},
							},
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.R,
								},
							},
							{
								Primitive: game.Sacrifice{
									Object: game.SourcePermanentReference(),
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										Negate:        true,
										Object:        opt.Val(game.SourcePermanentReference()),
										ObjectMatches: opt.Val(game.Selection{RequiredCounterCount: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 1}), RequiredCounter: counter.Depletion}),
									}),
								}),
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedWithCountersReplacement("This land enters tapped with two depletion counters on it.", game.CounterPlacement{Kind: counter.Depletion, Amount: 2}),
			},
			OracleText: `
			This land enters tapped with two depletion counters on it.
			{T}, Remove a depletion counter from this land: Add {R}{R}. If there are no depletion counters on this land, sacrifice it.
		`,
		},
	}
}
