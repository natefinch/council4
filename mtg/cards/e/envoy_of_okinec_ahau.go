package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EnvoyOfOkinecAhau is the card definition for Envoy of Okinec Ahau.
//
// Type: Creature — Cat Advisor
// Cost: {2}{W}
//
// Oracle text:
//
//	{4}{W}: Create a 1/1 colorless Gnome artifact creature token.
var EnvoyOfOkinecAhau = newEnvoyOfOkinecAhau

func newEnvoyOfOkinecAhau() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Envoy of Okinec Ahau",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat, types.Advisor},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{4}{W}: Create a 1/1 colorless Gnome artifact creature token.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(4),
						cost.W,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(envoyOfOkinecAhauToken),
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
			{4}{W}: Create a 1/1 colorless Gnome artifact creature token.
		`,
		},
	}
}

var envoyOfOkinecAhauToken = newEnvoyOfOkinecAhauToken()

func newEnvoyOfOkinecAhauToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Gnome",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Gnome},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
