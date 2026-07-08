package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PhantomWhelp is the card definition for Phantom Whelp.
//
// Type: Creature — Illusion Dog
// Cost: {1}{U}
//
// Oracle text:
//
//	When this creature attacks or blocks, return it to its owner's hand at end of combat. (Return it only if it's on the battlefield.)
var PhantomWhelp = newPhantomWhelp

func newPhantomWhelp() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Phantom Whelp",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Illusion, types.Dog},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
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
													Primitive: game.Bounce{
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
			When this creature attacks or blocks, return it to its owner's hand at end of combat. (Return it only if it's on the battlefield.)
		`,
		},
	}
}
