package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Technodrome is the card definition for Technodrome.
//
// Type: Artifact Creature — Construct
// Cost: {2}
//
// Oracle text:
//
//	Reach, trample
//	This creature can't attack or block unless its power is 6 or greater.
//	{T}, Sacrifice another artifact: Draw a card. Put a +1/+1 counter on this creature.
var Technodrome = newTechnodrome

func newTechnodrome() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Technodrome",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.TrampleStaticBody,
				game.StaticAbility{
					Text: "This creature can't attack or block unless its power is 6 or greater.",
					Condition: opt.Val(game.Condition{
						Text:   "unless its power is 6 or greater",
						Negate: true,
						Object: opt.Val(game.SourcePermanentReference()),
						ObjectMatches: opt.Val(game.Selection{Power: opt.Val(compare.Int{
							Op:    compare.GreaterOrEqual,
							Value: 6,
						})}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantAttack,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
						},
						game.RuleEffect{
							Kind:           game.RuleEffectCantBlock,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}, Sacrifice another artifact: Draw a card. Put a +1/+1 counter on this creature.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice another artifact",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Artifact,
							CounterKind:        counter.PlusOnePlusOne,
							ExcludeSource:      true,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.PlusOnePlusOne,
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
			OracleText: `
			Reach, trample
			This creature can't attack or block unless its power is 6 or greater.
			{T}, Sacrifice another artifact: Draw a card. Put a +1/+1 counter on this creature.
		`,
		},
	}
}
