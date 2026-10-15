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

// DaghatarTheAdamant is the card definition for Daghatar the Adamant.
//
// Type: Legendary Creature — Human Warrior
// Cost: {3}{W}
//
// Oracle text:
//
//	Vigilance
//	Daghatar enters with four +1/+1 counters on it.
//	{1}{B/G}{B/G}: Move a +1/+1 counter from target creature onto a second target creature.
var DaghatarTheAdamant = newDaghatarTheAdamant

func newDaghatarTheAdamant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Daghatar the Adamant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Warrior},
			Power:      opt.Val(game.PT{Value: 0}),
			Toughness:  opt.Val(game.PT{Value: 0}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}{B/G}{B/G}: Move a +1/+1 counter from target creature onto a second target creature.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
						cost.HybridMana(mana.B, mana.G),
						cost.HybridMana(mana.B, mana.G),
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
									},
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveCounters{
											Amount:      game.Fixed(1),
											Object:      game.TargetPermanentReference(1),
											CounterKind: counter.PlusOnePlusOne,
											Source: game.CounterSourceSpec{
												Kind:   game.CounterSourceTarget,
												Object: game.TargetPermanentReference(0),
											},
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
				game.EntersWithCountersReplacement("Daghatar enters with four +1/+1 counters on it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 4}),
			},
			OracleText: `
			Vigilance
			Daghatar enters with four +1/+1 counters on it.
			{1}{B/G}{B/G}: Move a +1/+1 counter from target creature onto a second target creature.
		`,
		},
	}
}
