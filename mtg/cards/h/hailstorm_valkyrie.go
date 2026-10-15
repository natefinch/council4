package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// HailstormValkyrie is the card definition for Hailstorm Valkyrie.
//
// Type: Snow Creature — Angel Wizard
// Cost: {3}{B}
//
// Oracle text:
//
//	Flying, trample
//	{S}{S}: This creature gets +2/+2 until end of turn. ({S} can be paid with one mana from a snow source.)
var HailstormValkyrie = newHailstormValkyrie

func newHailstormValkyrie() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Hailstorm Valkyrie",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Snow},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Angel, types.Wizard},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.TrampleStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{S}{S}: This creature gets +2/+2 until end of turn. ({S} can be paid with one mana from a snow source.)",
					ManaCost: opt.Val(cost.Mana{
						cost.S,
						cost.S,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ModifyPT{
											Object:         game.SourcePermanentReference(),
											PowerDelta:     game.Fixed(2),
											ToughnessDelta: game.Fixed(2),
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
			Flying, trample
			{S}{S}: This creature gets +2/+2 until end of turn. ({S} can be paid with one mana from a snow source.)
		`,
		},
	}
}
