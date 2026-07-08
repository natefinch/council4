package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CrumblingColossus is the card definition for Crumbling Colossus.
//
// Type: Artifact Creature — Golem
// Cost: {5}
//
// Oracle text:
//
//	Trample (This creature can deal excess combat damage to the player or planeswalker it's attacking.)
//	When this creature attacks, sacrifice it at end of combat.
var CrumblingColossus = newCrumblingColossus

func newCrumblingColossus() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Crumbling Colossus",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
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
			Trample (This creature can deal excess combat damage to the player or planeswalker it's attacking.)
			When this creature attacks, sacrifice it at end of combat.
		`,
		},
	}
}
