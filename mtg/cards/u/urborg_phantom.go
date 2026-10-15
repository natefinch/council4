package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// UrborgPhantom is the card definition for Urborg Phantom.
//
// Type: Creature — Spirit Minion
// Cost: {2}{B}
//
// Oracle text:
//
//	This creature can't block.
//	{U}: Prevent all combat damage that would be dealt to and dealt by this creature this turn.
var UrborgPhantom = newUrborgPhantom

func newUrborgPhantom() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Urborg Phantom",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit, types.Minion},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.CantBlockStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{U}: Prevent all combat damage that would be dealt to and dealt by this creature this turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.U,
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
			This creature can't block.
			{U}: Prevent all combat damage that would be dealt to and dealt by this creature this turn.
		`,
		},
	}
}
