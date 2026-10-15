package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ClanDefiance is the card definition for Clan Defiance.
//
// Type: Sorcery
// Cost: {X}{R}{G}
//
// Oracle text:
//
//	Choose one or more —
//	• Clan Defiance deals X damage to target creature with flying.
//	• Clan Defiance deals X damage to target creature without flying.
//	• Clan Defiance deals X damage to target player or planeswalker.
var ClanDefiance = newClanDefiance

func newClanDefiance() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Clan Defiance",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.R,
				cost.G,
			}),
			Colors: []color.Color{color.Green, color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Clan Defiance deals X damage to target creature with flying.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature with flying",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Keyword: game.Flying}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind: game.DynamicAmountX,
									}),
									Recipient: game.AnyTargetDamageRecipient(0),
								},
							},
						},
					},
					game.Mode{
						Text: "Clan Defiance deals X damage to target creature without flying.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature without flying",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ExcludedKeyword: game.Flying}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind: game.DynamicAmountX,
									}),
									Recipient: game.AnyTargetDamageRecipient(0),
								},
							},
						},
					},
					game.Mode{
						Text: "Clan Defiance deals X damage to target player or planeswalker.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target player or planeswalker",
								Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Planeswalker}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind: game.DynamicAmountX,
									}),
									Recipient: game.AnyTargetDamageRecipient(0),
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 3,
			}),
			OracleText: `
			Choose one or more —
			• Clan Defiance deals X damage to target creature with flying.
			• Clan Defiance deals X damage to target creature without flying.
			• Clan Defiance deals X damage to target player or planeswalker.
		`,
		},
	}
}
