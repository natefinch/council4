package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MysticMelting is the card definition for Mystic Melting.
//
// Type: Instant
// Cost: {3}{G}
//
// Oracle text:
//
//	Destroy target artifact or enchantment.
//	Draw a card at the beginning of the next turn's upkeep.
var MysticMelting = newMysticMelting

func newMysticMelting() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Mystic Melting",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target artifact or enchantment",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								Timing: game.DelayedAtBeginningOfNextUpkeep,
								Content: game.AbilityContent{
									Modes: []game.Mode{
										game.Mode{
											Sequence: []game.Instruction{
												game.Instruction{
													Primitive: game.Draw{
														Amount: game.Fixed(1),
														Player: game.ControllerReference(),
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
					},
				},
			}.Ability()),
			OracleText: `
			Destroy target artifact or enchantment.
			Draw a card at the beginning of the next turn's upkeep.
		`,
		},
	}
}
