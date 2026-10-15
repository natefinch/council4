package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DwarvenLandslide is the card definition for Dwarven Landslide.
//
// Type: Sorcery
// Cost: {3}{R}
//
// Oracle text:
//
//	Kicker—{2}{R}, Sacrifice a land. (You may pay {2}{R} and sacrifice a land in addition to any other costs as you cast this spell.)
//	Destroy target land. If this spell was kicked, destroy another target land.
var DwarvenLandslide = newDwarvenLandslide

func newDwarvenLandslide() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Dwarven Landslide",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Kicker ",
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{
							Cost: cost.Mana{
								cost.O(2),
								cost.R,
							},
							AdditionalCosts: []cost.Additional{
								cost.Additional{
									Kind:               cost.AdditionalSacrifice,
									Text:               "Sacrifice a land",
									Amount:             1,
									MatchPermanentType: true,
									PermanentType:      types.Land,
									CounterKind:        counter.PlusOnePlusOne,
								},
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target land",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}}),
					},
					game.TargetSpec{
						MinTargets:               1,
						MaxTargets:               1,
						Constraint:               "another target land",
						Allow:                    game.TargetAllowPermanent,
						Selection:                opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}}),
						DistinctFromPriorTargets: true,
						Gate:                     game.TargetGateSpellKicked,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(1),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								SpellWasKicked: true,
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Kicker—{2}{R}, Sacrifice a land. (You may pay {2}{R} and sacrifice a land in addition to any other costs as you cast this spell.)
			Destroy target land. If this spell was kicked, destroy another target land.
		`,
		},
	}
}
