package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ArgentSphinx is the card definition for Argent Sphinx.
//
// Type: Creature — Sphinx
// Cost: {2}{U}{U}
//
// Oracle text:
//
//	Flying
//	Metalcraft — {U}: Exile this creature. Return it to the battlefield under your control at the beginning of the next end step. Activate only if you control three or more artifacts.
var ArgentSphinx = newArgentSphinx

func newArgentSphinx() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Argent Sphinx",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Sphinx},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "Metalcraft — {U}: Exile this creature. Return it to the battlefield under your control at the beginning of the next end step. Activate only if you control three or more artifacts.",
					ManaCost:       opt.Val(cost.Mana{cost.U}),
					ZoneOfFunction: zone.Battlefield,
					ActivationCondition: opt.Val(game.Condition{
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
							MinCount:  3,
						}),
					}),
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Exile{
									Object:         game.SourcePermanentReference(),
									ExileLinkedKey: game.LinkedKey("delayed-self-blink"),
								},
							},
							{
								Primitive: game.CreateDelayedTrigger{
									Trigger: game.DelayedTriggerDef{
										Timing: game.DelayedAtBeginningOfNextEndStep,
										Content: game.Mode{
											Sequence: []game.Instruction{
												{
													Primitive: game.PutOnBattlefield{
														Source:    game.LinkedBattlefieldSource(game.LinkedKey("delayed-self-blink")),
														Recipient: opt.Val(game.ControllerReference()),
													},
												},
											},
										}.Ability(),
									},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			Metalcraft — {U}: Exile this creature. Return it to the battlefield under your control at the beginning of the next end step. Activate only if you control three or more artifacts.
		`,
		},
	}
}
