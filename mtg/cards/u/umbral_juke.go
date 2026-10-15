package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UmbralJuke is the card definition for Umbral Juke.
//
// Type: Instant
// Cost: {2}{B}
//
// Oracle text:
//
//	Choose one —
//	• Target player sacrifices a creature or planeswalker of their choice.
//	• Create a 2/1 white and black Inkling creature token with flying.
var UmbralJuke = newUmbralJuke

func newUmbralJuke() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Umbral Juke",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Target player sacrifices a creature or planeswalker of their choice.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "Target player",
								Allow:      game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.SacrificePermanents{
									Player:    game.TargetPlayerReference(0),
									Amount:    game.Fixed(1),
									Selection: game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}},
								},
							},
						},
					},
					game.Mode{
						Text: "Create a 2/1 white and black Inkling creature token with flying.",
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(umbralJukeToken),
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
			• Target player sacrifices a creature or planeswalker of their choice.
			• Create a 2/1 white and black Inkling creature token with flying.
		`,
		},
	}
}

var umbralJukeToken = newUmbralJukeToken()

func newUmbralJukeToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Inkling",
			Colors:    []color.Color{color.White, color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Inkling},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
