package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LoreholdCommand is the card definition for Lorehold Command.
//
// Type: Instant
// Cost: {3}{R}{W}
//
// Oracle text:
//
//	Choose two —
//	• Create a 3/2 red and white Spirit creature token.
//	• Creatures you control get +1/+0 and gain indestructible and haste until end of turn.
//	• Lorehold Command deals 3 damage to any target. Target player gains 3 life.
//	• Sacrifice a permanent, then draw two cards.
var LoreholdCommand = newLoreholdCommand

func newLoreholdCommand() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Lorehold Command",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.W,
			}),
			Colors: []color.Color{color.Red, color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Create a 3/2 red and white Spirit creature token.",
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(loreholdCommandToken),
								},
							},
						},
					},
					game.Mode{
						Text: "Creatures you control get +1/+0 and gain indestructible and haste until end of turn.",
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Duration:   game.DurationPermanent,
											Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											Layer:      game.LayerPowerToughnessModify,
											PowerDelta: 1,
										},
										game.ContinuousEffect{
											Duration:    game.DurationPermanent,
											Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											Layer:       game.LayerAbility,
											AddKeywords: []game.Keyword{game.Indestructible, game.Haste},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					},
					game.Mode{
						Text: "Lorehold Command deals 3 damage to any target. Target player gains 3 life.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "any target",
								Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle}}),
							},
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "Target player",
								Allow:      game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(3),
									Recipient: game.AnyTargetDamageRecipient(0),
								},
							},
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(3),
									Player: game.TargetPlayerReference(1),
								},
							},
						},
					},
					game.Mode{
						Text: "Sacrifice a permanent, then draw two cards.",
						Sequence: []game.Instruction{
							{
								Primitive: game.SacrificePermanents{
									Player: game.ControllerReference(),
									Amount: game.Fixed(1),
								},
							},
							{
								Primitive: game.Draw{
									Amount: game.Fixed(2),
									Player: game.ControllerReference(),
								},
							},
						},
					},
				},
				MinModes: 2,
				MaxModes: 2,
			}),
			OracleText: `
			Choose two —
			• Create a 3/2 red and white Spirit creature token.
			• Creatures you control get +1/+0 and gain indestructible and haste until end of turn.
			• Lorehold Command deals 3 damage to any target. Target player gains 3 life.
			• Sacrifice a permanent, then draw two cards.
		`,
		},
	}
}

var loreholdCommandToken = newLoreholdCommandToken()

func newLoreholdCommandToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Spirit",
			Colors:    []color.Color{color.Red, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
