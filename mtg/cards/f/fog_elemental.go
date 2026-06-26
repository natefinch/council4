package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FogElemental is the card definition for Fog Elemental.
//
// Type: Creature — Elemental
// Cost: {2}{U}
//
// Oracle text:
//
//	Flying (This creature can't be blocked except by creatures with flying or reach.)
//	When this creature attacks or blocks, sacrifice it at end of combat.
var FogElemental = newFogElemental()

func newFogElemental() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Fog Elemental",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:      game.EventAttackerDeclared,
							Source:     game.TriggerSourceSelf,
							UnionEvent: game.EventBlockerDeclared,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateDelayedTrigger{
									Trigger: game.DelayedTriggerDef{
										Timing: game.DelayedAtEndOfCombat,
										Content: game.Mode{
											Sequence: []game.Instruction{
												{
													Primitive: game.Sacrifice{
														Object: game.SourceCardPermanentReference(),
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
			Flying (This creature can't be blocked except by creatures with flying or reach.)
			When this creature attacks or blocks, sacrifice it at end of combat.
		`,
		},
	}
}
