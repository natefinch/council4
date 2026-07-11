package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Inkshield is the card definition for Inkshield.
//
// Type: Instant
// Cost: {3}{W}{B}
//
// Oracle text:
//
//	Prevent all combat damage that would be dealt to you this turn. For each 1 damage prevented this way, create a 2/1 white and black Inkling creature token with flying.
var Inkshield = newInkshield

func newInkshield() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Inkshield",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.PreventDamage{
							Player:     game.ControllerReference(),
							All:        true,
							CombatOnly: true,
						},
					},
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								Timing: game.DelayedAtBeginningOfNextEndStep,
								Content: game.Mode{
									Sequence: []game.Instruction{
										{
											Primitive: game.CreateToken{
												Amount: game.Dynamic(game.DynamicAmount{
													Kind:       game.DynamicAmountDamagePreventedThisWay,
													Multiplier: 1,
												}),
												Source: game.TokenDef(inkshieldToken),
											},
										},
									},
								}.Ability(),
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Prevent all combat damage that would be dealt to you this turn. For each 1 damage prevented this way, create a 2/1 white and black Inkling creature token with flying.
		`,
		},
	}
}

var inkshieldToken = newInkshieldToken()

func newInkshieldToken() *game.CardDef {
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
