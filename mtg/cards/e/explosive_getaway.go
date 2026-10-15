package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ExplosiveGetaway is the card definition for Explosive Getaway.
//
// Type: Sorcery
// Cost: {3}{R}{W}
//
// Oracle text:
//
//	Exile up to one target artifact or creature. Return it to the battlefield under its owner's control at the beginning of the next end step.
//	Explosive Getaway deals 4 damage to each creature.
var ExplosiveGetaway = newExplosiveGetaway

func newExplosiveGetaway() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Explosive Getaway",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.W,
			}),
			Colors: []color.Color{color.Red, color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 1,
						Constraint: "up to one target artifact or creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Object:        game.TargetPermanentReference(0),
							Destination:   zone.Exile,
							PublishLinked: game.LinkedKey("sequence-effect-0-product"),
						},
						LocalProducts: game.LocalProducts{
							Links: []game.LinkedKey{game.LinkedKey("sequence-effect-0-product")},
						},
						ClearLinkedBeforeGate: true,
					},
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								Timing: game.DelayedAtBeginningOfNextEndStep,
								Content: game.AbilityContent{
									Modes: []game.Mode{
										game.Mode{
											Sequence: []game.Instruction{
												game.Instruction{
													Primitive: game.PutOnBattlefield{
														Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceCaptured}),
													},
												},
											},
										},
									},
									MinModes: 1,
									MaxModes: 1,
								},
								CapturedCard: opt.Val(game.LinkedObjectReference("sequence-effect-0-product")),
							},
						},
					},
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(4),
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Exile up to one target artifact or creature. Return it to the battlefield under its owner's control at the beginning of the next end step.
			Explosive Getaway deals 4 damage to each creature.
		`,
		},
	}
}
