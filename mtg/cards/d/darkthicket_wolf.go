package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DarkthicketWolf is the card definition for Darkthicket Wolf.
//
// Type: Creature — Wolf
// Cost: {1}{G}
//
// Oracle text:
//
//	{2}{G}: This creature gets +2/+2 until end of turn. Activate only once each turn.
var DarkthicketWolf = newDarkthicketWolf

func newDarkthicketWolf() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Darkthicket Wolf",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wolf},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}{G}: This creature gets +2/+2 until end of turn. Activate only once each turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
						cost.G,
					}),
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.OncePerTurn,
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
			{2}{G}: This creature gets +2/+2 until end of turn. Activate only once each turn.
		`,
		},
	}
}
