package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MoonlightGeist is the card definition for Moonlight Geist.
//
// Type: Creature — Spirit
// Cost: {2}{W}
//
// Oracle text:
//
//	Flying
//	{3}{W}: Prevent all combat damage that would be dealt to and dealt by this creature this turn.
var MoonlightGeist = newMoonlightGeist

func newMoonlightGeist() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Moonlight Geist",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{3}{W}: Prevent all combat damage that would be dealt to and dealt by this creature this turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
						cost.W,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.PreventDamage{
											Object:     game.SourcePermanentReference(),
											All:        true,
											CombatOnly: true,
										},
									},
									game.Instruction{
										Primitive: game.PreventDamage{
											Object:     game.SourcePermanentReference(),
											All:        true,
											CombatOnly: true,
											BySource:   true,
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
			Flying
			{3}{W}: Prevent all combat damage that would be dealt to and dealt by this creature this turn.
		`,
		},
	}
}
