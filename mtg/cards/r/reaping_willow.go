package r

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

// ReapingWillow is the card definition for Reaping Willow.
//
// Type: Creature — Treefolk Cleric
// Cost: {1}{W/B}{W/B}{W/B}
//
// Oracle text:
//
//	Lifelink
//	This creature enters with two -1/-1 counters on it.
//	{1}{W/B}, Remove two counters from this creature: Return target creature card with mana value 3 or less from your graveyard to the battlefield. Activate only as a sorcery.
var ReapingWillow = newReapingWillow

func newReapingWillow() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Reaping Willow",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.HybridMana(mana.W, mana.B),
				cost.HybridMana(mana.W, mana.B),
				cost.HybridMana(mana.W, mana.B),
			}),
			Colors:    []color.Color{color.Black, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Treefolk, types.Cleric},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.LifelinkStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{1}{W/B}, Remove two counters from this creature: Return target creature card with mana value 3 or less from your graveyard to the battlefield. Activate only as a sorcery.",
					ManaCost: opt.Val(cost.Mana{cost.O(1), cost.HybridMana(mana.W, mana.B)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:           cost.AdditionalRemoveCounter,
							Text:           "Remove two counters from this creature",
							Amount:         2,
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
								Constraint: "target creature card with mana value 3 or less from your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3})}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
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
			Lifelink
			This creature enters with two -1/-1 counters on it.
			{1}{W/B}, Remove two counters from this creature: Return target creature card with mana value 3 or less from your graveyard to the battlefield. Activate only as a sorcery.
		`,
		},
	}
}
