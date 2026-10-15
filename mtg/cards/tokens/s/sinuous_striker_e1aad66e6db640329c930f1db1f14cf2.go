package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SinuousStriker is the card definition for Sinuous Striker.
//
// Type: Token Creature — Zombie Naga Warrior
//
// Oracle text:
//   {U}: Sinuous Striker gets +1/-1 until end of turn.

// SinuousStrikerTokene1aad66e6db640329c930f1db1f14cf2 is the card definition for Sinuous Striker.
var SinuousStrikerTokene1aad66e6db640329c930f1db1f14cf2 = newSinuousStrikerTokene1aad66e6db640329c930f1db1f14cf2()

func newSinuousStrikerTokene1aad66e6db640329c930f1db1f14cf2() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name:      "Sinuous Striker",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Sub("Naga"), types.Warrior},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{U}: Sinuous Striker gets +1/-1 until end of turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.U,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ModifyPT{
											Object:         game.SourcePermanentReference(),
											PowerDelta:     game.Fixed(1),
											ToughnessDelta: game.Fixed(-1),
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
			{U}: Sinuous Striker gets +1/-1 until end of turn.
		`,
		},
	}
}
