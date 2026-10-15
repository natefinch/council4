package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// JeskaiCharm is the card definition for Jeskai Charm.
//
// Type: Instant
// Cost: {U}{R}{W}
//
// Oracle text:
//
//	Choose one —
//	• Put target creature on top of its owner's library.
//	• Jeskai Charm deals 4 damage to target opponent or planeswalker.
//	• Creatures you control get +1/+1 and gain lifelink until end of turn.
var JeskaiCharm = newJeskaiCharm

func newJeskaiCharm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Jeskai Charm",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.R,
				cost.W,
			}),
			Colors: []color.Color{color.Red, color.Blue, color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Put target creature on top of its owner's library.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.MovePermanent{
									Object:      game.TargetPermanentReference(0),
									Destination: zone.Library,
								},
							},
						},
					},
					game.Mode{
						Text: "Jeskai Charm deals 4 damage to target opponent or planeswalker.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target opponent or planeswalker",
								Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Planeswalker}, Player: game.PlayerOpponent}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(4),
									Recipient: game.AnyTargetDamageRecipient(0),
								},
							},
						},
					},
					game.Mode{
						Text: "Creatures you control get +1/+1 and gain lifelink until end of turn.",
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Duration:       game.DurationPermanent,
											Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											Layer:          game.LayerPowerToughnessModify,
											PowerDelta:     1,
											ToughnessDelta: 1,
										},
										game.ContinuousEffect{
											Duration:    game.DurationPermanent,
											Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											Layer:       game.LayerAbility,
											AddKeywords: []game.Keyword{game.Lifelink},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
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
			• Put target creature on top of its owner's library.
			• Jeskai Charm deals 4 damage to target opponent or planeswalker.
			• Creatures you control get +1/+1 and gain lifelink until end of turn.
		`,
		},
	}
}
