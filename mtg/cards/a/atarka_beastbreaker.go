package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AtarkaBeastbreaker is the card definition for Atarka Beastbreaker.
//
// Type: Creature — Human Warrior
// Cost: {1}{G}
//
// Oracle text:
//
//	Formidable — {4}{G}: This creature gets +4/+4 until end of turn. Activate only if creatures you control have total power 8 or greater.
var AtarkaBeastbreaker = newAtarkaBeastbreaker

func newAtarkaBeastbreaker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Atarka Beastbreaker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Formidable — {4}{G}: This creature gets +4/+4 until end of turn. Activate only if creatures you control have total power 8 or greater.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(4),
						cost.G,
					}),
					ZoneOfFunction: zone.Battlefield,
					ActivationCondition: opt.Val(game.Condition{
						Text: "only if creatures you control have total power 8 or greater",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							TotalPower: opt.Val(compare.Int{
								Op:    compare.GreaterOrEqual,
								Value: 8,
							}),
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ModifyPT{
											Object:         game.SourcePermanentReference(),
											PowerDelta:     game.Fixed(4),
											ToughnessDelta: game.Fixed(4),
											Duration:       game.DurationUntilEndOfTurn,
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
			Formidable — {4}{G}: This creature gets +4/+4 until end of turn. Activate only if creatures you control have total power 8 or greater.
		`,
		},
	}
}
