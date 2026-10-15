package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CabarettiCharm is the card definition for Cabaretti Charm.
//
// Type: Instant
// Cost: {R}{G}{W}
//
// Oracle text:
//
//	Choose one —
//	• Cabaretti Charm deals damage equal to the number of creatures you control to target creature or planeswalker.
//	• Creatures you control get +1/+1 and gain trample until end of turn.
//	• Create two 1/1 green and white Citizen creature tokens.
var CabarettiCharm = newCabarettiCharm

func newCabarettiCharm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Cabaretti Charm",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
				cost.G,
				cost.W,
			}),
			Colors: []color.Color{color.Green, color.Red, color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Cabaretti Charm deals damage equal to the number of creatures you control to target creature or planeswalker.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature or planeswalker",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountSelector,
										Multiplier: 1,
										Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									}),
									Recipient: game.AnyTargetDamageRecipient(0),
								},
							},
						},
					},
					game.Mode{
						Text: "Creatures you control get +1/+1 and gain trample until end of turn.",
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
											AddKeywords: []game.Keyword{game.Trample},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					},
					game.Mode{
						Text: "Create two 1/1 green and white Citizen creature tokens.",
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(2),
									Source: game.TokenDef(cabarettiCharmToken),
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
			• Cabaretti Charm deals damage equal to the number of creatures you control to target creature or planeswalker.
			• Creatures you control get +1/+1 and gain trample until end of turn.
			• Create two 1/1 green and white Citizen creature tokens.
		`,
		},
	}
}

var cabarettiCharmToken = newCabarettiCharmToken()

func newCabarettiCharmToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Citizen",
			Colors:    []color.Color{color.Green, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Citizen},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
