package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PestilentHaze is the card definition for Pestilent Haze.
//
// Type: Sorcery
// Cost: {1}{B}{B}
//
// Oracle text:
//
//	Choose one —
//	• All creatures get -2/-2 until end of turn.
//	• Remove two loyalty counters from each planeswalker.
var PestilentHaze = newPestilentHaze()

func newPestilentHaze() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Pestilent Haze",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "All creatures get -2/-2 until end of turn.",
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:          game.LayerPowerToughnessModify,
											Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
											PowerDelta:     -2,
											ToughnessDelta: -2,
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					},
					game.Mode{
						Text: "Remove two loyalty counters from each planeswalker.",
						Sequence: []game.Instruction{
							{
								Primitive: game.RemoveCounter{
									Amount:      game.Fixed(2),
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Planeswalker}}),
									CounterKind: counter.Loyalty,
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• All creatures get -2/-2 until end of turn.
			• Remove two loyalty counters from each planeswalker.
		`,
		},
	}
}
